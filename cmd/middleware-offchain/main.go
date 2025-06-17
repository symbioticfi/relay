package main

import (
	"context"
	"log/slog"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-errors/errors"
	"github.com/libp2p/go-libp2p"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"middleware-offchain/core/client/evm"
	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/aggregator"
	keyprovider "middleware-offchain/core/usecase/key-provider"
	"middleware-offchain/core/usecase/signer"
	valsetDeriver "middleware-offchain/core/usecase/valset-deriver"
	"middleware-offchain/internal/client/p2p"
	"middleware-offchain/internal/client/repository/badger"
	aggregatorApp "middleware-offchain/internal/usecase/aggregator-app"
	signerApp "middleware-offchain/internal/usecase/signer-app"
	valsetGenerator "middleware-offchain/internal/usecase/valset-generator"
	valsetListener "middleware-offchain/internal/usecase/valset-listener"
	"middleware-offchain/pkg/log"
	"middleware-offchain/pkg/proof"
	"middleware-offchain/pkg/server"
	"middleware-offchain/pkg/signals"
)

// offchain_middleware --driver-address 0x1f5fE7682E49c20289C20a4cFc8b45d5EB410690 --rpc-url http://127.0.0.1:8545
func main() {
	slog.Info("Running offchain_middleware command", "args", os.Args)

	if err := runRootCMD(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("error executing command", "error", err)
		os.Exit(1)
	}
	slog.Info("Offchain middleware completed successfully")
}

func runRootCMD() error {
	rootCmd.PersistentFlags().StringVar(&cfg.rpcURL, "rpc-url", "", "RPC URL")
	rootCmd.PersistentFlags().StringVar(&cfg.driverAddress, "driver-address", "", "Driver contract address")
	rootCmd.PersistentFlags().StringVar(&cfg.logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&cfg.logMode, "log-mode", "text", "Log mode (text, pretty)")
	rootCmd.PersistentFlags().StringVar(&cfg.listenAddress, "p2p-listen", "", "P2P listen address")
	rootCmd.PersistentFlags().StringVar(&cfg.httpListenAddress, "http-listen", "", "Http listener address")
	rootCmd.PersistentFlags().StringVar(&cfg.secretKey, "secret-key", "", "Secret key for BLS signature generation")
	rootCmd.PersistentFlags().BoolVar(&cfg.isAggregator, "aggregator", false, "Is Aggregator")
	rootCmd.PersistentFlags().BoolVar(&cfg.isSigner, "signer", true, "Is Signer")
	rootCmd.PersistentFlags().BoolVar(&cfg.isCommitter, "committer", false, "Is Committer")
	rootCmd.PersistentFlags().StringVar(&cfg.storageDir, "storage-dir", ".data", "Dir to store data")

	if err := rootCmd.MarkPersistentFlagRequired("rpc-url"); err != nil {
		return errors.Errorf("failed to mark rpc-url as required: %w", err)
	}
	if err := rootCmd.MarkPersistentFlagRequired("driver-address"); err != nil {
		return errors.Errorf("failed to mark driver-address as required: %w", err)
	}
	if err := rootCmd.MarkPersistentFlagRequired("secret-key"); err != nil {
		return errors.Errorf("failed to mark secret-key as required: %w", err)
	}

	return rootCmd.Execute()
}

type config struct {
	rpcURL            string
	driverAddress     string
	logLevel          string
	logMode           string
	listenAddress     string
	httpListenAddress string
	secretKey         string
	isAggregator      bool
	isSigner          bool
	isCommitter       bool
	storageDir        string
}

