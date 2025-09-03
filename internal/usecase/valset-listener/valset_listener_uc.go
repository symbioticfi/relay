package valset_listener

import (
	"context"
	"log/slog"
	"time"

	"github.com/symbioticfi/relay/core/client/evm"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/pkg/log"
)

type repo interface {
	GetLatestValidatorSetHeader(_ context.Context) (entity.ValidatorSetHeader, error)
	SaveConfig(ctx context.Context, config entity.NetworkConfig, epoch uint64) error
	SaveValidatorSet(ctx context.Context, valset entity.ValidatorSet) error
}

type deriver interface {
	GetValidatorSet(ctx context.Context, epoch uint64, config entity.NetworkConfig) (entity.ValidatorSet, error)
}

type Config struct {
	EvmClient       evm.IEvmClient `validate:"required"`
	Repo            repo           `validate:"required"`
	Deriver         deriver        `validate:"required"`
	PollingInterval time.Duration  `validate:"required,gt=0"`
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

// LoadAllMissingEpochs runs tryLoadMissingEpochs until all missing epochs are loaded successfully
func (s *Service) LoadAllMissingEpochs(ctx context.Context) error {
	ctx = log.WithComponent(ctx, "listener")

	slog.InfoContext(ctx, "Loading all missing epochs before starting services")

	const maxRetries = 10
	retryCount := 0
	retryTimer := time.NewTimer(0)
	defer retryTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-retryTimer.C:
			if err := s.tryLoadMissingEpochs(ctx); err != nil {
				retryCount++
				if retryCount >= maxRetries {
					return errors.Errorf("failed to load missing epochs after %d retries: %w", maxRetries, err)
				}
				slog.ErrorContext(ctx, "Failed to load missing epochs, retrying", "error", err, "attempt", retryCount, "maxRetries", maxRetries)
				retryTimer.Reset(time.Second * 2)
				continue
			}
			slog.InfoContext(ctx, "Successfully loaded all missing epochs")
			return nil
		}
	}
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

	currentEpoch, err := s.cfg.EvmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return errors.Errorf("failed to get current epoch: %w", err)
	}

	latestHeader, err := s.cfg.Repo.GetLatestValidatorSetHeader(ctx)
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		return errors.Errorf("failed to get latest validator set header: %w", err)
	}

	nextEpoch := uint64(0)
	if err == nil {
		nextEpoch = latestHeader.Epoch + 1
	}

	for nextEpoch <= currentEpoch {
		epochStart, err := s.cfg.EvmClient.GetEpochStart(ctx, nextEpoch)
		if err != nil {
			return errors.Errorf("failed to get epoch start for epoch %d: %w", nextEpoch, err)
		}

		nextEpochConfig, err := s.cfg.EvmClient.GetConfig(ctx, epochStart)
		if err != nil {
			return errors.Errorf("failed to get network config for epoch %d: %w", nextEpoch, err)
		}

		nextValset, err := s.cfg.Deriver.GetValidatorSet(ctx, nextEpoch, nextEpochConfig)
		if err != nil {
			return errors.Errorf("failed to derive validator set extra for epoch %d: %w", nextEpoch, err)
		}

		if err := s.cfg.Repo.SaveConfig(ctx, nextEpochConfig, nextEpoch); err != nil {
			return errors.Errorf("failed to save validator set extra for epoch %d: %w", nextEpoch, err)
		}

		if err := s.cfg.Repo.SaveValidatorSet(ctx, nextValset); err != nil {
			return errors.Errorf("failed to save validator set extra for epoch %d: %w", nextEpoch, err)
		}

		slog.DebugContext(ctx, "Synced validator set", "epoch", nextEpoch, "config", nextEpochConfig, "valset", nextValset)

		nextEpoch = nextValset.Epoch + 1
	}

	slog.DebugContext(ctx, "All missing epochs loaded", "latestProcessedEpoch", currentEpoch)

	return nil
}
