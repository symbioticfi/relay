package badger

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) SaveProof(ctx context.Context, aggregationProof symbiotic.AggregationProof) error {
	requestID := aggregationProof.RequestID()
	return r.doUpdateInTxWithLock(ctx, "SaveProof", func(ctx context.Context) error {
		if err := r.writeAggregationProof(ctx, requestID, aggregationProof); err != nil {
			return errors.Errorf("failed to save aggregation proof: %w", err)
		}
		if err := r.removeAggregationProofPending(ctx, aggregationProof.Epoch, requestID); err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
			return errors.Errorf("failed to remove aggregation proof from pending collection: %w", err)
		}
		return nil
	}, &r.requestIDMutexMap, requestID)
}
