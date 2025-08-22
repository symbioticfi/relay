package valset_listener

import (
	"bytes"
	"context"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/symbioticfi/relay/core/client/evm"

	strategyTypes "github.com/symbioticfi/relay/core/usecase/growth-strategy/strategy-types"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/pkg/log"
)

type repo interface {
	GetLatestValidatorSet(ctx context.Context) (entity.ValidatorSet, error)
	SaveConfig(ctx context.Context, config entity.NetworkConfig, epoch uint64) error
	SaveValidatorSet(ctx context.Context, valset entity.ValidatorSet) error
}

type deriver interface {
	GetValidatorSet(ctx context.Context, epoch uint64, config entity.NetworkConfig) (entity.ValidatorSet, error)
	GetLastCommittedHeaderEpoch(ctx context.Context, config entity.NetworkConfig) (entity.CrossChainAddress, uint64, error)
}

type Config struct {
	EvmClient       evm.IEvmClient               `validate:"required"`
	Repo            repo                         `validate:"required"`
	Deriver         deriver                      `validate:"required"`
	GrowthStrategy  strategyTypes.GrowthStrategy `validate:"required"`
	PollingInterval time.Duration                `validate:"required,gt=0"`
}

var emptyValsetHeaderHash = common.HexToHash("0x868e09d528a16744c1f38ea3c10cc2251e01a456434f91172247695087d129b7")

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

		nextValset.Status, nextValset.PreviousHeaderHash, err = s.getStatusAndPreviousHash(ctx, nextEpoch, nextEpochConfig, nextValset)
		if err != nil {
			return errors.Errorf("failed to get status and previous hash for epoch %d: %w", nextEpoch, err)
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

	valset.Status, valset.PreviousHeaderHash, err = s.getStatusAndPreviousHash(ctx, epoch, config, valset)
	if err != nil {
		return errors.Errorf("failed to get status and previous hash for epoch %d: %w", epoch, err)
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

func (s *Service) getStatusAndPreviousHash(ctx context.Context, epoch uint64, config entity.NetworkConfig, valset entity.ValidatorSet) (entity.ValidatorSetStatus, common.Hash, error) {
	committedAddr, isValsetCommitted, err := s.isValsetHeaderCommitted(ctx, config, epoch)
	if err != nil {
		return 0, common.Hash{}, errors.Errorf("failed to check if validator committed at epoch %d: %w", epoch, err)
	}

	if isValsetCommitted {
		previousHeaderHash, err := s.cfg.EvmClient.GetPreviousHeaderHashAt(ctx, committedAddr, epoch)
		if err != nil {
			return 0, common.Hash{}, errors.Errorf("failed to get previous header hash: %w", err)
		}
		// valset integrity check
		valset.PreviousHeaderHash = previousHeaderHash
		committedHash, err := s.cfg.EvmClient.GetHeaderHashAt(ctx, committedAddr, epoch)
		if err != nil {
			return 0, common.Hash{}, errors.Errorf("failed to get header hash: %w", err)
		}
		valsetHeader, err := valset.GetHeader()
		if err != nil {
			return 0, common.Hash{}, errors.Errorf("failed to get header hash: %w", err)
		}
		calculatedHash, err := valsetHeader.Hash()
		if err != nil {
			return 0, common.Hash{}, errors.Errorf("failed to get header hash: %w", err)
		}

		if !bytes.Equal(committedHash[:], calculatedHash[:]) {
			slog.DebugContext(ctx, "Validator set integrity check failed", "committed hash", committedHash, "calculated hash", calculatedHash)
			return 0, common.Hash{}, errors.Errorf("validator set hash mistmach at epoch %d", epoch)
		}
		slog.DebugContext(ctx, "Validator set integrity check passed", "hash", committedHash)

		return entity.HeaderCommitted, previousHeaderHash, nil
	}

	// valset not committed

	lastCommittedAddr, latestCommittedEpoch, err := s.cfg.Deriver.GetLastCommittedHeaderEpoch(ctx, config)
	if err != nil {
		return 0, common.Hash{}, errors.Errorf("failed to get current valset epoch: %w", err)
	}

	if epoch < latestCommittedEpoch {
		slog.DebugContext(ctx, "Header is not committed [missed header]", "epoch", epoch)
		// zero PreviousHeaderHash cos header is orphaned
		return entity.HeaderMissed, emptyValsetHeaderHash, nil
	}

	// trying to link to latest committed header
	slog.DebugContext(ctx, "Header is not committed [new header]", "epoch", epoch)
	previousHeaderHash, err := s.cfg.EvmClient.GetHeaderHash(ctx, lastCommittedAddr)
	if err != nil {
		return 0, common.Hash{}, errors.Errorf("failed to get latest header hash: %w", err)
	}

	return entity.HeaderPending, previousHeaderHash, nil
}

func (s *Service) isValsetHeaderCommitted(ctx context.Context, config entity.NetworkConfig, epoch uint64) (entity.CrossChainAddress, bool, error) {
	for _, addr := range config.Replicas {
		isCommitted, err := s.cfg.EvmClient.IsValsetHeaderCommittedAt(ctx, addr, epoch)
		if err != nil {
			return entity.CrossChainAddress{}, false, errors.Errorf("failed to check if valset header is committed at epoch %d: %w", epoch, err)
		}
		if isCommitted {
			return addr, true, nil
		}
	}
	return entity.CrossChainAddress{}, false, nil
}
