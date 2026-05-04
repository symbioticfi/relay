package badger

import (
	"bytes"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/symbioticfi/relay/internal/client/repository/cached"
	"github.com/symbioticfi/relay/internal/client/repository/repoutil"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

var _ cached.Repository = (*Repository)(nil)

type Config struct {
	Dir                      string        `validate:"required"`
	Metrics                  metrics       `validate:"required"`
	MutexCleanupInterval     time.Duration // How often to run mutex cleanup (e.g., 1 hour). Zero disables cleanup.
	MutexCleanupStaleTimeout time.Duration // Remove mutexes not used for this duration, default 1 hour.
	ValueLogGCInterval       time.Duration // How often to run value log GC (e.g., 5m). Zero disables GC.
	ValueLogGCDiscardRatio   float64       // Discard ratio threshold for GC (0.0-1.0). Default 0.5.
	BlockCacheSize           int64         // -1 = use badger default, 0 = disabled, >0 = size in bytes
	MemTableSize             int64
	NumMemtables             int
	NumLevelZeroTables       int
	NumLevelZeroTablesStall  int
	CompactL0OnClose         bool
	NumCompactors            int
	ValueLogFileSize         int64
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("badger repository config validation failed: %w", err)
	}
	return nil
}

type metrics = repoutil.Metrics

type Repository struct {
	db      *badger.DB
	metrics metrics

	signatureMutexMap sync.Map // map[requestId]*mutexWithUseTime
	proofsMutexMap    sync.Map // map[requestId]*mutexWithUseTime
	valsetMutexMap    sync.Map // map[epoch]*mutexWithUseTime

	valueLogGCDiscardRatio float64

	done chan struct{}
}

func New(cfg Config) (*Repository, error) {
	if err := validator.New().Struct(cfg); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	opts := badger.DefaultOptions(cfg.Dir).
		WithLogger(&slogBadgerLogger{})
	applyBadgerTuning(&opts, cfg)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, errors.Errorf("failed to open badger database: %w", err)
	}

	discardRatio := cfg.ValueLogGCDiscardRatio
	if discardRatio == 0 {
		discardRatio = 0.5
	}
	repo := &Repository{
		db:                     db,
		metrics:                cfg.Metrics,
		valueLogGCDiscardRatio: discardRatio,
		done:                   make(chan struct{}),
	}

	repo.startMutexCleanup(cfg.MutexCleanupInterval, cfg.MutexCleanupStaleTimeout)
	repo.startValueLogGC(cfg.ValueLogGCInterval, cfg.ValueLogGCDiscardRatio)
	repo.startSizeReporter()

	return repo, nil
}

// applyBadgerTuning overrides badger.Options with config values.
// For most fields, zero means "use badger default". For BlockCacheSize,
// -1 means "use badger default", 0 means "disabled", >0 means explicit size.
func applyBadgerTuning(opts *badger.Options, cfg Config) {
	if cfg.BlockCacheSize >= 0 {
		opts.BlockCacheSize = cfg.BlockCacheSize
	}
	if cfg.MemTableSize != 0 {
		opts.MemTableSize = cfg.MemTableSize
	}
	if cfg.NumMemtables != 0 {
		opts.NumMemtables = cfg.NumMemtables
	}
	if cfg.NumLevelZeroTables != 0 {
		opts.NumLevelZeroTables = cfg.NumLevelZeroTables
	}
	if cfg.NumLevelZeroTablesStall != 0 {
		opts.NumLevelZeroTablesStall = cfg.NumLevelZeroTablesStall
	}
	if cfg.ValueLogFileSize != 0 {
		opts.ValueLogFileSize = cfg.ValueLogFileSize
	}
	if cfg.NumCompactors != 0 {
		opts.NumCompactors = cfg.NumCompactors
	}
	// CompactL0OnClose is a bool — always apply since the tuned default is true
	// and badger's default is false. When cfg comes from CLI flags, the default is true.
	// When cfg comes from tests (zero-value), this is a no-op (false == badger default).
	opts.CompactL0OnClose = cfg.CompactL0OnClose
}

func (r *Repository) Close() error {
	close(r.done)
	return r.db.Close()
}

