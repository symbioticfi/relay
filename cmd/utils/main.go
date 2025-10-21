package main

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/symbioticfi/relay/cmd/utils/root"
)

func main() {
	if err := root.NewRootCommand().Execute(); err != nil {
		pterm.Error.Println("Error executing command", err)
		os.Exit(1)
	}
}
