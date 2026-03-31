package bbolt

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) SaveSignature(ctx context.Context, signature symbiotic.Signature, validator symbiotic.Validator, activeIndex uint32) error {
	// 1. Persist signature to disk — parallel, batched across goroutines
	if err := r.saveSignature(ctx, activeIndex, signature); err != nil {
		return errors.Errorf("failed to save signature: %w", err)
	}

	// 2. Update in-memory signature map (fast, short lock)
	if err := r.updateSignatureMapInMemory(ctx, signature, activeIndex, validator.VotingPower); err != nil {
		return errors.Errorf("failed to update in-memory signature map: %w", err)
	}

	slog.DebugContext(ctx, "Saved signature for validator",
		"activeIndex", activeIndex,
		"requestId", signature.RequestID().Hex(),
		"epoch", signature.Epoch,
	)

	// 3. Handle pending aggregation proof management
	return r.handlePendingAggregationProof(ctx, signature)
}

func (r *Repository) updateSignatureMapInMemory(ctx context.Context, signature symbiotic.Signature, activeIndex uint32, votingPower symbiotic.VotingPower) error {
	requestID := signature.RequestID()

	entryI, loaded := r.signatureMapCache.LoadOrStore(requestID, &signatureMapEntry{})
	entry := entryI.(*signatureMapEntry)

	// Initialize on first access — I/O outside the lock to avoid blocking concurrent writers.
	// Double-init by concurrent goroutines is idempotent (same epoch, same validator count).
	if !loaded {
		totalActiveValidators, err := r.GetActiveValidatorCountByEpoch(ctx, signature.Epoch)
		if err != nil {
			return errors.Errorf("failed to get active validator count for epoch %d: %w", signature.Epoch, err)
		}
		entry.mu.Lock()
		entry.val = entity.NewSignatureMap(requestID, signature.Epoch, totalActiveValidators)
		entry.mu.Unlock()
	}

	entry.mu.Lock()
	// SetValidatorPresent returns error if already present — ignore it (idempotent)
	_ = entry.val.SetValidatorPresent(activeIndex, votingPower)
	entry.mu.Unlock()

	return nil
}

func (r *Repository) handlePendingAggregationProof(ctx context.Context, signature symbiotic.Signature) error {
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
		entryI, ok := r.signatureMapCache.Load(signature.RequestID())
		if ok {
			entry := entryI.(*signatureMapEntry)
			entry.mu.Lock()
			missingCount := len(entry.val.GetMissingValidators().ToArray())
			entry.mu.Unlock()

			if missingCount == 0 {
				err := r.RemoveAggregationProofPending(ctx, signature.Epoch, signature.RequestID())
				if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
					return errors.Errorf("failed to remove signature request from pending collection: %w", err)
				}
				return nil
			}
		}
		if err := r.saveAggregationProofPending(ctx, signature.RequestID(), signature.Epoch); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to save aggregation proof to pending collection: %w", err)
		}
	}
	return nil
}
