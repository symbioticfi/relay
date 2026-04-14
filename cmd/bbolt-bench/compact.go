package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

type CompactConfig struct {
	DBPath string
	Output string
}

var compactCmd = &cobra.Command{
	Use:   "compact",
	Short: "Run bolt.Compact and measure duration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := CompactConfig{
			DBPath: flagDB,
		}
		cfg.Output, _ = cmd.Flags().GetString("output")
		_, err := runCompact(cfg)
		return err
	},
}

func init() {
	compactCmd.Flags().String("output", "relay-test-compacted.db", "path to compacted output file")
}

func runCompact(cfg CompactConfig) (time.Duration, error) {
	sizeBefore := fileSize(cfg.DBPath)
	fmt.Printf("\nCompacting %s (%s) -> %s\n", cfg.DBPath, formatBytes(sizeBefore), cfg.Output)

	removeIfExists(cfg.Output)

	freelistType := bolt.FreelistType(flagFreelistType)
	src, err := bolt.Open(cfg.DBPath, 0600, &bolt.Options{
		ReadOnly: true, Timeout: 5 * time.Second, FreelistType: freelistType,
	})
	if err != nil {
		return 0, fmt.Errorf("open source db: %w", err)
	}
	defer src.Close()

	dst, err := bolt.Open(cfg.Output, 0600, &bolt.Options{
		Timeout: 5 * time.Second, FreelistType: freelistType,
	})
	if err != nil {
		return 0, fmt.Errorf("create destination db: %w", err)
	}

	const compactTxMaxSize = 64 << 20 // 64 MB per transaction to avoid OOM on large DBs

	start := time.Now()
	if err := bolt.Compact(dst, src, compactTxMaxSize); err != nil {
		dst.Close()
		return 0, fmt.Errorf("compact: %w", err)
	}
	duration := time.Since(start)

	dst.Close()
	src.Close()

	sizeAfter := fileSize(cfg.Output)

	fmt.Printf("Compact complete in %s\n", duration.Round(time.Millisecond))
	fmt.Printf("Size: %s -> %s (ratio: %.1fx)\n",
		formatBytes(sizeBefore), formatBytes(sizeAfter),
		float64(sizeBefore)/float64(sizeAfter))

	return duration, nil
}
