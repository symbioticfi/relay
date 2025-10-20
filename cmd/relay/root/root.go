package root

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var Version = "local"
var BuildTime = "unknown"

func NewRootCommand() *cobra.Command {
	slog.Info("Running relay_sidecar command",
		"version", Version,
		"buildTime", BuildTime,
		"args", os.Args,
	)

	addRootFlags(rootCmd)

	return rootCmd
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "relay_sidecar",
	Short:             "Relay sidecar for signature aggregation",
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
