package signer_app

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
)

type repo interface {
}

type p2pService interface {
	BroadcastSignatureGeneratedMessage(ctx context.Context, msg entity.SignatureHashMessage) error
}

type Config struct {
	P2PService p2pService  `validate:"required"`
	KeyPair    bls.KeyPair `validate:"required"`
	Repo       repo        `validate:"required"`
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("failed to validate config: %w", err)
	}

	return nil
}

type SignerApp struct {
	cfg Config
}

func NewSignerApp(cfg Config) (*SignerApp, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	return &SignerApp{
		cfg: cfg,
	}, nil
}

func (s *SignerApp) Sign(ctx context.Context, req entity.SignatureRequest) error {
	headerSignature, err := s.cfg.KeyPair.Sign(req.Message)
	if err != nil {
		return errors.Errorf("failed to sign valset header hash: %w", err)
	}

	slog.InfoContext(ctx, "valset header hash signed, sending via p2p", "headerSignature", headerSignature)

	err = s.cfg.P2PService.BroadcastSignatureGeneratedMessage(ctx, entity.SignatureHashMessage{
		MessageHash: req.Message,
		KeyTag:      req.KeyTag,
		Signature:   bls.SerializeG1(headerSignature),
		PublicKeyG1: bls.SerializeG1(&s.cfg.KeyPair.PublicKeyG1),
		PublicKeyG2: bls.SerializeG2(&s.cfg.KeyPair.PublicKeyG2),
		HashType:    entity.HashTypeValsetHeader,
		Epoch:       req.RequiredEpoch,
	})
	if err != nil {
		return errors.Errorf("failed to broadcast valset header: %w", err)
	}

	return nil
}
