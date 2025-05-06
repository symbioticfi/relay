package valset_header_generator_app

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"middleware-offchain/internal/client/p2p"
	"middleware-offchain/internal/entity"
	"middleware-offchain/valset/types"
)

type p2pService interface {
	Broadcast(typ entity.P2PMessageType, obj p2p.Data) error
}

type ethClient interface {
	GetCurrentPhase(ctx context.Context) (entity.Phase, error)
}

type valsetGenerator interface {
	GenerateValidatorSetHeader(ctx context.Context) (types.ValidatorSetHeader, error)
}

type Config struct {
	PollingInterval time.Duration   `validate:"required,gt=0"`
	ValsetGenerator valsetGenerator `validate:"required"`
	EthClient       ethClient       `validate:"required"`
	P2PService      p2pService      `validate:"required"`
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("failed to validate config: %w", err)
	}

	return nil
}

type SignerApp struct {
	cfg          Config
	currentPhase entity.Phase
}

func NewValsetHeaderGeneratorApp(cfg Config) (*SignerApp, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	return &SignerApp{
		cfg:          cfg,
		currentPhase: entity.IDLE,
	}, nil
}

func (s *SignerApp) Start(ctx context.Context) error {
	for {
		err := s.waitForCommitPhase(ctx)
		if err != nil {
			if errors.Is(err, entity.ErrPhaseFail) {
				return errors.Errorf("failed to wait for commit phase: %w", err)
			}
			if errors.Is(err, context.Canceled) {
				return err
			}

			return errors.Errorf("failed to wait for commit phase: %w", err)
		}

		slog.InfoContext(ctx, "commit phase started, generating valset header")

		header, err := s.cfg.ValsetGenerator.GenerateValidatorSetHeader(ctx)
		if err != nil {
			return errors.Errorf("failed to generate valset header: %w", err)
		}

		err = s.cfg.P2PService.Broadcast(entity.TypeValsetGenerated, &header)
		if err != nil {
			return errors.Errorf("failed to broadcast valset header: %w", err)
		}
	}
}

func (s *SignerApp) waitForCommitPhase(ctx context.Context) error {
	// todo ilya do not generate valset header if already generated once
	timer := time.NewTimer(0)
	defer timer.Stop()
	slog.InfoContext(ctx, "waiting for commit phase", "timeout", s.cfg.PollingInterval)

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "context canceled while waiting for commit phase, exiting")
			return ctx.Err()

		case <-timer.C:
			phase, err := s.cfg.EthClient.GetCurrentPhase(ctx)
			if err != nil {
				return errors.Errorf("failed to get current phase: %w", err)
			}

			slog.InfoContext(ctx, "got current phase", "phase", phase)

			switch phase {
			case entity.COMMIT:
				return nil
			case entity.FAIL:
				return errors.Errorf("current phase is FAIL: %w", entity.ErrPhaseFail)
			case entity.IDLE:
				slog.DebugContext(ctx, "current phase is IDLE, waiting for commit phase", "timeout", s.cfg.PollingInterval)
				timer.Reset(s.cfg.PollingInterval)
			default:
				return errors.Errorf("unknown phase: %v", phase)
			}
		}
	}
}
