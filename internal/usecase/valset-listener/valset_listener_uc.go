package valset_listener

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"middleware-offchain/core/entity"
	"middleware-offchain/pkg/log"
)

type eth interface {
	GetCurrentEpoch(ctx context.Context) (uint64, error)
	GetLastCommittedHeaderEpoch(ctx context.Context, addr entity.CrossChainAddress) (uint64, error)
	GetConfig(ctx context.Context, timestamp uint64) (entity.NetworkConfig, error)
	GetEpochStart(ctx context.Context, epoch uint64) (uint64, error)
}

type repo interface {
	GetLatestValidatorSet(ctx context.Context) (entity.ValidatorSet, error)
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

	latestProcessedEpoch uint64
}

func New(cfg Config) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Service{
		cfg:                  cfg,
		latestProcessedEpoch: 0,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	ctx = log.WithComponent(ctx, "listener")

	slog.InfoContext(ctx, "Starting valset listener service", "pollingInterval", s.cfg.PollingInterval)

	timer := time.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			if err := s.tryLoadMissingEpochs(ctx); err != nil {
				slog.ErrorContext(ctx, "Failed to process epochs", "error", err)
			}
			timer.Reset(s.cfg.PollingInterval)
		}
	}
}

func (s *Service) tryLoadMissingEpochs(ctx context.Context) error {
	slog.DebugContext(ctx, "Checking for missing epochs")

	currentEpoch, err := s.cfg.Eth.GetCurrentEpoch(ctx)
	if err != nil {
		return errors.Errorf("failed to get current epoch: %w", err)
	}
	currentEpochStart, err := s.cfg.Eth.GetEpochStart(ctx, currentEpoch)
	if err != nil {
		return errors.Errorf("failed to get current epoch start: %w", err)
	}
	config, err := s.cfg.Eth.GetConfig(ctx, currentEpochStart)
	if err != nil {
		return errors.Errorf("failed to get network config for current epoch: %w", err)
	}

	latestCommitedOnchainEpoch, err := s.getLastCommittedHeaderEpoch(ctx, config)
	if err != nil {
		return errors.Errorf("failed to get current epoch: %w", err)
	}

	latest, err := s.cfg.Repo.GetLatestValidatorSet(ctx)
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		return errors.Errorf("failed to get latest validator set extra: %w", err)
	}

	nextEpoch := uint64(0)
	if err == nil {
		nextEpoch = latest.Epoch + 1
	}

	if s.latestProcessedEpoch != 0 && nextEpoch <= s.latestProcessedEpoch {
		return nil
	}

	for latestCommitedOnchainEpoch >= nextEpoch {
		epochStart, err := s.cfg.Eth.GetEpochStart(ctx, nextEpoch)
		if err != nil {
			return errors.Errorf("failed to get epoch start for epoch %d: %w", nextEpoch, err)
		}

		nextEpochConfig, err := s.cfg.Eth.GetConfig(ctx, epochStart)
		if err != nil {
			return errors.Errorf("failed to get network config for epoch %d: %w", nextEpoch, err)
		}

		nextValset, err := s.cfg.Deriver.GetValidatorSet(ctx, nextEpoch, nextEpochConfig)
		if err != nil {
			return errors.Errorf("failed to derive validator set extra for epoch %d: %w", nextEpoch, err)
		}

		// TODO ilya: check valset integrity: valset.headerHash() == master.valsetHeaderHash(epoch)
		if err := s.cfg.Repo.SaveConfig(ctx, nextEpochConfig, nextEpoch); err != nil {
			return errors.Errorf("failed to save validator set extra for epoch %d: %w", nextEpoch, err)
		}

		if err := s.cfg.Repo.SaveValidatorSet(ctx, nextValset); err != nil {
			return errors.Errorf("failed to save validator set extra for epoch %d: %w", nextEpoch, err)
		}

		slog.DebugContext(ctx, "Synced validator set", "epoch", nextEpoch, "config", config, "valset", nextValset)

		s.latestProcessedEpoch = nextEpoch
		nextEpoch = nextValset.Epoch + 1
	}

	return nil
}

func (s *Service) getLastCommittedHeaderEpoch(ctx context.Context, config entity.NetworkConfig) (uint64, error) {
	maxEpoch := uint64(0)

	for _, addr := range config.Replicas {
		epoch, err := s.cfg.Eth.GetLastCommittedHeaderEpoch(ctx, addr)
		if err != nil {
			return 0, errors.Errorf("failed to get last committed header epoch for address %s: %w", addr.Address.Hex(), err)
		}

		if epoch >= maxEpoch {
			maxEpoch = epoch
		}
	}

	return maxEpoch, nil
}
