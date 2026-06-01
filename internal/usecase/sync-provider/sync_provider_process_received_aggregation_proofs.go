package sync_provider

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"go.opentelemetry.io/otel/attribute"

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/tracing"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// ProcessReceivedAggregationProofs processes aggregation proofs received from peers,
// filtering out any proofs that were not explicitly requested.
func (s *Syncer) ProcessReceivedAggregationProofs(ctx context.Context, response entity.WantAggregationProofsResponse, requestedIDs []common.Hash) (entity.AggregationProofProcessingStats, error) {
	ctx, span := tracing.StartSpan(ctx, "sync-provider.ProcessReceivedAggregationProofs",
		attribute.Int("request.proofs_count", len(response.Proofs)),
		attribute.Int("request.requested_ids_count", len(requestedIDs)),
	)
	defer span.End()

	stats := entity.AggregationProofProcessingStats{}

	requestedSet := make(map[common.Hash]struct{}, len(requestedIDs))
	for _, id := range requestedIDs {
		requestedSet[id] = struct{}{}
	}

	for requestID, proof := range response.Proofs {
		if _, ok := requestedSet[requestID]; !ok {
			stats.UnrequestedProofCount++
			continue
		}
		proofRequestID := proof.RequestID()
		if proofRequestID != requestID {
			slog.WarnContext(ctx, "Received aggregation proof with mismatched request ID",
				"requestId", requestID.Hex(),
				"proofRequestId", proofRequestID.Hex())
			stats.UnrequestedProofCount++
			continue
		}
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
	if err := s.cfg.EntityProcessor.ProcessAggregationProof(ctx, proof, false); err != nil {
		if errors.Is(err, entity.ErrEntityAlreadyExist) {
			stats.AlreadyExistCount++
			return
		}
		stats.ProcessingFailCount++
		return // Continue processing other proofs
	}

	stats.ProcessedCount++
}
