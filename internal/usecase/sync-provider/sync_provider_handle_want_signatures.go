package sync_provider

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/crypto"
)

// HandleWantSignaturesRequest processes a peer's request for missing signatures and returns
// the requested signatures that are available in local storage.
//
// The method performs the following operations:
// 1. Iterates through each signature request hash in the incoming request
// 2. For each request, retrieves all locally stored signatures from the repository
// 3. Filters signatures to only include those requested by the peer (based on validator indices)
// 4. Maps validator public keys to their active indices in the validator set
// 5. Builds a response containing validator signatures organized by request hash
//
// The response is limited by MaxResponseSignatureCount to prevent memory exhaustion
// and network congestion during P2P synchronization.
//
// Behavior:
//   - Processes requests in iteration order (map iteration is non-deterministic)
//   - Stops processing when MaxResponseSignatureCount limit is reached
//   - Skips signature requests that are not found in local storage
//   - Only includes signatures for validator indices specifically requested by the peer
//   - Returns empty signatures map for request hashes where no matching signatures are found
func (s *Syncer) HandleWantSignaturesRequest(ctx context.Context, request entity.WantSignaturesRequest) (entity.WantSignaturesResponse, error) {
	slog.InfoContext(ctx, "Handling want signatures request", "request_count", len(request.WantSignatures))

	response := entity.WantSignaturesResponse{
		Signatures: make(map[common.Hash][]entity.ValidatorSignature),
	}

	totalSignatureCount := 0

	for reqHash, requestedIndices := range request.WantSignatures {
		// Check signature count limit before processing each request
		if totalSignatureCount >= s.cfg.MaxResponseSignatureCount {
			return entity.WantSignaturesResponse{}, errors.Errorf("response signature limit exceeded")
		}

		// Get stored signatures for this request
		signatures, err := s.cfg.Repo.GetAllSignatures(ctx, reqHash)
		if err != nil {
			return entity.WantSignaturesResponse{}, errors.Errorf("failed to get signatures for request %s: %w", reqHash.Hex(), err)
		}

		// Get signature request for epoch info
		sigReq, err := s.cfg.Repo.GetSignatureRequest(ctx, reqHash)
		if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
			return entity.WantSignaturesResponse{}, errors.Errorf("failed to get signature request %s: %w", reqHash.Hex(), err)
		}
		if errors.Is(err, entity.ErrEntityNotFound) {
			// If we don't have the signature request, skip processing this request
			slog.DebugContext(ctx, "Signature request not found, skipping", "request_hash", reqHash.Hex())
			continue
		}

		var validatorSigs []entity.ValidatorSignature
		for _, sig := range signatures {
			// Check limit before processing each signature
			if totalSignatureCount >= s.cfg.MaxResponseSignatureCount {
				return entity.WantSignaturesResponse{}, errors.Errorf("response signature limit exceeded")
			}

			publicKey, err := crypto.NewPublicKey(sigReq.KeyTag.Type(), sig.PublicKey)
			if err != nil {
				return entity.WantSignaturesResponse{}, errors.Errorf("failed to get public key: %w", err)
			}

			// Map public key to validator index
			_, activeIndex, err := s.cfg.Repo.GetValidatorByKey(
				ctx,
				uint64(sigReq.RequiredEpoch),
				sigReq.KeyTag,
				publicKey.OnChain(),
			)
			if err != nil {
				return entity.WantSignaturesResponse{}, errors.Errorf("failed to get validator for key: %w", err)
			}

			// Only include if requested
			if requestedIndices.Contains(activeIndex) {
				validatorSigs = append(validatorSigs, entity.ValidatorSignature{
					ValidatorIndex: activeIndex,
					Signature:      sig,
				})
				totalSignatureCount++
			}
		}

		if len(validatorSigs) > 0 {
			response.Signatures[reqHash] = validatorSigs
		}
	}

	slog.InfoContext(ctx, "Want signatures request handled", "response_signatures", totalSignatureCount, "response_requests", len(response.Signatures))

	return response, nil
}
