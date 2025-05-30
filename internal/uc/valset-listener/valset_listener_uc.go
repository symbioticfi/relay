package valset_listener

import (
	"context"
	"log/slog"
	"math/big"
	"time"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/samber/mo"

	"middleware-offchain/internal/entity"
)

type eth interface {
	GetCurrentValsetEpoch(ctx context.Context) (*big.Int, error)
}

type repo interface {
	GetLatestValsetExtra(ctx context.Context) (mo.Option[entity.ValidatorSetExtra], error)
	SaveValsetExtra(ctx context.Context, extra entity.ValidatorSetExtra) error
}

type deriver interface {
	GetValidatorSetExtraForEpoch(ctx context.Context, epoch *big.Int) (entity.ValidatorSetExtra, error)
	MakeValidatorSetHeaderHash(ctx context.Context, extra entity.ValidatorSetExtra) ([]byte, error)
}

type Config struct {
	Eth             eth           `validate:"required"`
	Repo            repo          `validate:"required"`
	Deriver         deriver       `validate:"required"`
	PollingInterval time.Duration `validate:"required,gt=0"`
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("invalid config: %w", err)
	}

	return nil
}

type Service struct {
	cfg Config
}

func New(cfg Config) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Service{
		cfg: cfg,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	timer := time.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			if err := s.tryLoadMissingEpochs(ctx); err != nil {
				slog.ErrorContext(ctx, "failed to process epochs", "error", err)
			}
			timer.Reset(s.cfg.PollingInterval)
		}
	}
}

func (s *Service) tryLoadMissingEpochs(ctx context.Context) error {
	latestCommitedOnchainEpoch, err := s.cfg.Eth.GetCurrentValsetEpoch(ctx)
	if err != nil {
		return errors.Errorf("failed to get current epoch: %w", err)
	}

	latest, err := s.cfg.Repo.GetLatestValsetExtra(ctx)
	if err != nil {
		return errors.Errorf("failed to get latest validator set extra: %w", err)
	}

	latestEpoch := new(big.Int).SetInt64(1)
	if latest.IsPresent() {
		latestEpoch = latest.MustGet().Epoch
	}

	for new(big.Int).Sub(latestCommitedOnchainEpoch, latestEpoch).Cmp(big.NewInt(0)) > 0 {
		nextEpoch := new(big.Int).Add(latestEpoch, big.NewInt(1))

		nextValsetExtra, err := s.cfg.Deriver.GetValidatorSetExtraForEpoch(ctx, nextEpoch)
		if err != nil {
			return errors.Errorf("failed to derive validator set extra for epoch %s: %w", nextEpoch.String(), err)
		}

		// TODO ilya: check valset integrity: valset.headerHash() == master.valsetHeaderHash(epoch)

		if err := s.cfg.Repo.SaveValsetExtra(ctx, nextValsetExtra); err != nil {
			return errors.Errorf("failed to save validator set extra for epoch %s: %w", nextEpoch.String(), err)
		}

		latestEpoch = nextValsetExtra.Epoch
	}

	return nil
}
