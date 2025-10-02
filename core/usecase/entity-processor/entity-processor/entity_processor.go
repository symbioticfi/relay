package entity_processor

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/pkg/signals"
)

//go:generate mockgen -source=entity_processor.go -destination=mocks/entity_processor.go -package=mocks

// Repository defines the interface needed by the entity processor
type Repository interface {
	AddSignature(ctx context.Context, signature entity.SignatureExtended) error
	GetValidatorSetByEpoch(ctx context.Context, epoch entity.Epoch) (entity.ValidatorSet, error)

	AddProof(ctx context.Context, aggregationProof entity.AggregationProof) error
}

type Aggregator interface {
	Verify(valset entity.ValidatorSet, keyTag entity.KeyTag, aggregationProof entity.AggregationProof) (bool, error)
}

type AggProofSignal interface {
	Emit(payload entity.AggregationProof) error
}

type Config struct {
	Repo                     Repository                                `validate:"required"`
	Aggregator               Aggregator                                `validate:"required"`
	AggProofSignal           AggProofSignal                            `validate:"required"`
	SignatureProcessedSignal *signals.Signal[entity.SignatureExtended] `validate:"required"`
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
func (s *EntityProcessor) ProcessSignature(ctx context.Context, signature entity.SignatureExtended) error {
	slog.DebugContext(ctx, "Processing signature",
		"keyTag", signature.KeyTag,
		"requestId", signature.RequestID().Hex(),
		"epoch", signature.Epoch,
	)

	if err := s.cfg.Repo.AddSignature(ctx, signature); err != nil {
		return errors.Errorf("failed to add signature: %w", err)
	}

	// Emit signal after successful processing
	if err := s.cfg.SignatureProcessedSignal.Emit(signature); err != nil {
		return errors.Errorf("failed to emit signature processed signal: %w", err)
	}

	return nil
}

// ProcessAggregationProof processes an aggregation proof by saving it and removing from pending collection
func (s *EntityProcessor) ProcessAggregationProof(ctx context.Context, aggregationProof entity.AggregationProof) error {
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

	if err := s.cfg.Repo.AddProof(ctx, aggregationProof); err != nil {
		return errors.Errorf("failed to add aggregation proof: %w", err)
	}

	// Emit signal after successful save
	if err := s.cfg.AggProofSignal.Emit(aggregationProof); err != nil {
		return errors.Errorf("failed to emit aggregation proof signal: %w", err)
	}

	return nil
}
