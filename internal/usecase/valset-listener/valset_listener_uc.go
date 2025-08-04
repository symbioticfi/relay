package valset_listener

import (
	"context"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"

	strategyTypes "github.com/symbioticfi/relay/core/usecase/growth-strategy/strategy-types"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/pkg/log"
)

type evmClient interface {
	GetCurrentEpoch(ctx context.Context) (uint64, error)
	GetLastCommittedHeaderEpoch(ctx context.Context, addr entity.CrossChainAddress) (uint64, error)
	GetConfig(ctx context.Context, timestamp uint64) (entity.NetworkConfig, error)
	GetEpochStart(ctx context.Context, epoch uint64) (uint64, error)
	GetHeaderHashAt(ctx context.Context, addr entity.CrossChainAddress, epoch uint64) (common.Hash, error)
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
	EvmClient       evmClient                    `validate:"required"`
	Repo            repo                         `validate:"required"`
	Deriver         deriver                      `validate:"required"`
	GrowthStrategy  strategyTypes.GrowthStrategy `validate:"required"`
	PollingInterval time.Duration                `validate:"required,gt=0"`
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

	currentEpoch, err := s.cfg.EvmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return errors.Errorf("failed to get current epoch: %w", err)
	}
	currentEpochStart, err := s.cfg.EvmClient.GetEpochStart(ctx, currentEpoch)
	if err != nil {
		return errors.Errorf("failed to get current epoch start: %w", err)
	}
	config, err := s.cfg.EvmClient.GetConfig(ctx, currentEpochStart)
	if err != nil {
		return errors.Errorf("failed to get network config for current epoch: %w", err)
	}

	latestCommittedHash, latestCommittedEpoch, err := s.cfg.GrowthStrategy.GetLastCommittedHeaderHash(ctx, config)
	if err != nil {
		return errors.Errorf("failed to get latest committed header hash: %w", err)
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

	if err := s.validateHeaderHashAtLastCommittedEpoch(ctx, latestCommittedEpoch, latestCommittedHash); err != nil {
		return errors.Errorf("failed to validate header hash at last committed epoch: %w", err)
	}

	for latestCommittedEpoch >= nextEpoch {
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

		slog.DebugContext(ctx, "Synced validator set", "epoch", nextEpoch, "config", config, "valset", nextValset)

		s.latestProcessedEpoch = nextEpoch
		nextEpoch = nextValset.Epoch + 1
	}

	slog.DebugContext(ctx, "All missing epochs loaded", "latestProcessedEpoch", s.latestProcessedEpoch)

	return nil
}

func (s *Service) validateHeaderHashAtLastCommittedEpoch(ctx context.Context, epoch uint64, lastCommittedHash common.Hash) error {
	epochStart, err := s.cfg.EvmClient.GetEpochStart(ctx, epoch)
	if err != nil {
		return errors.Errorf("failed to get epoch start for epoch %d: %w", epochStart, err)
	}

	config, err := s.cfg.EvmClient.GetConfig(ctx, epochStart)
	if err != nil {
		return errors.Errorf("failed to get network config for epoch %d: %w", epoch, err)
	}

	valset, err := s.cfg.Deriver.GetValidatorSet(ctx, epoch, config)
	if err != nil {
		return errors.Errorf("failed to derive validator set extra for epoch %d: %w", epoch, err)
	}

	header, err := valset.GetHeader()
	if err != nil {
		return errors.Errorf("failed to get header for epoch %d: %w", epoch, err)
	}

	hash, err := header.Hash()
	if err != nil {
		return errors.Errorf("failed to get header hash for epoch %d: %w", epoch, err)
	}

	if lastCommittedHash != hash {
		return errors.Errorf("last committed header hash mismatch with derived hash for epoch %d, derived: %s, committed: %s", epoch, hash, lastCommittedHash)
	}

	return nil
}
