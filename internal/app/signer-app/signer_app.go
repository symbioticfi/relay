package signer_app

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
)

type p2pService interface {
	BroadcastSignatureGeneratedMessage(ctx context.Context, msg entity.SignatureHashMessage) error
}

type ethClient interface {
	GetCurrentPhase(ctx context.Context) (entity.Phase, error)
}

type valsetGenerator interface {
	GenerateCurrentValidatorSetHeader(ctx context.Context) (entity.ValidatorSetHeader, error)
	GenerateValidatorSetHeaderHash(validatorSetHeader entity.ValidatorSetHeader) ([]byte, error)
}

type Config struct {
	PollingInterval time.Duration   `validate:"required,gt=0"`
	ValsetGenerator valsetGenerator `validate:"required"`
	EthClient       ethClient       `validate:"required"`
	P2PService      p2pService      `validate:"required"`
	KeyPair         bls.KeyPair     `validate:"required"`
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("failed to validate config: %w", err)
	}

	return nil
}

type SignerApp struct {
	cfg           Config
	previousPhase entity.Phase
}

func NewSignerApp(cfg Config) (*SignerApp, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	return &SignerApp{
		cfg:           cfg,
		previousPhase: entity.IDLE,
	}, nil
}

func (s *SignerApp) Start(ctx context.Context) error {
	for {
		err := s.waitForCommitPhase(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}

			return errors.Errorf("failed to wait for commit phase: %w", err)
		}

		slog.InfoContext(ctx, "commit phase started, generating valset header")

		header, err := s.cfg.ValsetGenerator.GenerateCurrentValidatorSetHeader(ctx)
		if err != nil {
			return errors.Errorf("failed to generate valset header: %w", err)
		}

		slog.DebugContext(ctx, "valset header generated, generating hash", "header", header)

		headerHash, err := s.cfg.ValsetGenerator.GenerateValidatorSetHeaderHash(header)
		if err != nil {
			return errors.Errorf("failed to generate valset header hash: %w", err)
		}

		slog.DebugContext(ctx, "valset header hash generated, signing", "headerHash", headerHash)

		headerSignature, err := s.cfg.KeyPair.Sign(headerHash)
		if err != nil {
			return errors.Errorf("failed to sign valset header hash: %w", err)
		}

		slog.DebugContext(ctx, "valset header hash signed, sending via p2p", "headerSignature", headerSignature)

		err = s.cfg.P2PService.BroadcastSignatureGeneratedMessage(ctx, entity.SignatureHashMessage{
			MessageHash: headerHash,
			KeyTag:      15, // todo ilya: pass from config or from another place
			Signature:   bls.SerializeG1(headerSignature),
			PublicKeyG1: bls.SerializeG1(&s.cfg.KeyPair.PublicKeyG1),
			PublicKeyG2: bls.SerializeG2(&s.cfg.KeyPair.PublicKeyG2),
		})
		if err != nil {
			return errors.Errorf("failed to broadcast valset header: %w", err)
		}

		slog.DebugContext(ctx, "valset header sent p2p, waiting for the next cycle")
	}
}

func (s *SignerApp) waitForCommitPhase(ctx context.Context) error {
	timer := time.NewTimer(0)
	defer timer.Stop()
	slog.DebugContext(ctx, "waiting for commit phase", "timeout", s.cfg.PollingInterval)

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "context canceled while waiting for commit phase, exiting")
			return ctx.Err()

		case <-timer.C:
			slog.DebugContext(ctx, "trying to get current phase")
			phase, err := s.cfg.EthClient.GetCurrentPhase(ctx)
			if errors.Is(err, context.DeadlineExceeded) {
				slog.DebugContext(ctx, "context deadline exceeded while getting current phase, retrying")
				timer.Reset(s.cfg.PollingInterval)
				continue
			}
			if err != nil {
				return errors.Errorf("failed to get current phase: %w", err)
			}

			slog.DebugContext(ctx, "got current phase", "phase", phase)

			switch phase {
			case entity.COMMIT:
				if s.previousPhase == entity.COMMIT {
					slog.DebugContext(ctx, "current phase is COMMIT, waiting for next cycle")
					timer.Reset(s.cfg.PollingInterval)
					continue
				}
				s.previousPhase = entity.COMMIT
				return nil
			case entity.FAIL:
				s.previousPhase = entity.FAIL
				slog.DebugContext(ctx, "current phase is FAIL, waiting for commit phase", "timeout", s.cfg.PollingInterval)
				timer.Reset(s.cfg.PollingInterval)
			case entity.IDLE:
				s.previousPhase = entity.IDLE
				slog.DebugContext(ctx, "current phase is IDLE, waiting for commit phase", "timeout", s.cfg.PollingInterval)
				timer.Reset(s.cfg.PollingInterval)
			default:
				return errors.Errorf("unknown phase: %v", phase)
			}
		}
	}
}
