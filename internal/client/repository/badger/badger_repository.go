package badger

import (
	"bytes"
	"log/slog"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"google.golang.org/protobuf/proto"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type Config struct {
	Dir                      string        `validate:"required"`
	Metrics                  metrics       `validate:"required"`
	MutexCleanupInterval     time.Duration // How often to run mutex cleanup (e.g., 1 hour). Zero disables cleanup.
	MutexCleanupStaleTimeout time.Duration // Remove mutexes not used for this duration, default 1 hour.
	BlockCacheSize           int64
	MemTableSize             int64
	NumMemtables             int
	NumLevelZeroTables       int
	NumLevelZeroTablesStall  int
	CompactL0OnClose         bool
	ValueLogFileSize         int64
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("badger repository config validation failed: %w", err)
	}
	return nil
}

type metrics interface {
	ObserveRepoQueryDuration(queryName string, status string, d time.Duration)
	ObserveRepoQueryTotalDuration(queryName string, status string, d time.Duration)
}

type Repository struct {
	db      *badger.DB
	metrics metrics

	signatureMutexMap sync.Map // map[requestId]*mutexWithUseTime
	proofsMutexMap    sync.Map // map[requestId]*mutexWithUseTime
	valsetMutexMap    sync.Map // map[epoch]*mutexWithUseTime

	cleanupStop chan struct{}
}

func New(cfg Config) (*Repository, error) {
	if err := validator.New().Struct(cfg); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	opts := badger.DefaultOptions(cfg.Dir)
	opts.Logger = doNothingLog{}
	applyBadgerTuning(&opts, cfg)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, errors.Errorf("failed to open badger database: %w", err)
	}

	repo := &Repository{
		db:      db,
		metrics: cfg.Metrics,
	}

	// Start mutex cleanup goroutine if configured
	repo.startMutexCleanup(cfg.MutexCleanupInterval, cfg.MutexCleanupStaleTimeout)

	return repo, nil
}

// applyBadgerTuning overrides badger.Options with non-zero config values.
// Zero values are left as badger defaults, allowing tests to omit tuning fields.
func applyBadgerTuning(opts *badger.Options, cfg Config) {
	if cfg.BlockCacheSize != 0 {
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
	// CompactL0OnClose is a bool — always apply since the tuned default is true
	// and badger's default is false. When cfg comes from CLI flags, the default is true.
	// When cfg comes from tests (zero-value), this is a no-op (false == badger default).
	opts.CompactL0OnClose = cfg.CompactL0OnClose
}

func (r *Repository) Close() error {
	// Stop the mutex cleanup goroutine before closing the database
	r.stopMutexCleanup()
	return r.db.Close()
}

type doNothingLog struct{}

func (l doNothingLog) Errorf(s string, args ...any)   {}
func (l doNothingLog) Warningf(s string, args ...any) {}
func (l doNothingLog) Infof(s string, args ...any)    {}
func (l doNothingLog) Debugf(s string, args ...any)   {}

type DoNothingMetrics struct {
}

func (m DoNothingMetrics) ObserveRepoQueryDuration(queryName string, status string, d time.Duration) {
}
func (m DoNothingMetrics) ObserveRepoQueryTotalDuration(queryName string, status string, d time.Duration) {
}

func marshalProto(msg proto.Message) ([]byte, error) {
	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, errors.Errorf("failed to marshal proto: %v", err)
	}
	return data, nil
}

func unmarshalProto(data []byte, msg proto.Message) error {
	if err := proto.Unmarshal(data, msg); err != nil {
		return errors.Errorf("failed to unmarshal proto: %v", err)
	}
	return nil
}

// startMutexCleanup starts a background goroutine that periodically cleans up stale mutexes
func (r *Repository) startMutexCleanup(interval, staleTimeout time.Duration) {
	// If interval is 0, cleanup is disabled
	if interval == 0 {
		slog.Info("Mutex cleanup disabled (interval is 0)")
		return
	}

	r.cleanupStop = make(chan struct{})

	go func() {
		cleanupTicker := time.NewTicker(interval)
		defer func() {
			cleanupTicker.Stop()
		}()

		slog.Info("Starting mutex cleanup goroutine",
			"interval", interval,
			"staleTimeout", staleTimeout,
		)

		for {
			select {
			case <-cleanupTicker.C:
				r.cleanupStaleMutexes(staleTimeout)
			case <-r.cleanupStop:
				slog.Info("Stopping mutex cleanup goroutine")
				return
			}
		}
	}()
}

// stopMutexCleanup stops the background cleanup goroutine
func (r *Repository) stopMutexCleanup() {
	if r.cleanupStop != nil {
		close(r.cleanupStop)
	}
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
