package sync_provider

import (
	"context"

	"github.com/go-errors/errors"
	"go.opentelemetry.io/otel/attribute"

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/tracing"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// ProcessReceivedAggregationProofs processes aggregation proofs received from peers
func (s *Syncer) ProcessReceivedAggregationProofs(ctx context.Context, response entity.WantAggregationProofsResponse) (entity.AggregationProofProcessingStats, error) {
	ctx, span := tracing.StartSpan(ctx, "sync-provider.ProcessReceivedAggregationProofs",
		attribute.Int("request.proofs_count", len(response.Proofs)),
	)
	defer span.End()

	stats := entity.AggregationProofProcessingStats{}

	// Process each received aggregation proof
	for _, proof := range response.Proofs {
		s.processSingleAggregationProof(ctx, proof, &stats)
	}

	tracing.SetAttributes(span,
		attribute.Int("response.processed_count", stats.ProcessedCount),
		attribute.Int("response.already_exist_count", stats.AlreadyExistCount),
		attribute.Int("response.processing_fail_count", stats.ProcessingFailCount),
	)

	return stats, nil
}

// processSingleAggregationProof processes a single aggregation proof
func (s *Syncer) processSingleAggregationProof(ctx context.Context, proof symbiotic.AggregationProof, stats *entity.AggregationProofProcessingStats) {
	// Process the aggregation proof using the signature processor
	if err := s.cfg.EntityProcessor.ProcessAggregationProof(ctx, proof); err != nil {
		if errors.Is(err, entity.ErrEntityAlreadyExist) {
			stats.AlreadyExistCount++
			return
		}
		stats.ProcessingFailCount++
		return // Continue processing other proofs
	}

	stats.ProcessedCount++
}
