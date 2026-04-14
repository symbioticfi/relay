package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"

	bboltrepo "github.com/symbioticfi/relay/internal/client/repository/bbolt"
	"github.com/symbioticfi/relay/internal/client/repository/repoutil"
)

var (
	flagDB             string
	flagNoFreelistSync bool
	flagFreelistType   string
)

func main() {
	stop := installTimestampedStdout()
	defer stop()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// installTimestampedStdout pipes os.Stdout through a goroutine that prepends
// "HH:MM:SS.mmm " to every line. Returns a stop func that flushes and restores
// the original stdout.
func installTimestampedStdout() func() {
	orig := os.Stdout
	pr, pw, err := os.Pipe()
	if err != nil {
		return func() {}
	}
	os.Stdout = pw

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		br := bufio.NewReader(pr)
		atLineStart := true
		buf := make([]byte, 4096)
		for {
			n, err := br.Read(buf)
			if n > 0 {
				writeWithTimestamp(orig, buf[:n], &atLineStart)
			}
			if err != nil {
				return
			}
		}
	}()

	return func() {
		pw.Close()
		wg.Wait()
		os.Stdout = orig
	}
}

func writeWithTimestamp(w io.Writer, p []byte, atLineStart *bool) {
	for len(p) > 0 {
		if *atLineStart {
			io.WriteString(w, time.Now().Format("15:04:05.000 "))
			*atLineStart = false
		}
		nl := -1
		for i, b := range p {
			if b == '\n' {
				nl = i
				break
			}
		}
		if nl < 0 {
			w.Write(p)
			return
		}
		w.Write(p[:nl+1])
		p = p[nl+1:]
		*atLineStart = true
	}
}

var rootCmd = &cobra.Command{
	Use:   "bbolt-bench",
	Short: "Benchmark bbolt performance under various conditions",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagDB, "db", "relay-test.db", "path to database file")
	rootCmd.PersistentFlags().BoolVar(&flagNoFreelistSync, "no-freelist-sync", false, "bbolt NoFreelistSync setting")
	rootCmd.PersistentFlags().StringVar(&flagFreelistType, "freelist-type", "", "bbolt FreelistType: array (default) or hashmap")

	rootCmd.AddCommand(populateCmd)
	rootCmd.AddCommand(benchWriteCmd)
	rootCmd.AddCommand(pruneCmd)
	rootCmd.AddCommand(benchOpenCmd)
	rootCmd.AddCommand(compactCmd)
	rootCmd.AddCommand(runAllCmd)
}

type openRepoOpts struct {
	noSync        bool
	maxBatchDelay time.Duration
	maxBatchSize  int
}

func openRepo(dbPath string, noFreelistSync bool, opts ...openRepoOpts) (*bboltrepo.Repository, error) {
	dir := filepath.Dir(dbPath)
	if dir == "" {
		dir = "."
	}
	filename := filepath.Base(dbPath)

	cfg := bboltrepo.Config{
		Dir:            dir,
		DBFilename:     filename,
		Metrics:        repoutil.DoNothingMetrics{},
		NoFreelistSync: noFreelistSync,
		FreelistType:   bolt.FreelistType(flagFreelistType),
	}
	if len(opts) > 0 {
		cfg.NoSync = opts[0].noSync
		cfg.MaxBatchDelay = opts[0].maxBatchDelay
		cfg.MaxBatchSize = opts[0].maxBatchSize
	}
	start := time.Now()
	repo, err := bboltrepo.New(cfg)
	d := time.Since(start)
	size := fileSize(dbPath)
	fmt.Printf("    [open] %s (%s) in %s\n", filename, formatBytes(size), d.Round(time.Millisecond))
	return repo, err
}

func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func removeIfExists(path string) {
	if _, err := os.Stat(path); err == nil {
		os.Remove(path)
	}
}
