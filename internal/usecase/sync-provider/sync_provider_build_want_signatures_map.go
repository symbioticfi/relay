package sync_provider

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

func (s *Syncer) BuildWantSignaturesRequest(ctx context.Context) (entity.WantSignaturesRequest, error) {
	// Collect all pending signature requests across epochs
	wantSignatures, err := s.buildWantSignaturesMap(ctx)
	if err != nil {
		return entity.WantSignaturesRequest{}, errors.Errorf("failed to build want signatures map: %w", err)
	}

	return entity.WantSignaturesRequest{
		WantSignatures: wantSignatures,
	}, nil
}

// buildWantSignaturesMap constructs a map of signature request hashes to missing validator bitmaps
// for pending signature requests across multiple epochs.
//
// The method performs the following operations:
// 1. Determines the epoch range to scan (from latest epoch back to EpochsToSync epochs)
// 2. Iterates through epochs from newest to oldest to prioritize recent requests
// 3. For each epoch, fetches pending signature requests in batches
// 4. For each request, identifies validators that haven't provided signatures yet
// 5. Builds a map where keys are request hashes and values are bitmaps of missing validators
//
// The scanning is limited by MaxSignatureRequestsPerSync to prevent excessive memory usage
// and network overhead during synchronization.
//
// Behavior:
//   - Scans epochs in reverse order (newest first) to prioritize recent requests
//   - Stops scanning when MaxSignatureRequestsPerSync limit is reached
//   - Only includes requests that have missing signatures (non-empty bitmaps)
//   - Uses pagination to handle large numbers of requests per epoch
func (s *Syncer) buildWantSignaturesMap(ctx context.Context) (map[common.Hash]entity.Bitmap, error) {
	// Get the latest epoch
	latestEpoch, err := s.cfg.Repo.GetLatestValidatorSetEpoch(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to get latest epoch: %w", err)
	}

	// Calculate the starting epoch (go back EpochsToSync epochs)
	var startEpoch uint64
	if latestEpoch >= s.cfg.EpochsToSync {
		startEpoch = latestEpoch - s.cfg.EpochsToSync
	} else {
		startEpoch = 0
	}

	wantSignatures := make(map[common.Hash]entity.Bitmap)
	totalRequests := 0

	for epoch := latestEpoch; epoch >= startEpoch && totalRequests < s.cfg.MaxSignatureRequestsPerSync; epoch-- {
		var lastHash common.Hash
		remaining := s.cfg.MaxSignatureRequestsPerSync - totalRequests

		for remaining > 0 {
			requests, err := s.cfg.Repo.GetSignatureRequestsByEpochPending(ctx, entity.Epoch(epoch), remaining, lastHash)
			if err != nil {
				return nil, errors.Errorf("failed to get pending signature requests for epoch %d: %w", epoch, err)
			}

			if len(requests) == 0 {
				break
			}

			// Process each request to find missing signatures
			for _, req := range requests {
				reqSignatureID := req.SignatureTargetID

				// Get current signature map
				sigMap, err := s.cfg.Repo.GetSignatureMap(ctx, reqSignatureID)
				if err != nil {
					return nil, errors.Errorf("failed to get signature map for request %s: %w", reqSignatureID.Hex(), err)
				}

				// Get missing validators from signature map
				missingValidators := sigMap.GetMissingValidators()
				if !missingValidators.IsEmpty() {
					wantSignatures[reqSignatureID] = missingValidators
				}

				lastHash = reqSignatureID
			}

			totalRequests += len(requests)
			remaining = s.cfg.MaxSignatureRequestsPerSync - totalRequests

			// If we got fewer requests than requested, we've reached the end for this epoch
			if len(requests) < remaining {
				break
			}
		}

		if epoch == 0 {
			break
		}
	}

	return wantSignatures, nil
}
