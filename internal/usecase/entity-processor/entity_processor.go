package entity_processor

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/signals"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

//go:generate mockgen -source=entity_processor.go -destination=mocks/entity_processor.go -package=mocks

// Repository defines the interface needed by the entity processor
type Repository interface {
	SaveSignature(ctx context.Context, signature symbiotic.Signature, validator symbiotic.Validator, activeIndex uint32) error
	GetSignatureByIndex(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash, validatorIndex uint32) (symbiotic.Signature, error)
	GetValidatorByKey(ctx context.Context, epoch symbiotic.Epoch, keyTag symbiotic.KeyTag, publicKey []byte) (symbiotic.Validator, uint32, error)
	GetValidatorSetByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSet, error)
	GetAggregationProof(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) (symbiotic.AggregationProof, error)
	SaveProof(ctx context.Context, aggregationProof symbiotic.AggregationProof) error
}

type Aggregator interface {
	Verify(valset symbiotic.ValidatorSet, keyTag symbiotic.KeyTag, aggregationProof symbiotic.AggregationProof) (bool, error)
}

type AggProofSignal interface {
	Emit(payload symbiotic.AggregationProof) error
}

type Config struct {
	Repo                     Repository                           `validate:"required"`
	Aggregator               Aggregator                           `validate:"required"`
	AggProofSignal           AggProofSignal                       `validate:"required"`
	SignatureProcessedSignal *signals.Signal[symbiotic.Signature] `validate:"required"`
}

func (c Config) Validate() error {
	if err := validate.New().Struct(c); err != nil {
		return errors.Errorf("failed to validate config: %w", err)
	}

	return nil
}

// EntityProcessor handles both signature and aggregation proof processing with SignatureMap operations
type EntityProcessor struct {
	cfg Config
}

// NewEntityProcessor creates a new entity processor
func NewEntityProcessor(cfg Config) (*EntityProcessor, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	return &EntityProcessor{
		cfg: cfg,
	}, nil
}

// ProcessSignature processes a signature with SignatureMap operations and optionally saves SignatureRequest
func (s *EntityProcessor) ProcessSignature(ctx context.Context, signature symbiotic.Signature, self bool) error {
	slog.DebugContext(ctx, "Processing signature",
		"keyTag", signature.KeyTag,
		"requestId", signature.RequestID().Hex(),
		"epoch", signature.Epoch,
		"self", self,
	)

	validator, activeIndex, err := s.cfg.Repo.GetValidatorByKey(ctx, signature.Epoch, signature.KeyTag, signature.PublicKey.OnChain())
	if err != nil {
		return errors.Errorf("validator not found for public key %x, keyTag=%v, epoch=%v: %w", signature.PublicKey.OnChain(), signature.KeyTag, signature.Epoch, err)
	}

	// if self signature ignore validator check and signature existence check
	if !self {
		if !validator.IsActive {
			return errors.Errorf("validator %s is not active", validator.Operator.Hex())
		}

		// check if signature already exists
		_, err = s.cfg.Repo.GetSignatureByIndex(ctx, signature.Epoch, signature.RequestID(), activeIndex)
		if err == nil {
			return errors.Errorf("signature already exists for request ID %s and validator index %d: %w", signature.RequestID().Hex(), activeIndex, entity.ErrEntityAlreadyExist)
		}
		if !errors.Is(err, entity.ErrEntityNotFound) {
			return errors.Errorf("failed to check existing signature: %w", err)
		}

		err = signature.PublicKey.VerifyWithHash(signature.MessageHash, signature.Signature)
		if err != nil {
			return errors.Errorf("failed to verify signature: %w", err)
		}
	}

	if err := s.cfg.Repo.SaveSignature(ctx, signature, validator, activeIndex); err != nil {
		return errors.Errorf("failed to add signature: %w", err)
	}

	// Emit signal after successful processing
	if err := s.cfg.SignatureProcessedSignal.Emit(signature); err != nil {
		return errors.Errorf("failed to emit signature processed signal: %w", err)
	}

	return nil
}

// ProcessAggregationProof processes an aggregation proof by saving it and removing from pending collection
func (s *EntityProcessor) ProcessAggregationProof(ctx context.Context, aggregationProof symbiotic.AggregationProof) error {
	slog.DebugContext(ctx, "Processing proof",
		"keyTag", aggregationProof.KeyTag,
		"requestId", aggregationProof.RequestID().Hex(),
		"epoch", aggregationProof.Epoch,
	)

	_, err := s.cfg.Repo.GetAggregationProof(ctx, aggregationProof.Epoch, aggregationProof.RequestID())
	if err == nil {
		return errors.Errorf("aggregation proof already exists for request ID %s: %w", aggregationProof.RequestID().Hex(), entity.ErrEntityAlreadyExist)
	}
	if !errors.Is(err, entity.ErrEntityNotFound) {
		return errors.Errorf("failed to check existing aggregation proof: %w", err)
	}

	validatorSet, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, aggregationProof.Epoch)
	if err != nil {
		return errors.Errorf("failed to get validator set: %w", err)
	}

	ok, err := s.cfg.Aggregator.Verify(validatorSet, aggregationProof.KeyTag, aggregationProof)
	if err != nil {
		return errors.Errorf("failed to verify aggregation proof: %w", err)
	}
	if !ok {
		return errors.Errorf("aggregation proof invalid")
	}

	if err := s.cfg.Repo.SaveProof(ctx, aggregationProof); err != nil {
		return errors.Errorf("failed to add aggregation proof: %w", err)
	}

	// Emit signal after successful save
	if err := s.cfg.AggProofSignal.Emit(aggregationProof); err != nil {
		return errors.Errorf("failed to emit aggregation proof signal: %w", err)
	}

	return nil
}
