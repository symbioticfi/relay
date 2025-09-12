package sync_provider

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/crypto"
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
func (s *Syncer) ProcessReceivedSignatures(ctx context.Context, response entity.WantSignaturesResponse, wantSignatures map[common.Hash]entity.SignatureBitmap) entity.SignatureProcessingStats {
	var stats entity.SignatureProcessingStats

	for reqHash, signatures := range response.Signatures {
		for _, validatorSig := range signatures {
			// Validate that we actually requested this validator's signature
			requestedBitmap, exists := wantSignatures[reqHash]
			if !exists {
				slog.WarnContext(ctx, "Received signature for unrequested hash", "request_hash", reqHash.Hex())
				stats.UnrequestedHashCount++
				continue
			}

			if !requestedBitmap.Contains(validatorSig.ValidatorIndex) {
				slog.WarnContext(ctx, "Received unrequested signature",
					"request_hash", reqHash.Hex(),
					"validator_index", validatorSig.ValidatorIndex)
				stats.UnrequestedSignatureCount++
				continue
			}

			// Get the original signature request to extract epoch and other details
			sigReq, err := s.cfg.Repo.GetSignatureRequest(ctx, reqHash)
			if err != nil {
				slog.WarnContext(ctx, "Failed to get signature request for processing",
					"request_hash", reqHash.Hex(), "error", err)
				stats.SignatureRequestErrorCount++
				continue
			}

			publicKey, err := crypto.NewPublicKey(sigReq.KeyTag.Type(), validatorSig.Signature.PublicKey)
			if err != nil {
				slog.WarnContext(ctx, "Failed to create public key from signature",
					"request_hash", reqHash.Hex(),
					"validator_index", validatorSig.ValidatorIndex,
					"error", err)
				stats.PublicKeyErrorCount++
				continue
			}

			// Get validator info to extract voting power
			validatorInfo, index, err := s.cfg.Repo.GetValidatorByKey(
				ctx,
				uint64(sigReq.RequiredEpoch),
				sigReq.KeyTag,
				publicKey.OnChain(),
			)
			if err != nil {
				slog.WarnContext(ctx, "Failed to get validator info",
					"request_hash", reqHash.Hex(),
					"validator_index", validatorSig.ValidatorIndex,
					"error", err)
				stats.ValidatorInfoErrorCount++
				continue
			}

			if index != validatorSig.ValidatorIndex {
				slog.WarnContext(ctx, "Validator index mismatch",
					"request_hash", reqHash.Hex(),
					"expected_index", validatorSig.ValidatorIndex,
					"actual_index", index)
				stats.ValidatorIndexMismatchCount++
				continue
			}

			// Process the signature
			param := entity.SaveSignatureParam{
				RequestHash:      reqHash,
				Key:              validatorSig.Signature.PublicKey,
				Signature:        validatorSig.Signature,
				ActiveIndex:      validatorSig.ValidatorIndex,
				VotingPower:      validatorInfo.VotingPower,
				Epoch:            sigReq.RequiredEpoch,
				SignatureRequest: nil,
			}

			if err := s.cfg.SignatureProcessor.ProcessSignature(ctx, param); err != nil {
				if errors.Is(err, entity.ErrEntityAlreadyExist) {
					slog.DebugContext(ctx, "Signature already exists",
						"request_hash", reqHash.Hex(),
						"validator_index", validatorSig.ValidatorIndex)
					stats.AlreadyExistCount++
				} else {
					slog.WarnContext(ctx, "Failed to process received signature",
						"request_hash", reqHash.Hex(),
						"validator_index", validatorSig.ValidatorIndex,
						"error", err)
					stats.ProcessingErrorCount++
				}
				continue
			}

			err = s.cfg.SignatureReceivedSignal.Emit(entity.SignatureMessage{
				RequestHash: reqHash,
				KeyTag:      sigReq.KeyTag,
				Epoch:       sigReq.RequiredEpoch,
				Signature:   validatorSig.Signature,
			})
			if err != nil {
				slog.WarnContext(ctx, "Failed to emit signature received signal", "error", err)
			}

			slog.DebugContext(ctx, "Processed received signature",
				"request_hash", reqHash.Hex(),
				"epoch", uint64(sigReq.RequiredEpoch),
			)
			stats.ProcessedCount++
		}
	}

	return stats
}
