package prune

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-errors/errors"
	"github.com/pterm/pterm"
	"golang.org/x/term"

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
	totalStart := time.Now()

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Silence library slog so it doesn't interleave with pterm output and
	// break spinner animations. Errors propagate via returned error values.
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	if f.ValsetEpochs == 0 && f.ProofEpochs == 0 && f.SignatureEpochs == 0 && !f.Compact {
		return errors.New("nothing to do: pass at least one --retention.* flag or --compact")
	}

	// Mirror pruner.Config.Validate so we fail before opening the DB (which
	// takes an exclusive lock and may take seconds on large stores).
	if f.ValsetEpochs > 0 {
		if f.ProofEpochs == 0 || f.SignatureEpochs == 0 {
			return errors.New("--retention.valset-epochs requires --retention.proof-epochs and --retention.signature-epochs to also be set (otherwise pruning valset entities would orphan proof/signature data)")
		}
		if f.ProofEpochs > f.ValsetEpochs {
			return errors.Errorf("--retention.proof-epochs (%d) must be <= --retention.valset-epochs (%d) to avoid orphaning proofs whose valset has been pruned",
				f.ProofEpochs, f.ValsetEpochs)
		}
		if f.SignatureEpochs > f.ValsetEpochs {
			return errors.Errorf("--retention.signature-epochs (%d) must be <= --retention.valset-epochs (%d) to avoid orphaning signatures whose valset has been pruned",
				f.SignatureEpochs, f.ValsetEpochs)
		}
	}

	storageType, err := detectStorageType(f.StorageDir)
	if err != nil {
		return err
	}
	pterm.Info.Printf("Detected %s storage in %s\n", storageType, f.StorageDir)

	if err := confirmBackup(f, storageType); err != nil {
		return err
	}

	var runErr error
	switch storageType {
	case storageTypeBbolt:
		runErr = runBbolt(ctx, f)
	case storageTypeBadger:
		runErr = runBadger(ctx, f)
	default:
		return errors.Errorf("unsupported storage type: %s", storageType)
	}
	if runErr != nil {
		return runErr
	}
	pterm.Success.Printf("Total time: %s\n", time.Since(totalStart).Round(time.Millisecond))
	return nil
}

func runBbolt(ctx context.Context, f Flags) error {
	dbPath := filepath.Join(f.StorageDir, bboltDBFilename)

	hasPrune := f.ValsetEpochs > 0 || f.ProofEpochs > 0 || f.SignatureEpochs > 0
	if hasPrune {
		// Open repo once, run prune, and (if --compact) reuse the same handle
		// for compaction so we don't pay a second RW open + freelist build.
		return runBboltSession(ctx, f)
	}

	// Compact-only: skip our RW open entirely. CompactDB opens its own handle.
	if f.Compact {
		before, beforeErr := fileSize(dbPath)
		pterm.Info.Println("Compaction rewrites the entire database file — this may take a while, please wait.")
		spinner, _ := pterm.DefaultSpinner.Start("Compacting bbolt database…")
		start := time.Now()
		if err := bbolt.CompactDB(dbPath); err != nil {
			spinner.Fail("Compaction failed")
			return errors.Errorf("bbolt compaction failed: %w", err)
		}
		after, afterErr := fileSize(dbPath)
		compactDuration := time.Since(start).Round(time.Millisecond)
		spinner.Success(pterm.Sprintf("Compaction completed in %s", compactDuration))
		printSizeReport(before, beforeErr, after, afterErr, compactDuration)
	}

	return nil
}