// maxValueLogGCIterations bounds the value-log GC loop in Flatten. Each successful
// RunValueLogGC rewrites at most one vlog file, so this caps the total number of
// rewrites at a safe-but-large value to prevent pathological infinite loops if
// badger ever returns nil indefinitely.
const maxValueLogGCIterations = 100

// Flatten compacts all SST levels into the lowest possible LSM structure and then
// repeatedly runs value-log GC at the Config-supplied ValueLogGCDiscardRatio
// (default 0.5) until no more rewrites are needed (or the iteration cap is hit).
// Intended for offline maintenance (e.g. one-shot CLI) — must not run alongside
// active write traffic.
func (r *Repository) Flatten(workers int) error {
	if workers <= 0 {
		workers = 1
	}
	if err := r.db.Flatten(workers); err != nil {
		return errors.Errorf("badger flatten failed: %w", err)
	}
	for range maxValueLogGCIterations {
		if err := r.db.RunValueLogGC(r.valueLogGCDiscardRatio); err != nil {
			if errors.Is(err, badger.ErrNoRewrite) {
				return nil
			}
			return errors.Errorf("badger value log gc failed: %w", err)
		}
	}
	slog.Warn("badger value log GC reached iteration cap", "component", "badger", "cap", maxValueLogGCIterations)
	return nil
}

type slogBadgerLogger struct{}

func (l *slogBadgerLogger) Errorf(s string, args ...any) {
	slog.Error(fmt.Sprintf(s, args...), "component", "badger")
}
func (l *slogBadgerLogger) Warningf(s string, args ...any) {
	slog.Warn(fmt.Sprintf(s, args...), "component", "badger")
}
func (l *slogBadgerLogger) Infof(s string, args ...any) {
	slog.Info(fmt.Sprintf(s, args...), "component", "badger")
}
func (l *slogBadgerLogger) Debugf(string, ...any) {}

type DoNothingMetrics = repoutil.DoNothingMetrics

func (r *Repository) startMutexCleanup(interval, staleTimeout time.Duration) {
	if interval == 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.cleanupStaleMutexes(staleTimeout)
			case <-r.done:
				return
			}
		}
	}()
}

func (r *Repository) startValueLogGC(interval time.Duration, discardRatio float64) {
	if interval == 0 {
		return
	}
	if discardRatio == 0 {
		discardRatio = 0.5
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				for range 10 {
					if r.db.RunValueLogGC(discardRatio) == nil {
						break
					}
				}
			case <-r.done:
				return
			}
		}
	}()
}

func (r *Repository) startSizeReporter() {
	r.reportDBSize()

	go func() {
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
	}()
}

func (r *Repository) reportDBSize() {
	lsm, vlog := r.db.Size()
	r.metrics.SetDBSizeBytes(float64(lsm + vlog))
}

// cleanupStaleMutexes removes mutexes that haven't been used for longer than cleanupStaleAfter
func (r *Repository) cleanupStaleMutexes(staleTimeout time.Duration) {
	// Default stale timeout to 1 hour if not set
	if staleTimeout == 0 {
		staleTimeout = time.Hour
	}

	now := time.Now()
	staleThreshold := now.Add(-staleTimeout)

	signatureCount := cleanupMutexMap(&r.signatureMutexMap, staleThreshold)
	proofsCount := cleanupMutexMap(&r.proofsMutexMap, staleThreshold)
	valsetCount := cleanupMutexMap(&r.valsetMutexMap, staleThreshold)

	if signatureCount > 0 || proofsCount > 0 || valsetCount > 0 {
		slog.Info("Cleaned up stale mutexes",
			"signatureMutexes", signatureCount,
			"proofsMutexes", proofsCount,
			"valsetMutexes", valsetCount,
			"staleThreshold", staleThreshold,
		)
	}
}

// cleanupMutexMap removes stale mutexes from a single sync.Map using double-check pattern
func cleanupMutexMap(mutexMap *sync.Map, staleThreshold time.Time) int {
	var count int

	mutexMap.Range(func(key, value any) bool {
		mutex := value.(*mutexWithUseTime)

		// First check: if recently accessed, skip
		if !mutex.lastAccess().Before(staleThreshold) {
			return true
		}

		// Try to acquire the lock to ensure it's not in use
		if !mutex.tryLock() {
			return true
		}
		defer mutex.unlock()

		// Double-check last access time after acquiring lock
		// This handles the race where updateAccess() was called between the first check and TryLock
		if !mutex.lastAccess().Before(staleThreshold) {
			return true
		}

		// Safe to delete now
		mutexMap.Delete(key)
		count++

		return true
	})

	return count
}

