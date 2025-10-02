package badger

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

func (r *Repository) AddProof(ctx context.Context, aggregationProof entity.AggregationProof) error {
	requestID := aggregationProof.RequestID()

	return r.doUpdateInTx(ctx, "ProcessProof", func(ctx context.Context) error {
		// Save the aggregation proof
		err := r.saveAggregationProof(ctx, requestID, aggregationProof)
		if err != nil {
			return errors.Errorf("failed to save aggregation proof: %w", err)
		}

		// Remove from pending collection
		err = r.removeAggregationProofPending(ctx, aggregationProof.Epoch, requestID)
		if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
			return errors.Errorf("failed to remove aggregation proof from pending collection: %w", err)
		}
		// If ErrEntityNotFound, it means it was already removed or never added - that's ok

		return nil
	})
}
