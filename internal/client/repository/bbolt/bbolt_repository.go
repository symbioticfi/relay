package bbolt

import (
	"context"
	"encoding/binary"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/sync/singleflight"

	"github.com/symbioticfi/relay/internal/client/repository/cached"
	"github.com/symbioticfi/relay/internal/client/repository/repoutil"
	"github.com/symbioticfi/relay/pkg/tracing"
)

var _ cached.Repository = (*Repository)(nil)

type Config struct {
	Dir              string           `validate:"required"`
	Metrics          repoutil.Metrics `validate:"required"`
	DBFilename       string
	InitialMmapSize  int
	StatsLogInterval time.Duration
	PrunePause       time.Duration
	CompactOnStartup bool
	NoFreelistSync   bool
	NoSync           bool
	MaxBatchDelay    time.Duration
	MaxBatchSize     int
}

var (
	bucketSignatures          = []byte("signatures")
	bucketSignatureRequests   = []byte("signature_requests")
	bucketSignaturePending    = []byte("signature_pending")
	bucketRequestIDIndex      = []byte("request_id_index")
	bucketRequestIDEpochs     = []byte("request_id_epochs")
	bucketAggregationProofs   = []byte("aggregation_proofs")
	bucketAggProofPending     = []byte("agg_proof_pending")
	bucketAggProofCommits     = []byte("agg_proof_commits")
	bucketValidatorSetHeaders = []byte("validator_set_headers")
	bucketValidatorSetStatus  = []byte("validator_set_status")
	bucketValidatorSetMeta    = []byte("validator_set_metadata")
	bucketValidators          = []byte("validators")
	bucketValidatorKeyLookups = []byte("validator_key_lookups")
	bucketActiveValCounts     = []byte("active_validator_counts")
	bucketNetworkConfigs      = []byte("network_configs")
	bucketMeta                = []byte("meta")
)

var allBuckets = [][]byte{
	bucketSignatures, bucketSignatureRequests, bucketSignaturePending,
	bucketRequestIDIndex, bucketRequestIDEpochs, bucketAggregationProofs, bucketAggProofPending,
	bucketAggProofCommits, bucketValidatorSetHeaders, bucketValidatorSetStatus, bucketValidatorSetMeta,
	bucketValidators, bucketValidatorKeyLookups, bucketActiveValCounts, bucketNetworkConfigs,
	bucketMeta,
}

type Repository struct {
	db      *bolt.DB
	metrics repoutil.Metrics

	signatureMapCache  sync.Map // map[common.Hash]entity.SignatureMap
	signatureMapLoader singleflight.Group

	prunePause time.Duration

	done chan struct{}
	bgWg sync.WaitGroup
}

func compactDB(dbPath string) error {
	info, err := os.Stat(dbPath)
	if err != nil {
		return nil // DB doesn't exist yet, nothing to compact
	}
	origSize := info.Size()

	start := time.Now()

	// Open source in writable mode to acquire an exclusive flock for the whole
	// compaction window. This prevents any other process from opening the DB
	// between Compact and Rename, which would otherwise observe an inode swap.
	src, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 5 * time.Second, FreelistType: bolt.FreelistArrayType})
	if err != nil {
		return errors.Errorf("failed to acquire exclusive lock on source db: %w", err)
	}
	defer src.Close()

	tmpPath := dbPath + ".compact"
	// Drop any stale tmp file from a previously crashed compaction so we don't
	// merge into pre-existing data.
	if err := os.Remove(tmpPath); err != nil && !os.IsNotExist(err) {
		return errors.Errorf("failed to remove stale compact tmp file: %w", err)
	}
	dst, err := bolt.Open(tmpPath, 0600, &bolt.Options{Timeout: 5 * time.Second, FreelistType: bolt.FreelistArrayType})
	if err != nil {
		return errors.Errorf("failed to create compact db: %w", err)
	}

	const compactTxMaxSize = 64 << 20 // 64 MB per transaction to limit memory usage
	if err := bolt.Compact(dst, src, compactTxMaxSize); err != nil {
		dst.Close()
		os.Remove(tmpPath)
		return errors.Errorf("compaction failed: %w", err)
	}
	if err := dst.Close(); err != nil {
		os.Remove(tmpPath)
		return errors.Errorf("failed to close compacted db: %w", err)
	}

	// Rename happens while src still holds the exclusive flock on dbPath, so
	// no concurrent opener can race in and observe the swapped inode. The
	// deferred src.Close() releases the lock on the now-unlinked old inode.
	if err := os.Rename(tmpPath, dbPath); err != nil {
		os.Remove(tmpPath)
		return errors.Errorf("failed to replace db with compacted version: %w", err)
	}

	compactInfo, err := os.Stat(dbPath)
	if err != nil {
		return errors.Errorf("failed to stat compacted db: %w", err)
	}
	compactSize := compactInfo.Size()

	logAttrs := []any{
		"component", "bbolt",
		"duration", time.Since(start),
		"before", origSize,
		"after", compactSize,
	}
	if compactSize > 0 {
		logAttrs = append(logAttrs, "ratio", float64(origSize)/float64(compactSize))
	}
	slog.Info("DB compacted", logAttrs...)
	return nil
}

