package badger

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) PruneProofsByRequestIDs(ctx context.Context, epoch symbiotic.Epoch, requestIDs []common.Hash, batchSize int) error {
	for _, requestID := range requestIDs {
		if err := r.pruneAggregationProof(ctx, epoch, requestID); err != nil {
			return errors.Errorf("failed to prune aggregation proof for request %s: %w", requestID.Hex(), err)
		}

		r.proofsMutexMap.Delete(requestID)
	}

	return nil
}

func (r *Repository) PruneSignaturesByRequestIDs(ctx context.Context, epoch symbiotic.Epoch, requestIDs []common.Hash, batchSize int) error {
	for _, requestID := range requestIDs {
		if err := r.pruneSignatureEntities(ctx, epoch, requestID); err != nil {
			return errors.Errorf("failed to prune signature entities for request %s: %w", requestID.Hex(), err)
		}

		r.signatureMutexMap.Delete(requestID)
	}

	return nil
}

func (r *Repository) PruneEpochIndicesByRequestIDs(ctx context.Context, epoch symbiotic.Epoch, requestIDs []common.Hash, batchSize int) error {
	for _, requestID := range requestIDs {
		if err := r.deleteRequestIDEpochIndex(ctx, epoch, requestID); err != nil {
			return errors.Errorf("failed to delete request ID epoch index for request %s: %w", requestID.Hex(), err)
		}
	}

	return nil
}

func (r *Repository) GetRequestIDsByEpoch(ctx context.Context, epoch symbiotic.Epoch, limit int) ([]common.Hash, error) {
	return r.getRequestIDsByEpoch(ctx, epoch, limit)
}

func (r *Repository) PruneProofCommits(ctx context.Context, epoch symbiotic.Epoch) error {
	return r.pruneProofCommits(ctx, epoch)
}
