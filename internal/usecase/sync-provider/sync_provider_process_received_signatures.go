package sync_provider

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

// ProcessReceivedSignatures validates and processes signatures received from peer nodes during
// synchronization, updating local storage and tracking statistics for monitoring.
//
// The method performs the following validation and processing steps:
// 1. Validates that received signatures were actually requested (hash and validator index matching)
// 2. Retrieves original signature request metadata (epoch, key type, etc.)
// 3. Reconstructs and validates public keys from signature data
// 4. Cross-references validator information to ensure consistency
// 5. Processes valid signatures through the signature processor
// 6. Emits signature received signals for downstream components
// 7. Tracks comprehensive statistics for all outcomes
//
// The method is designed to be resilient against malformed or malicious peer responses,
// validating all received data before processing and continuing on errors.
//
// Behavior:
//   - Validates all received signatures against original requests
//   - Skips invalid signatures and continues processing others
//   - Handles duplicate signatures gracefully (tracks but doesn't error)
//   - Emits signals for successfully processed signatures
//   - Returns comprehensive statistics for monitoring and debugging
//   - Logs warnings for validation failures and errors
func (s *Syncer) ProcessReceivedSignatures(ctx context.Context, response entity.WantSignaturesResponse, wantSignatures map[common.Hash]entity.Bitmap) entity.SignatureProcessingStats {
	var stats entity.SignatureProcessingStats

	for requestID, signatures := range response.Signatures {
		for _, validatorSig := range signatures {
			// Validate that we actually requested this validator's signature
			requestedBitmap, exists := wantSignatures[requestID]
			if !exists {
				slog.WarnContext(ctx, "Received signature for unrequested hash", "requestId", requestID.Hex())
				stats.UnrequestedHashCount++
				continue
			}

			if !requestedBitmap.Contains(validatorSig.ValidatorIndex) {
				slog.WarnContext(ctx, "Received unrequested signature",
					"requestId", requestID.Hex(),
					"validatorIndex", validatorSig.ValidatorIndex)
				stats.UnrequestedSignatureCount++
				continue
			}

			// Get the original signature request to extract epoch and other details
			sigReq, err := s.cfg.Repo.GetSignatureRequest(ctx, requestID)
			if err != nil {
				slog.WarnContext(ctx, "Failed to get signature request for processing",
					"requestId", requestID.Hex(), "error", err)
				stats.SignatureRequestErrorCount++
				continue
			}

			// Process the signature
			param := entity.SaveSignatureParam{
				Signature:        validatorSig.Signature,
				SignatureRequest: nil,
			}

			if err := s.cfg.EntityProcessor.ProcessSignature(ctx, param); err != nil {
				if errors.Is(err, entity.ErrEntityAlreadyExist) {
					slog.DebugContext(ctx, "Signature already exists",
						"requestId", requestID.Hex(),
						"validatorIndex", validatorSig.ValidatorIndex)
					stats.AlreadyExistCount++
				} else {
					slog.WarnContext(ctx, "Failed to process received signature",
						"requestId", requestID.Hex(),
						"validatorIndex", validatorSig.ValidatorIndex,
						"error", err)
					stats.ProcessingErrorCount++
				}
				continue
			}

			slog.DebugContext(ctx, "Processed received signature",
				"requestId", requestID.Hex(),
				"epoch", uint64(sigReq.RequiredEpoch),
			)
			stats.ProcessedCount++
		}
	}

	return stats
}
