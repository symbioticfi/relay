package pruner

import (
	"context"
	stderrors "errors"
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

type NoopMetrics struct{}

func (NoopMetrics) IncPrunedEpochsCount(string) {}

type repo interface {
	GetOldestValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error)
	GetLatestValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error)
	PruneValsetEntities(ctx context.Context, epoch symbiotic.Epoch, batchSize int) error
	PruneProofEntities(ctx context.Context, epoch symbiotic.Epoch, batchSize int) error
	PruneSignatureEntitiesForEpoch(ctx context.Context, epoch symbiotic.Epoch, batchSize int) error
	PruneRequestIDEpochIndices(ctx context.Context, epoch symbiotic.Epoch, batchSize int) error
}

type Config struct {
	Repo                     repo    `validate:"required"`
	Metrics                  metrics `validate:"required"`
	Enabled                  bool
	Interval                 time.Duration `validate:"gte=0"`
	ValsetRetentionEpochs    uint64
	ProofRetentionEpochs     uint64
	SignatureRetentionEpochs uint64
	PruneBatchSize           int `validate:"gte=0"`
	// ProgressFn is invoked once with current=0 when a non-empty entity-type
	// pass starts (so the caller knows the total) and again after each pruned
	// epoch with current=1..total. Optional; nil disables progress reporting.
	ProgressFn func(entityType string, current, total uint64)
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("pruner config validation failed: %w", err)
	}
	if c.Enabled && c.Interval <= 0 {
		return errors.New("pruner interval must be greater than zero when enabled")
	}
	// Pruning a valset epoch without also pruning the proof / signature data
	// for that epoch leaves orphans (the dependent rows still reference an
	// epoch whose validator set is gone). Require proof + signature retention
	// whenever valset retention is set, and keep them no larger than valset
	// retention so the dependents are removed first.
	if c.ValsetRetentionEpochs > 0 {
		if c.ProofRetentionEpochs == 0 || c.SignatureRetentionEpochs == 0 {
			return errors.New("retention.valset-epochs requires retention.proof-epochs and retention.signature-epochs to also be set (otherwise pruning valset entities would orphan proof/signature data)")
		}
		if c.ProofRetentionEpochs > c.ValsetRetentionEpochs {
			return errors.Errorf("retention.proof-epochs (%d) must be <= retention.valset-epochs (%d) to avoid orphaning proofs whose valset has been pruned",
				c.ProofRetentionEpochs, c.ValsetRetentionEpochs)
		}
		if c.SignatureRetentionEpochs > c.ValsetRetentionEpochs {
			return errors.Errorf("retention.signature-epochs (%d) must be <= retention.valset-epochs (%d) to avoid orphaning signatures whose valset has been pruned",
				c.SignatureRetentionEpochs, c.ValsetRetentionEpochs)
		}
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
		"pruneBatchSize", s.cfg.PruneBatchSize,
		"valsetRetentionEpochs", s.cfg.ValsetRetentionEpochs,
		"proofRetentionEpochs", s.cfg.ProofRetentionEpochs,
		"signatureRetentionEpochs", s.cfg.SignatureRetentionEpochs,
	)

	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.RunOnce(ctx); err != nil {
				slog.ErrorContext(ctx, "Pruning failed", "error", err)
			}
		case <-ctx.Done():
			slog.InfoContext(ctx, "Pruner stopped")
			return
		}
	}
}

// RunOnce executes a single pruning pass over the repository: prunes valset, proof,
// signature and request-id-epoch entities according to the configured retention values.
// Safe to call without Start (e.g. from a one-shot CLI).
func (s *Service) RunOnce(ctx context.Context) error {
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

	// Each entity-type pass is independent: a failure of one does not invalidate
	// the others (e.g. proof pruning can succeed even if a single signature
	// epoch fails). Continue on error and join failures so callers (sidecar
	// loop logger, CLI exit code) can react. Context cancellation, however,
	// short-circuits the remaining passes.
	//
	// Order matters for crash safety: the valset pass MUST be last. Subsequent
	// runs use GetOldestValidatorSetEpoch as the lower bound for proof /
	// signature / requestIdEpochIndex loops; if valset gets pruned first and we
	// crash before the dependent passes finish, the next run's lower bound will
	// jump forward and the orphaned dependents will never be reclaimed.
	var errs []error
	var proofCount, signatureCount, indexCount, valsetCount uint64

	proofCount, err = s.pruneProofEntities(ctx, latestEpoch, oldestStoredEpoch)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to prune proof entities", "error", err)
		errs = append(errs, errors.Errorf("proof: %w", err))
	}

	if ctx.Err() == nil {
		signatureCount, err = s.pruneSignatureEntities(ctx, latestEpoch, oldestStoredEpoch)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to prune signature entities", "error", err)
			errs = append(errs, errors.Errorf("signature: %w", err))
		}
	}

	if ctx.Err() == nil {
		indexCount, err = s.pruneRequestIDEpochIndices(ctx, latestEpoch, oldestStoredEpoch)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to prune request ID epoch indices", "error", err)
			errs = append(errs, errors.Errorf("requestIdEpochIndex: %w", err))
		}
	}

	if ctx.Err() == nil {
		valsetCount, err = s.pruneValsetEntities(ctx, latestEpoch, oldestStoredEpoch)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to prune valset entities", "error", err)
			errs = append(errs, errors.Errorf("valset: %w", err))
		}
	}

	slog.InfoContext(ctx, "Pruning completed",
		"valsetEpochs", valsetCount,
		"proofEpochs", proofCount,
		"signatureEpochs", signatureCount,
		"indexCleanupEpochs", indexCount,
		"duration", time.Since(start),
	)

	return stderrors.Join(errs...)
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
	pruneFunc func(context.Context, symbiotic.Epoch, int) error,
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

	total := uint64(oldestToKeep - oldestStoredEpoch)
	if s.cfg.ProgressFn != nil {
		s.cfg.ProgressFn(entityType, 0, total)
	}

	count := uint64(0)
	for epoch := oldestStoredEpoch; epoch < oldestToKeep; epoch++ {
		if err := ctx.Err(); err != nil {
			return count, err
		}

		slog.DebugContext(ctx, "Pruning entities", "entityType", entityType, "epoch", epoch)

		if err := pruneFunc(ctx, epoch, s.cfg.PruneBatchSize); err != nil {
			tracing.RecordError(span, err)
			return count, errors.Errorf("failed to prune %s entities for epoch %d: %w", entityType, epoch, err)
		}

		count++
		s.cfg.Metrics.IncPrunedEpochsCount(entityType)
		if s.cfg.ProgressFn != nil {
			s.cfg.ProgressFn(entityType, count, total)
		}
	}

	tracing.SetAttributes(span, tracing.AttrEpochCount.Int(int(count)))

	return count, nil
}
