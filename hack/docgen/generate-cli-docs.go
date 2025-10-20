package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	relay "github.com/symbioticfi/relay/cmd/relay/root"
	utils "github.com/symbioticfi/relay/cmd/utils/root"
)

func main1() {
	fmt.Println("=== Generating Utils CLI Documentation ===")

	// Create docs/cli directory
	docsDir := "docs/cli"
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		log.Fatalf("Failed to create docs directory: %v", err)
	}

	// Generate documentation for each command
	if err := generateDocs(utils.NewRootCommand(), docsDir, "utils"); err != nil {
		log.Fatalf("Failed to generate documentation: %v", err)
	}

	if err := generateDocs(relay.NewRootCommand(), docsDir, "relay"); err != nil {
		log.Fatalf("Failed to generate documentation: %v", err)
	}

	fmt.Println("\n✓ Utils CLI documentation generated successfully in", docsDir)
}

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

	err := doc.GenMarkdownTreeCustom(utils.NewRootCommand(), "docs/cli/utils", headerPrepender, identity)
	if err != nil {
		log.Fatal(err)
	}

	err = doc.GenMarkdownTreeCustom(relay.NewRootCommand(), "docs/cli/relay", headerPrepender, identity)
	if err != nil {
		log.Fatal(err)
	}
}

// generateDocs recursively generates documentation for the command and all its subcommands
func generateDocs(cmd *cobra.Command, baseDir string, rootName string) error {
	// Generate markdown for this command
	cmdPath := strings.TrimPrefix(cmd.CommandPath(), cmd.Root().Use+" ")

	var cmdDir string
	if cmdPath == "" {
		// For root command
		cmdDir = filepath.Join(baseDir, rootName)
	} else {
		// For subcommands
		cmdDir = filepath.Join(baseDir, strings.ReplaceAll(cmdPath, " ", "/"))
	}

	// Create directory for this command
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", cmdDir, err)
	}

	// Generate markdown documentation
	docFile := filepath.Join(cmdDir, "doc.md")

	// Create the file
	f, err := os.Create(docFile)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", docFile, err)
	}
	defer f.Close()

	// Use custom link handler to generate proper links between command docs
	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, filepath.Ext(name))
		// Convert command name to path (e.g., "utils_keys_add" -> "../keys/add/doc.md")
		parts := strings.Split(base, "_")
		if len(parts) > 1 {
			// Skip the first part (utils) and build relative path
			relPath := strings.Join(parts[1:], "/")
			return "../" + relPath + "/doc.md"
		}
		return ""
	}

	// Generate markdown with custom settings
	if err := doc.GenMarkdownCustom(cmd, f, linkHandler); err != nil {
		return fmt.Errorf("failed to generate markdown for %s: %w", cmd.CommandPath(), err)
	}

	fmt.Printf("  ✓ Generated: %s\n", docFile)

	// Recursively generate docs for subcommands
	for _, subCmd := range cmd.Commands() {
		if subCmd.Hidden {
			continue
		}
		// For subcommands, pass baseDir without rootName since cmdPath already includes parent
		subBaseDir := baseDir
		if cmdPath == "" {
			// If this is root, prefix subcommands with rootName
			subBaseDir = filepath.Join(baseDir, rootName)
		}
		if err := generateDocs(subCmd, subBaseDir, rootName); err != nil {
			return err
		}
	}

	return nil
}
