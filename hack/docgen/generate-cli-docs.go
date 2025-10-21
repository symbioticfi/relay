package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra/doc"
	relay "github.com/symbioticfi/relay/cmd/relay/root"
	utils "github.com/symbioticfi/relay/cmd/utils/root"
)

func main() {
	identity := func(s string) string { return s }
	headerPrepender := func(filename string) string {
		// The default header looks like `Argocd app get`. The leading capital letter is off-putting.
		// This header overrides the default. It's better visually and for search results.
		filename = filepath.Base(filename)
		filename = filename[:len(filename)-3] // Drop the '.md'
		return fmt.Sprintf("# `%s` Command Reference\n\n", strings.ReplaceAll(filename, "_", " "))
	}
	// Create docs/cli directory
	docsDir := "docs/cli"
	if err := os.MkdirAll(docsDir+"/utils", 0755); err != nil {
		log.Fatalf("Failed to create docs directory: %v", err)
	}

	if err := os.MkdirAll(docsDir+"/relay", 0755); err != nil {
		log.Fatalf("Failed to create docs directory: %v", err)
	}

	// Disable auto-generated timestamp line
	utilsCmd := utils.NewRootCommand()
	utilsCmd.DisableAutoGenTag = true
	err := doc.GenMarkdownTreeCustom(utilsCmd, "docs/cli/utils", headerPrepender, identity)
	if err != nil {
		log.Fatal(err)
	}

	relayCmd := relay.NewRootCommand()
	relayCmd.DisableAutoGenTag = true
	err = doc.GenMarkdownTreeCustom(relayCmd, "docs/cli/relay", headerPrepender, identity)
	if err != nil {
		log.Fatal(err)
	}
}
