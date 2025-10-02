package badger

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-playground/validator/v10"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-errors/errors"
)

type Config struct {
	Dir     string  `validate:"required"`
	Metrics metrics `validate:"required"`
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

	signatureMutexMu  sync.Mutex
	signatureMutexMap map[common.Hash]*sync.Mutex
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

	return &Repository{
		db:                db,
		metrics:           cfg.Metrics,
		signatureMutexMu:  sync.Mutex{},
		signatureMutexMap: make(map[common.Hash]*sync.Mutex),
	}, nil
}

func (r *Repository) Close() error {
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
