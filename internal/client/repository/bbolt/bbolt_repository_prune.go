package bbolt

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	bolt "go.etcd.io/bbolt"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) PruneValsetEntities(ctx context.Context, epoch symbiotic.Epoch) error {
	return r.doUpdate(ctx, "PruneValsetEntities", func(tx *bolt.Tx) error {
		ek := epochBytes(uint64(epoch))

		// Delete network config
		if err := tx.Bucket(bucketNetworkConfigs).Delete(ek); err != nil {
			return errors.Errorf("failed to delete network config: %w", err)
		}

		// Delete static validator set keys
		for _, bucket := range [][]byte{
			bucketValidatorSetHeaders,
			bucketValidatorSetStatus,
			bucketValidatorSetMeta,
			bucketActiveValCounts,
		} {
			if err := tx.Bucket(bucket).Delete(ek); err != nil {
				return errors.Errorf("failed to delete from bucket %s: %w", bucket, err)
			}
		}

		// Delete all validators for this epoch
		prefix := epochBytes(uint64(epoch))
		if err := deletePrefixedKeys(tx.Bucket(bucketValidators), prefix); err != nil {
			return errors.Errorf("failed to delete validators: %w", err)
		}
		if err := deletePrefixedKeys(tx.Bucket(bucketValidatorKeyLookups), prefix); err != nil {
			return errors.Errorf("failed to delete validator key lookups: %w", err)
		}

		return nil
	})
}

func (r *Repository) PruneProofEntities(ctx context.Context, epoch symbiotic.Epoch) error {
	return r.doUpdate(ctx, "PruneProofEntities", func(tx *bolt.Tx) error {
		ek := epochBytes(uint64(epoch))

		// Delete proof commits
		if err := tx.Bucket(bucketAggProofCommits).Delete(ek); err != nil {
			return errors.Errorf("failed to delete proof commits: %w", err)
		}

		// Find all request IDs for this epoch
		requestIDs := getRequestIDsByEpochTx(tx, epoch)

		for _, requestID := range requestIDs {
			// Delete aggregation proof
			if err := tx.Bucket(bucketAggregationProofs).Delete(requestID.Bytes()); err != nil {
				return errors.Errorf("failed to delete aggregation proof: %w", err)
			}

			// Delete aggregation proof pending
			pendingKey := epochHashKey(uint64(epoch), requestID.Bytes())
			if err := tx.Bucket(bucketAggProofPending).Delete(pendingKey); err != nil {
				return errors.Errorf("failed to delete pending agg proof: %w", err)
			}
		}

		return nil
	})
}

func (r *Repository) PruneSignatureEntitiesForEpoch(ctx context.Context, epoch symbiotic.Epoch) error {
	requestIDs, err := r.getRequestIDsByEpoch(ctx, epoch)
	if err != nil {
		return errors.Errorf("failed to get request IDs for epoch %d: %w", epoch, err)
	}
	slog.DebugContext(ctx, "Pruning signature entities", "requestCount", len(requestIDs))

	return r.doUpdate(ctx, "PruneSignatureEntitiesForEpoch", func(tx *bolt.Tx) error {
		for _, requestID := range requestIDs {
			// Delete all signatures for this requestID
			sigPrefix := requestID.Bytes()
			if err := deletePrefixedKeys(tx.Bucket(bucketSignatures), sigPrefix); err != nil {
				return errors.Errorf("failed to delete signatures: %w", err)
			}

			// Delete signature map
			if err := tx.Bucket(bucketSignatureMaps).Delete(requestID.Bytes()); err != nil {
				return errors.Errorf("failed to delete signature map: %w", err)
			}

			// Delete signature request
			reqKey := epochHashKey(uint64(epoch), requestID.Bytes())
			if err := tx.Bucket(bucketSignatureRequests).Delete(reqKey); err != nil {
				return errors.Errorf("failed to delete signature request: %w", err)
			}

			// Delete signature pending
			if err := tx.Bucket(bucketSignaturePending).Delete(reqKey); err != nil {
				return errors.Errorf("failed to delete signature pending: %w", err)
			}

			// Delete request ID index
			if err := tx.Bucket(bucketRequestIDIndex).Delete(requestID.Bytes()); err != nil {
				return errors.Errorf("failed to delete request ID index: %w", err)
			}

			r.signatureMutexMap.Delete(requestID)
		}
		return nil
	})
}

func (r *Repository) PruneRequestIDEpochIndices(ctx context.Context, epoch symbiotic.Epoch) error {
	requestIDs, err := r.getRequestIDsByEpoch(ctx, epoch)
	if err != nil {
		return errors.Errorf("failed to get request IDs for epoch %d: %w", epoch, err)
	}
	slog.DebugContext(ctx, "Pruning request ID epoch indices", "epoch", epoch, "requestCount", len(requestIDs))

	return r.doUpdate(ctx, "PruneRequestIDEpochIndices", func(tx *bolt.Tx) error {
		for _, requestID := range requestIDs {
			// Check if aggregation proof still exists
			if tx.Bucket(bucketAggregationProofs).Get(requestID.Bytes()) != nil {
				continue
			}
			// Check if request ID index still exists
			if tx.Bucket(bucketRequestIDIndex).Get(requestID.Bytes()) != nil {
				continue
			}

			// Both gone, safe to delete
			epochKey := epochHashKey(uint64(epoch), requestID.Bytes())
			if err := tx.Bucket(bucketRequestIDEpochs).Delete(epochKey); err != nil {
				return errors.Errorf("failed to delete request ID epoch index: %w", err)
			}
		}
		return nil
	})
}

func (r *Repository) getRequestIDsByEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]common.Hash, error) {
	var requestIDs []common.Hash
	err := r.doView(ctx, "getRequestIDsByEpoch", func(tx *bolt.Tx) error {
		requestIDs = getRequestIDsByEpochTx(tx, epoch)
		return nil
	})
	return requestIDs, err
}

func getRequestIDsByEpochTx(tx *bolt.Tx, epoch symbiotic.Epoch) []common.Hash {
	var requestIDs []common.Hash
	prefix := epochBytes(uint64(epoch))
	c := tx.Bucket(bucketRequestIDEpochs).Cursor()

	for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
		if len(k) < 40 {
			continue
		}
		requestIDs = append(requestIDs, common.BytesToHash(k[8:40]))
	}
	return requestIDs
}

func deletePrefixedKeys(b *bolt.Bucket, prefix []byte) error {
	c := b.Cursor()
	for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Seek(prefix) {
		if err := c.Delete(); err != nil {
			return err
		}
	}
	return nil
}
