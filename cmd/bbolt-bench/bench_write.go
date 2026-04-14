package main

import (
	"context"
	"fmt"
	"os"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/spf13/cobra"

	bboltrepo "github.com/symbioticfi/relay/internal/client/repository/bbolt"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type BenchWriteConfig struct {
	DBPath         string
	NoFreelistSync bool
	Requests       int
	Validators     int
	Concurrency    int
	WriteEpoch     symbiotic.Epoch
	MaxBatchDelay  time.Duration
	MaxBatchSize   int
}

type BenchWriteResult struct {
	Samples        []time.Duration
	TotalDuration  time.Duration
	RequestCount   int
	ValidatorCount int
}

func (r BenchWriteResult) SigPerSec() float64 {
	if r.TotalDuration == 0 {
		return 0
	}
	return float64(r.RequestCount*r.ValidatorCount) / r.TotalDuration.Seconds()
}

var benchWriteCmd = &cobra.Command{
	Use:   "bench-write",
	Short: "Measure SaveSignature latency on current DB state",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := BenchWriteConfig{
			DBPath:         flagDB,
			NoFreelistSync: flagNoFreelistSync,
		}
		cfg.Requests, _ = cmd.Flags().GetInt("requests")
		cfg.Validators, _ = cmd.Flags().GetInt("validators")
		cfg.Concurrency, _ = cmd.Flags().GetInt("concurrency")
		cfg.MaxBatchDelay, _ = cmd.Flags().GetDuration("max-batch-delay")
		cfg.MaxBatchSize, _ = cmd.Flags().GetInt("max-batch-size")

		if cpuprofile, _ := cmd.Flags().GetString("cpuprofile"); cpuprofile != "" {
			f, err := os.Create(cpuprofile)
			if err != nil {
				return fmt.Errorf("create cpuprofile: %w", err)
			}
			defer f.Close()
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}

		_, err := runBenchWrite(cfg, "bench-write")
		return err
	},
}

func init() {
	benchWriteCmd.Flags().Int("requests", 1000, "number of requestIDs to write")
	benchWriteCmd.Flags().Int("validators", 10, "number of validators per request")
	benchWriteCmd.Flags().Int("concurrency", 10, "number of concurrent writer goroutines")
	benchWriteCmd.Flags().Duration("max-batch-delay", 0, "bbolt MaxBatchDelay (0 = bbolt default 10ms)")
	benchWriteCmd.Flags().Int("max-batch-size", 0, "bbolt MaxBatchSize (0 = bbolt default 1000)")
	benchWriteCmd.Flags().String("cpuprofile", "", "write CPU profile to file")
}

type requestWork struct {
	index      int
	messageHash []byte
	sigs       []symbiotic.Signature
	epoch      symbiotic.Epoch
}

func runBenchWrite(cfg BenchWriteConfig, label string) (BenchWriteResult, error) {
	var opts []openRepoOpts
	if cfg.MaxBatchDelay > 0 || cfg.MaxBatchSize > 0 {
		opts = append(opts, openRepoOpts{
			maxBatchDelay: cfg.MaxBatchDelay,
			maxBatchSize:  cfg.MaxBatchSize,
		})
	}
	repo, err := openRepo(cfg.DBPath, cfg.NoFreelistSync, opts...)
	if err != nil {
		return BenchWriteResult{}, fmt.Errorf("open repo: %w", err)
	}
	defer repo.Close()
	return runBenchWriteOnRepo(repo, cfg, label)
}

