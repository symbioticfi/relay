package main

import (
	"context"
	"log/slog"
	"time"

	signature_processor "github.com/symbioticfi/relay/core/usecase/signature-processor"
	signatureListener "github.com/symbioticfi/relay/internal/usecase/signature-listener"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	valsetStatusTracker "github.com/symbioticfi/relay/internal/usecase/valset-status-tracker"
	"golang.org/x/sync/errgroup"

	"github.com/symbioticfi/relay/core/client/evm"
	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/aggregator"
	symbioticCrypto "github.com/symbioticfi/relay/core/usecase/crypto"
	keyprovider "github.com/symbioticfi/relay/core/usecase/key-provider"
	valsetDeriver "github.com/symbioticfi/relay/core/usecase/valset-deriver"
	"github.com/symbioticfi/relay/internal/client/p2p"
	"github.com/symbioticfi/relay/internal/client/repository/badger"
	aggregatorApp "github.com/symbioticfi/relay/internal/usecase/aggregator-app"
	api_server "github.com/symbioticfi/relay/internal/usecase/api-server"
	"github.com/symbioticfi/relay/internal/usecase/metrics"
	signerApp "github.com/symbioticfi/relay/internal/usecase/signer-app"
	valsetGenerator "github.com/symbioticfi/relay/internal/usecase/valset-generator"
	valsetListener "github.com/symbioticfi/relay/internal/usecase/valset-listener"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/proof"
	"github.com/symbioticfi/relay/pkg/signals"
)

