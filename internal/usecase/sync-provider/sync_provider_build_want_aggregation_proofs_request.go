package sync_provider

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

// BuildWantAggregationProofsRequest builds a request for missing aggregation proofs from recent epochs
func (s *Syncer) BuildWantAggregationProofsRequest(ctx context.Context) (entity.WantAggregationProofsRequest, error) {
	// Get the latest epoch
	latestEpoch, err := s.cfg.Repo.GetLatestValidatorSetEpoch(ctx)
	if err != nil {
		return entity.WantAggregationProofsRequest{}, errors.Errorf("failed to get latest epoch: %w", err)
	}

	startEpoch := uint64(0)
	if latestEpoch >= s.cfg.EpochsToSync {
		startEpoch = latestEpoch - s.cfg.EpochsToSync
	}

	var allSignatureTargetIDs []common.Hash
	totalRequests := 0

	// Iterate through epochs from newest to oldest to prioritize recent requests
	for epoch := latestEpoch; epoch >= startEpoch && totalRequests < s.cfg.MaxAggProofRequestsPerSync; epoch-- {
		var lastHash common.Hash
		remaining := s.cfg.MaxAggProofRequestsPerSync - totalRequests

		// Paginate through signature requests without aggregation proofs for this epoch
		for remaining > 0 {
			requests, err := s.cfg.Repo.GetSignatureRequestsWithoutAggregationProof(ctx, entity.Epoch(epoch), remaining, lastHash)
			if err != nil {
				return entity.WantAggregationProofsRequest{}, errors.Errorf("failed to get signature requests without aggregation proof for epoch %d: %w", epoch, err)
			}

			if len(requests) == 0 {
				break // No more requests for this epoch
			}

			// Collect request hashes
			for _, req := range requests {
				allSignatureTargetIDs = append(allSignatureTargetIDs, req.SignatureTargetID)
				totalRequests++
				lastHash = req.SignatureTargetID // Update for pagination
			}

			remaining = s.cfg.MaxAggProofRequestsPerSync - totalRequests
		}

		// Handle epoch == 0 to avoid underflow in unsigned arithmetic
		if epoch == 0 {
			break
		}
	}

	return entity.WantAggregationProofsRequest{
		SignatureTargetIDs: allSignatureTargetIDs,
	}, nil
}
