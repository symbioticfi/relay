package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"

	bboltrepo "github.com/symbioticfi/relay/internal/client/repository/bbolt"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type PopulateConfig struct {
	DBPath         string
	NoFreelistSync bool
	Epochs         int
	Validators     int
	ReqInterval    time.Duration
}

var populateCmd = &cobra.Command{
	Use:   "populate",
	Short: "Fill database with N epochs of realistic data",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := PopulateConfig{
			DBPath:         flagDB,
			NoFreelistSync: flagNoFreelistSync,
		}
		cfg.Epochs, _ = cmd.Flags().GetInt("epochs")
		cfg.Validators, _ = cmd.Flags().GetInt("validators")
		cfg.ReqInterval, _ = cmd.Flags().GetDuration("req-interval")
		return runPopulate(cfg)
	},
}

func init() {
	populateCmd.Flags().Int("epochs", 20, "number of epochs to populate")
	populateCmd.Flags().Int("validators", 10, "number of validators")
	populateCmd.Flags().Duration("req-interval", 100*time.Millisecond, "simulated request interval (determines requests per epoch)")
}

func runPopulate(cfg PopulateConfig) error {
	fmt.Printf("Generating test data (%d validators, BLS bn254 keys)...\n", cfg.Validators)
	td, err := GenerateTestData(cfg.Validators)
	if err != nil {
		return fmt.Errorf("generate test data: %w", err)
	}

	// NoSync skips fsync — commits are instant, so minimize batch delay
	// to unblock workers faster. Large batches don't help when commit is free.
	repo, err := openRepo(cfg.DBPath, cfg.NoFreelistSync, openRepoOpts{
		noSync:        true,
		maxBatchDelay: 500 * time.Microsecond,
		maxBatchSize:  10000,
	})
	if err != nil {
		return fmt.Errorf("open repo: %w", err)
	}
	defer repo.Close()

	requestsPerEpoch := int(time.Hour / cfg.ReqInterval)

	fmt.Printf("Populating %d epochs, %d requests/epoch, %d validators\n",
		cfg.Epochs, requestsPerEpoch, cfg.Validators)
	fmt.Printf("Total signatures: %d\n", cfg.Epochs*requestsPerEpoch*cfg.Validators)

	totalStart := time.Now()

	for epoch := 1; epoch <= cfg.Epochs; epoch++ {
		if err := populateSingleEpoch(repo, td, epoch, requestsPerEpoch, cfg.DBPath, cfg.Epochs); err != nil {
			return err
		}
	}

	fmt.Printf("\nPopulate complete in %s\n", time.Since(totalStart).Round(time.Second))
	printDBStats(cfg.DBPath, repo.Stats())
	return nil
}

func populateSingleEpoch(repo *bboltrepo.Repository, td *TestData, epoch int, requestsPerEpoch int, dbPath string, totalEpochs int) error {
	epochStart := time.Now()
	e := symbiotic.Epoch(epoch)
	ctx := context.Background()

	if err := repo.SaveNextValsetData(ctx, td.MakeNextValsetData(e)); err != nil {
		return fmt.Errorf("save valset data epoch %d: %w", epoch, err)
	}

	const workers = 32
	reqCh := make(chan int, workers)
	var wg sync.WaitGroup
	var firstErr atomic.Value
	var completed atomic.Int64

	for range workers {
		wg.Go(func() {
			for r := range reqCh {
				if firstErr.Load() != nil {
					return
				}

				messageHash := randomBytes32()
				sigs := td.MakeSignatures(e, messageHash)
				requestID := sigs[0].RequestID()

				sigReq := symbiotic.SignatureRequest{
					KeyTag:        symbiotic.KeyTag(15),
					RequiredEpoch: e,
					Message:       messageHash,
				}
				if err := repo.SaveSignatureRequest(ctx, requestID, sigReq); err != nil {
					firstErr.CompareAndSwap(nil, fmt.Errorf("save sig request epoch %d req %d: %w", epoch, r, err))
					return
				}

				for v, sig := range sigs {
					if err := repo.SaveSignature(ctx, sig, td.Validators[v], uint32(v)); err != nil {
						firstErr.CompareAndSwap(nil, fmt.Errorf("save signature epoch %d req %d val %d: %w", epoch, r, v, err))
						return
					}
				}

				completed.Add(1)
			}
		})
	}

	progressDone := make(chan struct{})
	go func() {
		defer close(progressDone)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				done := int(completed.Load())
				elapsed := time.Since(epochStart)
				reqPerSec := float64(done) / elapsed.Seconds()
				remaining := time.Duration(float64(requestsPerEpoch-done)/reqPerSec) * time.Second
				pctDone := float64(done) / float64(requestsPerEpoch) * 100
				fmt.Printf("  epoch %d: %d/%d (%.0f%%) %.0f req/s, ETA %s\n",
					epoch, done, requestsPerEpoch, pctDone, reqPerSec, remaining.Round(time.Second))
			case <-progressDone:
				return
			}
		}
	}()

	for r := range requestsPerEpoch {
		if firstErr.Load() != nil {
			break
		}
		reqCh <- r
	}
	close(reqCh)
	wg.Wait()
	progressDone <- struct{}{}

	if errVal := firstErr.Load(); errVal != nil {
		return errVal.(error)
	}

	fmt.Printf("Epoch %d/%d done in %s, file size: %s\n",
		epoch, totalEpochs,
		time.Since(epochStart).Round(time.Millisecond),
		formatBytes(fileSize(dbPath)))
	return nil
}
