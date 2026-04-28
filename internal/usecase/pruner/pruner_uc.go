package pruner

import (
	"context"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
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
	GetRequestIDsByEpoch(ctx context.Context, epoch symbiotic.Epoch, limit int) ([]common.Hash, error)
	PruneValsetEntities(ctx context.Context, epoch symbiotic.Epoch, batchSize int) error
	PruneProofCommits(ctx context.Context, epoch symbiotic.Epoch) error
	PruneProofsByRequestIDs(ctx context.Context, epoch symbiotic.Epoch, requestIDs []common.Hash, batchSize int) error
	PruneSignaturesByRequestIDs(ctx context.Context, epoch symbiotic.Epoch, requestIDs []common.Hash, batchSize int) error
	PruneEpochIndicesByRequestIDs(ctx context.Context, epoch symbiotic.Epoch, requestIDs []common.Hash, batchSize int) error
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

	slog.InfoContext(ctx, "Pruning tick started")

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

	s.pruneEntityType(ctx, latestEpoch, oldestStoredEpoch, s.cfg.ProofRetentionEpochs, "proof", s.pruneProofBatch)
	s.pruneEntityType(ctx, latestEpoch, oldestStoredEpoch, s.cfg.SignatureRetentionEpochs, "signature", s.pruneSignatureBatch)

	maxRetention := max(s.cfg.SignatureRetentionEpochs, s.cfg.ProofRetentionEpochs)
	s.pruneEntityType(ctx, latestEpoch, oldestStoredEpoch, maxRetention, "requestIdEpochIndex", s.pruneIndexBatch)

	// Valset must be pruned last — GetOldestValidatorSetEpoch drives epoch iteration for all entity types,
	// so deleting a valset before its request data is fully pruned would skip the remaining work.
	s.pruneEntityType(ctx, latestEpoch, oldestStoredEpoch, s.cfg.ValsetRetentionEpochs, "valset", s.pruneValsetEpoch)

	slog.InfoContext(ctx, "Pruning tick completed",
		"latestEpoch", latestEpoch,
		"oldestStoredEpoch", oldestStoredEpoch,
		"duration", time.Since(start),
	)

	return nil
}

// pruneEntityType finds the oldest epoch outside the retention window and calls batchFunc.
// Stateless: completed epochs have their data removed, so getRequestIDsByEpoch returns fewer results next tick.
func (s *Service) pruneEntityType(
	ctx context.Context,
	latestEpoch, oldestStoredEpoch symbiotic.Epoch,
	retentionEpochs uint64,
	entityType string,
	batchFunc func(ctx context.Context, epoch symbiotic.Epoch) (done bool, err error),
) {
	if retentionEpochs == 0 {
		return
	}

	retentionWindow := symbiotic.Epoch(retentionEpochs)
	if latestEpoch < retentionWindow {
		return
	}

	oldestToKeep := latestEpoch - retentionWindow + 1
	if oldestStoredEpoch >= oldestToKeep {
		return
	}

	for epoch := oldestStoredEpoch; epoch < oldestToKeep; epoch++ {
		done, err := batchFunc(ctx, epoch)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to prune entities",
				"entityType", entityType, "epoch", epoch, "error", err)
			return
		}

		if done {
			slog.InfoContext(ctx, "Epoch pruned",
				"entityType", entityType, "epoch", epoch)
			s.cfg.Metrics.IncPrunedEpochsCount(entityType)
			continue
		}

		slog.InfoContext(ctx, "Epoch partially pruned, continuing next tick",
			"entityType", entityType, "epoch", epoch)
		return
	}
}

func (s *Service) pruneProofBatch(ctx context.Context, epoch symbiotic.Epoch) (bool, error) {
	if err := s.cfg.Repo.PruneProofCommits(ctx, epoch); err != nil {
		return false, errors.Errorf("failed to prune proof commits: %w", err)
	}

	requestIDs, err := s.cfg.Repo.GetRequestIDsByEpoch(ctx, epoch, s.cfg.PruneBatchSize)
	if err != nil {
		return false, errors.Errorf("failed to get request IDs: %w", err)
	}

	if len(requestIDs) == 0 {
		return true, nil
	}

	if err := s.cfg.Repo.PruneProofsByRequestIDs(ctx, epoch, requestIDs, s.cfg.PruneBatchSize); err != nil {
		return false, errors.Errorf("failed to prune proofs: %w", err)
	}

	slog.InfoContext(ctx, "Pruned proof batch",
		"epoch", epoch, "requestIds", len(requestIDs))

	return false, nil
}

func (s *Service) pruneSignatureBatch(ctx context.Context, epoch symbiotic.Epoch) (bool, error) {
	requestIDs, err := s.cfg.Repo.GetRequestIDsByEpoch(ctx, epoch, s.cfg.PruneBatchSize)
	if err != nil {
		return false, errors.Errorf("failed to get request IDs: %w", err)
	}

	if len(requestIDs) == 0 {
		return true, nil
	}

	if err := s.cfg.Repo.PruneSignaturesByRequestIDs(ctx, epoch, requestIDs, s.cfg.PruneBatchSize); err != nil {
		return false, errors.Errorf("failed to prune signatures: %w", err)
	}

	slog.InfoContext(ctx, "Pruned signature batch",
		"epoch", epoch, "requestIds", len(requestIDs))

	return false, nil
}

func (s *Service) pruneIndexBatch(ctx context.Context, epoch symbiotic.Epoch) (bool, error) {
	requestIDs, err := s.cfg.Repo.GetRequestIDsByEpoch(ctx, epoch, s.cfg.PruneBatchSize)
	if err != nil {
		return false, errors.Errorf("failed to get request IDs: %w", err)
	}

	if len(requestIDs) == 0 {
		return true, nil
	}

	if err := s.cfg.Repo.PruneEpochIndicesByRequestIDs(ctx, epoch, requestIDs, s.cfg.PruneBatchSize); err != nil {
		return false, errors.Errorf("failed to prune epoch indices: %w", err)
	}

	slog.InfoContext(ctx, "Pruned index batch",
		"epoch", epoch, "requestIds", len(requestIDs))

	return false, nil
}

func (s *Service) pruneValsetEpoch(ctx context.Context, epoch symbiotic.Epoch) (bool, error) {
	requestIDs, err := s.cfg.Repo.GetRequestIDsByEpoch(ctx, epoch, 1)
	if err != nil {
		return false, errors.Errorf("failed to check remaining request IDs: %w", err)
	}

	if len(requestIDs) > 0 {
		slog.InfoContext(ctx, "Valset pruning deferred, request data still exists", "epoch", epoch)
		return false, nil
	}

	if err := s.cfg.Repo.PruneValsetEntities(ctx, epoch, 0); err != nil {
		return false, errors.Errorf("failed to prune valset entities: %w", err)
	}
	return true, nil
}
