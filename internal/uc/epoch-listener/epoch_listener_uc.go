package epoch_listener

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
}

type deriver interface {
	GetValidatorSetExtraForEpoch(ctx context.Context, epoch *big.Int) (entity.ValidatorSetExtra, error)
	MakeValidatorSetHeaderHash(ctx context.Context, extra entity.ValidatorSetExtra) ([]byte, error)
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
			if err := s.processEpochs(ctx); err != nil {
				slog.ErrorContext(ctx, "failed to process epochs", "error", err)
			}
			timer.Reset(s.cfg.PollingInterval)
		}
	}
}

func (s *Service) processEpochs(ctx context.Context) error {
	newValsetExtra, err := s.tryLoadMissingEpochs(ctx)
	if err != nil {
		return errors.Errorf("failed to load missing epochs: %w", err)
	}
	if !newValsetExtra.IsPresent() {
		// no new validator set extra found, nothing to do
		return nil
	}

	headerHash, err := s.cfg.Deriver.MakeValidatorSetHeaderHash(ctx, newValsetExtra.MustGet())
	if err != nil {
		return errors.Errorf("failed to make validator set header hash: %w", err)
	}

	// todo ilya what to do if we failed to sign?
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

func (s *Service) tryLoadMissingEpochs(ctx context.Context) (mo.Option[entity.ValidatorSetExtra], error) {
	phase, err := s.cfg.Eth.GetCurrentPhase(ctx)
	if err != nil {
		return mo.None[entity.ValidatorSetExtra](), errors.Errorf("failed to get current phase: %w", err)
	}

	if phase == entity.FAIL {
		return mo.None[entity.ValidatorSetExtra](), errors.New(entity.ErrPhaseFail)
	}

	currentOnchainEpoch, err := s.cfg.Eth.GetCurrentEpoch(ctx)
	if err != nil {
		return mo.None[entity.ValidatorSetExtra](), errors.Errorf("failed to get current epoch: %w", err)
	}

	latest, err := s.cfg.Repo.GetLatestValsetExtra(ctx)
	if err != nil {
		return mo.None[entity.ValidatorSetExtra](), errors.Errorf("failed to get latest validator set extra: %w", err)
	}

	currentStoredEpoch := new(big.Int).SetInt64(1)
	if latest.IsPresent() {
		currentStoredEpoch = latest.MustGet().Epoch
	}

	for {
		latestEpoch := new(big.Int).SetInt64(1)
		if latest.IsPresent() {
			latestEpoch = latest.MustGet().Epoch
		}

		diff := new(big.Int).Sub(currentOnchainEpoch, latestEpoch)
		if diff.Cmp(big.NewInt(0)) <= 0 {
			break
		}
		nextEpoch := new(big.Int).Add(latestEpoch, big.NewInt(1))

		nextValsetExtra, err := s.cfg.Deriver.GetValidatorSetExtraForEpoch(ctx, nextEpoch)
		if err != nil {
			return mo.None[entity.ValidatorSetExtra](), errors.Errorf("failed to derive validator set extra for epoch %s: %w", nextEpoch.String(), err)
		}

		// TODO ilya: check valset integrity: valset.headerHash() == master.valsetHeaderHash(epoch)

		if err := s.cfg.Repo.SaveValsetExtra(ctx, nextValsetExtra); err != nil {
			return mo.None[entity.ValidatorSetExtra](), errors.Errorf("failed to save validator set extra for epoch %s: %w", nextEpoch.String(), err)
		}

		latest = mo.Some(nextValsetExtra)
	}

	if currentOnchainEpoch.Cmp(currentStoredEpoch) != 0 {
		return latest, nil
	}

	return mo.None[entity.ValidatorSetExtra](), nil
}