func New(cfg Config) (*Repository, error) {
	if err := validator.New().Struct(cfg); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	filename := cfg.DBFilename
	if filename == "" {
		filename = "relay.db"
	}
	dbPath := filepath.Join(cfg.Dir, filename)

	if cfg.CompactOnStartup {
		if err := compactDB(dbPath); err != nil {
			return nil, errors.Errorf("startup db compaction failed: %w", err)
		}
	}

	opts := &bolt.Options{
		Timeout:         1 * time.Second,
		InitialMmapSize: cfg.InitialMmapSize,
		FreelistType:    bolt.FreelistArrayType,
	}

	db, err := bolt.Open(dbPath, 0600, opts)
	if err != nil {
		return nil, errors.Errorf("failed to open bbolt database: %w", err)
	}

	db.NoFreelistSync = cfg.NoFreelistSync
	db.NoSync = cfg.NoSync
	if cfg.MaxBatchDelay > 0 {
		db.MaxBatchDelay = cfg.MaxBatchDelay
	}
	if cfg.MaxBatchSize > 0 {
		db.MaxBatchSize = cfg.MaxBatchSize
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		for _, name := range allBuckets {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return errors.Errorf("failed to create bucket %s: %w", name, err)
			}
		}
		return nil
	}); err != nil {
		db.Close()
		return nil, errors.Errorf("failed to initialize buckets: %w", err)
	}

	repo := &Repository{
		db:                 db,
		metrics:            cfg.Metrics,
		done:               make(chan struct{}),
		signatureMapCache:  sync.Map{},
		signatureMapLoader: singleflight.Group{},
		prunePause:         cfg.PrunePause,
	}
	repo.startStatsLogger(cfg.StatsLogInterval)
	repo.startSizeReporter()

	return repo, nil
}

func (r *Repository) Close() error {
	close(r.done)
	r.bgWg.Wait()
	return r.db.Close()
}

func (r *Repository) Stats() bolt.Stats {
	return r.db.Stats()
}

func (r *Repository) startStatsLogger(interval time.Duration) {
	if interval == 0 {
		return
	}

	r.bgWg.Go(func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		var prevTxStats bolt.TxStats

		for {
			select {
			case <-ticker.C:
				r.logStats(&prevTxStats)
			case <-r.done:
				return
			}
		}
	})
}

func (r *Repository) startSizeReporter() {
	r.reportDBSize() // report once at startup

	r.bgWg.Go(func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.reportDBSize()
			case <-r.done:
				return
			}
		}
	})
}

func (r *Repository) reportDBSize() {
	info, err := os.Stat(r.db.Path())
	if err != nil {
		return
	}
	r.metrics.SetDBSizeBytes(float64(info.Size()))
}