var errCorruptedRequestIDEpochLink = errors.New("corrupted request id epoch link")

func keyRequestIDEpoch(epoch symbiotic.Epoch, requestID common.Hash) []byte {
	return append(keyRequestIDEpochPrefix(epoch), requestID.Bytes()...)
}

func keyRequestIDEpochPrefix(epoch symbiotic.Epoch) []byte {
	return append(keyRequestIDEpochAll(), epoch.Bytes()...)
}

func keyRequestIDEpochAll() []byte {
	return []byte("request_id_epoch")
}

const (
	epochLen   = 8
	hashLen    = 32
	hashHexLen = 66
	colonByte  = byte(':')
)

// extractRequestIDFromEpochKey extracts request ID from the epoch key link
// Key format: "request_id_epoch" (16 bytes) + epoch (8 bytes) + requestID (32 bytes)
func extractRequestIDFromEpochKey(key []byte) (common.Hash, error) {
	prefixLen := len(keyRequestIDEpochAll())

	if len(key) < prefixLen+epochLen+hashLen {
		return common.Hash{}, errors.New("invalid key length")
	}

	return common.BytesToHash(key[prefixLen+epochLen:]), nil
}

func epochKey(prefix string, epoch symbiotic.Epoch) []byte {
	epochBytes := epoch.Bytes()
	key := make([]byte, len(prefix)+len(epochBytes))
	copy(key, prefix)
	copy(key[len(prefix):], epochBytes)
	return key
}

func epochKeyWithColon(prefix string, epoch symbiotic.Epoch) []byte {
	key := epochKey(prefix, epoch)
	return append(key, colonByte)
}

func extractRequestIDFromEpochDelimitedKey(key []byte, prefix string) (common.Hash, error) {
	prefixBytes := []byte(prefix)
	prefixLen := len(prefixBytes)

	if len(key) < prefixLen+epochLen+1+hashHexLen {
		return common.Hash{}, errors.Errorf("invalid key length for prefix %s", prefix)
	}

	if !bytes.HasPrefix(key, prefixBytes) {
		return common.Hash{}, errors.Errorf("invalid key prefix: %s", prefix)
	}

	delimiterIndex := prefixLen + epochLen
	if key[delimiterIndex] != colonByte {
		return common.Hash{}, errors.Errorf("missing delimiter for prefix %s", prefix)
	}

	hashBytes := key[delimiterIndex+1:]
	if len(hashBytes) != hashHexLen {
		return common.Hash{}, errors.Errorf("unexpected hash length for prefix %s", prefix)
	}

	return common.HexToHash(string(hashBytes)), nil
}

// extractEpochFromKey extracts epoch from a key with format: prefix + epoch
func extractEpochFromKey(key []byte, prefix string) (symbiotic.Epoch, error) {
	prefixBytes := []byte(prefix)
	prefixLen := len(prefixBytes)

	if len(key) != prefixLen+epochLen {
		return 0, errors.Errorf("invalid key length for prefix %s: expected %d, got %d", prefix, prefixLen+epochLen, len(key))
	}

	if !bytes.HasPrefix(key, prefixBytes) {
		return 0, errors.Errorf("invalid key prefix: expected %s", prefix)
	}

	epochBytes := key[prefixLen:]
	epoch, err := symbiotic.EpochFromBytes(epochBytes)
	if err != nil {
		return 0, errors.Errorf("failed to decode epoch from key: %w", err)
	}
	return epoch, nil
}

// extractEpochFromValue extracts epoch from a stored value (8-byte big-endian uint64)
func extractEpochFromValue(value []byte) (symbiotic.Epoch, error) {
	if len(value) != epochLen {
		return 0, errors.Errorf("invalid value length for epoch: expected %d, got %d", epochLen, len(value))
	}

	epoch, err := symbiotic.EpochFromBytes(value)
	if err != nil {
		return 0, errors.Errorf("failed to decode epoch from value: %w", err)
	}
	return epoch, nil
}
