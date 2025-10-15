package badger

import (
	"context"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-errors/errors"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) SaveProof(ctx context.Context, aggregationProof symbiotic.AggregationProof) error {
	requestID := aggregationProof.RequestID()

	err := r.saveAggregationProof(ctx, requestID, aggregationProof)
	if err != nil {
		return errors.Errorf("failed to save aggregation proof: %w", err)
	}

	// Remove from pending collection
	err = r.RemoveAggregationProofPending(ctx, aggregationProof.Epoch, requestID)
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) && !errors.Is(err, entity.ErrTxConflict) && !errors.Is(err, badger.ErrConflict) {
		return errors.Errorf("failed to remove aggregation proof from pending collection: %w", err)
	}
	// If ErrEntityNotFound or ErrTxConflict, it means it was already processed or is being processed, so we can ignore it

	return nil
}
