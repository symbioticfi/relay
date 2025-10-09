package sync_provider

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// BuildWantAggregationProofsRequest builds a request for missing aggregation proofs from recent epochs
func (s *Syncer) BuildWantAggregationProofsRequest(ctx context.Context) (entity.WantAggregationProofsRequest, error) {
	// Get the latest epoch
	latestEpoch, err := s.cfg.Repo.GetLatestValidatorSetEpoch(ctx)
	if err != nil {
		return entity.WantAggregationProofsRequest{}, errors.Errorf("failed to get latest epoch: %w", err)
	}

	startEpoch := symbiotic.Epoch(0)
	if latestEpoch >= symbiotic.Epoch(s.cfg.EpochsToSync) {
		startEpoch = latestEpoch - symbiotic.Epoch(s.cfg.EpochsToSync)
	}

	var allRequestIDs []common.Hash
	totalRequests := 0

	// Iterate through epochs from newest to oldest to prioritize recent requests
	for epoch := latestEpoch; epoch >= startEpoch && totalRequests < s.cfg.MaxAggProofRequestsPerSync; epoch-- {
		var lastHash common.Hash
		remaining := s.cfg.MaxAggProofRequestsPerSync - totalRequests

		// Paginate through signature requests without aggregation proofs for this epoch
		for remaining > 0 {
			requests, err := s.cfg.Repo.GetSignatureRequestsWithoutAggregationProof(ctx, epoch, remaining, lastHash)
			if err != nil {
				return entity.WantAggregationProofsRequest{}, errors.Errorf("failed to get signature requests without aggregation proof for epoch %d: %w", epoch, err)
			}

			if len(requests) == 0 {
				break // No more requests for this epoch
			}

			// Collect request ids
			for _, req := range requests {
				if !req.KeyTag.Type().AggregationKey() {
					continue // Skip non-aggregation requests
				}
				// check if proof exists
				_, err := s.cfg.Repo.GetAggregationProof(ctx, req.RequestID)
				if err == nil {
					// remove pending from db
					err = s.cfg.Repo.RemoveAggregationProofPending(ctx, req.RequiredEpoch, req.RequestID)
					// ignore not found and tx conflict errors, as they indicate the proof was already processed or is being processed
					if err != nil && !errors.Is(err, entity.ErrEntityNotFound) && !errors.Is(err, entity.ErrTxConflict) {
						return entity.WantAggregationProofsRequest{}, errors.Errorf("failed to remove aggregation proof from pending collection: %w", err)
					}
					continue // Proof already exists, skip
				}
				allRequestIDs = append(allRequestIDs, req.RequestID)
				totalRequests++
				lastHash = req.RequestID // Update for pagination
			}

			remaining = s.cfg.MaxAggProofRequestsPerSync - totalRequests
		}

		// Handle epoch == 0 to avoid underflow in unsigned arithmetic
		if epoch == 0 {
			break
		}
	}

	return entity.WantAggregationProofsRequest{
		RequestIDs: allRequestIDs,
	}, nil
}
