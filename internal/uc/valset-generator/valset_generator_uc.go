package valset_generator

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

type signer interface {
	Sign(ctx context.Context, req entity.SignatureRequest) error
}

type eth interface {
	GetCurrentEpoch(ctx context.Context) (*big.Int, error)
	GetCurrentPhase(ctx context.Context) (entity.Phase, error)
}

type repo interface {
	GetLatestValsetExtra(ctx context.Context) (mo.Option[entity.ValidatorSetExtra], error)
	SaveValsetExtra(ctx context.Context, extra entity.ValidatorSetExtra) error
	SaveLatestSignedValsetExtra(ctx context.Context, extra entity.ValidatorSetExtra) error
	GetLatestSignedValsetExtra(_ context.Context) (entity.ValidatorSetExtra, error)
}

type deriver interface {
	GetValidatorSetExtraForEpoch(ctx context.Context, epoch *big.Int) (entity.ValidatorSetExtra, error)
	MakeValidatorSetHeaderHash(ctx context.Context, extra entity.ValidatorSetExtra, data []entity.ExtraData) ([]byte, error)
	MakeValsetHeader(ctx context.Context, extra entity.ValidatorSetExtra) (entity.ValidatorSetHeader, error)
	GenerateExtraData(ctx context.Context, valsetHeader entity.ValidatorSetHeader, verificationType uint32) ([]entity.ExtraData, error)
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
			if err := s.process(ctx); err != nil {
				slog.ErrorContext(ctx, "failed to process epochs", "error", err)
			}
			timer.Reset(s.cfg.PollingInterval)
		}
	}
}

func (s *Service) process(ctx context.Context) error {
	newValsetExtra, err := s.tryDetectNewEpochToCommit(ctx)
	if err != nil {
		return errors.Errorf("failed to detect new epoch to commit: %w", err)
	}
	if !newValsetExtra.IsPresent() {
		// no new validator set extra found, nothing to do
		return nil
	}

	valsetHeader, err := s.cfg.Deriver.MakeValsetHeader(ctx, newValsetExtra.MustGet())
	if err != nil {
		return errors.Errorf("failed to generate validator set header: %w", err)
	}
	extraData, err := s.cfg.Deriver.GenerateExtraData(ctx, valsetHeader, entity.ZkVerificationType)
	if err != nil {
		return errors.Errorf("failed to generate extra data: %w", err)
	}

	headerHash, err := s.cfg.Deriver.MakeValidatorSetHeaderHash(ctx, newValsetExtra.MustGet(), extraData)
	if err != nil {
		return errors.Errorf("failed to make validator set header hash: %w", err)
	}

	err = s.cfg.Signer.Sign(ctx, entity.SignatureRequest{
		KeyTag:        entity.ValsetHeaderKeyTag,
		RequiredEpoch: newValsetExtra.MustGet().Epoch,
		Message:       headerHash,
	})
	if err != nil {
		return errors.Errorf("failed to sign new validator set extra: %w", err)
	}

	if err := s.cfg.Repo.SaveLatestSignedValsetExtra(ctx, newValsetExtra.MustGet()); err != nil {
		return errors.Errorf("failed to save latest signed validator set extra: %w", err)
	}

	return nil
}

func (s *Service) tryDetectNewEpochToCommit(ctx context.Context) (mo.Option[entity.ValidatorSetExtra], error) {
	phase, err := s.cfg.Eth.GetCurrentPhase(ctx)
	if err != nil {
		return mo.None[entity.ValidatorSetExtra](), errors.Errorf("failed to get current phase: %w", err)
	}

	if phase != entity.COMMIT {
		return mo.None[entity.ValidatorSetExtra](), errors.New(entity.ErrPhaseNotCommit)
	}

	currentOnchainEpoch, err := s.cfg.Eth.GetCurrentEpoch(ctx)
	if err != nil {
		return mo.None[entity.ValidatorSetExtra](), errors.Errorf("failed to get current epoch: %w", err)
	}

	latest, err := s.cfg.Repo.GetLatestSignedValsetExtra(ctx)
	if err != nil {
		return mo.None[entity.ValidatorSetExtra](), errors.Errorf("failed to getlatest validator set extra: %w", err)
	}

	if isGreaterOrEqual(latest.Epoch, currentOnchainEpoch) {
		return mo.None[entity.ValidatorSetExtra](), nil
	}

	newValset, err := s.cfg.Deriver.GetValidatorSetExtraForEpoch(ctx, currentOnchainEpoch)
	if err != nil {
		return mo.None[entity.ValidatorSetExtra](), errors.Errorf("failed to get validator set extra for epoch %s: %w", currentOnchainEpoch, err)
	}

	return mo.Some(newValset), nil
}

func isGreaterOrEqual(latest *big.Int, current *big.Int) bool {
	return latest.Cmp(current) >= 0
}
