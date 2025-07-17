package main

import (
	"os"
	"runtime"

	"github.com/symbiotic/relay/cmd/utils/keys"
	"github.com/symbiotic/relay/cmd/utils/network"
	"github.com/symbiotic/relay/cmd/utils/operator"
	"github.com/symbiotic/relay/pkg/log"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type config struct {
	logLevel string
	logMode  string
}

var Version = "local"
var BuildTime = "unknown"

var cfg config

func main() {
	rootCmd.PersistentFlags().StringVar(&cfg.logLevel, "log-level", "info", "log level")
	rootCmd.PersistentFlags().StringVar(&cfg.logMode, "log-mode", "debug", "log mode")

	rootCmd.AddCommand(keys.NewKeysCmd())
	rootCmd.AddCommand(network.NewNetworkCmd())
	rootCmd.AddCommand(operator.NewOperatorCmd())
	rootCmd.AddCommand(versionCommand)

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

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "Print the version of the utils tool",
	Run: func(cmd *cobra.Command, args []string) {
		pterm.Info.Println("Utils tool version:", Version)
		pterm.Info.Println("Go version:", runtime.Version())
		pterm.Info.Println("OS/Arch:", runtime.GOOS+"/"+runtime.GOARCH)
		pterm.Info.Println("Build time:", BuildTime)
	},
}
