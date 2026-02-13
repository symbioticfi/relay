package root

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-errors/errors"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"golang.org/x/sync/errgroup"

	"github.com/symbioticfi/relay/internal/client/p2p"
	"github.com/symbioticfi/relay/internal/client/repository/badger"
	"github.com/symbioticfi/relay/internal/entity"
	aggregationPolicy "github.com/symbioticfi/relay/internal/usecase/aggregation-policy"
	aggregatorApp "github.com/symbioticfi/relay/internal/usecase/aggregator-app"
	api_server "github.com/symbioticfi/relay/internal/usecase/api-server"
	entity_processor "github.com/symbioticfi/relay/internal/usecase/entity-processor"
	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	"github.com/symbioticfi/relay/internal/usecase/metrics"
	"github.com/symbioticfi/relay/internal/usecase/pruner"
	signatureListener "github.com/symbioticfi/relay/internal/usecase/signature-listener"
	signerApp "github.com/symbioticfi/relay/internal/usecase/signer-app"
	sync_provider "github.com/symbioticfi/relay/internal/usecase/sync-provider"
	sync_runner "github.com/symbioticfi/relay/internal/usecase/sync-runner"
	valsetListener "github.com/symbioticfi/relay/internal/usecase/valset-listener"
	valsetStatusTracker "github.com/symbioticfi/relay/internal/usecase/valset-status-tracker"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/proof"
	"github.com/symbioticfi/relay/pkg/signals"
	"github.com/symbioticfi/relay/pkg/tracing"
	"github.com/symbioticfi/relay/symbiotic/client/evm"
	"github.com/symbioticfi/relay/symbiotic/client/votingpower"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/aggregator"
	symbioticCrypto "github.com/symbioticfi/relay/symbiotic/usecase/crypto"
	valsetDeriver "github.com/symbioticfi/relay/symbiotic/usecase/valset-deriver"
)

