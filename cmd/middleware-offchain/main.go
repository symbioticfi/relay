package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
)

// offchain_middleware --driver-address 0x1f5fE7682E49c20289C20a4cFc8b45d5EB410690 --rpc-url http://127.0.0.1:8545
func main() {
	slog.Info("Running offchain_middleware command", "args", os.Args)

	if err := runRootCMD(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("Error executing command", "error", err)
		os.Exit(1)
	}
	slog.Info("Offchain middleware completed successfully")
}

func runRootCMD() error {
	addRootFlags(rootCmd)

	return rootCmd.Execute()
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "middleware_offchain",
	Short:             "Offchain middleware for signature aggregation",
	Long:              "A P2P service for collecting and aggregating signatures for Ethereum contracts.",
	SilenceUsage:      true,
	SilenceErrors:     true,
	PersistentPreRunE: initConfig,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runApp(signalContext(cmd.Context()))
	},
}

// signalContext returns a context that is canceled if either SIGTERM or SIGINT signal is received.
func signalContext(ctx context.Context) context.Context {
	cnCtx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-c
		slog.Info("Received signal", "signal", sig)
		cancel()
	}()

	return cnCtx
}
