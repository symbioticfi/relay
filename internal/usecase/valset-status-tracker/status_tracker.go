package valsetStatusTracker

import (
	"context"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/symbioticfi/relay/core/client/evm"
	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/pkg/log"
)

var zeroHeaderHash = common.HexToHash("0x00000")

type repo interface {
	GetConfigByEpoch(_ context.Context, epoch uint64) (entity.NetworkConfig, error)
	GetValidatorSetByEpoch(_ context.Context, epoch uint64) (entity.ValidatorSet, error)
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

type Service struct {
	cfg                   Config
	firstUncommittedEpoch uint64
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
		cfg:                   cfg,
		firstUncommittedEpoch: 0,
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
			if err := s.trackCommittedEpochs(ctx); err != nil {
				return errors.Errorf("failed to track committed epochs: %w", err)
			}
			timer.Reset(s.cfg.PollingInterval)
		}
	}
}

func (s *Service) HandleProofAggregated(ctx context.Context, msg entity.AggregatedSignatureMessage) error {
	valset, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, uint64(msg.Epoch))
	if err != nil {
		return errors.Errorf("failed to get validator set: %w", err) // if not found then it's failure case
	}

	if valset.Status != entity.HeaderDerived {
		slog.InfoContext(ctx, "Validator set is already aggregated or committed", "epoch", valset.Epoch)
		return nil
	}

	valset.Status = entity.HeaderAggregated
	if err := s.cfg.Repo.SaveValidatorSet(ctx, valset); err != nil {
		return errors.Errorf("failed to save validator set: %w", err)
	}

	if valset.Epoch == s.firstUncommittedEpoch+1 {
		s.firstUncommittedEpoch++
	}

	return nil
}

// ignore TODO: need to create an algorithm for effective uncommitted valsets tracking, either bitmaps either store reference to next uncommitted epoch
func (s *Service) trackCommittedEpochs(ctx context.Context) error {
	epoch := s.firstUncommittedEpoch
	for {
		valset, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, epoch)
		if err != nil {
			if errors.Is(err, entity.ErrEntityNotFound) {
				slog.InfoContext(ctx, "Valset not found, waiting...", "epoch", epoch)
				break
			}
			return errors.Errorf("failed to get validator set for epoch %d: %w", epoch, err)
		}

		epoch++

		if valset.Status == entity.HeaderCommitted {
			if valset.Epoch == s.firstUncommittedEpoch+1 {
				s.firstUncommittedEpoch++
			}
			continue
		}

		config, err := s.cfg.Repo.GetConfigByEpoch(ctx, valset.Epoch)
		if err != nil {
			return errors.Errorf("failed to get config for epoch %d: %w", epoch, err)
		}

		if len(config.Replicas) == 0 {
			if valset.Epoch == s.firstUncommittedEpoch+1 {
				s.firstUncommittedEpoch++
			}
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

		for _, replica := range config.Replicas {
			committedHash, err := s.cfg.EvmClient.GetHeaderHashAt(ctx, replica, valset.Epoch)
			if err != nil {
				return errors.Errorf("failed to get header hash for epoch %d: %w", epoch, err)
			}

			if committedHash == zeroHeaderHash {
				continue
			}

			if hash != committedHash {
				return errors.Errorf("header hash for epoch %d is not equal to committed hash, derived: %s, committed: %s", epoch, hash.Hex(), committedHash.Hex())
			}
		}

		valset.Status = entity.HeaderCommitted
		if err := s.cfg.Repo.SaveValidatorSet(ctx, valset); err != nil {
			return errors.Errorf("failed to save validator set: %w", err)
		}

		slog.InfoContext(ctx, "Validator set is committed", "epoch", epoch)

		if valset.Epoch == s.firstUncommittedEpoch+1 {
			s.firstUncommittedEpoch++
			continue
		}
	}

	return nil
}
