package main

import (
	"context"
	"log/slog"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"middleware-offchain/internal/usecase/aggregator"
	"middleware-offchain/pkg/proof"

	"github.com/go-errors/errors"
	"github.com/libp2p/go-libp2p"
	"github.com/maniartech/signals"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"middleware-offchain/internal/client/p2p"
	"middleware-offchain/internal/client/repository/memory"
	"middleware-offchain/internal/client/symbiotic"
	"middleware-offchain/internal/entity"
	aggregatorApp "middleware-offchain/internal/usecase/aggregator-app"
	keyprovider "middleware-offchain/internal/usecase/key-provider"
	"middleware-offchain/internal/usecase/signer"
	signerApp "middleware-offchain/internal/usecase/signer-app"
	valsetDeriver "middleware-offchain/internal/usecase/valset-deriver"
	valsetGenerator "middleware-offchain/internal/usecase/valset-generator"
	valsetListener "middleware-offchain/internal/usecase/valset-listener"
	"middleware-offchain/pkg/log"
	"middleware-offchain/pkg/server"
)

// offchain_middleware --master-address 0x1f5fE7682E49c20289C20a4cFc8b45d5EB410690 --rpc-url http://127.0.0.1:8545
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
	rootCmd.PersistentFlags().StringVar(&cfg.masterAddress, "master-address", "", "Master contract address")
	rootCmd.PersistentFlags().StringVar(&cfg.logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&cfg.logMode, "log-mode", "text", "Log mode (text, pretty)")
	rootCmd.PersistentFlags().StringVar(&cfg.listenAddress, "p2p-listen", "", "P2P listen address")
	rootCmd.PersistentFlags().StringVar(&cfg.httpListenAddress, "http-listen", "", "Http listener address")
	rootCmd.PersistentFlags().StringVar(&cfg.secretKey, "secret-key", "", "Secret key for BLS signature generation")
	rootCmd.PersistentFlags().BoolVar(&cfg.isAggregator, "aggregator", false, "Is Aggregator")
	rootCmd.PersistentFlags().BoolVar(&cfg.isSigner, "signer", true, "Is Signer")
	rootCmd.PersistentFlags().BoolVar(&cfg.isCommitter, "committer", false, "Is Committer")

	if err := rootCmd.MarkPersistentFlagRequired("rpc-url"); err != nil {
		return errors.Errorf("failed to mark rpc-url as required: %w", err)
	}
	if err := rootCmd.MarkPersistentFlagRequired("master-address"); err != nil {
		return errors.Errorf("failed to mark master-address as required: %w", err)
	}
	if err := rootCmd.MarkPersistentFlagRequired("secret-key"); err != nil {
		return errors.Errorf("failed to mark secret-key as required: %w", err)
	}

	return rootCmd.Execute()
}

type config struct {
	rpcURL            string
	masterAddress     string
	logLevel          string
	logMode           string
	listenAddress     string
	httpListenAddress string
	secretKey         string
	isAggregator      bool
	isSigner          bool
	isCommitter       bool
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

		ethClient, err := symbiotic.NewEVMClient(symbiotic.Config{
			MasterRPCURL:   cfg.rpcURL,
			MasterAddress:  cfg.masterAddress,
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

		repo, err := memory.New()
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

		// todo ilya extract to lib package in order to get rid of vendor lock on specific lib
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

		aggProofReadySignal.AddListener(func(ctx context.Context, msg entity.AggregatedSignatureMessage) {
			err := generator.HandleProofAggregated(ctx, msg)
			if err != nil {
				slog.ErrorContext(ctx, "failed to handle proof aggregated", "error", err)
			} else {
				slog.DebugContext(ctx, "handled proof aggregated", "request", msg)
			}
		})

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
			_, err := aggregatorApp.NewAggregatorApp(aggregatorApp.Config{
				Repo:       repo,
				P2PClient:  p2pService,
				Aggregator: aggregator,
			})
			if err != nil {
				return errors.Errorf("failed to create aggregator app: %w", err)
			}
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
