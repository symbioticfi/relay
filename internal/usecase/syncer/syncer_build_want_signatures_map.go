package syncer

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

func (s *Syncer) askSignatures(ctx context.Context) error {
	// Create context with timeout for the entire sync operation
	syncCtx, cancel := context.WithTimeout(ctx, s.cfg.SyncTimeout)
	defer cancel()

	slog.InfoContext(syncCtx, "Starting signature sync")

	// Collect all pending signature requests across epochs
	wantSignatures, err := s.buildWantSignaturesMap(syncCtx)
	if err != nil {
		return errors.Errorf("failed to build want signatures map: %w", err)
	}

	// If no signatures needed, log and return
	if len(wantSignatures) == 0 {
		slog.InfoContext(syncCtx, "No pending signature requests found")
		return nil
	}

	slog.InfoContext(syncCtx, "Found pending signature requests", "count", len(wantSignatures))

	// Send request to peer
	request := entity.WantSignaturesRequest{
		WantSignatures: wantSignatures,
	}

	response, err := s.cfg.P2PService.SendWantSignaturesRequest(syncCtx, request)
	if err != nil {
		return errors.Errorf("failed to send want signatures request: %w", err)
	}

	slog.InfoContext(syncCtx, "Received signature response", "signatures_count", len(response.Signatures))

	stats := s.processReceivedSignatures(syncCtx, response, wantSignatures)

	slog.InfoContext(syncCtx, "Signature sync completed",
		"processed", stats.ProcessedCount,
		"total_errors", stats.TotalErrors(),
		"unrequested_signatures", stats.UnrequestedSignatureCount,
		"unrequested_hashes", stats.UnrequestedHashCount,
		"signature_request_errors", stats.SignatureRequestErrorCount,
		"public_key_errors", stats.PublicKeyErrorCount,
		"validator_info_errors", stats.ValidatorInfoErrorCount,
		"processing_errors", stats.ProcessingErrorCount,
		"already_exist", stats.AlreadyExistCount,
	)

	return nil
}

func (s *Syncer) buildWantSignaturesMap(ctx context.Context) (map[common.Hash]entity.SignatureBitmap, error) {
	// Get the latest epoch
	latestEpoch, err := s.cfg.Repo.GetLatestValidatorSetEpoch(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to get latest epoch: %w", err)
	}

	// Calculate the starting epoch (go back EpochsToSync epochs)
	var startEpoch uint64
	if latestEpoch >= uint64(s.cfg.EpochsToSync) {
		startEpoch = latestEpoch - uint64(s.cfg.EpochsToSync)
	} else {
		startEpoch = 0
	}

	wantSignatures := make(map[common.Hash]entity.SignatureBitmap)
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
				reqHash := req.Hash()

				// Get current signature map
				sigMap, err := s.cfg.Repo.GetSignatureMap(ctx, reqHash)
				if err != nil {
					return nil, errors.Errorf("failed to get signature map for request %s: %w", reqHash.Hex(), err)
				}

				// Get missing validators from signature map
				missingValidators := sigMap.GetMissingValidators()
				if !missingValidators.IsEmpty() {
					wantSignatures[reqHash] = missingValidators
				}

				lastHash = reqHash
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
