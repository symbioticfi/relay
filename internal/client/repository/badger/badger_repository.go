package badger

import (
	"github.com/go-playground/validator/v10"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-errors/errors"
)

type Config struct {
	Dir string `validate:"required"`
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("badger repository config validation failed: %w", err)
	}
	return nil
}

type Repository struct {
	db *badger.DB
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

	db.Subscribe()
	return &Repository{
		db: db,
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
