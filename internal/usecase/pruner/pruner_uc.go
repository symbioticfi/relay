package pruner

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/tracing"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

//go:generate mockgen -source=pruner_uc.go -destination=mocks/pruner_mocks.go -package=mocks

type metrics interface {
	IncPrunedEpochsCount(entityType string)
}

type repo interface {
	GetOldestValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error)
	GetLatestValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error)
	PruneValsetEntities(ctx context.Context, epoch symbiotic.Epoch) error
	PruneProofEntities(ctx context.Context, epoch symbiotic.Epoch) error
	PruneSignatureEntitiesForEpoch(ctx context.Context, epoch symbiotic.Epoch) error
	PruneRequestIDEpochIndices(ctx context.Context, epoch symbiotic.Epoch) error
}

type Config struct {
	Repo                     repo    `validate:"required"`
	Metrics                  metrics `validate:"required"`
	Enabled                  bool
	Interval                 time.Duration `validate:"gte=0"`
	ValsetRetentionEpochs    uint64
	ProofRetentionEpochs     uint64
	SignatureRetentionEpochs uint64
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("pruner config validation failed: %w", err)
	}
	if c.Enabled && c.Interval <= 0 {
		return errors.New("pruner interval must be greater than zero when enabled")
	}
	return nil
}

type Service struct {
	cfg Config
}

func New(cfg Config) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	return &Service{
		cfg: cfg,
	}, nil
}

func (s *Service) Start(ctx context.Context) {
	ctx = log.WithComponent(ctx, "pruner")

	// Check if any retention is configured
	hasRetention := s.cfg.ValsetRetentionEpochs > 0 ||
		s.cfg.ProofRetentionEpochs > 0 ||
		s.cfg.SignatureRetentionEpochs > 0

	if !s.cfg.Enabled || !hasRetention {
		slog.InfoContext(ctx, "Pruner disabled")
		return
	}

	slog.InfoContext(ctx, "Starting pruner",
		"interval", s.cfg.Interval,
		"valsetRetentionEpochs", s.cfg.ValsetRetentionEpochs,
		"proofRetentionEpochs", s.cfg.ProofRetentionEpochs,
		"signatureRetentionEpochs", s.cfg.SignatureRetentionEpochs,
	)

	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.runPruning(ctx); err != nil {
				slog.ErrorContext(ctx, "Pruning failed", "error", err)
			}
		case <-ctx.Done():
			slog.InfoContext(ctx, "Pruner stopped")
			return
		}
	}
}

func (s *Service) runPruning(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx, "pruner.RunPruning")
	defer span.End()

	start := time.Now()

	latestEpoch, err := s.cfg.Repo.GetLatestValidatorSetEpoch(ctx)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			slog.DebugContext(ctx, "Pruning skipped", "reason", "no validator sets in storage yet")
			return nil
		}
		tracing.RecordError(span, err)
		return errors.Errorf("failed to get latest validator set epoch: %w", err)
	}

	tracing.SetAttributes(span, tracing.AttrEpoch.Int64(int64(latestEpoch)))

	oldestStoredEpoch, err := s.cfg.Repo.GetOldestValidatorSetEpoch(ctx)
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to get oldest validator set epoch: %w", err)
	}

	valsetCount, err := s.pruneValsetEntities(ctx, latestEpoch, oldestStoredEpoch)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to prune valset entities", "error", err)
	}

	proofCount, err := s.pruneProofEntities(ctx, latestEpoch, oldestStoredEpoch)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to prune proof entities", "error", err)
	}

	signatureCount, err := s.pruneSignatureEntities(ctx, latestEpoch, oldestStoredEpoch)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to prune signature entities", "error", err)
	}

	indexCount, err := s.pruneRequestIDEpochIndices(ctx, latestEpoch, oldestStoredEpoch)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to prune request ID epoch indices", "error", err)
	}

	slog.InfoContext(ctx, "Pruning completed",
		"valsetEpochs", valsetCount,
		"proofEpochs", proofCount,
		"signatureEpochs", signatureCount,
		"indexCleanupEpochs", indexCount,
		"duration", time.Since(start),
	)

	return nil
}

