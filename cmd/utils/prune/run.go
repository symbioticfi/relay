package prune

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-errors/errors"
	"github.com/pterm/pterm"

	"github.com/symbioticfi/relay/internal/client/repository/badger"
	"github.com/symbioticfi/relay/internal/client/repository/bbolt"
	"github.com/symbioticfi/relay/internal/client/repository/repoutil"
	"github.com/symbioticfi/relay/internal/usecase/pruner"
)

const (
	storageTypeBadger = "badger"
	storageTypeBbolt  = "bbolt"

	bboltDBFilename = "relay.db"
)

var (
	badgerFilePatterns = []string{"*.vlog", "MANIFEST"}
	bboltFilePatterns  = []string{bboltDBFilename}
)

func run(ctx context.Context, f Flags) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if f.ValsetEpochs == 0 && f.ProofEpochs == 0 && f.SignatureEpochs == 0 && !f.Compact {
		return errors.New("nothing to do: pass at least one --retention.* flag or --compact")
	}

	storageType, err := detectStorageType(f.StorageDir)
	if err != nil {
		return err
	}
	pterm.Info.Printf("Detected %s storage in %s\n", storageType, f.StorageDir)

	switch storageType {
	case storageTypeBbolt:
		return runBbolt(ctx, f)
	case storageTypeBadger:
		return runBadger(ctx, f)
	default:
		return errors.Errorf("unsupported storage type: %s", storageType)
	}
}

func runBbolt(ctx context.Context, f Flags) error {
	dbPath := filepath.Join(f.StorageDir, bboltDBFilename)

	repo, err := bbolt.New(bbolt.Config{
		Dir:        f.StorageDir,
		DBFilename: bboltDBFilename,
		Metrics:    repoutil.DoNothingMetrics{},
		// Speed-tuned for offline use: no fsync per write, no inter-batch
		// yielding, no startup compaction (handled below). The prune path
		// uses db.Batch internally; with a single offline writer no peers
		// join the batch, so MaxBatchDelay = per-call flush wait. Drop from
		// bbolt's 10ms default to 1ms.
		PrunePause:       0,
		MaxBatchDelay:    time.Millisecond,
		NoSync:           true,
		NoFreelistSync:   true,
		CompactOnStartup: false,
	})
	if err != nil {
		return errors.Errorf("failed to open bbolt repository at %s (relay still running, or directory locked?): %w", f.StorageDir, err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = repo.Close()
		}
	}()

	if err := runPruneOnce(ctx, pruner.Config{
		Repo:                     repo,
		Metrics:                  pruner.NoopMetrics{},
		Enabled:                  true,
		Interval:                 time.Hour,
		ValsetRetentionEpochs:    f.ValsetEpochs,
		ProofRetentionEpochs:     f.ProofEpochs,
		SignatureRetentionEpochs: f.SignatureEpochs,
		PruneBatchSize:           1000,
	}, f); err != nil {
		return err
	}

	closed = true
	if err := repo.Close(); err != nil {
		return errors.Errorf("failed to close bbolt repository: %w", err)
	}

	if f.Compact {
		before, beforeErr := fileSize(dbPath)
		spinner, _ := pterm.DefaultSpinner.Start("Compacting bbolt database…")
		start := time.Now()
		if err := bbolt.CompactDB(dbPath); err != nil {
			spinner.Fail("Compaction failed")
			return errors.Errorf("bbolt compaction failed: %w", err)
		}
		after, afterErr := fileSize(dbPath)
		spinner.Success("Compaction done")
		printSizeReport(before, beforeErr, after, afterErr, time.Since(start))
	}

	return nil
}

