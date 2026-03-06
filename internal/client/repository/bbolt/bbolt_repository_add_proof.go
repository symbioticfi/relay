package bbolt

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) SaveProof(ctx context.Context, aggregationProof symbiotic.AggregationProof) error {
	requestID := aggregationProof.RequestID()

	if err := r.saveAggregationProof(ctx, requestID, aggregationProof); err != nil {
		return errors.Errorf("failed to save aggregation proof: %w", err)
	}

	err := r.RemoveAggregationProofPending(ctx, aggregationProof.Epoch, requestID)
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		return errors.Errorf("failed to remove aggregation proof from pending collection: %w", err)
	}

	return nil
}
