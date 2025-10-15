package badger

import (
	"log/slog"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"google.golang.org/protobuf/proto"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-errors/errors"
)

type Config struct {
	Dir                      string        `validate:"required"`
	Metrics                  metrics       `validate:"required"`
	MutexCleanupInterval     time.Duration // How often to run mutex cleanup (e.g., 1 hour). Zero disables cleanup.
	MutexCleanupStaleTimeout time.Duration // Remove mutexes not used for this duration, default 1 hour.
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

	cleanupTicker     *time.Ticker
	cleanupStop       chan struct{}
	cleanupStaleAfter time.Duration
}

func New(cfg Config) (*Repository, error) {
	if err := validator.New().Struct(cfg); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	opts := badger.DefaultOptions(cfg.Dir)
	opts.Logger = doNothingLog{}

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

func (r *Repository) Close() error {
	// Stop the mutex cleanup goroutine before closing the database
	r.stopMutexCleanup()
	return r.db.Close()
}

type doNothingLog struct{}

func (l doNothingLog) Errorf(s string, args ...interface{})   {}
func (l doNothingLog) Warningf(s string, args ...interface{}) {}
func (l doNothingLog) Infof(s string, args ...interface{})    {}
func (l doNothingLog) Debugf(s string, args ...interface{})   {}

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
	// Default stale timeout to 1 hour if not set
	if staleTimeout == 0 {
		staleTimeout = time.Hour
	}
	r.cleanupStaleAfter = staleTimeout

	// If interval is 0, cleanup is disabled
	if interval == 0 {
		slog.Info("Mutex cleanup disabled (interval is 0)")
		return
	}

	r.cleanupTicker = time.NewTicker(interval)
	r.cleanupStop = make(chan struct{})

	go func() {
		slog.Info("Starting mutex cleanup goroutine",
			"interval", interval,
			"staleTimeout", staleTimeout,
		)

		for {
			select {
			case <-r.cleanupTicker.C:
				r.cleanupStaleMutexes()
			case <-r.cleanupStop:
				slog.Info("Stopping mutex cleanup goroutine")
				return
			}
		}
	}()
}

// stopMutexCleanup stops the background cleanup goroutine
func (r *Repository) stopMutexCleanup() {
	if r.cleanupTicker != nil {
		r.cleanupTicker.Stop()
	}
	if r.cleanupStop != nil {
		close(r.cleanupStop)
	}
}

// cleanupStaleMutexes removes mutexes that haven't been used for longer than cleanupStaleAfter
func (r *Repository) cleanupStaleMutexes() {
	now := time.Now()
	staleThreshold := now.Add(-r.cleanupStaleAfter)

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
		defer mutex.mutex.Unlock()

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
