package syncer

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/crypto"
)

func (s *Syncer) processReceivedSignatures(ctx context.Context, response entity.WantSignaturesResponse, wantSignatures map[common.Hash]entity.SignatureBitmap) SignatureProcessingStats {
	var stats SignatureProcessingStats

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
			validatorInfo, _, err := s.cfg.Repo.GetValidatorByKey(
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

			stats.ProcessedCount++
		}
	}

	return stats
}
