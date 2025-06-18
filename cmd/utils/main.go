package main

import (
	"github.com/spf13/cobra"
	"log/slog"
	"middleware-offchain/cmd/utils/keys"
	"os"
)

func main() {
	keysCmd, err := keys.NewKeysCmd()
	if err != nil {
		slog.Error("error creating keys command", "error", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(keysCmd)
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
}