func (s *Service) pruneValsetEntities(ctx context.Context, latestEpoch, oldestStoredEpoch symbiotic.Epoch) (uint64, error) {
	return s.pruneEntities(
		ctx,
		latestEpoch,
		oldestStoredEpoch,
		s.cfg.ValsetRetentionEpochs,
		"valset",
		s.cfg.Repo.PruneValsetEntities,
	)
}

func (s *Service) pruneProofEntities(ctx context.Context, latestEpoch, oldestStoredEpoch symbiotic.Epoch) (uint64, error) {
	return s.pruneEntities(
		ctx,
		latestEpoch,
		oldestStoredEpoch,
		s.cfg.ProofRetentionEpochs,
		"proof",
		s.cfg.Repo.PruneProofEntities,
	)
}

func (s *Service) pruneSignatureEntities(ctx context.Context, latestEpoch, oldestStoredEpoch symbiotic.Epoch) (uint64, error) {
	return s.pruneEntities(
		ctx,
		latestEpoch,
		oldestStoredEpoch,
		s.cfg.SignatureRetentionEpochs,
		"signature",
		s.cfg.Repo.PruneSignatureEntitiesForEpoch,
	)
}

// pruneRequestIDEpochIndices cleans up request ID epoch indices for old epochs.
// It uses the maximum retention window of proofs and signatures to determine which epochs
// might have indices to clean up. The actual deletion only happens if both the aggregation
// proof and signatures have been pruned for a given requestID.
func (s *Service) pruneRequestIDEpochIndices(ctx context.Context, latestEpoch, oldestStoredEpoch symbiotic.Epoch) (uint64, error) {
	// Use the maximum of proof and signature retention to determine the range to scan
	maxRetention := max(s.cfg.SignatureRetentionEpochs, s.cfg.ProofRetentionEpochs)

	return s.pruneEntities(
		ctx,
		latestEpoch,
		oldestStoredEpoch,
		maxRetention,
		"requestIdEpochIndex",
		s.cfg.Repo.PruneRequestIDEpochIndices,
	)
}

// pruneEntities is a common utility function that implements the pruning logic for all entity types.
// It calculates the retention window and iterates through epochs to delete, calling the provided
// pruneFunc for each epoch.
func (s *Service) pruneEntities(
	ctx context.Context,
	latestEpoch, oldestStoredEpoch symbiotic.Epoch,
	retentionEpochs uint64,
	entityType string,
	pruneFunc func(context.Context, symbiotic.Epoch) error,
) (uint64, error) {
	ctx, span := tracing.StartSpan(ctx, "pruner.pruneEntities")
	defer span.End()

	tracing.SetAttributes(span,
		tracing.AttrEpoch.Int64(int64(latestEpoch)),
	)

	if retentionEpochs == 0 {
		tracing.AddEvent(span, "skipped_no_retention")
		return 0, nil
	}

	retentionWindow := symbiotic.Epoch(retentionEpochs)
	if latestEpoch < retentionWindow {
		tracing.AddEvent(span, "skipped_insufficient_epochs")
		return 0, nil
	}

	oldestToKeep := latestEpoch - retentionWindow + 1
	if oldestStoredEpoch >= oldestToKeep {
		tracing.AddEvent(span, "skipped_no_epochs_to_prune")
		return 0, nil
	}

	count := uint64(0)
	for epoch := oldestStoredEpoch; epoch < oldestToKeep; epoch++ {
		slog.DebugContext(ctx, "Pruning entities", "entityType", entityType, "epoch", epoch)

		if err := pruneFunc(ctx, epoch); err != nil {
			tracing.RecordError(span, err)
			return count, errors.Errorf("failed to prune %s entities for epoch %d: %w", entityType, epoch, err)
		}

		count++
		s.cfg.Metrics.IncPrunedEpochsCount(entityType)
	}

	tracing.SetAttributes(span, tracing.AttrEpochCount.Int(int(count)))

	return count, nil
}