func runBboltSession(ctx context.Context, f Flags) error {
	openSpinner, _ := pterm.DefaultSpinner.Start("Opening bbolt database…")
	openStart := time.Now()
	repo, err := bbolt.New(bbolt.Config{
		Dir:              f.StorageDir,
		DBFilename:       bboltDBFilename,
		Metrics:          repoutil.DoNothingMetrics{},
		PrunePause:       0,
		MaxBatchDelay:    time.Millisecond,
		MaxBatchSize:     0,
		InitialMmapSize:  0,
		StatsLogInterval: 0,
		NoSync:           false,
		NoFreelistSync:   false,
		CompactOnStartup: false,
	})
	if err != nil {
		openSpinner.Fail("Failed to open bbolt database")
		return errors.Errorf("failed to open bbolt repository at %s (relay still running, or directory locked?): %w", f.StorageDir, err)
	}
	openSpinner.Success(pterm.Sprintf("Opened bbolt database in %s", time.Since(openStart).Round(time.Millisecond)))
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
		PruneBatchSize:           f.PruneBatchSize,
	}, f); err != nil {
		return err
	}

	dbPath := filepath.Join(f.StorageDir, bboltDBFilename)
	if f.Compact && ctx.Err() == nil {
		before, beforeErr := fileSize(dbPath)
		pterm.Info.Println("Compaction rewrites the entire database file — this may take a while, please wait.")
		spinner, _ := pterm.DefaultSpinner.Start("Compacting bbolt database…")
		start := time.Now()
		err := repo.CompactAndClose()
		closed = true // CompactAndClose always closes the handle, success or failure.
		if err != nil {
			spinner.Fail("Compaction failed")
			return errors.Errorf("bbolt compaction failed: %w", err)
		}
		after, afterErr := fileSize(dbPath)
		compactDuration := time.Since(start).Round(time.Millisecond)
		spinner.Success(pterm.Sprintf("Compaction completed in %s", compactDuration))
		printSizeReport(before, beforeErr, after, afterErr, compactDuration)
		return nil
	}

	closeSpinner, _ := pterm.DefaultSpinner.Start("Closing bbolt database…")
	closeStart := time.Now()
	closed = true
	if err := repo.Close(); err != nil {
		closeSpinner.Fail("Failed to close bbolt database")
		return errors.Errorf("failed to close bbolt repository: %w", err)
	}
	closeSpinner.Success(pterm.Sprintf("Closed bbolt database in %s", time.Since(closeStart).Round(time.Millisecond)))
	return nil
}

func runBadger(ctx context.Context, f Flags) error {
	openSpinner, _ := pterm.DefaultSpinner.Start("Opening badger database…")
	openStart := time.Now()
	repo, err := badger.New(badger.Config{
		Dir:                      f.StorageDir,
		Metrics:                  repoutil.DoNothingMetrics{},
		BlockCacheSize:           -1, // -1 = badger default; 0 means "disabled"
		CompactL0OnClose:         true,
		MutexCleanupInterval:     0,
		MutexCleanupStaleTimeout: 0,
		ValueLogGCInterval:       0,
		ValueLogGCDiscardRatio:   0,
		MemTableSize:             0,
		NumMemtables:             0,
		NumLevelZeroTables:       0,
		NumLevelZeroTablesStall:  0,
		NumCompactors:            0,
		ValueLogFileSize:         0,
	})
	if err != nil {
		openSpinner.Fail("Failed to open badger database")
		return errors.Errorf("failed to open badger repository at %s (relay still running, or directory locked?): %w", f.StorageDir, err)
	}
	openSpinner.Success(pterm.Sprintf("Opened badger database in %s", time.Since(openStart).Round(time.Millisecond)))
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
		PruneBatchSize:           f.PruneBatchSize,
	}, f); err != nil {
		return err
	}

	if f.Compact {
		before, beforeErr := dirSize(f.StorageDir)
		pterm.Info.Println("Compaction rewrites LSM levels and runs value-log GC — this may take a while, please wait.")
		spinner, _ := pterm.DefaultSpinner.Start("Flattening badger LSM + value log GC…")
		start := time.Now()
		capHit, err := repo.Flatten(ctx, f.BadgerFlattenWorkers)
		if err != nil {
			spinner.Fail("Flatten failed")
			return errors.Errorf("badger flatten failed: %w", err)
		}
		spinner.Success(pterm.Sprintf("Flatten completed in %s", time.Since(start).Round(time.Millisecond)))
		if capHit {
			pterm.Warning.Printf("value-log GC hit iteration cap (%d) — re-run with --compact to continue reclaiming space\n",
				badger.MaxValueLogGCIterations)
		}
		// Final L0 compaction happens in Close (CompactL0OnClose=true).
		// Mark closed before the error check so the deferred Close can't run
		// on a partially-closed DB.
		closeSpinner, _ := pterm.DefaultSpinner.Start("Closing badger database (final L0 compaction)…")
		closeStart := time.Now()
		closed = true
		if err := repo.Close(); err != nil {
			closeSpinner.Fail("Failed to close badger database")
			return errors.Errorf("failed to close badger repository: %w", err)
		}
		closeSpinner.Success(pterm.Sprintf("Closed badger database in %s", time.Since(closeStart).Round(time.Millisecond)))
		after, afterErr := dirSize(f.StorageDir)
		compactDuration := time.Since(start).Round(time.Millisecond)
		pterm.Success.Printf("Compaction completed in %s\n", compactDuration)
		printSizeReport(before, beforeErr, after, afterErr, compactDuration)
	}

	return nil
}

