package sync_provider

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

// HandleWantAggregationProofsRequest handles incoming requests for aggregation proofs from peers
func (s *Syncer) HandleWantAggregationProofsRequest(ctx context.Context, request entity.WantAggregationProofsRequest) (entity.WantAggregationProofsResponse, error) {
	proofs := make(map[common.Hash]entity.AggregationProof)
	responseCount := 0

	// Process each requested hash
	for _, reqHash := range request.RequestHashes {
		// Stop if we've reached the maximum response count
		if responseCount >= s.cfg.MaxResponseAggProofCount {
			break
		}

		// Try to get the aggregation proof
		proof, err := s.cfg.Repo.GetAggregationProof(ctx, reqHash)
		if err != nil {
			if errors.Is(err, entity.ErrEntityNotFound) {
				// Aggregation proof not found, skip this request
				continue
			}
			return entity.WantAggregationProofsResponse{}, errors.Errorf("failed to get aggregation proof for hash %s: %w", reqHash.Hex(), err)
		}

		// Add the proof to the response
		proofs[reqHash] = proof
		responseCount++
	}

	return entity.WantAggregationProofsResponse{
		Proofs: proofs,
	}, nil
}
