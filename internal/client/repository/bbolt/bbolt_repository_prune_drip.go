package bbolt

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"
	bolt "go.etcd.io/bbolt"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) PruneProofsByRequestIDs(ctx context.Context, epoch symbiotic.Epoch, requestIDs []common.Hash, batchSize int) error {
	for _, chunk := range lo.Chunk(requestIDs, pruneBatchSize(batchSize, len(requestIDs))) {
		if err := r.doUpdate(ctx, "PruneProofsByRequestIDs:batch", func(tx *bolt.Tx) error {
			for _, requestID := range chunk {
				if err := tx.Bucket(bucketAggregationProofs).Delete(requestID.Bytes()); err != nil {
					return errors.Errorf("failed to delete aggregation proof: %w", err)
				}

				pendingKey := epochHashKey(uint64(epoch), requestID.Bytes())
				if err := tx.Bucket(bucketAggProofPending).Delete(pendingKey); err != nil {
					return errors.Errorf("failed to delete pending agg proof: %w", err)
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) PruneSignaturesByRequestIDs(ctx context.Context, epoch symbiotic.Epoch, requestIDs []common.Hash, batchSize int) error {
	for _, chunk := range lo.Chunk(requestIDs, pruneBatchSize(batchSize, len(requestIDs))) {
		if err := r.doUpdate(ctx, "PruneSignaturesByRequestIDs:batch", func(tx *bolt.Tx) error {
			for _, requestID := range chunk {
				sigPrefix := requestID.Bytes()
				if err := deletePrefixedKeys(tx.Bucket(bucketSignatures), sigPrefix); err != nil {
					return errors.Errorf("failed to delete signatures: %w", err)
				}

				if err := tx.Bucket(bucketSignatureMaps).Delete(requestID.Bytes()); err != nil {
					return errors.Errorf("failed to delete signature map: %w", err)
				}

				reqKey := epochHashKey(uint64(epoch), requestID.Bytes())
				if err := tx.Bucket(bucketSignatureRequests).Delete(reqKey); err != nil {
					return errors.Errorf("failed to delete signature request: %w", err)
				}

				if err := tx.Bucket(bucketSignaturePending).Delete(reqKey); err != nil {
					return errors.Errorf("failed to delete signature pending: %w", err)
				}

				if err := tx.Bucket(bucketRequestIDIndex).Delete(requestID.Bytes()); err != nil {
					return errors.Errorf("failed to delete request ID index: %w", err)
				}
			}
			return nil
		}); err != nil {
			return err
		}

		for _, requestID := range chunk {
			r.signatureMapCache.Delete(requestID)
		}
	}

	return nil
}

func (r *Repository) PruneEpochIndicesByRequestIDs(ctx context.Context, epoch symbiotic.Epoch, requestIDs []common.Hash, batchSize int) error {
	for _, chunk := range lo.Chunk(requestIDs, pruneBatchSize(batchSize, len(requestIDs))) {
		if err := r.doUpdate(ctx, "PruneEpochIndicesByRequestIDs:batch", func(tx *bolt.Tx) error {
			for _, requestID := range chunk {
				if tx.Bucket(bucketAggregationProofs).Get(requestID.Bytes()) != nil {
					continue
				}
				if tx.Bucket(bucketRequestIDIndex).Get(requestID.Bytes()) != nil {
					continue
				}

				epochKey := epochHashKey(uint64(epoch), requestID.Bytes())
				if err := tx.Bucket(bucketRequestIDEpochs).Delete(epochKey); err != nil {
					return errors.Errorf("failed to delete request ID epoch index: %w", err)
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) GetRequestIDsByEpoch(ctx context.Context, epoch symbiotic.Epoch, limit int) ([]common.Hash, error) {
	return r.getRequestIDsByEpoch(ctx, epoch, limit)
}

func (r *Repository) PruneProofCommits(ctx context.Context, epoch symbiotic.Epoch) error {
	return r.doUpdate(ctx, "PruneProofCommits", func(tx *bolt.Tx) error {
		ek := epochBytes(uint64(epoch))
		if err := tx.Bucket(bucketAggProofCommits).Delete(ek); err != nil {
			return errors.Errorf("failed to delete proof commits: %w", err)
		}
		return nil
	})
}
