package entity_processor

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/signals"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

//go:generate mockgen -source=entity_processor.go -destination=mocks/entity_processor.go -package=mocks

// Repository defines the interface needed by the entity processor
type Repository interface {
	SaveSignature(ctx context.Context, signature symbiotic.Signature, validator symbiotic.Validator, activeIndex uint32) error
	GetSignatureByIndex(ctx context.Context, requestID common.Hash, validatorIndex uint32) (symbiotic.Signature, error)
	GetValidatorByKey(ctx context.Context, epoch symbiotic.Epoch, keyTag symbiotic.KeyTag, publicKey []byte) (symbiotic.Validator, uint32, error)
	GetValidatorSetByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSet, error)
	GetAggregationProof(ctx context.Context, requestID common.Hash) (symbiotic.AggregationProof, error)
	SaveProof(ctx context.Context, aggregationProof symbiotic.AggregationProof) error
	UpdateValidatorSetStatus(ctx context.Context, epoch symbiotic.Epoch, item symbiotic.ValidatorSetStatusItem) error
	GetLatestAggregatedValsetHeader(ctx context.Context) (symbiotic.ValidatorSetHeader, error)
}

type Aggregator interface {
	Verify(valset symbiotic.ValidatorSet, keyTag symbiotic.KeyTag, aggregationProof symbiotic.AggregationProof) (bool, error)
}

type AggProofSignal interface {
	Emit(payload symbiotic.AggregationProof) error
}

type Metrics interface {
	ObserveEpoch(epochType string, epochNumber uint64)
}

type Config struct {
	Repo                     Repository                           `validate:"required"`
	Aggregator               Aggregator                           `validate:"required"`
	AggProofSignal           AggProofSignal                       `validate:"required"`
	SignatureProcessedSignal *signals.Signal[symbiotic.Signature] `validate:"required"`
	Metrics                  Metrics                              `validate:"required"`
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
	ctx = log.WithAttrs(ctx,
		slog.String("requestId", signature.RequestID().Hex()),
		slog.Uint64("epoch", uint64(signature.Epoch)),
		slog.Uint64("keyTag", uint64(signature.KeyTag)),
	)
	slog.DebugContext(ctx, "Started processing signature", "self", self)

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
		_, err = s.cfg.Repo.GetSignatureByIndex(ctx, signature.RequestID(), activeIndex)
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
	ctx = log.WithAttrs(ctx,
		slog.String("requestId", aggregationProof.RequestID().Hex()),
		slog.Uint64("epoch", uint64(aggregationProof.Epoch)),
		slog.Uint64("keyTag", uint64(aggregationProof.KeyTag)),
	)
	slog.DebugContext(ctx, "Started processing aggregation proof")

	_, err := s.cfg.Repo.GetAggregationProof(ctx, aggregationProof.RequestID())
	if err == nil {
		return errors.Errorf("aggregation proof already exists for request ID %s: %w", aggregationProof.RequestID().Hex(), entity.ErrEntityAlreadyExist)
	}
	if !errors.Is(err, entity.ErrEntityNotFound) {
		return errors.Errorf("failed to check existing aggregation proof: %w", err)
	}

	validatorSet, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, aggregationProof.Epoch)
	if err != nil {
		return errors.Errorf("failed to get validator set for epoch %d: %w", aggregationProof.Epoch, err)
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

	if err := s.cfg.Repo.UpdateValidatorSetStatus(ctx, aggregationProof.Epoch, symbiotic.HeaderAggregated); err != nil {
		return errors.Errorf("failed to update validator set status: %w", err)
	}
	latestAggregatedHeader, err := s.cfg.Repo.GetLatestAggregatedValsetHeader(ctx)
	if latestFound := err == nil; latestFound {
		s.cfg.Metrics.ObserveEpoch("aggregated", uint64(latestAggregatedHeader.Epoch))
	} else if !errors.Is(err, entity.ErrEntityNotFound) {
		return errors.Errorf("failed to get latest aggregated valset header: %w", err)
	}

	slog.DebugContext(ctx, "Proof saved")

	// Emit signal after successful save
	if err := s.cfg.AggProofSignal.Emit(aggregationProof); err != nil {
		return errors.Errorf("failed to emit aggregation proof signal: %w", err)
	}

	return nil
}
