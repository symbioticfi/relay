package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/go-errors/errors"
	"github.com/symbioticfi/relay/cmd/relay/root"
)

func main() {
	if err := root.NewRootCommand().Execute(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("Error executing command", "error", err)
		os.Exit(1)
	}
	slog.Info("Relay sidecar completed successfully")
}
