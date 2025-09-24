package sync_provider

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

// HandleWantSignaturesRequest processes a peer's request for missing signatures and returns
// the requested signatures that are available in local storage.
//
// The method performs the following operations:
// 1. Iterates through each signature request hash in the incoming request
// 2. For each requested validator index, directly retrieves the signature using GetSignatureByIndex
// 3. Builds a response containing validator signatures organized by request hash
//
// The response is limited by MaxResponseSignatureCount to prevent memory exhaustion
// and network congestion during P2P synchronization.
//
// Behavior:
//   - Processes requests in iteration order (map iteration is non-deterministic)
//   - Stops processing when MaxResponseSignatureCount limit is reached
//   - Skips validator indices where signatures are not found locally
//   - Returns empty signatures map for request hashes where no matching signatures are found
func (s *Syncer) HandleWantSignaturesRequest(ctx context.Context, request entity.WantSignaturesRequest) (entity.WantSignaturesResponse, error) {
	slog.InfoContext(ctx, "Handling want signatures request", "request_count", len(request.WantSignatures))

	response := entity.WantSignaturesResponse{
		Signatures: make(map[common.Hash][]entity.ValidatorSignature),
	}

	totalSignatureCount := 0

	for signatureTargetID, requestedIndices := range request.WantSignatures {
		// Check signature count limit before processing each request
		if totalSignatureCount >= s.cfg.MaxResponseSignatureCount {
			return entity.WantSignaturesResponse{}, errors.Errorf("response signature limit exceeded")
		}

		var validatorSigs []entity.ValidatorSignature

		// Iterate over requested validator indices and get signatures directly
		it := requestedIndices.Iterator()
		for it.HasNext() {
			validatorIndex := it.Next()
			// Check limit before processing each signature
			if totalSignatureCount >= s.cfg.MaxResponseSignatureCount {
				return entity.WantSignaturesResponse{}, errors.Errorf("response signature limit exceeded")
			}

			// Get signature by validator index directly
			sig, err := s.cfg.Repo.GetSignatureByIndex(ctx, signatureTargetID, validatorIndex)
			if err != nil {
				if errors.Is(err, entity.ErrEntityNotFound) {
					// Signature not found for this validator index, skip
					continue
				}
				return entity.WantSignaturesResponse{}, errors.Errorf("failed to get signature for validator %d in request %s: %w", validatorIndex, signatureTargetID.Hex(), err)
			}

			validatorSigs = append(validatorSigs, entity.ValidatorSignature{
				ValidatorIndex: validatorIndex,
				Signature:      sig,
			})
			totalSignatureCount++
		}

		if len(validatorSigs) > 0 {
			response.Signatures[signatureTargetID] = validatorSigs
		}
	}

	slog.InfoContext(ctx, "Want signatures request handled", "response_signatures", totalSignatureCount, "response_requests", len(response.Signatures))

	return response, nil
}