func runApp(ctx context.Context) error {
	cfg := cfgFromCtx(ctx)
	log.Init(cfg.LogLevel, cfg.LogMode)
	mtr := metrics.New(metrics.Config{})

	var keyProvider keyprovider.KeyProvider
	if cfg.KeyStore.Path != "" {
		var err error
		keyProvider, err = keyprovider.NewKeystoreProvider(cfg.KeyStore.Path, cfg.KeyStore.Password)
		if err != nil {
			return errors.Errorf("failed to create keystore provider from keystore file: %w", err)
		}
	} else {
		simpleKeyProvider, err := keyprovider.NewSimpleKeystoreProvider()
		if err != nil {
			return errors.Errorf("failed to create keystore provider: %w", err)
		}

		for _, key := range cfg.SecretKeys {
			keyBytes := common.FromHex(key.Secret)
			if len(keyBytes) == 0 {
				return errors.Errorf("invalid key bytes for key %s/%d/%d/%s", key.Namespace, key.KeyType, key.KeyId, keyBytes)
			}
			pk, err := symbioticCrypto.NewPrivateKey(entity.KeyType(key.KeyType), keyBytes)
			if err != nil {
				return errors.Errorf("failed to create private key: %w", err)
			}
			err = simpleKeyProvider.AddKeyByNamespaceTypeId(key.Namespace, entity.KeyType(key.KeyType), key.KeyId, pk)
			if err != nil {
				return errors.Errorf("failed to add key to keystore: %w", err)
			}
		}
		keyProvider = simpleKeyProvider
	}

	evmClient, err := evm.NewEvmClient(ctx, evm.Config{
		ChainURLs: cfg.Chains,
		DriverAddress: entity.CrossChainAddress{
			ChainId: cfg.Driver.ChainID,
			Address: common.HexToAddress(cfg.Driver.Address),
		},
		RequestTimeout: time.Second * 5,
		KeyProvider:    keyProvider,
		Metrics:        mtr,
		MaxCalls:       cfg.MaxCalls,
	})
	if err != nil {
		return errors.Errorf("failed to create symbiotic client: %w", err)
	}

	deriver, err := valsetDeriver.NewDeriver(evmClient)
	if err != nil {
		return errors.Errorf("failed to create valset deriver: %w", err)
	}

	baseRepo, err := badger.New(badger.Config{Dir: cfg.StorageDir})
	if err != nil {
		return errors.Errorf("failed to create badger repository: %w", err)
	}

	repo, err := badger.NewCached(baseRepo, badger.CachedConfig{
		NetworkConfigCacheSize: cfg.Cache.NetworkConfigCacheSize,
		ValidatorSetCacheSize:  cfg.Cache.ValidatorSetCacheSize,
	})
	if err != nil {
		return errors.Errorf("failed to create cached repository: %w", err)
	}

	listener, err := valsetListener.New(valsetListener.Config{
		EvmClient:       evmClient,
		Repo:            repo,
		Deriver:         deriver,
		PollingInterval: time.Second * 5,
	})
	if err != nil {
		return errors.Errorf("failed to create epoch listener: %w", err)
	}

	// Load all missing epochs before starting services
	if err := listener.LoadAllMissingEpochs(ctx); err != nil {
		return errors.Errorf("failed to load missing epochs: %w", err)
	}

	currentOnchainEpoch, err := evmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return errors.Errorf("failed to get current epoch: %w", err)
	}

	captureTimestamp, err := evmClient.GetEpochStart(ctx, currentOnchainEpoch)
	if err != nil {
		return errors.Errorf("failed to get capture timestamp: %w", err)
	}

	config, err := evmClient.GetConfig(ctx, captureTimestamp)
	if err != nil {
		return errors.Errorf("failed to get config: %w", err)
	}

	var prover *proof.ZkProver
	if config.VerificationType == entity.VerificationTypeBlsBn254ZK {
		prover = proof.NewZkProver(cfg.CircuitsDir)
	}
	agg, err := aggregator.NewAggregator(config.VerificationType, prover)
	if err != nil {
		return errors.Errorf("failed to create aggregator: %w", err)
	}

	p2pService, discoveryService, err := initP2PService(ctx, cfg, keyProvider, mtr)
	if err != nil {
		return errors.Errorf("failed to create p2p service: %w", err)
	}
	defer p2pService.Close()

	slog.InfoContext(ctx, "Created discovery service", "listenAddr", cfg.P2PListenAddress)
	if err := discoveryService.Start(ctx); err != nil {
		return errors.Errorf("failed to start discovery service: %w", err)
	}
	defer discoveryService.Close(ctx)

	slog.InfoContext(ctx, "Started discovery service", "listenAddr", cfg.P2PListenAddress)

	aggProofReadySignal := signals.New[entity.AggregatedSignatureMessage](cfg.SignalCfg, "aggProofReady", nil)

	signatureProcessor, err := signature_processor.NewSignatureProcessor(signature_processor.Config{
		Repo: repo,
	})
	if err != nil {
		return errors.Errorf("failed to create signature processor: %w", err)
	}

	signerApp, err := signerApp.NewSignerApp(signerApp.Config{
		P2PService:         p2pService,
		KeyProvider:        keyProvider,
		Repo:               repo,
		SignatureProcessor: signatureProcessor,
		AggProofSignal:     aggProofReadySignal,
		Aggregator:         agg,
		Metrics:            mtr,
	})
	if err != nil {
		return errors.Errorf("failed to create signer app: %w", err)
	}
	if err := p2pService.StartSignaturesAggregatedMessageListener(signerApp.HandleSignaturesAggregatedMessage); err != nil {
		return errors.Errorf("failed to start signatures aggregated message listener: %w", err)
	}

	slog.InfoContext(ctx, "Created signer app, starting")

	generator, err := valsetGenerator.New(valsetGenerator.Config{
		Signer:          signerApp,
		EvmClient:       evmClient,
		Repo:            repo,
		Deriver:         deriver,
		Aggregator:      agg,
		PollingInterval: time.Second * 5,
		IsCommitter:     cfg.IsCommitter,
	})
	if err != nil {
		return errors.Errorf("failed to create epoch listener: %w", err)
	}

	statusTracker, err := valsetStatusTracker.New(valsetStatusTracker.Config{
		EvmClient:       evmClient,
		Repo:            repo,
		Deriver:         deriver,
		PollingInterval: time.Second * 5,
	})
	if err != nil {
		return errors.Errorf("failed to create valset status tracker: %w", err)
	}

	if err := aggProofReadySignal.SetHandler(func(ctx context.Context, msg entity.AggregatedSignatureMessage) error {
		if err := generator.HandleProofAggregated(ctx, msg); err != nil {
			return errors.Errorf("failed to handle proof aggregated: %w", err)
		}
		if err := statusTracker.HandleProofAggregated(ctx, msg); err != nil {
			return errors.Errorf("failed to handle proof aggregated: %w", err)
		}
		slog.DebugContext(ctx, "Handled proof aggregated", "request", msg)

		return nil
	}); err != nil {
		return errors.Errorf("failed to set agg proof ready signal handler: %w", err)
	}
	if err := aggProofReadySignal.StartWorkers(ctx); err != nil {
		return errors.Errorf("failed to start agg proof ready signal workers: %w", err)
	}

	signListener, err := signatureListener.New(signatureListener.Config{
		Repo:               repo,
		SignatureProcessor: signatureProcessor,
		SignalCfg:          cfg.SignalCfg,
		SelfP2PID:          p2pService.ID(),
	})
	if err != nil {
		return errors.Errorf("failed to create signature listener: %w", err)
	}
	if err := p2pService.StartSignatureMessageListener(signListener.HandleSignatureReceivedMessage); err != nil {
		return errors.Errorf("failed to start signature message listener: %w", err)
	}

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err := listener.Start(egCtx); err != nil {
			return errors.Errorf("failed to start valset listener: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		if err := generator.Start(egCtx); err != nil {
			return errors.Errorf("failed to start valset generator: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		if err := statusTracker.Start(egCtx); err != nil {
			return errors.Errorf("failed to start valset status tracker: %w", err)
		}
		return nil
	})

	var aggApp *aggregatorApp.AggregatorApp
	if cfg.IsAggregator {
		aggApp, err = aggregatorApp.NewAggregatorApp(aggregatorApp.Config{
			Repo:       repo,
			P2PClient:  p2pService,
			Aggregator: agg,
			Metrics:    mtr,
		})
		if err != nil {
			return errors.Errorf("failed to create aggregator app: %w", err)
		}

		if err := signListener.StartSignatureSavedMessageListener(ctx, aggApp.HandleSignatureGeneratedMessage); err != nil {
			return errors.Errorf("failed to start signature saved message listener: %w", err)
		}

		slog.DebugContext(ctx, "Created aggregator app, starting")
	}

	serveMetricsOnAPIAddress := cfg.HTTPListenAddr == cfg.MetricsListenAddr || cfg.MetricsListenAddr == ""

	api, err := api_server.NewSymbioticServer(ctx, api_server.Config{
		Address:           cfg.HTTPListenAddr,
		ShutdownTimeout:   time.Second * 5,
		ReadHeaderTimeout: time.Second,
		Signer:            signerApp,
		Repo:              repo,
		EvmClient:         evmClient,
		Aggregator:        aggApp,
		Deriver:           deriver,
		ServeMetrics:      serveMetricsOnAPIAddress,
	})
	if err != nil {
		return errors.Errorf("failed to create api app: %w", err)
	}

	eg.Go(func() error {
		if err := api.Start(egCtx); err != nil {
			return errors.Errorf("failed to start api server: %w", err)
		}
		return nil
	})

	if !serveMetricsOnAPIAddress {
		mtrApp, err := metrics.NewApp(metrics.AppConfig{
			Address:           cfg.MetricsListenAddr,
			ReadHeaderTimeout: time.Second * 5,
		})
		if err != nil {
			return errors.Errorf("failed to create metrics app: %w", err)
		}

		slog.DebugContext(ctx, "Created metrics app, starting")
		eg.Go(func() error {
			if err := mtrApp.Start(egCtx); err != nil {
				return errors.Errorf("failed to start metrics server: %w", err)
			}
			return nil
		})
	}

	return eg.Wait()
}

