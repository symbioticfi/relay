package bbolt

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) SaveSignature(ctx context.Context, signature symbiotic.Signature, validator symbiotic.Validator, activeIndex uint32) error {
	if err := r.saveSignature(ctx, activeIndex, signature); err != nil {
		return errors.Errorf("failed to save signature: %w", err)
	}

	signatureMap, err := r.addToSignatureMapCache(ctx, signature.RequestID(), signature.Epoch, activeIndex, validator.VotingPower)
	if err != nil {
		return errors.Errorf("failed to update signature map cache: %w", err)
	}

	slog.DebugContext(ctx, "Saved signature for validator",
		"activeIndex", activeIndex,
		"requestId", signature.RequestID().Hex(),
		"epoch", signature.Epoch,
		"totalSignatures", signatureMap.SignedValidatorsBitmap.GetCardinality(),
		"presentValidators", signatureMap.SignedValidatorsBitmap.ToArray(),
	)

	if signature.KeyTag.Type().AggregationKey() {
		_, err := r.GetAggregationProof(ctx, signature.RequestID())
		if err != nil {
			if !errors.Is(err, entity.ErrEntityNotFound) {
				return errors.Errorf("failed to get aggregation proof: %w", err)
			}
			if err := r.saveAggregationProofPending(ctx, signature.RequestID(), signature.Epoch); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
				return errors.Errorf("failed to save aggregation proof to pending collection: %w", err)
			}
		}
	} else {
		if len(signatureMap.GetMissingValidators().ToArray()) == 0 {
			err := r.RemoveAggregationProofPending(ctx, signature.Epoch, signature.RequestID())
			if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
				return errors.Errorf("failed to remove signature request from pending collection: %w", err)
			}
		} else {
			if err := r.saveAggregationProofPending(ctx, signature.RequestID(), signature.Epoch); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
				return errors.Errorf("failed to save aggregation proof to pending collection: %w", err)
			}
		}
	}

	return nil
}

func (r *Repository) addToSignatureMapCache(ctx context.Context, requestID common.Hash, epoch symbiotic.Epoch, activeIndex uint32, votingPower symbiotic.VotingPower) (entity.SignatureMap, error) {
	for {
		raw, ok := r.signatureMapCache.Load(requestID)
		if !ok {
			totalActive, err := r.GetActiveValidatorCountByEpoch(ctx, epoch)
			if err != nil {
				return entity.SignatureMap{}, errors.Errorf("failed to get active validator count: %w", err)
			}
			sm := entity.NewSignatureMap(requestID, epoch, totalActive)
			_ = sm.SetValidatorPresent(activeIndex, votingPower)

			actual, loaded := r.signatureMapCache.LoadOrStore(requestID, sm)
			if !loaded {
				return sm, nil
			}
			raw = actual
		}

		old := raw.(entity.SignatureMap)
		if old.SignedValidatorsBitmap.Contains(activeIndex) {
			return old, nil
		}

		cloned := old.Clone()
		_ = cloned.SetValidatorPresent(activeIndex, votingPower)
		if r.signatureMapCache.CompareAndSwap(requestID, raw, cloned) {
			return cloned, nil
		}
	}
}
