package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

type BenchOpenConfig struct {
	DBPath         string
	NoFreelistSync bool
	Iterations     int
}

var benchOpenCmd = &cobra.Command{
	Use:   "bench-open",
	Short: "Measure database open time",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := BenchOpenConfig{
			DBPath:         flagDB,
			NoFreelistSync: flagNoFreelistSync,
		}
		cfg.Iterations, _ = cmd.Flags().GetInt("iterations")
		_, err := runBenchOpen(cfg)
		return err
	},
}

func init() {
	benchOpenCmd.Flags().Int("iterations", 1, "number of open/close cycles")
}

func runBenchOpen(cfg BenchOpenConfig) (time.Duration, error) {
	fmt.Printf("\nBenchmarking DB open (iterations=%d, no-freelist-sync=%v)...\n",
		cfg.Iterations, cfg.NoFreelistSync)
	fmt.Printf("File size: %s\n", formatBytes(fileSize(cfg.DBPath)))

	durations := make([]time.Duration, cfg.Iterations)

	for i := range cfg.Iterations {
		start := time.Now()
		repo, err := openRepo(cfg.DBPath, cfg.NoFreelistSync)
		if err != nil {
			return 0, fmt.Errorf("open repo iteration %d: %w", i, err)
		}

		stats := repo.Stats()
		openDuration := time.Since(start)
		durations[i] = openDuration

		repo.Close()

		fmt.Printf("  Iteration %d: %s\n", i+1, openDuration.Round(time.Millisecond))
		if i == 0 {
			printDBStats(cfg.DBPath, stats)
		}
	}

	if cfg.Iterations > 1 {
		var total time.Duration
		minD, maxD := durations[0], durations[0]
		for _, d := range durations {
			total += d
			if d < minD {
				minD = d
			}
			if d > maxD {
				maxD = d
			}
		}
		avg := total / time.Duration(cfg.Iterations)
		fmt.Printf("\nOpen time: min=%s avg=%s max=%s\n",
			minD.Round(time.Millisecond), avg.Round(time.Millisecond), maxD.Round(time.Millisecond))
	}

	return durations[0], nil
}