func runBadger(ctx context.Context, f Flags) error {
	// Use badger's own defaults for sizing/compaction knobs (zero values
	// fall through in applyBadgerTuning). Hard-coding here would silently
	// diverge from the sidecar config and create vlog files of mismatched
	// sizes on the next live-run. CompactL0OnClose=true ensures the final
	// L0 compaction runs when we Close after Flatten.
	repo, err := badger.New(badger.Config{
		Dir:              f.StorageDir,
		Metrics:          repoutil.DoNothingMetrics{},
		BlockCacheSize:   -1, // -1 = badger default; 0 means "disabled"
		CompactL0OnClose: true,
	})
	if err != nil {
		return errors.Errorf("failed to open badger repository at %s (relay still running, or directory locked?): %w", f.StorageDir, err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = repo.Close()
		}
	}()

	if err := runPruneOnce(ctx, pruner.Config{
		Repo:                     repo,
		Metrics:                  pruner.NoopMetrics{},
		Enabled:                  true,
		Interval:                 time.Hour,
		ValsetRetentionEpochs:    f.ValsetEpochs,
		ProofRetentionEpochs:     f.ProofEpochs,
		SignatureRetentionEpochs: f.SignatureEpochs,
		PruneBatchSize:           1000,
	}, f); err != nil {
		return err
	}

	if f.Compact {
		before, beforeErr := dirSize(f.StorageDir)
		spinner, _ := pterm.DefaultSpinner.Start("Flattening badger LSM + value log GC…")
		start := time.Now()
		capHit, err := repo.Flatten(f.BadgerFlattenWorkers)
		if err != nil {
			spinner.Fail("Flatten failed")
			return errors.Errorf("badger flatten failed: %w", err)
		}
		spinner.Success("Flatten done")
		if capHit {
			pterm.Warning.Printf("value-log GC hit iteration cap (%d) — re-run with --compact to continue reclaiming space\n",
				badger.MaxValueLogGCIterations)
		}
		// Final L0 compaction happens in Close (CompactL0OnClose=true).
		// Mark closed before the error check so the deferred Close can't run
		// on a partially-closed DB.
		closed = true
		if err := repo.Close(); err != nil {
			return errors.Errorf("failed to close badger repository: %w", err)
		}
		after, afterErr := dirSize(f.StorageDir)
		printSizeReport(before, beforeErr, after, afterErr, time.Since(start))
	}

	return nil
}

func runPruneOnce(ctx context.Context, cfg pruner.Config, f Flags) error {
	if f.ValsetEpochs == 0 && f.ProofEpochs == 0 && f.SignatureEpochs == 0 {
		pterm.Info.Println("Skipping pruning (no --retention.* flags set)")
		return nil
	}
	svc, err := pruner.New(cfg)
	if err != nil {
		return errors.Errorf("failed to construct pruner: %w", err)
	}
	pterm.Info.Printf("Pruning (valset=%d proof=%d signature=%d)…\n",
		f.ValsetEpochs, f.ProofEpochs, f.SignatureEpochs)
	start := time.Now()
	if err := svc.RunOnce(ctx); err != nil {
		return errors.Errorf("pruning failed: %w", err)
	}
	pterm.Success.Printf("Pruning completed in %s\n", time.Since(start).Round(time.Millisecond))
	return nil
}

func detectStorageType(dir string) (string, error) {
	bb, err := matchAny(dir, bboltFilePatterns)
	if err != nil {
		return "", err
	}
	bd, err := matchAny(dir, badgerFilePatterns)
	if err != nil {
		return "", err
	}
	switch {
	case bb && bd:
		return "", errors.Errorf("storage directory %q contains both bbolt and badger files; please clean up first", dir)
	case bb:
		return storageTypeBbolt, nil
	case bd:
		return storageTypeBadger, nil
	default:
		return "", errors.Errorf("storage directory %q does not contain a recognized database (looked for %v and %v)", dir, bboltFilePatterns, badgerFilePatterns)
	}
}

func matchAny(dir string, patterns []string) (bool, error) {
	for _, p := range patterns {
		matches, err := filepath.Glob(filepath.Join(dir, p))
		if err != nil {
			return false, errors.Errorf("failed to glob %s: %w", p, err)
		}
		if len(matches) > 0 {
			return true, nil
		}
	}
	return false, nil
}

func fileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func dirSize(dir string) (int64, error) {
	var total int64
	err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total, err
}

func printSizeReport(before int64, beforeErr error, after int64, afterErr error, duration time.Duration) {
	d := duration.Round(time.Millisecond)
	if beforeErr != nil || afterErr != nil {
		pterm.Info.Printf("Compaction took %s (size unavailable: before=%v after=%v)\n", d, beforeErr, afterErr)
		return
	}
	pterm.Info.Printf("Size: %s → %s (%.1f%% reduction) in %s\n",
		humanBytes(before), humanBytes(after), reductionPct(before, after), d)
}

func reductionPct(before, after int64) float64 {
	if before == 0 {
		return 0
	}
	return float64(before-after) / float64(before) * 100
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return pterm.Sprintf("%d B", n)
	}
	const suffixes = "KMGTPE"
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit && exp < len(suffixes)-1; x /= unit {
		div *= unit
		exp++
	}
	return pterm.Sprintf("%.2f %ciB", float64(n)/float64(div), suffixes[exp])
}