func runApp(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	cfg := cfgFromCtx(ctx)
	log.Init(cfg.Log.Level, cfg.Log.Mode)
	mtr := metrics.New(metrics.Config{})

	var keyProvider *keyprovider.CacheKeyProvider
	if cfg.KeyStore.Path != "" {
		var err error
		kp, err := keyprovider.NewKeystoreProvider(cfg.KeyStore.Path, cfg.KeyStore.Password)
		if err != nil {
			return errors.Errorf("failed to create keystore provider from keystore file: %w", err)
		}
		keyProvider = keyprovider.NewCacheKeyProvider(kp)
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
			pk, err := symbioticCrypto.NewPrivateKey(symbiotic.KeyType(key.KeyType), keyBytes)
			if err != nil {
				return errors.Errorf("failed to create private key: %w", err)
			}
			err = simpleKeyProvider.AddKeyByNamespaceTypeId(key.Namespace, symbiotic.KeyType(key.KeyType), key.KeyId, pk)
			if err != nil {
				return errors.Errorf("failed to add key to keystore: %w", err)
			}
		}
		keyProvider = keyprovider.NewCacheKeyProvider(simpleKeyProvider)
	}

	if cfg.KeyCache.Enabled {
		if err := symbioticCrypto.InitializePubkeyCache(cfg.KeyCache.Size); err != nil {
			return errors.Errorf("failed to initialize public key cache: %w", err)
		}
		slog.DebugContext(ctx, "Initialized public key cache", "size", cfg.KeyCache.Size)
	}

	evmClient, err := evm.NewEvmClient(ctx, evm.Config{
		ChainURLs: cfg.Evm.Chains,
		DriverAddress: symbiotic.CrossChainAddress{
			ChainId: cfg.Driver.ChainID,
			Address: common.HexToAddress(cfg.Driver.Address),
		},
		RequestTimeout:    time.Second * 5,
		KeyProvider:       keyProvider,
		Metrics:           mtr,
		MaxCalls:          cfg.Evm.MaxCalls,
		FallbackGasPrices: cfg.Evm.FallbackGasPrices,
	})
	if err != nil {
		return errors.Errorf("failed to create symbiotic client: %w", err)
	}

	var externalVPClient *votingpower.Client
	if len(cfg.ExternalVotingPowerProviders) > 0 {
		externalVPClient, err = votingpower.NewClient(ctx, cfg.ExternalVotingPowerProviders)
		if err != nil {
			return errors.Errorf("failed to create external voting power client: %w", err)
		}
		defer func() {
			if err := externalVPClient.Close(); err != nil {
				slog.WarnContext(ctx, "Failed to close external voting power client", "error", err)
			}
		}()
	}

	deriver, err := valsetDeriver.NewDeriver(evmClient, externalVPClient)
	if err != nil {
		return errors.Errorf("failed to create valset deriver: %w", err)
	}

	baseRepo, err := badger.New(badger.Config{
		Dir:                      cfg.StorageDir,
		Metrics:                  mtr,
		MutexCleanupInterval:     time.Hour,
		MutexCleanupStaleTimeout: time.Hour - time.Minute,
	})
	if err != nil {
		return errors.Errorf("failed to create badger repository: %w", err)
	}
	defer baseRepo.Close()

	repo, err := badger.NewCached(baseRepo, badger.CachedConfig{
		NetworkConfigCacheSize: cfg.Cache.NetworkConfigCacheSize,
		ValidatorSetCacheSize:  cfg.Cache.ValidatorSetCacheSize,
	})
	if err != nil {
		return errors.Errorf("failed to create cached repository: %w", err)
	}

	currentOnchainEpoch, err := evmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return errors.Errorf("failed to get current epoch: %w", err)
	}

	captureTimestamp, err := evmClient.GetEpochStart(ctx, currentOnchainEpoch)
	if err != nil {
		return errors.Errorf("failed to get capture timestamp: %w", err)
	}

	config, err := evmClient.GetConfig(ctx, captureTimestamp, currentOnchainEpoch)
	if err != nil {
		return errors.Errorf("failed to get config: %w", err)
	}

	var prover *proof.ZkProver
	if config.VerificationType == symbiotic.VerificationTypeBlsBn254ZK {
		prover = proof.NewZkProver(cfg.CircuitsDir)
	}
	agg, err := aggregator.NewAggregator(config.VerificationType, prover)
	if err != nil {
		return errors.Errorf("failed to create aggregator: %w", err)
	}

	signatureProcessedSignal := signals.New[symbiotic.Signature](cfg.SignalCfg, "signatureProcessed", nil)
	aggProofReadySignal := signals.New[symbiotic.AggregationProof](cfg.SignalCfg, "aggProofReady", nil)
	validatorSetSignal := signals.New[symbiotic.ValidatorSet](cfg.SignalCfg, "validatorSet", nil)

	entityProcessor, err := entity_processor.NewEntityProcessor(entity_processor.Config{
		Repo:                     repo,
		Aggregator:               agg,
		AggProofSignal:           aggProofReadySignal,
		SignatureProcessedSignal: signatureProcessedSignal,
		Metrics:                  mtr,
	})
	if err != nil {
		return errors.Errorf("failed to create entity processor: %w", err)
	}
	syncProvider, err := sync_provider.New(sync_provider.Config{
		Repo:                        repo,
		EntityProcessor:             entityProcessor,
		EpochsToSync:                cfg.Sync.EpochsToSync,
		MaxSignatureRequestsPerSync: 1000,
		MaxResponseSignatureCount:   1000,
		MaxAggProofRequestsPerSync:  500,
		MaxResponseAggProofCount:    500,
	})
	if err != nil {
		return errors.Errorf("failed to create syncer: %w", err)
	}

	signer, err := signerApp.NewSignerApp(signerApp.Config{
		KeyProvider:     keyProvider,
		Repo:            repo,
		EntityProcessor: entityProcessor,
		Metrics:         mtr,
	})
	if err != nil {
		return errors.Errorf("failed to create signer app: %w", err)
	}

	listener, err := valsetListener.New(valsetListener.Config{
		EvmClient:           evmClient,
		Repo:                repo,
		Deriver:             deriver,
		PollingInterval:     time.Second * 5,
		ValidatorSet:        validatorSetSignal,
		Signer:              signer,
		Aggregator:          agg,
		KeyProvider:         keyProvider,
		Metrics:             mtr,
		ForceCommitter:      cfg.ForceRole.Committer,
		EpochRetentionCount: cfg.Retention.ValSetEpochs,
	})
	if err != nil {
		return errors.Errorf("failed to create epoch listener: %w", err)
	}

	// Load all missing epochs before starting services
	if err := listener.LoadAllMissingEpochs(ctx); err != nil {
		return errors.Errorf("failed to load missing epochs: %w", err)
	}

	eg, egCtx := errgroup.WithContext(ctx)

	// also start monitoring for new epochs immediately so that we don't miss any epochs while starting other services
	eg.Go(func() error {
		err := listener.Start(egCtx)
		if err != nil && !errors.Is(err, context.Canceled) {
			slog.ErrorContext(ctx, "Valset listener failed", "error", err)
			return errors.Errorf("failed to start valset listener: %w", err)
		}
		slog.InfoContext(ctx, "Valset listener stopped")
		return nil
	})

	p2pService, discoveryService, err := initP2PService(ctx, cfg, keyProvider, syncProvider, mtr)
	if err != nil {
		return errors.Errorf("failed to create p2p service: %w", err)
	}
	defer p2pService.Close()

	// Initialize tracing with instance ID from P2P service
	if cfg.Tracing.Enabled {
		tracer, err := tracing.New(ctx, tracing.Config{
			Enabled:    cfg.Tracing.Enabled,
			Endpoint:   cfg.Tracing.Endpoint,
			SampleRate: cfg.Tracing.SampleRate,
			InstanceID: p2pService.ID(), // Use P2P ID as unique instance identifier
			Version:    Version,
		})
		if err != nil {
			return errors.Errorf("failed to create tracer: %w", err)
		}
		defer func() {
			shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
			defer shutdownCancel()
			if err := tracer.Shutdown(shutdownCtx); err != nil {
				slog.ErrorContext(ctx, "Failed to shutdown tracer", "error", err)
			}
		}()
		slog.InfoContext(ctx, "Tracing enabled",
			"endpoint", cfg.Tracing.Endpoint,
			"service", tracing.ServiceName,
			"instanceId", p2pService.ID(),
			"sampleRate", cfg.Tracing.SampleRate,
		)
	}

	eg.Go(func() error {
		err := signer.HandleSignatureRequests(egCtx, cfg.SignalCfg.WorkerCount, p2pService)
		if err != nil && !errors.Is(err, context.Canceled) {
			slog.ErrorContext(ctx, "Signature requests handler failed", "error", err)
			return errors.Errorf("failed to handle missing self signatures: %w", err)
		}
		slog.InfoContext(ctx, "Signature requests handler stopped")
		return nil
	})

	syncRunner, err := sync_runner.New(sync_runner.Config{
		Enabled:     cfg.Sync.Enabled,
		P2PService:  p2pService,
		Provider:    syncProvider,
		SyncPeriod:  cfg.Sync.Period,
		SyncTimeout: cfg.Sync.Timeout,
		Metrics:     mtr,
	})
	if err != nil {
		return errors.Errorf("failed to create sync runner: %w", err)
	}

	prunerService, err := pruner.New(pruner.Config{
		Repo:                     repo,
		Metrics:                  mtr,
		Enabled:                  cfg.Pruner.Enabled,
		Interval:                 cfg.Pruner.Interval,
		ValsetRetentionEpochs:    cfg.Retention.ValSetEpochs,
		ProofRetentionEpochs:     cfg.Retention.ProofEpochs,
		SignatureRetentionEpochs: cfg.Retention.SignatureEpochs,
	})
	if err != nil {
		return errors.Errorf("failed to create pruner: %w", err)
	}

	slog.InfoContext(ctx, "Created discovery service", "listenAddr", cfg.P2P.ListenAddress)
	if err := discoveryService.Start(ctx); err != nil {
		return errors.Errorf("failed to start discovery service: %w", err)
	}
	defer discoveryService.Close(ctx)

	slog.InfoContext(ctx, "Started discovery service", "listenAddr", cfg.P2P.ListenAddress)

	if err := p2pService.StartSignaturesAggregatedMessageListener(signer.HandleSignaturesAggregatedMessage); err != nil {
		return errors.Errorf("failed to start signatures aggregated message listener: %w", err)
	}

	slog.InfoContext(ctx, "Created signer app, starting")

	statusTracker, err := valsetStatusTracker.New(valsetStatusTracker.Config{
		EvmClient:            evmClient,
		Repo:                 repo,
		PollingInterval:      time.Second * 5,
		EpochPollingInterval: time.Minute,
		Metrics:              mtr,
	})
	if err != nil {
		return errors.Errorf("failed to create valset status tracker: %w", err)
	}

	if err := statusTracker.TrackMissingEpochsStatuses(ctx); err != nil {
		return errors.Errorf("failed to track missing epochs statuses: %w", err)
	}

	signListener, err := signatureListener.New(signatureListener.Config{
		Repo:            repo,
		EntityProcessor: entityProcessor,
		SignalCfg:       cfg.SignalCfg,
		SelfP2PID:       p2pService.ID(),
	})
	if err != nil {
		return errors.Errorf("failed to create signature listener: %w", err)
	}

	if err := p2pService.StartSignatureMessageListener(signListener.HandleSignatureReceivedMessage); err != nil {
		return errors.Errorf("failed to start signature message listener: %w", err)
	}

	eg.Go(func() error {
		err := statusTracker.Start(egCtx)
		if err != nil && !errors.Is(err, context.Canceled) {
			slog.ErrorContext(ctx, "Valset status tracker failed", "error", err)
			return errors.Errorf("failed to start valset status tracker: %w", err)
		}
		slog.InfoContext(ctx, "Valset status tracker stopped")
		return nil
	})

	eg.Go(func() error {
		err := listener.StartCommitterLoop(egCtx)
		if err != nil && !errors.Is(err, context.Canceled) {
			slog.ErrorContext(ctx, "Valset listener committer loop failed", "error", err)
			return errors.Errorf("failed to start committer loop: %w", err)
		}
		slog.InfoContext(ctx, "Valset listener committer loop stopped")
		return nil
	})

	aggPolicyType := symbiotic.AggregationPolicyLowLatency
	if config.VerificationType == symbiotic.VerificationTypeBlsBn254Simple {
		aggPolicyType = symbiotic.AggregationPolicyLowCost
	}
	aggPolicy, err := aggregationPolicy.NewAggregationPolicy(aggPolicyType, cfg.MaxUnsigners)
	if err != nil {
		return errors.Errorf("failed to create aggregator policy: %w", err)
	}

	var aggApp *aggregatorApp.AggregatorApp
	aggApp, err = aggregatorApp.NewAggregatorApp(aggregatorApp.Config{
		Repo:              repo,
		P2PClient:         p2pService,
		Aggregator:        agg,
		Metrics:           mtr,
		AggregationPolicy: aggPolicy,
		KeyProvider:       keyProvider,
		ForceAggregator:   cfg.ForceRole.Aggregator,
	})
	if err != nil {
		return errors.Errorf("failed to create aggregator app: %w", err)
	}

	serveMetricsOnAPIAddress := cfg.API.ListenAddress == cfg.Metrics.ListenAddress || cfg.Metrics.ListenAddress == ""

	api, err := api_server.NewSymbioticServer(ctx, api_server.Config{
		Address:                cfg.API.ListenAddress,
		ShutdownTimeout:        time.Second * 5,
		ReadHeaderTimeout:      time.Second,
		Signer:                 signer,
		Repo:                   repo,
		EvmClient:              evmClient,
		KeyProvider:            keyProvider,
		Aggregator:             aggApp,
		Deriver:                deriver,
		Metrics:                mtr,
		ServeMetrics:           serveMetricsOnAPIAddress,
		ServePprof:             cfg.Metrics.PprofEnabled,
		ServeHTTPGateway:       cfg.API.HTTPGateway,
		VerboseLogging:         cfg.API.VerboseLogging,
		MaxAllowedStreamsCount: int(cfg.API.MaxAllowedStreams),
	})
	if err != nil {
		return errors.Errorf("failed to create api app: %w", err)
	}

	if err := validatorSetSignal.SetHandlers(api.HandleValidatorSet()); err != nil {
		return errors.Errorf("failed to set validator set set message handler: %w", err)
	}
	if err := validatorSetSignal.StartWorkers(ctx); err != nil {
		return errors.Errorf("failed to start validator set set signal workers: %w", err)
	}

	if err := signatureProcessedSignal.SetHandlers(
		aggApp.HandleSignatureProcessedMessage,
		api.HandleSignatureProcessed(),
	); err != nil {
		return errors.Errorf("failed to set signature received message handler: %w", err)
	}
	if err := signatureProcessedSignal.StartWorkers(ctx); err != nil {
		return errors.Errorf("failed to start signature received signal workers: %w", err)
	}

	err = aggProofReadySignal.SetHandlers(
		api.HandleProofAggregated(),
	)
	if err != nil {
		return errors.Errorf("failed to set agg proof ready signal handler: %w", err)
	}
	if err := aggProofReadySignal.StartWorkers(ctx); err != nil {
		return errors.Errorf("failed to start agg proof ready signal workers: %w", err)
	}

	slog.DebugContext(ctx, "Created aggregator app, starting")

	eg.Go(func() error {
		err := api.Start(egCtx)
		if err != nil && !errors.Is(err, context.Canceled) {
			slog.ErrorContext(ctx, "API server failed", "error", err)
			return errors.Errorf("failed to start api server: %w", err)
		}
		slog.InfoContext(ctx, "API server stopped")
		return nil
	})

	eg.Go(func() error {
		err := syncRunner.Start(egCtx)
		if err != nil && !errors.Is(err, context.Canceled) {
			slog.ErrorContext(ctx, "Sync runner failed", "error", err)
			return errors.Errorf("failed to start sync runner: %w", err)
		}
		slog.InfoContext(ctx, "Sync finished stopped")
		return nil
	})

	eg.Go(func() error {
		err := statusTracker.RunEpochTracker(egCtx)
		if err != nil && !errors.Is(err, context.Canceled) {
			slog.ErrorContext(ctx, "Epoch tracker failed", "error", err)
			return errors.Errorf("failed to start epoch tracker: %w", err)
		}
		slog.InfoContext(ctx, "Epoch tracker stopped")
		return nil
	})

	eg.Go(func() error {
		return aggApp.TryAggregateRequestsWithoutProof(ctx)
	})

	eg.Go(func() error {
		prunerService.Start(egCtx)
		slog.InfoContext(ctx, "Pruner stopped")
		return nil
	})

	eg.Go(func() error {
		err := p2pService.StartGRPCServer(egCtx)
		if err != nil && !errors.Is(err, context.Canceled) {
			slog.ErrorContext(ctx, "P2P grpc server failed", "error", err)
			return errors.Errorf("failed to start p2p grpc server: %w", err)
		}
		slog.InfoContext(ctx, "P2P grpc server stopped")
		return nil
	})

	if !serveMetricsOnAPIAddress {
		mtrApp, err := metrics.NewApp(metrics.AppConfig{
			Address:           cfg.Metrics.ListenAddress,
			ReadHeaderTimeout: time.Second * 5,
		})
		if err != nil {
			return errors.Errorf("failed to create metrics app: %w", err)
		}

		slog.DebugContext(ctx, "Created metrics app, starting")
		eg.Go(func() error {
			err := mtrApp.Start(egCtx)
			if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, http.ErrServerClosed) {
				slog.ErrorContext(ctx, "Metrics server failed", "error", err)
				return errors.Errorf("failed to start metrics server: %w", err)
			}
			slog.InfoContext(ctx, "Metrics server stopped")
			return nil
		})
	}

	return eg.Wait()
}

