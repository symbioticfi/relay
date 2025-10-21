package root

import (
	"runtime"

	"github.com/symbioticfi/relay/cmd/utils/keys"
	"github.com/symbioticfi/relay/cmd/utils/network"
	"github.com/symbioticfi/relay/cmd/utils/operator"
	"github.com/symbioticfi/relay/pkg/log"

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

func NewRootCommand() *cobra.Command {
	rootCmd.PersistentFlags().StringVar(&cfg.logLevel, "log.level", "info", "log level(info, debug, warn, error)")
	rootCmd.PersistentFlags().StringVar(&cfg.logMode, "log.mode", "text", "log mode(pretty, text, json)")

	rootCmd.AddCommand(keys.NewKeysCmd())
	rootCmd.AddCommand(network.NewNetworkCmd())
	rootCmd.AddCommand(operator.NewOperatorCmd())
	rootCmd.AddCommand(versionCommand)

	return rootCmd
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
