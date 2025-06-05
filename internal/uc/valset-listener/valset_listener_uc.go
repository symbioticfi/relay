package valset_listener

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"middleware-offchain/internal/entity"
)

type eth interface {
	GetLastCommittedHeaderEpoch(ctx context.Context) (uint64, error)
	GetConfig(ctx context.Context, timestamp uint64) (entity.NetworkConfig, error)
	GetEpochStart(ctx context.Context, epoch uint64) (uint64, error)
}

type repo interface {
	GetLatestValset(ctx context.Context) (entity.ValidatorSet, error)
	SaveConfig(ctx context.Context, config entity.NetworkConfig, epoch uint64) error
	SaveValidatorSet(ctx context.Context, valset entity.ValidatorSet) error
}

type deriver interface {
	GetValidatorSet(ctx context.Context, epoch uint64, config entity.NetworkConfig) (entity.ValidatorSet, error)
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
	latestCommitedOnchainEpoch, err := s.cfg.Eth.GetLastCommittedHeaderEpoch(ctx)
	if err != nil {
		return errors.Errorf("failed to get current epoch: %w", err)
	}

	latest, err := s.cfg.Repo.GetLatestValset(ctx)
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		return errors.Errorf("failed to get latest validator set extra: %w", err)
	}

	latestEpoch := uint64(1)
	if err == nil {
		latestEpoch = latest.Epoch
	}

	for latestCommitedOnchainEpoch > latestEpoch {
		nextEpoch := latestEpoch + 1

		epochStart, err := s.cfg.Eth.GetEpochStart(ctx, nextEpoch)
		if err != nil {
			return errors.Errorf("failed to get epoch start for epoch %d: %w", nextEpoch, err)
		}

		config, err := s.cfg.Eth.GetConfig(ctx, epochStart)
		if err != nil {
			return errors.Errorf("failed to get network config for epoch %d: %w", nextEpoch, err)
		}

		nextValset, err := s.cfg.Deriver.GetValidatorSet(ctx, nextEpoch, config)
		if err != nil {
			return errors.Errorf("failed to derive validator set extra for epoch %d: %w", nextEpoch, err)
		}

		// TODO ilya: check valset integrity: valset.headerHash() == master.valsetHeaderHash(epoch)

		if err := s.cfg.Repo.SaveConfig(ctx, config, nextEpoch); err != nil {
			return errors.Errorf("failed to save validator set extra for epoch %d: %w", nextEpoch, err)
		}

		if err := s.cfg.Repo.SaveValidatorSet(ctx, nextValset); err != nil {
			return errors.Errorf("failed to save validator set extra for epoch %d: %w", nextEpoch, err)
		}

		latestEpoch = nextValset.Epoch
	}

	return nil
}
