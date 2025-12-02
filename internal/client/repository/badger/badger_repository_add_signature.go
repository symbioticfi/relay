package badger

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) SaveSignature(ctx context.Context, signature symbiotic.Signature, validator symbiotic.Validator, activeIndex uint32) error {
	var (
		signatureMap entity.SignatureMap
		err          error
	)

	if err := r.doUpdateInTxWithLock(ctx, "SaveSignature", func(ctx context.Context) error {
		signatureMap, err = r.GetSignatureMap(ctx, signature.RequestID())
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

		if err = signatureMap.SetValidatorPresent(activeIndex, validator.VotingPower); err != nil {
			return errors.Errorf("failed to set validator present for request id %s: %w", signature.RequestID().Hex(), err)
		}

		if err = r.UpdateSignatureMap(ctx, signatureMap); err != nil {
			return errors.Errorf("failed to update valset signature map: %w", err)
		}

		if err = r.saveSignature(ctx, activeIndex, signature); err != nil {
			return errors.Errorf("failed to save signature: %w", err)
		}

		slog.DebugContext(ctx, "Saved signature for validator",
			"activeIndex", activeIndex,
			"requestId", signature.RequestID().Hex(),
			"epoch", signature.Epoch,
			"totalSignatures", signatureMap.SignedValidatorsBitmap.GetCardinality(),
			"presentValidators", signatureMap.SignedValidatorsBitmap.ToArray(),
		)

		return nil
	}, &r.signatureMutexMap, signature.RequestID()); err != nil {
		return err
	}

	// outside previous transaction, check if we can remove from pending collection
	if signature.KeyTag.Type().AggregationKey() {
		_, err := r.GetAggregationProof(ctx, signature.RequestID())
		if err != nil {
			if !errors.Is(err, entity.ErrEntityNotFound) {
				return errors.Errorf("failed to get aggregation proof: %v", err)
			}
			// Blindly save to pending aggregation proof collection
			// syncer will remove it from collection once proof is found
			if err := r.saveAggregationProofPending(ctx, signature.RequestID(), signature.Epoch); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) && !errors.Is(err, entity.ErrTxConflict) {
				// ignore ErrEntityAlreadyExist and ErrTxConflict - it means it's already there or being processed
				return errors.Errorf("failed to save aggregation proof to pending collection: %v", err)
			}
		}
	} else {
		if len(signatureMap.GetMissingValidators().ToArray()) == 0 {
			// for non aggregation keys, we wait for all validators to sign and then remove
			// the pending aggregation marker to stop syncing signatures for this request
			err := r.RemoveAggregationProofPending(ctx, signature.Epoch, signature.RequestID())
			if err != nil && !errors.Is(err, entity.ErrEntityNotFound) && !errors.Is(err, entity.ErrTxConflict) {
				return errors.Errorf("failed to remove signature request from pending collection: %v", err)
			}
		} else {
			// Save to pending aggregation proof collection, to sync for missing signatures
			// syncer will remove it from collection once all signatures are found
			if err := r.saveAggregationProofPending(ctx, signature.RequestID(), signature.Epoch); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) && !errors.Is(err, entity.ErrTxConflict) {
				// ignore ErrEntityAlreadyExist and ErrTxConflict - it means it's already there or being processed
				return errors.Errorf("failed to save aggregation proof to pending collection: %v", err)
			}
		}
	}
	return nil
}