func runPruneOnce(ctx context.Context, cfg pruner.Config, f Flags) error {
	if f.ValsetEpochs == 0 && f.ProofEpochs == 0 && f.SignatureEpochs == 0 {
		pterm.Info.Println("Skipping pruning (no --retention.* flags set)")
		return nil
	}
	pterm.Info.Printf("Pruning: keeping last %d valset / %d proof / %d signature epochs; older epochs will be deleted…\n",
		f.ValsetEpochs, f.ProofEpochs, f.SignatureEpochs)

	progress := newProgressReporter()
	cfg.ProgressFn = progress.Report

	svc, err := pruner.New(cfg)
	if err != nil {
		return errors.Errorf("failed to construct pruner: %w", err)
	}
	start := time.Now()
	if err := svc.RunOnce(ctx); err != nil {
		progress.Stop()
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			pterm.Warning.Printf("Pruning interrupted after %s\n", time.Since(start).Round(time.Millisecond))
			return err
		}
		return errors.Errorf("pruning failed: %w", err)
	}
	progress.Stop()
	pterm.Success.Printf("Pruning completed in %s\n", time.Since(start).Round(time.Millisecond))
	return nil
}

// progressReporter drives a sequence of pterm progress bars: one bar per
// entity-type, each starting at 0 and advancing as epochs are pruned. The
// bars stack vertically since pruner pass ordering is sequential.
type progressReporter struct {
	bar           *pterm.ProgressbarPrinter
	currentEntity string
	currentValue  uint64
}

func newProgressReporter() *progressReporter {
	return &progressReporter{}
}

func (p *progressReporter) Report(entityType string, current, total uint64) {
	if entityType != p.currentEntity {
		p.Stop()
		bar, err := pterm.DefaultProgressbar.
			WithTotal(int(total)).
			WithTitle(pterm.Sprintf("Deleting %d %s epochs", total, entityType)).
			WithShowElapsedTime(true).
			Start()
		if err != nil {
			return
		}
		p.bar = bar
		p.currentEntity = entityType
		p.currentValue = 0
	}
	if p.bar != nil && current > p.currentValue {
		p.bar.Add(int(current - p.currentValue))
		p.currentValue = current
	}
}

func (p *progressReporter) Stop() {
	if p.bar != nil {
		_, _ = p.bar.Stop()
		p.bar = nil
	}
}

func confirmBackup(f Flags, storageType string) error {
	absPath, err := filepath.Abs(f.StorageDir)
	if err != nil {
		absPath = f.StorageDir
	}

	pterm.Warning.Println("This command rewrites the relay storage in place.")
	switch storageType {
	case storageTypeBbolt:
		pterm.Warning.Println("bbolt compaction writes to a tmp file and atomically renames, so a crash leaves the original DB intact — but a backup is still recommended.")
	case storageTypeBadger:
		pterm.Warning.Println("badger compaction relies on its WAL/manifest for crash recovery; SIGKILL or disk failure mid-flatten can leave the store in a state requiring badger's recovery path.")
	}
	pterm.Warning.Printf("Storage path: %s\n", absPath)

	if f.AssumeYes {
		pterm.Info.Println("Skipping backup confirmation (--yes).")
		return nil
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return errors.New("backup confirmation requires an interactive terminal; rerun with --yes once you have backed up the storage directory")
	}

	confirmed, err := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(false).
		WithDefaultText(pterm.Sprintf("Have you taken a backup of %s?", absPath)).
		Show()
	if err != nil {
		return errors.Errorf("failed to read backup confirmation: %w", err)
	}
	if !confirmed {
		return errors.New("aborted: take a backup of the storage directory and re-run, or pass --yes to skip this prompt")
	}
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
