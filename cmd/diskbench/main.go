package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	dir, err := os.MkdirTemp("", "diskbench-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	hostname, _ := os.Hostname()
	fmt.Printf("Host: %s\n", hostname)
	fmt.Printf("Dir:  %s\n\n", dir)

	benchSingleFsync(dir)
	fmt.Println()
	benchBboltContention(dir)

	return nil
}

// benchSingleFsync measures raw fsync latency with no contention.
func benchSingleFsync(dir string) {
	const iterations = 100
	path := filepath.Join(dir, "single")
	data := make([]byte, 4096)
	durations := make([]time.Duration, iterations)

	for i := range iterations {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}

		start := time.Now()
		_, _ = f.Write(data)
		_ = f.Sync()
		durations[i] = time.Since(start)
		f.Close()
	}

	printStats("Single write+fsync (no contention)", durations)
}

// benchBboltContention simulates bbolt's single-writer model:
// N goroutines compete for a single mutex, each doing write+fsync while holding it.
// This is exactly what bbolt db.Update does internally.
func benchBboltContention(dir string) {
	fmt.Println("=== Simulated bbolt contention (mutex + write + fsync) ===")
	fmt.Println("Each goroutine: lock mutex → write 4KB → fsync → unlock")
	fmt.Println()

	for _, writers := range []int{1, 5, 10, 20} {
		benchContentionN(dir, writers)
	}
}

func benchContentionN(dir string, numWriters int) {
	const requestsPerWriter = 10
	data := make([]byte, 4096)

	var mu sync.Mutex
	durations := make([]time.Duration, numWriters*requestsPerWriter)
	var idx int
	var idxMu sync.Mutex

	files := make([]*os.File, numWriters)
	for i := range numWriters {
		path := filepath.Join(dir, fmt.Sprintf("contention-%d-%d", numWriters, i))
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
		files[i] = f
	}

	var wg sync.WaitGroup
	wg.Add(numWriters)

	wallStart := time.Now()

	for w := range numWriters {
		go func() {
			defer wg.Done()
			f := files[w]

			for range requestsPerWriter {
				start := time.Now()

				mu.Lock()
				_, _ = f.Write(data)
				_ = f.Sync()
				mu.Unlock()

				d := time.Since(start)

				idxMu.Lock()
				durations[idx] = d
				idx++
				idxMu.Unlock()
			}
		}()
	}

	wg.Wait()
	wallDuration := time.Since(wallStart)

	for _, f := range files {
		f.Close()
	}

	fmt.Printf("--- %d concurrent writers (%d ops each, %d total) wall=%v ---\n",
		numWriters, requestsPerWriter, numWriters*requestsPerWriter, wallDuration.Round(time.Microsecond))
	printStats("  Per-operation (includes wait)", durations[:idx])
	fmt.Println()
}

func printStats(label string, durations []time.Duration) {
	slices.Sort(durations)

	n := len(durations)
	var total time.Duration
	for _, d := range durations {
		total += d
	}

	fmt.Printf("%s (%d samples):\n", label, n)
	fmt.Printf("  avg=%-12v min=%-12v max=%v\n",
		(total / time.Duration(n)).Round(time.Microsecond),
		durations[0].Round(time.Microsecond),
		durations[n-1].Round(time.Microsecond))
	fmt.Printf("  p50=%-12v p95=%-12v p99=%v\n",
		durations[n*50/100].Round(time.Microsecond),
		durations[n*95/100].Round(time.Microsecond),
		durations[n*99/100].Round(time.Microsecond))
}