var cfg config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "middleware_offchain",
	Short:         "Offchain middleware for signature aggregation",
	Long:          "A P2P service for collecting and aggregating signatures for Ethereum contracts.",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Init(cfg.logLevel, cfg.logMode)

		ctx := signalContext(context.Background())

		b, ok := new(big.Int).SetString(cfg.secretKey, 10)
		if !ok {
			return errors.Errorf("failed to parse secret key as big.Int")
		}

		pkBytes := [32]byte{}
		b.FillBytes(pkBytes[:])

		ethClient, err := evm.NewEVMClient(evm.Config{
			MasterRPCURL:   cfg.rpcURL,
			DriverAddress:  cfg.driverAddress,
			RequestTimeout: time.Second * 5,
			PrivateKey:     pkBytes[:],
		})
		if err != nil {
			return errors.Errorf("failed to create symbiotic client: %w", err)
		}

		deriver, err := valsetDeriver.NewDeriver(ethClient)
		if err != nil {
			return errors.Errorf("failed to create valset deriver: %w", err)
		}

		var opts []libp2p.Option
		if cfg.listenAddress != "" {
			opts = append(opts, libp2p.ListenAddrStrings(cfg.listenAddress))
		}
		h, err := libp2p.New(opts...)
		if err != nil {
			return errors.Errorf("failed to create libp2p host: %w", err)
		}

		p2pService, err := p2p.NewService(ctx, h)
		if err != nil {
			return errors.Errorf("failed to create p2p service: %w", err)
		}
		slog.InfoContext(ctx, "created p2p service", "listenAddr", cfg.listenAddress)
		defer p2pService.Close()

		discoveryService, err := p2p.NewDiscoveryService(ctx, p2pService, h)
		if err != nil {
			return errors.Errorf("failed to create discovery service: %w", err)
		}
		defer discoveryService.Close()
		slog.InfoContext(ctx, "created discovery service", "listenAddr", cfg.listenAddress)
		if err := discoveryService.Start(); err != nil {
			return errors.Errorf("failed to start discovery service: %w", err)
		}
		slog.InfoContext(ctx, "started discovery service", "listenAddr", cfg.listenAddress)

		repo, err := badger.New(badger.Config{Dir: cfg.storageDir})
		if err != nil {
			return errors.Errorf("failed to create memory repository: %w", err)
		}

		keystoreProvider, err := keyprovider.NewSimpleKeystoreProvider()
		if err != nil {
			return errors.Errorf("failed to create keystore provider: %w", err)
		}
		err = keystoreProvider.AddKey(15, b.Bytes())
		if err != nil {
			return errors.Errorf("failed to add key to keystore provider: %w", err)
		}

		aggregator := aggregator.NewAggregator(proof.NewZkProver())

		signerLib := signer.NewSigner(keystoreProvider)

		aggProofReadySignal := signals.New[entity.AggregatedSignatureMessage]()

		signerApp, err := signerApp.NewSignerApp(signerApp.Config{
			P2PService:     p2pService,
			Signer:         signerLib,
			Repo:           repo,
			AggProofSignal: aggProofReadySignal,
			Aggregator:     aggregator,
		})
		if err != nil {
			return errors.Errorf("failed to create signer app: %w", err)
		}
		p2pService.AddSignaturesAggregatedMessageListener(signerApp.HandleSignaturesAggregatedMessage, "signerAppSignaturesAggregatedListener")

		slog.InfoContext(ctx, "created signer app, starting")

		listener, err := valsetListener.New(valsetListener.Config{
			Eth:             ethClient,
			Repo:            repo,
			Deriver:         deriver,
			PollingInterval: time.Second * 5,
		})
		if err != nil {
			return errors.Errorf("failed to create epoch listener: %w", err)
		}

		generator, err := valsetGenerator.New(valsetGenerator.Config{
			Signer:          signerApp,
			Eth:             ethClient,
			Repo:            repo,
			Deriver:         deriver,
			Aggregator:      aggregator,
			PollingInterval: time.Second * 5,
			IsCommitter:     cfg.isCommitter,
		})
		if err != nil {
			return errors.Errorf("failed to create epoch listener: %w", err)
		}

		aggProofReadySignal.AddListener(func(ctx context.Context, msg entity.AggregatedSignatureMessage) error {
			err := generator.HandleProofAggregated(ctx, msg)
			if err != nil {
				return errors.Errorf("failed to handle proof aggregated: %w", err)
			}
			slog.DebugContext(ctx, "handled proof aggregated", "request", msg)

			return nil
		}, "aggregatedProofReadySignalListener")

		srv, err := server.New(server.Config{
			Address:           cfg.httpListenAddress,
			ReadHeaderTimeout: time.Second,
			ShutdownTimeout:   time.Second * 5,
			Prefix:            "/api/v1",
			APIHandler:        signerApp.Handler(),
		})
		if err != nil {
			return errors.Errorf("failed to create server: %w", err)
		}
		slog.InfoContext(ctx, "created server, starting")

		eg, egCtx := errgroup.WithContext(ctx)
		eg.Go(func() error {
			logCtx := log.WithComponent(egCtx, "listener")
			if err := listener.Start(logCtx); err != nil {
				return errors.Errorf("failed to start valset listener: %w", err)
			}
			return nil
		})

		if cfg.isSigner {
			eg.Go(func() error {
				logCtx := log.WithComponent(egCtx, "generator")

				if err := generator.Start(logCtx); err != nil {
					return errors.Errorf("failed to start valset generator: %w", err)
				}
				return nil
			})

			eg.Go(func() error {
				if err := srv.Serve(egCtx); err != nil {
					return errors.Errorf("failed to start epoch listener server: %w", err)
				}
				return nil
			})
		}

		if cfg.isAggregator {
			aggApp, err := aggregatorApp.NewAggregatorApp(aggregatorApp.Config{
				Repo:       repo,
				P2PClient:  p2pService,
				Aggregator: aggregator,
				Verifier:   signerLib,
			})
			if err != nil {
				return errors.Errorf("failed to create aggregator app: %w", err)
			}
			p2pService.AddSignatureMessageListener(aggApp.HandleSignatureGeneratedMessage, "aggregatorAppSignatureGeneratedListener")

			slog.DebugContext(ctx, "created aggregator app, starting")
		}

		return eg.Wait()
	},
}

// signalContext returns a context that is canceled if either SIGTERM or SIGINT signal is received.
func signalContext(ctx context.Context) context.Context {
	cnCtx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-c
		slog.Info("received signal", "signal", sig)
		cancel()
	}()

	return cnCtx
}
