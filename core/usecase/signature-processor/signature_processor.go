package signature_processor

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
)

// Repository defines the interface needed by the signature processor
type Repository interface {
	DoUpdateInTx(ctx context.Context, f func(ctx context.Context) error) error
	GetSignatureMap(ctx context.Context, reqHash common.Hash) (entity.SignatureMap, error)
	UpdateSignatureMap(ctx context.Context, vm entity.SignatureMap) error
	SaveSignature(ctx context.Context, reqHash common.Hash, key entity.RawPublicKey, sig entity.SignatureExtended) error
	SaveSignatureRequest(ctx context.Context, req entity.SignatureRequest) error
	RemoveSignatureRequestPending(ctx context.Context, epoch entity.Epoch, reqHash common.Hash) error
	GetValidatorSetHeaderByEpoch(ctx context.Context, epoch uint64) (entity.ValidatorSetHeader, error)
	GetActiveValidatorCountByEpoch(ctx context.Context, epoch uint64) (uint32, error)
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
type SignatureProcessor struct {
	cfg Config
}

// NewSignatureProcessor creates a new signature processor
func NewSignatureProcessor(cfg Config) (*SignatureProcessor, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	return &SignatureProcessor{
		cfg: cfg,
	}, nil
}

// ProcessSignature processes a signature with SignatureMap operations and optionally saves SignatureRequest
func (s *SignatureProcessor) ProcessSignature(ctx context.Context, param entity.SaveSignatureParam) error {
	return s.cfg.Repo.DoUpdateInTx(ctx, func(ctx context.Context) error {
		signatureMap, err := s.cfg.Repo.GetSignatureMap(ctx, param.RequestHash)
		if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
			return errors.Errorf("failed to get valset signature map: %w", err)
		}
		if errors.Is(err, entity.ErrEntityNotFound) {
			// Get the total number of active validators for this epoch
			totalActiveValidators, err := s.cfg.Repo.GetActiveValidatorCountByEpoch(ctx, uint64(param.Epoch))
			if err != nil {
				return errors.Errorf("failed to get active validator count for epoch %d: %w", param.Epoch, err)
			}

			signatureMap = entity.NewSignatureMap(param.RequestHash, param.Epoch, totalActiveValidators)
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
			// Save to pending collection as well
			if param.KeyTag.Type().AggregationKey() {
				if err := s.cfg.Repo.SaveSignatureRequestPending(ctx, *param.SignatureRequest); err != nil {
					return errors.Errorf("failed to save signature request to pending collection: %v", err)
				}
			}
		}

		if param.KeyTag.Type().AggregationKey() {
			// Check if quorum is reached and remove from pending collection if so
			validatorSetHeader, err := s.cfg.Repo.GetValidatorSetHeaderByEpoch(ctx, uint64(param.Epoch))
			if err != nil {
				return errors.Errorf("failed to get validator set header: %v", err)
			}

			// todo check quorum threshold from signature request
			if signatureMap.ThresholdReached(validatorSetHeader.QuorumThreshold) {
				// Remove from pending collection since quorum is reached
				err := s.cfg.Repo.RemoveSignatureRequestPending(ctx, param.Epoch, param.RequestHash)
				if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
					return errors.Errorf("failed to remove signature request from pending collection: %v", err)
				}
				// If ErrEntityNotFound, it means it was already removed or never added - that's ok
			}
		}

		return nil
	})
}
