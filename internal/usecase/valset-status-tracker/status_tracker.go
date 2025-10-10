package valsetStatusTracker

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/symbiotic/client/evm"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

var zeroHeaderHash = common.HexToHash("0x868e09d528a16744c1f38ea3c10cc2251e01a456434f91172247695087d129b7")

type repo interface {
	GetConfigByEpoch(_ context.Context, epoch symbiotic.Epoch) (symbiotic.NetworkConfig, error)
	GetValidatorSetByEpoch(_ context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSet, error)
	UpdateValidatorSetStatus(ctx context.Context, valset symbiotic.ValidatorSet) error
	GetFirstUncommittedValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error)
	SaveFirstUncommittedValidatorSetEpoch(_ context.Context, epoch symbiotic.Epoch) error
}

type Config struct {
	EvmClient       evm.IEvmClient `validate:"required"`
	Repo            repo           `validate:"required"`
	PollingInterval time.Duration  `validate:"required,gt=0"`
}

type Service struct {
	cfg Config
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("invalid config: %w", err)
	}

	return nil
}

func New(cfg Config) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Service{
		cfg: cfg,
	}, nil
}

// TrackMissingEpochsStatuses runs trackCommittedEpochs until all missing epochs statuses are loaded successfully
func (s *Service) TrackMissingEpochsStatuses(ctx context.Context) error {
	ctx = log.WithComponent(ctx, "status_tracker")

	slog.InfoContext(ctx, "Track statuses of all missing epochs before starting services")

	const maxRetries = 10
	retryCount := 0
	retryTimer := time.NewTimer(0)
	defer retryTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-retryTimer.C:
			if err := s.trackCommittedEpochs(ctx); err != nil {
				retryCount++
				if retryCount >= maxRetries {
					return errors.Errorf("failed to track statuses of missing epochs after %d retries: %w", maxRetries, err)
				}
				slog.ErrorContext(ctx, "Failed to track statuses of missing epochs, retrying", "error", err, "attempt", retryCount, "maxRetries", maxRetries)
				retryTimer.Reset(time.Second * 2)
				continue
			}
			slog.InfoContext(ctx, "Successfully tracked statuses of all missing epochs")
			return nil
		}
	}
}

func (s *Service) Start(ctx context.Context) error {
	ctx = log.WithComponent(ctx, "status_tracker")

	slog.InfoContext(ctx, "Starting status tracker service", "pollingInterval", s.cfg.PollingInterval)

	timer := time.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			if err := s.trackCommittedEpochs(ctx); err != nil {
				return errors.Errorf("failed to track committed epochs: %w", err)
			}
			timer.Reset(s.cfg.PollingInterval)
		}
	}
}

func (s *Service) HandleProofAggregated(ctx context.Context, msg symbiotic.AggregationProof) error {
	valset, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, msg.Epoch)
	if err != nil {
		return errors.Errorf("failed to get validator set: %w", err) // if not found then it's failure case
	}

	if valset.Status != symbiotic.HeaderDerived {
		slog.DebugContext(ctx, "Validator set is already aggregated or committed", "epoch", valset.Epoch)
		return nil
	}

	valset.Status = symbiotic.HeaderAggregated
	if err := s.cfg.Repo.UpdateValidatorSetStatus(ctx, valset); err != nil {
		return errors.Errorf("failed to save validator set: %w", err)
	}

	slog.InfoContext(ctx, "Validator set is aggregated", "epoch", valset.Epoch)

	return nil
}

