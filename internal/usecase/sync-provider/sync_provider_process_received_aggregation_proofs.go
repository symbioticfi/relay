package sync_provider

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

// ProcessReceivedAggregationProofs processes aggregation proofs received from peers
func (s *Syncer) ProcessReceivedAggregationProofs(ctx context.Context, response entity.WantAggregationProofsResponse) (entity.AggregationProofProcessingStats, error) {
	stats := entity.AggregationProofProcessingStats{}

	// Process each received aggregation proof
	for _, proof := range response.Proofs {
		s.processSingleAggregationProof(ctx, proof, &stats)
	}

	return stats, nil
}

// processSingleAggregationProof processes a single aggregation proof
func (s *Syncer) processSingleAggregationProof(ctx context.Context, proof entity.AggregationProof, stats *entity.AggregationProofProcessingStats) {
	// Process the aggregation proof using the signature processor
	if err := s.cfg.EntityProcessor.ProcessAggregationProof(ctx, proof); err != nil {
		if errors.Is(err, entity.ErrEntityAlreadyExist) {
			stats.AlreadyExistCount++
			return
		}
		stats.ProcessingErrorCount++
		return // Continue processing other proofs
	}

	stats.ProcessedCount++
}