func runBenchWriteOnRepo(repo *bboltrepo.Repository, cfg BenchWriteConfig, label string) (BenchWriteResult, error) {
	fmt.Printf("\nGenerating test data (%d validators)...\n", cfg.Validators)
	td, err := GenerateTestData(cfg.Validators)
	if err != nil {
		return BenchWriteResult{}, fmt.Errorf("generate test data: %w", err)
	}

	ctx := context.Background()

	writeEpoch := cfg.WriteEpoch
	if writeEpoch == 0 {
		latestEpoch, err := repo.GetLatestValidatorSetEpoch(ctx)
		if err != nil {
			return BenchWriteResult{}, fmt.Errorf("get latest epoch: %w", err)
		}
		writeEpoch = latestEpoch + 1
		if err := repo.SaveNextValsetData(ctx, td.MakeNextValsetData(writeEpoch)); err != nil {
			return BenchWriteResult{}, fmt.Errorf("save valset for write epoch: %w", err)
		}
	}

	// Pre-generate all request data
	work := make([]requestWork, cfg.Requests)
	for r := range cfg.Requests {
		messageHash := randomBytes32()
		work[r] = requestWork{
			index:       r,
			messageHash: messageHash,
			sigs:        td.MakeSignatures(writeEpoch, messageHash),
			epoch:       writeEpoch,
		}
	}

	concurrency := cfg.Concurrency
	if concurrency < 1 {
		concurrency = 1
	}

	statsBefore := repo.Stats()

	samples := make([]time.Duration, cfg.Requests)
	var mu sync.Mutex
	var firstErr error

	totalStart := time.Now()

	if concurrency == 1 {
		// Sequential path — same as before
		for _, w := range work {
			reqStart := time.Now()
			requestID := w.sigs[0].RequestID()

			sigReq := symbiotic.SignatureRequest{
				KeyTag:        symbiotic.KeyTag(15),
				RequiredEpoch: w.epoch,
				Message:       w.messageHash,
			}
			if err := repo.SaveSignatureRequest(ctx, requestID, sigReq); err != nil {
				return BenchWriteResult{}, fmt.Errorf("save sig request %d: %w", w.index, err)
			}

			for v, sig := range w.sigs {
				if err := repo.SaveSignature(ctx, sig, td.Validators[v], uint32(v)); err != nil {
					return BenchWriteResult{}, fmt.Errorf("save signature req %d val %d: %w", w.index, v, err)
				}
			}

			samples[w.index] = time.Since(reqStart)
		}
	} else {
		// Concurrent path
		ch := make(chan requestWork, concurrency)
		var wg sync.WaitGroup

		for range concurrency {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for w := range ch {
					reqStart := time.Now()
					requestID := w.sigs[0].RequestID()

					sigReq := symbiotic.SignatureRequest{
						KeyTag:        symbiotic.KeyTag(15),
						RequiredEpoch: w.epoch,
						Message:       w.messageHash,
					}
					if err := repo.SaveSignatureRequest(ctx, requestID, sigReq); err != nil {
						mu.Lock()
						if firstErr == nil {
							firstErr = fmt.Errorf("save sig request %d: %w", w.index, err)
						}
						mu.Unlock()
						return
					}

					for v, sig := range w.sigs {
						if err := repo.SaveSignature(ctx, sig, td.Validators[v], uint32(v)); err != nil {
							mu.Lock()
							if firstErr == nil {
								firstErr = fmt.Errorf("save signature req %d val %d: %w", w.index, v, err)
							}
							mu.Unlock()
							return
						}
					}

					samples[w.index] = time.Since(reqStart)
				}
			}()
		}

		for _, w := range work {
			ch <- w
		}
		close(ch)
		wg.Wait()

		if firstErr != nil {
			return BenchWriteResult{}, firstErr
		}
	}

	totalDuration := time.Since(totalStart)
	statsAfter := repo.Stats()

	sampleSlice := make([]time.Duration, 0, cfg.Requests)
	for _, s := range samples {
		if s > 0 {
			sampleSlice = append(sampleSlice, s)
		}
	}

	printLatencyStats(label, sampleSlice, totalDuration, cfg.Requests, cfg.Validators)
	if concurrency > 1 {
		fmt.Printf("Concurrency: %d\n", concurrency)
	}
	printDBStats(cfg.DBPath, statsAfter)
	printTxStatsDelta(statsBefore.TxStats, statsAfter.TxStats)

	return BenchWriteResult{
		Samples:        sampleSlice,
		TotalDuration:  totalDuration,
		RequestCount:   cfg.Requests,
		ValidatorCount: cfg.Validators,
	}, nil
}