func initP2PService(ctx context.Context, cfg config, keyProvider keyprovider.KeyProvider, mtr *metrics.Metrics) (*p2p.Service, *p2p.DiscoveryService, error) {
	swarmPK, err := keyProvider.GetPrivateKeyByNamespaceTypeId(keyprovider.P2P_KEY_NAMESPACE, entity.KeyTypeEcdsaSecp256k1, keyprovider.P2P_SWARM_NETWORK_KEY_ID)
	if err != nil {
		return nil, nil, errors.Errorf("failed to get P2P swarm private key: %w", err)
	}

	p2pIdentityPKRaw, err := keyProvider.GetPrivateKeyByNamespaceTypeId(keyprovider.P2P_KEY_NAMESPACE, entity.KeyTypeEcdsaSecp256k1, keyprovider.P2P_HOST_IDENTITY_KEY_ID)
	if err != nil && !errors.Is(err, keyprovider.ErrKeyNotFound) {
		return nil, nil, errors.Errorf("failed to get P2P identity private key: %w", err)
	}
	if errors.Is(err, keyprovider.ErrKeyNotFound) {
		slog.WarnContext(ctx, "P2P identity private key not found, generating a new one")
		p2pIdentityPKRaw, err = symbioticCrypto.GeneratePrivateKey(entity.KeyTypeEcdsaSecp256k1)
		if err != nil {
			return nil, nil, errors.Errorf("failed to create P2P identity private key: %w", err)
		}
	}

	p2pIdentityPK, err := crypto.UnmarshalSecp256k1PrivateKey(p2pIdentityPKRaw.Bytes())
	if err != nil {
		return nil, nil, errors.Errorf("failed to unmarshal P2P identity private key: %w", err)
	}

	opts := []libp2p.Option{
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.PrivateNetwork(swarmPK.Bytes()), // Use a private network with the provided swarm key
		libp2p.Identity(p2pIdentityPK),         // Use the provided identity private key to sign messages that will be sent over the P2P gossip sub
		libp2p.Security(noise.ID, noise.New),
		libp2p.DefaultMuxers,
	}
	if cfg.P2PListenAddress != "" {
		opts = append(opts, libp2p.ListenAddrStrings(cfg.P2PListenAddress))
	}
	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, nil, errors.Errorf("failed to create libp2p host: %w", err)
	}

	p2pCfg := p2p.Config{
		Host:      h,
		Metrics:   mtr,
		Discovery: p2p.DefaultDiscoveryConfig(),
	}
	if len(cfg.Bootnodes) > 0 {
		p2pCfg.Discovery.BootstrapPeers = cfg.Bootnodes
	}
	p2pCfg.Discovery.DHTMode = cfg.DHTMode
	p2pCfg.Discovery.EnableMDNS = cfg.MDnsEnabled

	p2pService, err := p2p.NewService(ctx, p2pCfg, cfg.SignalCfg)
	if err != nil {
		return nil, nil, errors.Errorf("failed to create p2p service: %w", err)
	}
	slog.InfoContext(ctx, "Created p2p service", "listenAddr", h.Addrs(), "id", h.ID().String())

	discoveryService, err := p2p.NewDiscoveryService(p2pCfg)
	if err != nil {
		return nil, nil, errors.Errorf("failed to create discovery service: %w", err)
	}

	return p2pService, discoveryService, nil
}
