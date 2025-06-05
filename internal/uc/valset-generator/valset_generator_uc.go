package valset_generator

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"middleware-offchain/internal/entity"
)

type signer interface {
	Sign(ctx context.Context, req entity.SignatureRequest) error
}

type eth interface {
	GetCurrentEpoch(ctx context.Context) (uint64, error)
	GetCurrentPhase(ctx context.Context) (entity.Phase, error)
	GetConfig(ctx context.Context, timestamp uint64) (entity.NetworkConfig, error)
	GetEpochStart(ctx context.Context, epoch uint64) (uint64, error)
}

type repo interface {
	GetLatestSignedValset(_ context.Context) (entity.ValidatorSet, error)
	GetLatestValset(ctx context.Context) (entity.ValidatorSet, error)
}

type deriver interface {
	GetValidatorSet(ctx context.Context, epoch uint64, config entity.NetworkConfig) (entity.ValidatorSet, error)
	GetNetworkData(ctx context.Context) (entity.NetworkData, error)
	GenerateExtraData(
		valset *entity.ValidatorSet,
		config *entity.NetworkConfig,
	) ([]entity.ExtraData, error)
	HeaderCommitmentHash(networkData entity.NetworkData, header entity.ValidatorSetHeader, extraData []entity.ExtraData) ([]byte, error)
}

type Config struct {
	Signer          signer        `validate:"required"`
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
	cfg            Config
	generatedEpoch uint64
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
			if err := s.process(ctx); err != nil {
				slog.ErrorContext(ctx, "failed to process epochs", "error", err)
			}
			timer.Reset(s.cfg.PollingInterval)
		}
	}
}

func (s *Service) process(ctx context.Context) error {
	valSet, config, err := s.tryDetectNewEpochToCommit(ctx)
	if err != nil {
		return errors.Errorf("failed to detect new epoch to commit: %w", err)
	}
	if valSet == nil {
		// no new validator set extra found, nothing to do
		return nil
	}

	if s.generatedEpoch >= valSet.Epoch {
		slog.DebugContext(ctx, "no new epoch to commit, already generated for this epoch", "epoch", valSet.Epoch)
		return nil
	}

	networkData, err := s.cfg.Deriver.GetNetworkData(ctx)
	if err != nil {
		return errors.Errorf("failed to get network data: %w", err)
	}

	extraData, err := s.cfg.Deriver.GenerateExtraData(valSet, config)
	if err != nil {
		return errors.Errorf("failed to generate extra data: %w", err)
	}

	header, err := valSet.GetHeader()
	if err != nil {
		return errors.Errorf("failed to get validator set header: %w", err)
	}
	hash, err := s.cfg.Deriver.HeaderCommitmentHash(networkData, header, extraData)
	if err != nil {
		return errors.Errorf("failed to get header commitment hash: %w", err)
	}

	latestValset, err := s.cfg.Repo.GetLatestValset(ctx)
	if err != nil {
		return errors.Errorf("failed to get latest validator set extra: %w", err)
	}

	err = s.cfg.Signer.Sign(ctx, entity.SignatureRequest{
		KeyTag:        entity.ValsetHeaderKeyTag,
		RequiredEpoch: latestValset.Epoch,
		Message:       hash,
	})
	if err != nil {
		return errors.Errorf("failed to sign new validator set extra: %w", err)
	}

	s.generatedEpoch = header.Epoch

	return nil
}

func (s *Service) tryDetectNewEpochToCommit(ctx context.Context) (*entity.ValidatorSet, *entity.NetworkConfig, error) {
	phase, err := s.cfg.Eth.GetCurrentPhase(ctx)
	if err != nil {
		return nil, nil, errors.Errorf("failed to get current phase: %w", err)
	}

	if phase == entity.IDLE {
		slog.DebugContext(ctx, "current phase is IDLE, no new epoch to commit")
		return nil, nil, nil // no new epoch to commit, idle phase
	}

	if phase != entity.COMMIT {
		return nil, nil, errors.Errorf("current phase is not COMMIT, got: %d: %w", phase, entity.ErrPhaseNotCommit)
	}

	currentOnchainEpoch, err := s.cfg.Eth.GetCurrentEpoch(ctx)
	if err != nil {
		return nil, nil, errors.Errorf("failed to get current epoch: %w", err)
	}

	epochStart, err := s.cfg.Eth.GetEpochStart(ctx, currentOnchainEpoch)
	if err != nil {
		return nil, nil, errors.Errorf("failed to get epoch start for epoch %d: %w", currentOnchainEpoch, err)
	}

	config, err := s.cfg.Eth.GetConfig(ctx, epochStart)
	if err != nil {
		return nil, nil, errors.Errorf("failed to get network config for epoch %d: %w", currentOnchainEpoch, err)
	}

	newValset, err := s.cfg.Deriver.GetValidatorSet(ctx, currentOnchainEpoch, config)
	if err != nil {
		return nil, nil, errors.Errorf("failed to get validator set extra for epoch %s: %w", currentOnchainEpoch, err)
	}

	return &newValset, &config, nil
}