func initP2PService(ctx context.Context, cfg config, keyProvider keyprovider.KeyProvider, provider *sync_provider.Syncer, mtr *metrics.Metrics) (*p2p.Service, *p2p.DiscoveryService, error) {
	swarmPSK, err := hexutil.Decode(cfg.Driver.Address)
	if err != nil {
		return nil, nil, errors.Errorf("failed to get P2P swarm psk: %w", err)
	}
	// pad to make 20 byte to 32 bytes
	swarmPSK = append(swarmPSK, make([]byte, 32-len(swarmPSK))...)

	if len(swarmPSK) != 32 {
		return nil, nil, errors.Errorf("invalid swarm psk length: %d, expected 20", len(swarmPSK))
	}

	// TODO: include p2p key in valset
	p2pIdentityPKRaw, err := keyProvider.GetPrivateKeyByNamespaceTypeId(keyprovider.P2P_KEY_NAMESPACE, symbiotic.KeyTypeEcdsaSecp256k1, keyprovider.P2P_HOST_IDENTITY_KEY_ID)
	if err != nil && !errors.Is(err, entity.ErrKeyNotFound) {
		return nil, nil, errors.Errorf("failed to get P2P identity private key: %w", err)
	}
	if errors.Is(err, entity.ErrKeyNotFound) {
		slog.WarnContext(ctx, "P2P identity private key not found, generating a new one")
		p2pIdentityPKRaw, err = symbioticCrypto.GeneratePrivateKey(symbiotic.KeyTypeEcdsaSecp256k1)
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
		libp2p.PrivateNetwork(swarmPSK), // Use a private network with the provided swarm key
		libp2p.Identity(p2pIdentityPK),  // Use the provided identity private key to sign messages that will be sent over the P2P gossip sub
		libp2p.Security(noise.ID, noise.New),
		libp2p.DefaultMuxers,
	}
	if cfg.P2P.ListenAddress != "" {
		opts = append(opts, libp2p.ListenAddrStrings(cfg.P2P.ListenAddress))
	}
	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, nil, errors.Errorf("failed to create libp2p host: %w", err)
	}

	p2pCfg := p2p.Config{
		Host:      h,
		Metrics:   mtr,
		Discovery: p2p.DefaultDiscoveryConfig(),
		Handler:   p2p.NewP2PHandler(provider),
	}
	if len(cfg.P2P.Bootnodes) > 0 {
		p2pCfg.Discovery.BootstrapPeers = cfg.P2P.Bootnodes
	}
	p2pCfg.Discovery.DHTMode = cfg.P2P.DHTMode
	p2pCfg.Discovery.EnableMDNS = cfg.P2P.MDnsEnabled

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
