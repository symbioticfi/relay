package main

import (
	"log/slog"
	"middleware-offchain/cmd/utils/keys"
	"middleware-offchain/cmd/utils/network"
	"middleware-offchain/cmd/utils/operator"
	"middleware-offchain/pkg/log"
	"os"

	"github.com/spf13/cobra"
)

type config struct {
	logLevel string
	logMode  string
}

var cfg config

func main() {
	keysCmd, err := keys.NewKeysCmd()
	if err != nil {
		slog.Error("error creating keys command", "error", err)
		os.Exit(1)
	}

	networkCmd, err := network.NewNetworkCmd()
	if err != nil {
		slog.Error("error creating network command", "error", err)
		os.Exit(1)
	}

	operatorCmd, err := operator.NewOperatorCmd()
	if err != nil {
		slog.Error("error creating network command", "error", err)
		os.Exit(1)
	}

	rootCmd.PersistentFlags().StringVar(&cfg.logLevel, "log-level", "info", "log level")
	rootCmd.PersistentFlags().StringVar(&cfg.logMode, "log-mode", "debug", "log mode")

	rootCmd.AddCommand(keysCmd)
	rootCmd.AddCommand(networkCmd)
	rootCmd.AddCommand(operatorCmd)

	if err := run(); err != nil {
		slog.Error("error executing command", "error", err)
		os.Exit(1)
	}
}

func run() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "utils",
	Short: "Utils tool",
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Init(cfg.logLevel, cfg.logMode)
	},
}
