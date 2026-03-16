package bbolt

import (
	"context"
	"encoding/binary"
	"log/slog"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	bolt "go.etcd.io/bbolt"

	"github.com/symbioticfi/relay/internal/client/repository/cached"
	"github.com/symbioticfi/relay/internal/client/repository/repoutil"
	"github.com/symbioticfi/relay/pkg/tracing"
)

var _ cached.Repository = (*Repository)(nil)

type Config struct {
	Dir                      string           `validate:"required"`
	Metrics                  repoutil.Metrics `validate:"required"`
	InitialMmapSize          int
	MutexCleanupInterval     time.Duration
	MutexCleanupStaleTimeout time.Duration
}

var (
	bucketSignatures          = []byte("signatures")
	bucketSignatureMaps       = []byte("signature_maps")
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
	bucketSignatures, bucketSignatureMaps, bucketSignatureRequests, bucketSignaturePending,
	bucketRequestIDIndex, bucketRequestIDEpochs, bucketAggregationProofs, bucketAggProofPending,
	bucketAggProofCommits, bucketValidatorSetHeaders, bucketValidatorSetStatus, bucketValidatorSetMeta,
	bucketValidators, bucketValidatorKeyLookups, bucketActiveValCounts, bucketNetworkConfigs,
	bucketMeta,
}

type mutexWithUseTime struct {
	mutex        sync.Mutex
	lastAccessNs atomic.Int64
}

func (m *mutexWithUseTime) lock() {
	m.mutex.Lock()
	m.lastAccessNs.Store(time.Now().UnixNano())
}

func (m *mutexWithUseTime) unlock() {
	m.mutex.Unlock()
}

func (m *mutexWithUseTime) lastAccess() time.Time {
	return time.Unix(0, m.lastAccessNs.Load())
}

func (m *mutexWithUseTime) tryLock() bool {
	return m.mutex.TryLock()
}

type Repository struct {
	db      *bolt.DB
	metrics repoutil.Metrics

	signatureMutexMap sync.Map // map[common.Hash]*mutexWithUseTime

	cleanupStop chan struct{}
}

func New(cfg Config) (*Repository, error) {
	if err := validator.New().Struct(cfg); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	opts := &bolt.Options{
		Timeout:         1 * time.Second,
		InitialMmapSize: cfg.InitialMmapSize,
	}

	dbPath := filepath.Join(cfg.Dir, "relay.db")
	db, err := bolt.Open(dbPath, 0600, opts)
	if err != nil {
		return nil, errors.Errorf("failed to open bbolt database: %w", err)
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
		db:                db,
		metrics:           cfg.Metrics,
		signatureMutexMap: sync.Map{},
		cleanupStop:       make(chan struct{}),
	}

	repo.startMutexCleanup(cfg.MutexCleanupInterval, cfg.MutexCleanupStaleTimeout)

	return repo, nil
}

func (r *Repository) Close() error {
	r.stopMutexCleanup()
	return r.db.Close()
}

func (r *Repository) startMutexCleanup(interval, staleTimeout time.Duration) {
	if interval == 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		slog.Info("Starting mutex cleanup goroutine",
			"component", "bbolt",
			"interval", interval,
			"staleTimeout", staleTimeout,
		)

		for {
			select {
			case <-ticker.C:
				r.cleanupStaleMutexes(staleTimeout)
			case <-r.cleanupStop:
				return
			}
		}
	}()
}

func (r *Repository) stopMutexCleanup() {
	close(r.cleanupStop)
}

func (r *Repository) cleanupStaleMutexes(staleTimeout time.Duration) {
	if staleTimeout == 0 {
		staleTimeout = time.Hour
	}

	staleThreshold := time.Now().Add(-staleTimeout)
	count := cleanupMutexMap(&r.signatureMutexMap, staleThreshold)

	if count > 0 {
		slog.Info("Cleaned up stale mutexes",
			"component", "bbolt",
			"signatureMutexes", count,
		)
	}
}

func cleanupMutexMap(mutexMap *sync.Map, staleThreshold time.Time) int {
	var count int

	mutexMap.Range(func(key, value any) bool {
		mutex := value.(*mutexWithUseTime)

		if !mutex.lastAccess().Before(staleThreshold) {
			return true
		}

		if !mutex.tryLock() {
			return true
		}
		defer mutex.unlock()

		// Double-check after acquiring lock
		if !mutex.lastAccess().Before(staleThreshold) {
			return true
		}

		mutexMap.Delete(key)
		count++
		return true
	})

	return count
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

func (r *Repository) doUpdate(ctx context.Context, name string, fn func(tx *bolt.Tx) error) error {
	ctx, span := tracing.StartSpan(ctx, "bbolt.update:"+name)
	defer span.End()

	start := time.Now()
	var err error
	nested := false

	if tx := getTx(ctx); tx != nil && tx.Writable() {
		nested = true
		err = fn(tx)
	} else {
		err = r.db.Update(fn)
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

func (r *Repository) doUpdateWithLock(ctx context.Context, name string, fn func(ctx context.Context) error, lockMap *sync.Map, key any) error {
	ctx, span := tracing.StartSpan(ctx, "bbolt.updateWithLock:"+name)
	defer span.End()

	mutexInterface, ok := lockMap.Load(key)
	if !ok {
		newMutex := &mutexWithUseTime{}
		newMutex.lastAccessNs.Store(time.Now().UnixNano())
		mutexInterface, _ = lockMap.LoadOrStore(key, newMutex)
	}
	activeMutex := mutexInterface.(*mutexWithUseTime)

	activeMutex.lock()
	defer activeMutex.unlock()

	start := time.Now()

	err := r.db.Update(func(tx *bolt.Tx) error {
		return fn(withTx(ctx, tx))
	})

	status := "ok"
	if err != nil {
		status = statusError
		tracing.RecordError(span, err)
	}

	d := time.Since(start)
	r.metrics.ObserveRepoQueryDuration(name, status, d)
	r.metrics.ObserveRepoQueryTotalDuration(name, status, d)
	return err
}
