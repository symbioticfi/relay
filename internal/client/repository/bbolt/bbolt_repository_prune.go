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
		tx.Bucket(bucketNetworkConfigs).Delete(ek) //nolint:errcheck // bbolt Delete only errors on readonly tx

		// Delete static validator set keys
		for _, bucket := range [][]byte{
			bucketValidatorSetHeaders,
			bucketValidatorSetStatus,
			bucketValidatorSetMeta,
			bucketActiveValCounts,
		} {
			tx.Bucket(bucket).Delete(ek) //nolint:errcheck // bbolt Delete only errors on readonly tx
		}

		// Delete all validators for this epoch
		prefix := epochBytes(uint64(epoch))
		deletePrefixedKeys(tx.Bucket(bucketValidators), prefix)
		deletePrefixedKeys(tx.Bucket(bucketValidatorKeyLookups), prefix)

		return nil
	})
}

func (r *Repository) PruneProofEntities(ctx context.Context, epoch symbiotic.Epoch) error {
	return r.doUpdate(ctx, "PruneProofEntities", func(tx *bolt.Tx) error {
		ek := epochBytes(uint64(epoch))

		// Delete proof commits
		tx.Bucket(bucketAggProofCommits).Delete(ek) //nolint:errcheck // bbolt Delete only errors on readonly tx

		// Find all request IDs for this epoch
		requestIDs := getRequestIDsByEpochTx(tx, epoch)

		for _, requestID := range requestIDs {
			// Delete aggregation proof
			tx.Bucket(bucketAggregationProofs).Delete(requestID.Bytes()) //nolint:errcheck // bbolt Delete only errors on readonly tx

			// Delete aggregation proof pending
			pendingKey := epochHashKey(uint64(epoch), requestID.Bytes())
			tx.Bucket(bucketAggProofPending).Delete(pendingKey) //nolint:errcheck // bbolt Delete only errors on readonly tx
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
			deletePrefixedKeys(tx.Bucket(bucketSignatures), sigPrefix)

			// Delete signature map
			tx.Bucket(bucketSignatureMaps).Delete(requestID.Bytes()) //nolint:errcheck // bbolt Delete only errors on readonly tx

			// Delete signature request
			reqKey := epochHashKey(uint64(epoch), requestID.Bytes())
			tx.Bucket(bucketSignatureRequests).Delete(reqKey) //nolint:errcheck // bbolt Delete only errors on readonly tx

			// Delete signature pending
			tx.Bucket(bucketSignaturePending).Delete(reqKey) //nolint:errcheck // bbolt Delete only errors on readonly tx

			// Delete request ID index
			tx.Bucket(bucketRequestIDIndex).Delete(requestID.Bytes()) //nolint:errcheck // bbolt Delete only errors on readonly tx

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
			tx.Bucket(bucketRequestIDEpochs).Delete(epochKey) //nolint:errcheck // bbolt Delete only errors on readonly tx
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

func deletePrefixedKeys(b *bolt.Bucket, prefix []byte) {
	c := b.Cursor()
	for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Seek(prefix) {
		c.Delete() //nolint:errcheck // bbolt cursor Delete only errors on readonly tx
	}
}
