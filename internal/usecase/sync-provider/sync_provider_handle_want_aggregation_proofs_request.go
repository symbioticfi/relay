package sync_provider

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"go.opentelemetry.io/otel/attribute"

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/tracing"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// HandleWantAggregationProofsRequest handles incoming requests for aggregation proofs from peers
func (s *Syncer) HandleWantAggregationProofsRequest(ctx context.Context, request entity.WantAggregationProofsRequest) (entity.WantAggregationProofsResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "sync-provider.HandleWantAggregationProofsRequest",
		attribute.Int("request.request_ids_count", len(request.RequestIDs)),
	)
	defer span.End()

	proofs := make(map[common.Hash]symbiotic.AggregationProof)
	responseCount := 0

	// Process each requested hash
	for _, requestID := range request.RequestIDs {
		// Stop if we've reached the maximum response count
		if responseCount >= s.cfg.MaxResponseAggProofCount {
			break
		}

		// Try to get the aggregation proof
		proof, err := s.cfg.Repo.GetAggregationProof(ctx, requestID)
		if err != nil {
			if errors.Is(err, entity.ErrEntityNotFound) {
				// Aggregation proof not found, skip this request
				continue
			}
			return entity.WantAggregationProofsResponse{}, errors.Errorf("failed to get aggregation proof for hash %s: %w", requestID.Hex(), err)
		}

		// Add the proof to the response
		proofs[requestID] = proof
		responseCount++
	}

	tracing.SetAttributes(span, attribute.Int("response.proofs_count", len(proofs)))

	return entity.WantAggregationProofsResponse{
		Proofs: proofs,
	}, nil
}
