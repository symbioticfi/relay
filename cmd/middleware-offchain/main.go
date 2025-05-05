package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	app "middleware-offchain/internal/app/signer-app"
	"middleware-offchain/internal/client/eth"
	"middleware-offchain/internal/client/p2p"
	"middleware-offchain/pkg/log"
)

func main() {
	cobra.OnInitialize(initConfig)

	log.Init()
	// Start command flags
	startCmd.Flags().String("listen", "/ip4/127.0.0.1/tcp/8000", "Address to listen on")

	// Bind flags to viper
	viper.BindPFlag("listen", startCmd.Flags().Lookup("listen"))

	// Add commands
	rootCmd.AddCommand(startCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("error executing command", "error", err)
		os.Exit(1)
	}
	slog.Info("Offchain middleware completed successfully")
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "offchain-middleware",
	Short: "Offchain middleware for signature aggregation",
	Long:  `A P2P service for collecting and aggregating signatures for Ethereum contracts.`,
}

type config struct {
	EthEndpoint   string `mapstructure:"eth_endpoint"`
	ContractAddr  string `mapstructure:"contract_addr"`
	EthPrivateKey string `mapstructure:"eth_private_key"`
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the offchain middleware service",
	Long:  "Start the offchain middleware service with the specified configuration.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := signalContext(context.Background())
		listenAddr := viper.GetString("listen")

		var cfg config
		err := viper.Unmarshal(&cfg)
		if err != nil {
			return errors.Errorf("failed to unmarshal config: %w", err)
		}

		ethClient, err := eth.NewEthClient(eth.Config{
			MasterRPCURL:  cfg.EthEndpoint,
			MasterAddress: cfg.ContractAddr,
			PrivateKey:    []byte(cfg.EthPrivateKey),
		})
		if err != nil {
			return errors.Errorf("failed to create eth client: %w", err)
		}

		p2pService, err := p2p.NewService(ctx, listenAddr)
		if err != nil {
			return errors.Errorf("failed to create p2p service: %w", err)
		}
		slog.InfoContext(ctx, "created p2p service", "listenAddr", listenAddr)
		defer p2pService.Close()

		discoveryService, err := p2p.NewDiscoveryService(ctx, p2pService, listenAddr)
		if err != nil {
			return errors.Errorf("failed to create discovery service: %w", err)
		}
		defer discoveryService.Close()
		slog.InfoContext(ctx, "created discovery service", "listenAddr", listenAddr)
		if err := discoveryService.Start(); err != nil {
			return errors.Errorf("failed to start discovery service: %w", err)
		}
		slog.InfoContext(ctx, "started discovery service", "listenAddr", listenAddr)

		signerApp, err := app.NewSignerApp(app.Config{
			PollingInterval: time.Second * 30,
			EthClient:       ethClient,
			P2PService:      p2pService,
		})
		if err != nil {
			return errors.Errorf("failed to create signer app: %w", err)
		}
		slog.InfoContext(ctx, "created signer app, starting")

		if err := signerApp.Start(ctx); err != nil {
			return errors.Errorf("failed to start app: %w", err)
		}

		return nil
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

// initConfig reads in config file and ENV variables if set
func initConfig() {
	viper.SetConfigType("yaml")

	viper.AddConfigPath(".")
	viper.SetConfigName("middleware-offchain.config.yaml")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in
	err := viper.ReadInConfig()
	if err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
