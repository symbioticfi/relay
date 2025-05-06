package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-errors/errors"
	"github.com/libp2p/go-libp2p"
	"github.com/spf13/cobra"

	"middleware-offchain/internal/client/p2p"
	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/log"
)

// p2p_listener
func main() {
	fmt.Println(os.Args)

	if err := run(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("error executing command", "error", err)
		os.Exit(1)
	}
	slog.Info("Offchain middleware completed successfully")
}

func run() error {
	rootCmd.PersistentFlags().StringVar(&cfg.logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&cfg.listenAddress, "p2p-listen", "", "P2P listen address, for example '/ip4/127.0.0.1/tcp/8000'")

	return rootCmd.Execute()
}

type config struct {
	logLevel      string
	listenAddress string
}

var cfg config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "p2p_listener",
	Short:         "P2P Listener listens for incoming P2P messages",
	Long:          "P2P Listener listens for incoming P2P messages about new validator set headers",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Init(cfg.logLevel)

		ctx := signalContext(context.Background())

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

		p2pService.SetMessageHandler(func(msg entity.P2PMessage) error {
			slog.InfoContext(ctx, "received message", "message", msg)
			return nil
		})
		slog.InfoContext(ctx, "p2p listener created, waiting for messages")

		<-ctx.Done()

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