func (r *Repository) logStats(prevTxStats *bolt.TxStats) {
	stats := r.db.Stats()
	txDiff := stats.TxStats.Sub(prevTxStats)
	*prevTxStats = stats.TxStats

	// Only log when there was write activity
	if txDiff.GetWriteTime() == 0 && txDiff.GetSpillTime() == 0 {
		return
	}

	slog.Info("BBolt stats",
		"component", "bbolt",
		// DB-level: open read txns can block page reclamation and slow writes
		"openReadTxns", stats.OpenTxN,
		"freePages", stats.FreePageN,
		"pendingPages", stats.PendingPageN,
		// Write breakdown since last log:
		// spillTime = B-tree restructuring (CPU-bound)
		// writeTime = disk write + fsync (IO-bound) — this is the fsync cost
		"spillTime", txDiff.GetSpillTime(),
		"writeTime", txDiff.GetWriteTime(),
		// How many write operations happened
		"writes", txDiff.GetWrite(),
		"spills", txDiff.GetSpill(),
		"splits", txDiff.GetSplit(),
		"rebalances", txDiff.GetRebalance(),
		"rebalanceTime", txDiff.GetRebalanceTime(),
	)
}

// Key encoding helpers — all use raw bytes for efficiency.

func epochBytes(epoch uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, epoch)
	return b
}

func uint32Bytes(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

// epochHashKey returns epoch(8) + hash(32) = 40 bytes
func epochHashKey(epoch uint64, hash []byte) []byte {
	key := make([]byte, 8+len(hash))
	binary.BigEndian.PutUint64(key, epoch)
	copy(key[8:], hash)
	return key
}

// epochOperatorKey returns epoch(8) + operator(20) = 28 bytes
func epochOperatorKey(epoch uint64, operator []byte) []byte {
	key := make([]byte, 8+len(operator))
	binary.BigEndian.PutUint64(key, epoch)
	copy(key[8:], operator)
	return key
}

// hashIndexKey returns requestID(32) + validatorIndex(4) = 36 bytes
func signatureKey(requestID []byte, validatorIndex uint32) []byte {
	key := make([]byte, 32+4)
	copy(key, requestID)
	binary.BigEndian.PutUint32(key[32:], validatorIndex)
	return key
}

// validatorKeyLookupKey returns epoch(8) + keyTag(4) + pubKeyHash(32) = 44 bytes
func validatorKeyLookupKey(epoch uint64, keyTag uint32, pubKeyHash []byte) []byte {
	key := make([]byte, 8+4+32)
	binary.BigEndian.PutUint64(key, epoch)
	binary.BigEndian.PutUint32(key[8:], keyTag)
	copy(key[12:], pubKeyHash)
	return key
}

const statusError = "error"

type txKey struct{}

func withTx(ctx context.Context, tx *bolt.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func getTx(ctx context.Context) *bolt.Tx {
	if tx, ok := ctx.Value(txKey{}).(*bolt.Tx); ok {
		return tx
	}
	return nil
}

func (r *Repository) doView(ctx context.Context, name string, fn func(tx *bolt.Tx) error) error {
	ctx, span := tracing.StartSpan(ctx, "bbolt.view:"+name)
	defer span.End()

	start := time.Now()
	var err error
	nested := false

	if tx := getTx(ctx); tx != nil {
		nested = true
		err = fn(tx)
	} else {
		err = r.db.View(fn)
	}

	status := "ok"
	if err != nil {
		status = statusError
		tracing.RecordError(span, err)
	}

	d := time.Since(start)
	r.metrics.ObserveRepoQueryDuration(name, status, d)
	if !nested {
		r.metrics.ObserveRepoQueryTotalDuration(name, status, d)
	}
	return err
}

// doBatch wraps db.Batch, which coalesces concurrent writers into a single bbolt
// transaction. The callback fn MUST be idempotent: bbolt re-runs it if the batch
// transaction commits but a peer's callback panics, or on internal retry. Do not
// rely on side effects performed before the final, committed call.
func (r *Repository) doBatch(ctx context.Context, name string, fn func(tx *bolt.Tx) error) error {
	ctx, span := tracing.StartSpan(ctx, "bbolt.batch:"+name)
	defer span.End()

	start := time.Now()
	var err error
	nested := false

	if tx := getTx(ctx); tx != nil && tx.Writable() {
		nested = true
		err = fn(tx)
	} else {
		err = r.db.Batch(fn)
	}

	status := "ok"
	if err != nil {
		status = statusError
		tracing.RecordError(span, err)
	}

	d := time.Since(start)
	r.metrics.ObserveRepoQueryDuration(name, status, d)
	if !nested {
		r.metrics.ObserveRepoQueryTotalDuration(name, status, d)
	}
	return err
}
