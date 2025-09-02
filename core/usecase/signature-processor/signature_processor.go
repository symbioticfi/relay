package signature_processor

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
)

//go:generate mockgen -source=signature_processor.go -destination=mocks/signature_processor.go -package=mocks

// Repository defines the interface needed by the signature processor
type Repository interface {
	DoUpdateInTx(ctx context.Context, f func(ctx context.Context) error) error
	GetSignatureMap(ctx context.Context, reqHash common.Hash) (entity.SignatureMap, error)
	UpdateSignatureMap(ctx context.Context, vm entity.SignatureMap) error
	SaveSignature(ctx context.Context, reqHash common.Hash, key entity.RawPublicKey, sig entity.SignatureExtended) error
	SaveSignatureRequest(ctx context.Context, req entity.SignatureRequest) error
}

type Config struct {
	Repo Repository `validate:"required"`
}

func (c Config) Validate() error {
	if err := validate.New().Struct(c); err != nil {
		return errors.Errorf("failed to validate config: %w", err)
	}

	return nil
}

// SignatureProcessor handles signature processing with SignatureMap operations
type SignatureProcessor interface {
	ProcessSignature(ctx context.Context, param entity.SaveSignatureParam) error
}

type signatureProcessor struct {
	cfg Config
}

// NewSignatureProcessor creates a new signature processor
func NewSignatureProcessor(cfg Config) (SignatureProcessor, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	return &signatureProcessor{
		cfg: cfg,
	}, nil
}

// ProcessSignature processes a signature with SignatureMap operations and optionally saves SignatureRequest
func (s *signatureProcessor) ProcessSignature(ctx context.Context, param entity.SaveSignatureParam) error {
	return s.cfg.Repo.DoUpdateInTx(ctx, func(ctx context.Context) error {
		signatureMap, err := s.cfg.Repo.GetSignatureMap(ctx, param.RequestHash)
		if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
			return errors.Errorf("failed to get valset signature map: %w", err)
		}
		if errors.Is(err, entity.ErrEntityNotFound) {
			signatureMap = entity.NewSignatureMap(param.RequestHash, param.Epoch)
		}

		if err := signatureMap.SetValidatorPresent(param.ActiveIndex, param.VotingPower); err != nil {
			return errors.Errorf("failed to set validator present for request %s: %w", param.RequestHash.Hex(), err)
		}

		if err := s.cfg.Repo.UpdateSignatureMap(ctx, signatureMap); err != nil {
			return errors.Errorf("failed to update valset signature map: %w", err)
		}

		if err := s.cfg.Repo.SaveSignature(ctx, param.RequestHash, param.Key, param.Signature); err != nil {
			return errors.Errorf("failed to save signature: %w", err)
		}

		if param.SignatureRequest != nil {
			if err := s.cfg.Repo.SaveSignatureRequest(ctx, *param.SignatureRequest); err != nil {
				return errors.Errorf("failed to save signature request: %w", err)
			}
		}

		return nil
	})
}
