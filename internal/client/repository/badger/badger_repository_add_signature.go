package badger

import (
	"context"
	"log/slog"
	"sync"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/crypto"
)

func (r *Repository) AddSignature(ctx context.Context, signature entity.SignatureExtended) error {
	publicKey, err := crypto.NewPublicKey(signature.KeyTag.Type(), signature.PublicKey)
	if err != nil {
		return errors.Errorf("failed to get public key: %w", err)
	}
	err = publicKey.VerifyWithHash(signature.MessageHash, signature.Signature)
	if err != nil {
		return errors.Errorf("failed to verify signature: %w", err)
	}

	validator, activeIndex, err := r.GetValidatorByKey(ctx, signature.Epoch, signature.KeyTag, publicKey.OnChain())
	if err != nil {
		return errors.Errorf("validator not found for public key %x: %w", signature.PublicKey, err)
	}

	if !validator.IsActive {
		return errors.Errorf("validator %s is not active", validator.Operator.Hex())
	}

	slog.DebugContext(ctx, "Found active validator", "validator", validator)

	// Ensure only one goroutine is processing signatures for this request ID at a time
	r.signatureMutexMu.Lock()
	if _, exists := r.signatureMutexMap[signature.RequestID()]; !exists {
		r.signatureMutexMap[signature.RequestID()] = &sync.Mutex{}
	}
	activeMutex := r.signatureMutexMap[signature.RequestID()]
	r.signatureMutexMu.Unlock()

	activeMutex.Lock()
	defer activeMutex.Unlock()

	return r.doUpdateInTx(ctx, "ProcessSignature", func(ctx context.Context) error {
		signatureMap, err := r.GetSignatureMap(ctx, signature.RequestID())
		if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
			return errors.Errorf("failed to get valset signature map: %w", err)
		}
		if errors.Is(err, entity.ErrEntityNotFound) {
			// Get the total number of active validators for this epoch
			totalActiveValidators, err := r.GetActiveValidatorCountByEpoch(ctx, signature.Epoch)
			if err != nil {
				return errors.Errorf("failed to get active validator count for epoch %d: %w", signature.Epoch, err)
			}

			signatureMap = entity.NewSignatureMap(signature.RequestID(), signature.Epoch, totalActiveValidators)
		}

		if err := signatureMap.SetValidatorPresent(activeIndex, validator.VotingPower); err != nil {
			return errors.Errorf("failed to set validator present for request id %s: %w", signature.RequestID().Hex(), err)
		}

		if err := r.UpdateSignatureMap(ctx, signatureMap); err != nil {
			return errors.Errorf("failed to update valset signature map: %w", err)
		}

		if err := r.SaveSignature(ctx, activeIndex, signature); err != nil {
			return errors.Errorf("failed to save signature: %w", err)
		}

		slog.DebugContext(ctx, "Saved signature for validator",
			"activeIndex", activeIndex,
			"requestId", signature.RequestID().Hex(),
			"epoch", signature.Epoch,
			"totalSignatures", signatureMap.SignedValidatorsBitmap.GetCardinality(),
			"presentValidators", signatureMap.SignedValidatorsBitmap.ToArray(),
		)

		if signature.KeyTag.Type().AggregationKey() {
			// Check if quorum is reached and remove from pending collection if so
			validatorSetHeader, err := r.GetValidatorSetHeaderByEpoch(ctx, signature.Epoch)
			if err != nil {
				return errors.Errorf("failed to get validator set header: %v", err)
			}

			// todo check quorum threshold from signature request
			if signatureMap.ThresholdReached(validatorSetHeader.QuorumThreshold) {
				// Remove from pending collection since quorum is reached
				err := r.RemoveSignatureRequestPending(ctx, signature.Epoch, signature.RequestID())
				if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
					return errors.Errorf("failed to remove signature request from pending collection: %v", err)
				}
				// If ErrEntityNotFound, it means it was already removed or never added - that's ok
			}
		}

		return nil
	})
}
