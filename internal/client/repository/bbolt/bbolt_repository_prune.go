package bbolt

import (
	"bytes"
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	bolt "go.etcd.io/bbolt"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) PruneValsetEntities(ctx context.Context, epoch symbiotic.Epoch, batchSize int) error {
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

func (r *Repository) getRequestIDsByEpoch(ctx context.Context, epoch symbiotic.Epoch, limit int) ([]common.Hash, error) {
	var requestIDs []common.Hash
	err := r.doView(ctx, "getRequestIDsByEpoch", func(tx *bolt.Tx) error {
		requestIDs = getRequestIDsByEpochTx(tx, epoch, limit)
		return nil
	})
	return requestIDs, err
}

func getRequestIDsByEpochTx(tx *bolt.Tx, epoch symbiotic.Epoch, limit int) []common.Hash {
	var requestIDs []common.Hash
	prefix := epochBytes(uint64(epoch))
	c := tx.Bucket(bucketRequestIDEpochs).Cursor()

	for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
		if len(k) < 40 {
			continue
		}
		requestIDs = append(requestIDs, common.BytesToHash(k[8:40]))
		if limit > 0 && len(requestIDs) >= limit {
			break
		}
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

func pruneBatchSize(batchSize, total int) int {
	if batchSize <= 0 {
		return max(total, 1)
	}
	return batchSize
}
