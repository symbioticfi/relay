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

	"middleware-offchain/bls"
	aggregator "middleware-offchain/internal/app/aggregator-app"
	app "middleware-offchain/internal/app/signer-app"
	"middleware-offchain/internal/client/eth"
	"middleware-offchain/internal/client/p2p"
	"middleware-offchain/pkg/log"
	"middleware-offchain/valset"
)

// offchain_middleware --master-address 0x5081a39b8A5f0E35a8D959395a630b68B74Dd30f --rpc-url http://127.0.0.1:8545
func main() {
	slog.Info("Running offchain_middleware command", "args", os.Args)

	if err := run(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("error executing command", "error", err)
		os.Exit(1)
	}
	slog.Info("Offchain middleware completed successfully")
}

func run() error {
	rootCmd.PersistentFlags().StringVar(&cfg.rpcURL, "rpc-url", "", "RPC URL")
	rootCmd.PersistentFlags().StringVar(&cfg.masterAddress, "master-address", "", "Master contract address")
	rootCmd.PersistentFlags().StringVar(&cfg.logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&cfg.listenAddress, "p2p-listen", "/ip4/127.0.0.1/tcp/8000", "P2P listen address")
	rootCmd.PersistentFlags().StringVar(&cfg.secretKey, "secret-key", "", "Secret key for BLS signature generation")
	rootCmd.PersistentFlags().BoolVar(&cfg.isAggregator, "aggregator", false, "Is Aggregator")
	rootCmd.PersistentFlags().BoolVar(&cfg.isSigner, "signer", true, "Is Signer")

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
	rpcURL        string
	masterAddress string
	logLevel      string
	listenAddress string
	secretKey     string
	isAggregator  bool
	isSigner      bool
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
		log.Init(cfg.logLevel)

		ctx := signalContext(context.Background())

		b := new(big.Int).SetBytes([]byte(cfg.secretKey))
		keyPair := bls.ComputeKeyPair(b.Bytes())

		ethClient, err := eth.NewEthClient(eth.Config{
			MasterRPCURL:   cfg.rpcURL,
			MasterAddress:  cfg.masterAddress,
			RequestTimeout: time.Second * 5,
		})
		if err != nil {
			return errors.Errorf("failed to create eth client: %w", err)
		}

		deriver, err := valset.NewValsetDeriver(ethClient)
		if err != nil {
			return errors.Errorf("failed to create valset deriver: %w", err)
		}

		generator, err := valset.NewValsetGenerator(deriver, ethClient)
		if err != nil {
			return errors.Errorf("failed to create valset generator: %w", err)
		}

		h, err := libp2p.New(libp2p.ListenAddrStrings(cfg.listenAddress))
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

		signerApp, err := app.NewSignerApp(app.Config{
			PollingInterval: time.Second * 10,
			ValsetGenerator: generator,
			EthClient:       ethClient,
			P2PService:      p2pService,
			KeyPair:         keyPair,
		})
		if err != nil {
			return errors.Errorf("failed to create signer app: %w", err)
		}
		slog.InfoContext(ctx, "created signer app, starting")

		eg, egCtx := errgroup.WithContext(ctx)
		if cfg.isSigner {
			eg.Go(func() error {
				if err := signerApp.Start(egCtx); err != nil {
					return errors.Errorf("failed to start signer app: %w", err)
				}
				return nil
			})
		}

		if cfg.isAggregator {
			aggregatorApp, err := aggregator.NewAggregatorApp(ctx, aggregator.Config{
				EthClient:     ethClient,
				ValsetDeriver: deriver,
			})
			if err != nil {
				return errors.Errorf("failed to create aggregator app: %w", err)
			}
			slog.DebugContext(ctx, "created aggregator app, starting")
			p2pService.SetSignatureHashMessageHandler(aggregatorApp.HandleSignatureGeneratedMessage)

			eg.Go(func() error {
				if err := aggregatorApp.Start(egCtx); err != nil {
					return errors.Errorf("failed to start aggregator app: %w", err)
				}
				return nil
			})
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