func (s *Service) trackCommittedEpochs(ctx context.Context) error {
	fce, err := s.cfg.Repo.GetFirstUncommittedValidatorSetEpoch(ctx)
	if err != nil {
		return errors.Errorf("failed to get first uncommitted validator set epoch: %w", err)
	}

	firstUncommittedEpoch := uint64(fce)

	settlements, err := s.findLatestNonZeroSettlements(ctx)
	if err != nil {
		return errors.Errorf("failed to find latest settlements: %w", err)
	}

	if len(settlements) == 0 {
		slog.InfoContext(ctx, "No settlements found, nothing to do")
		return nil
	}

	var lastCommittedEpoch uint64 = math.MaxUint64

	for _, settlement := range settlements {
		lce, err := s.cfg.EvmClient.GetLastCommittedHeaderEpoch(ctx, settlement)
		if err != nil {
			return errors.Errorf("failed to get last committed header epoch: %w", err)
		}

		lastCommittedEpoch = min(lastCommittedEpoch, uint64(lce))
	}

	for epoch := firstUncommittedEpoch; epoch <= lastCommittedEpoch; epoch++ {
		valset, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, symbiotic.Epoch(epoch))
		if err != nil {
			if errors.Is(err, entity.ErrEntityNotFound) {
				slog.DebugContext(ctx, "No uncommitted valset found, waiting...", "epoch", epoch)
				break
			}
			return errors.Errorf("failed to get validator set for epoch %d: %w", epoch, err)
		}

		if valset.Status == symbiotic.HeaderCommitted {
			continue
		}

		config, err := s.cfg.Repo.GetConfigByEpoch(ctx, valset.Epoch)
		if err != nil {
			return errors.Errorf("failed to get config for epoch %d: %w", epoch, err)
		}

		if len(config.Settlements) == 0 {
			continue
		}

		header, err := valset.GetHeader()
		if err != nil {
			return errors.Errorf("failed to get validator set header: %w", err)
		}

		hash, err := header.Hash()
		if err != nil {
			return errors.Errorf("failed to hash validator set header: %w", err)
		}

		isCommitted := true
		for _, settlement := range config.Settlements {
			committedHash, err := s.cfg.EvmClient.GetHeaderHashAt(ctx, settlement, valset.Epoch)
			if err != nil {
				return errors.Errorf("failed to get header hash for epoch %d: %w", epoch, err)
			}

			if committedHash == zeroHeaderHash {
				isCommitted = false
				continue
			}

			if hash != committedHash {
				return errors.Errorf("header hash for epoch %d is not equal to committed hash, derived: %s, committed: %s", epoch, hash.Hex(), committedHash.Hex())
			}
		}

		if !isCommitted {
			continue
		}

		valset.Status = symbiotic.HeaderCommitted
		if err := s.cfg.Repo.UpdateValidatorSetStatus(ctx, valset); err != nil {
			return errors.Errorf("failed to save validator set: %w", err)
		}

		slog.InfoContext(ctx, "Validator set is committed", "epoch", epoch)
	}

	if err := s.cfg.Repo.SaveFirstUncommittedValidatorSetEpoch(ctx, symbiotic.Epoch(lastCommittedEpoch+1)); err != nil {
		return errors.Errorf("failed to save last uncommitted validator set: %w", err)
	}

	return nil
}

func (s *Service) findLatestNonZeroSettlements(ctx context.Context) ([]symbiotic.CrossChainAddress, error) {
	currentEpoch, err := s.cfg.EvmClient.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to get current epoch: %w", err)
	}

	for epoch := currentEpoch; ; epoch-- {
		config, err := s.cfg.Repo.GetConfigByEpoch(ctx, epoch)
		if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
			return nil, errors.Errorf("failed to get config for epoch %d: %w", epoch, err)
		}

		if errors.Is(err, entity.ErrEntityNotFound) {
			epochStart, err := s.cfg.EvmClient.GetEpochStart(ctx, epoch)
			if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
				return nil, errors.Errorf("failed to get epoch %d: %w", epoch, err)
			}
			config, err = s.cfg.EvmClient.GetConfig(ctx, epochStart)
			if err != nil {
				return nil, errors.Errorf("failed to get config for epoch %d: %w", epoch, err)
			}
		}

		if len(config.Settlements) != 0 {
			return config.Settlements, nil
		}

		if epoch == 0 {
			break
		}
	}

	return nil, nil
}
