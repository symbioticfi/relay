package main

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"

	bboltrepo "github.com/symbioticfi/relay/internal/client/repository/bbolt"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

var runAllCmd = &cobra.Command{
	Use:   "run-all",
	Short: "Run full benchmark pipeline with sync=ON and sync=OFF, print comparison table",
	RunE: func(cmd *cobra.Command, args []string) error {
		epochs, _ := cmd.Flags().GetInt("epochs")
		cycles, _ := cmd.Flags().GetInt("cycles")
		validators, _ := cmd.Flags().GetInt("validators")
		reqInterval, _ := cmd.Flags().GetDuration("req-interval")
		requests, _ := cmd.Flags().GetInt("requests")
		batchSize, _ := cmd.Flags().GetInt("batch-size")
		maxBatchDelay, _ := cmd.Flags().GetDuration("max-batch-delay")
		maxBatchSize, _ := cmd.Flags().GetInt("max-batch-size")

		base := strings.TrimSuffix(flagDB, filepath.Ext(flagDB))
		ext := filepath.Ext(flagDB)
		if ext == "" {
			ext = ".db"
		}

		requestsPerEpoch := int(time.Hour / reqInterval)

		runMode := func(noFreelistSync bool) (RunAllResults, error) {
			mode := "sync-on"
			if noFreelistSync {
				mode = "sync-off"
			}
			dbPath := base + "-" + mode + ext
			compactedPath := base + "-" + mode + "-compacted" + ext

			removeIfExists(dbPath)
			removeIfExists(compactedPath)

			modeStart := time.Now()
			fmt.Printf("\n%s\n", strings.Repeat("=", 60))
			fmt.Printf("=== Mode: no-freelist-sync=%v (%s) — started at %s ===\n",
				noFreelistSync, mode, modeStart.Format("15:04:05"))
			fmt.Printf("%s\n", strings.Repeat("=", 60))

			td, err := GenerateTestData(validators)
			if err != nil {
				return RunAllResults{}, fmt.Errorf("generate test data (%s): %w", mode, err)
			}

			// Resolve bench-write batch settings (0 means use bbolt default).
			benchDelay := maxBatchDelay
			if benchDelay == 0 {
				benchDelay = 10 * time.Millisecond
			}
			benchSize := maxBatchSize
			if benchSize == 0 {
				benchSize = 1000
			}
			switchToPopulate := func(repo *bboltrepo.Repository) {
				repo.ApplyBatchOpts(true, 500*time.Microsecond, 10000)
			}
			switchToBench := func(repo *bboltrepo.Repository) {
				repo.ApplyBatchOpts(false, benchDelay, benchSize)
			}

			// --- Step 1: Bench-write on empty DB ---
			// Single repo open for seed + bench-write, then close and wipe.
			step1Start := time.Now()
			fmt.Printf("\n--- Step 1: Bench-write (empty DB) ---\n")
			emptyRepo, err := openRepo(dbPath, noFreelistSync, openRepoOpts{
				maxBatchDelay: benchDelay, maxBatchSize: benchSize,
			})
			if err != nil {
				return RunAllResults{}, fmt.Errorf("open empty db (%s): %w", mode, err)
			}
			if err := emptyRepo.SaveNextValsetData(context.Background(), td.MakeNextValsetData(1)); err != nil {
				emptyRepo.Close()
				return RunAllResults{}, fmt.Errorf("seed empty db (%s): %w", mode, err)
			}
			writeEmpty, err := runBenchWriteOnRepo(emptyRepo, BenchWriteConfig{
				DBPath: dbPath, NoFreelistSync: noFreelistSync,
				Requests: requests, Validators: validators,
				MaxBatchDelay: benchDelay, MaxBatchSize: benchSize,
			}, "bench-write (empty DB, "+mode+")")
			emptyRepo.Close()
			if err != nil {
				return RunAllResults{}, fmt.Errorf("bench-write empty (%s): %w", mode, err)
			}

			removeIfExists(dbPath)
			fmt.Printf("[step 1 done in %s]\n", time.Since(step1Start).Round(time.Millisecond))

			// Open a single repo for steps 2 + 3. Avoids repeated re-opens which
			// rebuild the freelist on every open when NoFreelistSync is on.
			workRepo, err := openRepo(dbPath, noFreelistSync, openRepoOpts{
				maxBatchDelay: benchDelay, maxBatchSize: benchSize,
			})
			if err != nil {
				return RunAllResults{}, fmt.Errorf("open work repo (%s): %w", mode, err)
			}

			// --- Step 2: Populate epoch by epoch with bench-write after each ---
			step2Start := time.Now()
			fmt.Printf("\n--- Step 2: Populate + bench-write per epoch ---\n")
			var populateWrites []EpochWriteResult

			for epoch := 1; epoch <= epochs; epoch++ {
				fmt.Printf("\n  [populate epoch %d/%d]\n", epoch, epochs)
				switchToPopulate(workRepo)
				if err := populateSingleEpoch(workRepo, td, epoch, requestsPerEpoch, dbPath, epochs); err != nil {
					workRepo.Close()
					return RunAllResults{}, fmt.Errorf("populate epoch %d (%s): %w", epoch, mode, err)
				}

				fmt.Printf("  [bench-write after epoch %d]\n", epoch)
				switchToBench(workRepo)
				result, err := runBenchWriteOnRepo(workRepo, BenchWriteConfig{
					DBPath: dbPath, NoFreelistSync: noFreelistSync,
					Requests: requests, Validators: validators,
					WriteEpoch:    symbiotic.Epoch(epoch),
					MaxBatchDelay: benchDelay, MaxBatchSize: benchSize,
				}, fmt.Sprintf("after populate epoch %d (%s)", epoch, mode))
				if err != nil {
					workRepo.Close()
					return RunAllResults{}, fmt.Errorf("bench-write after populate epoch %d (%s): %w", epoch, mode, err)
				}
				totalSigs := requests * validators
				populateWrites = append(populateWrites, EpochWriteResult{
					Epoch:     epoch,
					P95:       pct(sortDurations(result.Samples), 0.95),
					P50:       pct(sortDurations(result.Samples), 0.50),
					FileSize:  fileSize(dbPath),
					Duration:  result.TotalDuration,
					SigPerSec: float64(totalSigs) / result.TotalDuration.Seconds(),
				})
			}

			// Collect stats after all populate (no reopen needed).
			fileSizeAfterPopulate := fileSize(dbPath)
			statsAfterPopulate := workRepo.Stats()
			fmt.Printf("[step 2 done in %s]\n", time.Since(step2Start).Round(time.Second))

			// --- Step 3: Steady-state cycles (prune oldest + populate new) ---
			// Simulates production: DB keeps ~N epochs; each cycle drops the
			// oldest and appends a new one. Bench-write is measured after each
			// prune AND after each populate so we see the full signal.
			step3Start := time.Now()
			fmt.Printf("\n--- Step 3: Steady-state cycles (%d × prune+populate) ---\n", cycles)

			latestEpoch, err := workRepo.GetLatestValidatorSetEpoch(context.Background())
			if err != nil {
				workRepo.Close()
				return RunAllResults{}, fmt.Errorf("get latest epoch (%s): %w", mode, err)
			}
			oldestEpoch, err := workRepo.GetOldestValidatorSetEpoch(context.Background())
			if err != nil {
				workRepo.Close()
				return RunAllResults{}, fmt.Errorf("get oldest epoch (%s): %w", mode, err)
			}

			var pruneWrites []EpochWriteResult

			for cycle := 1; cycle <= cycles; cycle++ {
				// --- Prune oldest epoch ---
				pruneEpoch := oldestEpoch
				fmt.Printf("\n  [cycle %d/%d: prune epoch %d]\n", cycle, cycles, pruneEpoch)

				switchToBench(workRepo)
				pruneStart := time.Now()
				if err := workRepo.PruneProofEntities(context.Background(), pruneEpoch, batchSize); err != nil {
					workRepo.Close()
					return RunAllResults{}, fmt.Errorf("prune proof epoch %d (%s): %w", pruneEpoch, mode, err)
				}
				if err := workRepo.PruneSignatureEntitiesForEpoch(context.Background(), pruneEpoch, batchSize); err != nil {
					workRepo.Close()
					return RunAllResults{}, fmt.Errorf("prune sig epoch %d (%s): %w", pruneEpoch, mode, err)
				}
				if err := workRepo.PruneValsetEntities(context.Background(), pruneEpoch, batchSize); err != nil {
					workRepo.Close()
					return RunAllResults{}, fmt.Errorf("prune valset epoch %d (%s): %w", pruneEpoch, mode, err)
				}
				if err := workRepo.PruneRequestIDEpochIndices(context.Background(), pruneEpoch, batchSize); err != nil {
					workRepo.Close()
					return RunAllResults{}, fmt.Errorf("prune reqid epoch %d (%s): %w", pruneEpoch, mode, err)
				}
				fmt.Printf("  Pruned epoch %d in %s\n", pruneEpoch, time.Since(pruneStart).Round(time.Millisecond))
				oldestEpoch++

				fmt.Printf("  [bench-write after pruning epoch %d]\n", pruneEpoch)
				pruneResult, err := runBenchWriteOnRepo(workRepo, BenchWriteConfig{
					DBPath: dbPath, NoFreelistSync: noFreelistSync,
					WriteEpoch: latestEpoch,
					Requests:   requests, Validators: validators,
					MaxBatchDelay: benchDelay, MaxBatchSize: benchSize,
				}, fmt.Sprintf("after pruning epoch %d (%s)", pruneEpoch, mode))
				if err != nil {
					workRepo.Close()
					return RunAllResults{}, fmt.Errorf("bench-write after prune epoch %d (%s): %w", pruneEpoch, mode, err)
				}
				pruneWrites = append(pruneWrites, EpochWriteResult{
					Epoch:     int(pruneEpoch),
					P95:       pct(sortDurations(pruneResult.Samples), 0.95),
					P50:       pct(sortDurations(pruneResult.Samples), 0.50),
					FileSize:  fileSize(dbPath),
					Duration:  pruneResult.TotalDuration,
					SigPerSec: float64(requests*validators) / pruneResult.TotalDuration.Seconds(),
				})

				// --- Populate next epoch ---
				newEpoch := int(latestEpoch) + 1
				fmt.Printf("\n  [cycle %d/%d: populate epoch %d]\n", cycle, cycles, newEpoch)
				switchToPopulate(workRepo)
				if err := populateSingleEpoch(workRepo, td, newEpoch, requestsPerEpoch, dbPath, int(latestEpoch)+cycles); err != nil {
					workRepo.Close()
					return RunAllResults{}, fmt.Errorf("populate epoch %d (%s): %w", newEpoch, mode, err)
				}
				latestEpoch = symbiotic.Epoch(newEpoch)

				fmt.Printf("  [bench-write after populate epoch %d]\n", newEpoch)
				switchToBench(workRepo)
				popResult, err := runBenchWriteOnRepo(workRepo, BenchWriteConfig{
					DBPath: dbPath, NoFreelistSync: noFreelistSync,
					Requests: requests, Validators: validators,
					WriteEpoch:    latestEpoch,
					MaxBatchDelay: benchDelay, MaxBatchSize: benchSize,
				}, fmt.Sprintf("after populate epoch %d (%s)", newEpoch, mode))
				if err != nil {
					workRepo.Close()
					return RunAllResults{}, fmt.Errorf("bench-write after populate epoch %d (%s): %w", newEpoch, mode, err)
				}
				totalSigs := requests * validators
				populateWrites = append(populateWrites, EpochWriteResult{
					Epoch:     newEpoch,
					P95:       pct(sortDurations(popResult.Samples), 0.95),
					P50:       pct(sortDurations(popResult.Samples), 0.50),
					FileSize:  fileSize(dbPath),
					Duration:  popResult.TotalDuration,
					SigPerSec: float64(totalSigs) / popResult.TotalDuration.Seconds(),
				})
			}

			// Stats after all prune (no reopen needed). Close before step 4 so
			// bench-open measures a real cold open.
			statsAfterPrune := workRepo.Stats()
			fileSizeBeforeCompact := fileSize(dbPath)
			workRepo.Close()
			fmt.Printf("[step 3 done in %s]\n", time.Since(step3Start).Round(time.Second))

			// --- Step 4: Bench-open ---
			step4Start := time.Now()
			fmt.Printf("\n--- Step 4: Bench-open ---\n")
			openTime, err := runBenchOpen(BenchOpenConfig{
				DBPath: dbPath, NoFreelistSync: noFreelistSync, Iterations: 1,
			})
			if err != nil {
				return RunAllResults{}, fmt.Errorf("bench-open (%s): %w", mode, err)
			}
			fmt.Printf("[step 4 done in %s]\n", time.Since(step4Start).Round(time.Millisecond))

			// --- Step 5: Compact ---
			step5Start := time.Now()
			fmt.Printf("\n--- Step 5: Compact ---\n")
			compactDuration, err := runCompact(CompactConfig{DBPath: dbPath, Output: compactedPath})
			if err != nil {
				return RunAllResults{}, fmt.Errorf("compact (%s): %w", mode, err)
			}
			fileSizeAfterCompact := fileSize(compactedPath)
			fmt.Printf("[step 5 done in %s]\n", time.Since(step5Start).Round(time.Second))

			// --- Step 6: Bench-open on compacted ---
			step6Start := time.Now()
			fmt.Printf("\n--- Step 6: Bench-open (compacted) ---\n")
			openTimeCompacted, err := runBenchOpen(BenchOpenConfig{
				DBPath: compactedPath, NoFreelistSync: noFreelistSync, Iterations: 1,
			})
			if err != nil {
				return RunAllResults{}, fmt.Errorf("bench-open compacted (%s): %w", mode, err)
			}
			fmt.Printf("[step 6 done in %s]\n", time.Since(step6Start).Round(time.Millisecond))

			// --- Step 7: Bench-write after compact ---
			step7Start := time.Now()
			fmt.Printf("\n--- Step 7: Bench-write (after compact) ---\n")
			writeAfterCompact, err := runBenchWrite(BenchWriteConfig{
				DBPath: compactedPath, NoFreelistSync: noFreelistSync,
				Requests: requests, Validators: validators,
				WriteEpoch:    latestEpoch,
				MaxBatchDelay: maxBatchDelay, MaxBatchSize: maxBatchSize,
			}, "bench-write (after compact, "+mode+")")
			if err != nil {
				return RunAllResults{}, fmt.Errorf("bench-write after compact (%s): %w", mode, err)
			}

			writeEmpty.Samples = sortDurations(writeEmpty.Samples)
			writeAfterCompact.Samples = sortDurations(writeAfterCompact.Samples)
			fmt.Printf("[step 7 done in %s]\n", time.Since(step7Start).Round(time.Millisecond))

			fmt.Printf("\n=== Mode %s finished — total %s ===\n",
				mode, time.Since(modeStart).Round(time.Second))

			// Cleanup DB files
			removeIfExists(dbPath)
			removeIfExists(compactedPath)

			return RunAllResults{
				WriteEmpty:            writeEmpty,
				WritePerPopulateEpoch: populateWrites,
				WritePerPruneEpoch:    pruneWrites,
				WriteAfterCompact:     writeAfterCompact,
				OpenTime:              openTime,
				OpenTimeCompacted:     openTimeCompacted,
				CompactDuration:       compactDuration,
				StatsAfterPopulate:    statsAfterPopulate,
				StatsAfterPrune:       statsAfterPrune,
				FileSizeAfterPopulate: fileSizeAfterPopulate,
				FileSizeBeforeCompact: fileSizeBeforeCompact,
				FileSizeAfterCompact:  fileSizeAfterCompact,
			}, nil
		}

		overallStart := time.Now()

		onStart := time.Now()
		onResults, err := runMode(false)
		if err != nil {
			return err
		}
		onDuration := time.Since(onStart)

		offStart := time.Now()
		offResults, err := runMode(true)
		if err != nil {
			return err
		}
		offDuration := time.Since(offStart)

		fmt.Printf("\n%s\n", strings.Repeat("=", 60))
		fmt.Printf("=== SUMMARY ===\n")
		fmt.Printf("%s\n", strings.Repeat("=", 60))
		fmt.Printf("Total time: %s (sync-on: %s, sync-off: %s)\n",
			time.Since(overallStart).Round(time.Second),
			onDuration.Round(time.Second),
			offDuration.Round(time.Second))
		printSummaryTable(onResults, offResults)
		return nil
	},
}

func init() {
	runAllCmd.Flags().Int("epochs", 30, "number of epochs to populate initially (step 2)")
	runAllCmd.Flags().Int("cycles", 10, "number of steady-state prune+populate cycles (step 3)")
	runAllCmd.Flags().Int("validators", 10, "number of validators")
	runAllCmd.Flags().Duration("req-interval", 100*time.Millisecond, "simulated request interval")
	runAllCmd.Flags().Int("requests", 1000, "number of requestIDs for bench-write")
	runAllCmd.Flags().Int("batch-size", 100, "prune batch size")
	runAllCmd.Flags().Duration("max-batch-delay", 0, "bbolt MaxBatchDelay for bench-write (0 = bbolt default 10ms)")
	runAllCmd.Flags().Int("max-batch-size", 0, "bbolt MaxBatchSize for bench-write (0 = bbolt default 1000)")
}

func sortDurations(samples []time.Duration) []time.Duration {
	cp := make([]time.Duration, len(samples))
	copy(cp, samples)
	slices.Sort(cp)
	return cp
}
