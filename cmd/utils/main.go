package main

import (
	"middleware-offchain/cmd/utils/keys"
	"middleware-offchain/cmd/utils/network"
	"middleware-offchain/cmd/utils/operator"
	"middleware-offchain/pkg/log"
	"os"

	"github.com/pterm/pterm"

	"github.com/spf13/cobra"
)

type config struct {
	logLevel string
	logMode  string
}

var cfg config

func main() {
	rootCmd.PersistentFlags().StringVar(&cfg.logLevel, "log-level", "info", "log level")
	rootCmd.PersistentFlags().StringVar(&cfg.logMode, "log-mode", "debug", "log mode")

	rootCmd.AddCommand(keys.NewKeysCmd())
	rootCmd.AddCommand(network.NewNetworkCmd())
	rootCmd.AddCommand(operator.NewOperatorCmd())

	if err := run(); err != nil {
		pterm.Error.Println("Error executing command", err)
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
