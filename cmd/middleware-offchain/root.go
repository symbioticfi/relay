package main

import (
	"context"
	"log/slog"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/libp2p/go-libp2p"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"

	"middleware-offchain/core/client/evm"
	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/aggregator"
	symbioticCrypto "middleware-offchain/core/usecase/crypto"
	keyprovider "middleware-offchain/core/usecase/key-provider"
	valsetDeriver "middleware-offchain/core/usecase/valset-deriver"
	"middleware-offchain/internal/client/p2p"
	"middleware-offchain/internal/client/repository/badger"
	aggregatorApp "middleware-offchain/internal/usecase/aggregator-app"
	apiApp "middleware-offchain/internal/usecase/api-app"
	"middleware-offchain/internal/usecase/metrics"
	signerApp "middleware-offchain/internal/usecase/signer-app"
	valsetGenerator "middleware-offchain/internal/usecase/valset-generator"
	valsetListener "middleware-offchain/internal/usecase/valset-listener"
	"middleware-offchain/pkg/log"
	"middleware-offchain/pkg/proof"
	"middleware-offchain/pkg/signals"
)

func runApp(ctx context.Context) error {
	cfg := cfgFromCtx(ctx)
	log.Init(cfg.LogLevel, cfg.LogMode)
	mtr := metrics.New(metrics.Config{})

	chains := lo.Map(cfg.Chains, func(chain CMDChain, _ int) entity.ChainURL {
		return entity.ChainURL{
			ChainID: chain.ChainID,
			RPCURL:  chain.URL,
		}
	})

	// TODO if keystore is used use another keystore and ignore keys from flags
	keyProvider, err := keyprovider.NewSimpleKeystoreProvider()
	if err != nil {
		return errors.Errorf("failed to create keystore provider: %w", err)
	}

	for _, key := range cfg.SecretKeys {
		keyBytes, ok := new(big.Int).SetString(key.Secret, 10)
		if !ok {
			return errors.Errorf("failed to parse secret key as big.Int")
		}
		pk, err := symbioticCrypto.NewPrivateKey(entity.KeyType(key.KeyType), keyBytes.Bytes())
		if err != nil {
			return errors.Errorf("failed to create private key: %w", err)
		}
		err = keyProvider.AddKeyByNamespaceTypeId(key.Namespace, entity.KeyType(key.KeyType), key.KeyId, pk)
		if err != nil {
			return errors.Errorf("failed to add key to keystore: %w", err)
		}
	}

	evmClient, err := evm.NewEVMClient(ctx, evm.Config{
		Chains: chains,
		DriverAddress: entity.CrossChainAddress{
			ChainId: cfg.Driver.ChainID,
			Address: common.HexToAddress(cfg.Driver.Address),
		},
		RequestTimeout: time.Second * 5,
		KeyProvider:    keyProvider,
		Metrics:        mtr,
	})
	if err != nil {
		return errors.Errorf("failed to create symbiotic client: %w", err)
	}

	deriver, err := valsetDeriver.NewDeriver(evmClient)
	if err != nil {
		return errors.Errorf("failed to create valset deriver: %w", err)
	}

	var opts []libp2p.Option
	if cfg.P2PListenAddress != "" {
		opts = append(opts, libp2p.ListenAddrStrings(cfg.P2PListenAddress))
	}
	h, err := libp2p.New(opts...)
	if err != nil {
		return errors.Errorf("failed to create libp2p host: %w", err)
	}

	p2pService, err := p2p.NewService(ctx, p2p.Config{
		Host:        h,
		SendTimeout: time.Second * 10,
		Metrics:     mtr,
	})
	if err != nil {
		return errors.Errorf("failed to create p2p service: %w", err)
	}
	slog.InfoContext(ctx, "Created p2p service", "listenAddr", cfg.P2PListenAddress)
	defer p2pService.Close()

	discoveryService, err := p2p.NewDiscoveryService(ctx, p2pService, h)
	if err != nil {
		return errors.Errorf("failed to create discovery service: %w", err)
	}
	defer discoveryService.Close()
	slog.InfoContext(ctx, "Created discovery service", "listenAddr", cfg.P2PListenAddress)
	if err := discoveryService.Start(); err != nil {
		return errors.Errorf("failed to start discovery service: %w", err)
	}
	slog.InfoContext(ctx, "Started discovery service", "listenAddr", cfg.P2PListenAddress)

	repo, err := badger.New(badger.Config{Dir: cfg.StorageDir})
	if err != nil {
		return errors.Errorf("failed to create memory repository: %w", err)
	}

	aggregatorLib := aggregator.NewAggregator(proof.NewZkProver())

	aggProofReadySignal := signals.New[entity.AggregatedSignatureMessage]()

	signerApp, err := signerApp.NewSignerApp(signerApp.Config{
		P2PService:     p2pService,
		KeyProvider:    keyProvider,
		Repo:           repo,
		AggProofSignal: aggProofReadySignal,
		Aggregator:     aggregatorLib,
		Metrics:        mtr,
	})
	if err != nil {
		return errors.Errorf("failed to create signer app: %w", err)
	}
	p2pService.AddSignaturesAggregatedMessageListener(signerApp.HandleSignaturesAggregatedMessage, "signerAppSignaturesAggregatedListener")

	slog.InfoContext(ctx, "Created signer app, starting")

	listener, err := valsetListener.New(valsetListener.Config{
		Eth:             evmClient,
		Repo:            repo,
		Deriver:         deriver,
		PollingInterval: time.Second * 5,
	})
	if err != nil {
		return errors.Errorf("failed to create epoch listener: %w", err)
	}

	generator, err := valsetGenerator.New(valsetGenerator.Config{
		Signer:          signerApp,
		Eth:             evmClient,
		Repo:            repo,
		Deriver:         deriver,
		Aggregator:      aggregatorLib,
		PollingInterval: time.Second * 5,
		IsCommitter:     cfg.IsCommitter,
	})
	if err != nil {
		return errors.Errorf("failed to create epoch listener: %w", err)
	}

	aggProofReadySignal.AddListener(func(ctx context.Context, msg entity.AggregatedSignatureMessage) error {
		err := generator.HandleProofAggregated(ctx, msg)
		if err != nil {
			return errors.Errorf("failed to handle proof aggregated: %w", err)
		}
		slog.DebugContext(ctx, "Handled proof aggregated", "request", msg)

		return nil
	}, "aggregatedProofReadySignalListener")

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err := listener.Start(egCtx); err != nil {
			return errors.Errorf("failed to start valset listener: %w", err)
		}
		return nil
	})

	if cfg.IsSigner {
		eg.Go(func() error {
			if err := generator.Start(egCtx); err != nil {
				return errors.Errorf("failed to start valset generator: %w", err)
			}
			return nil
		})
	}

	var aggApp *aggregatorApp.AggregatorApp
	if cfg.IsAggregator {
		aggApp, err = aggregatorApp.NewAggregatorApp(aggregatorApp.Config{
			Repo:       repo,
			P2PClient:  p2pService,
			Aggregator: aggregatorLib,
			Metrics:    mtr,
		})
		if err != nil {
			return errors.Errorf("failed to create aggregator app: %w", err)
		}
		p2pService.AddSignatureMessageListener(aggApp.HandleSignatureGeneratedMessage, "aggregatorAppSignatureGeneratedListener")

		slog.DebugContext(ctx, "Created aggregator app, starting")
	}

	serveMetricsOnAPIAddress := cfg.HTTPListenAddr == cfg.MetricsListenAddr || cfg.MetricsListenAddr == ""
	api, err := apiApp.NewAPIApp(apiApp.Config{
		Address:           cfg.HTTPListenAddr,
		ReadHeaderTimeout: time.Second,
		ShutdownTimeout:   time.Second * 5,
		Prefix:            "/api/v1",
		Signer:            signerApp,
		Repo:              repo,
		EVMClient:         evmClient,
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
