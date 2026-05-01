package bbolt

import (
	"bytes"
	"context"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"
	bolt "go.etcd.io/bbolt"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (r *Repository) PruneValsetEntities(ctx context.Context, epoch symbiotic.Epoch, batchSize int) error {
	return r.doBatch(ctx, "PruneValsetEntities", func(tx *bolt.Tx) error {
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

func (r *Repository) PruneProofEntities(ctx context.Context, epoch symbiotic.Epoch, batchSize int) error {
	if err := r.doBatch(ctx, "PruneProofEntities:commits", func(tx *bolt.Tx) error {
		ek := epochBytes(uint64(epoch))
		if err := tx.Bucket(bucketAggProofCommits).Delete(ek); err != nil {
			return errors.Errorf("failed to delete proof commits: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	requestIDs, err := r.getRequestIDsByEpoch(ctx, epoch)
	if err != nil {
		return errors.Errorf("failed to get request IDs for epoch %d: %w", epoch, err)
	}

	for _, chunk := range lo.Chunk(requestIDs, pruneBatchSize(batchSize, len(requestIDs))) {
		if err := r.doBatch(ctx, "PruneProofEntities:batch", func(tx *bolt.Tx) error {
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

		if r.pausePrune(ctx, epoch) {
			break
		}
	}

	return nil
}

func (r *Repository) PruneSignatureEntitiesForEpoch(ctx context.Context, epoch symbiotic.Epoch, batchSize int) error {
	requestIDs, err := r.getRequestIDsByEpoch(ctx, epoch)
	if err != nil {
		return errors.Errorf("failed to get request IDs for epoch %d: %w", epoch, err)
	}
	slog.DebugContext(ctx, "Pruning signature entities", "requestCount", len(requestIDs))

	for _, chunk := range lo.Chunk(requestIDs, pruneBatchSize(batchSize, len(requestIDs))) {
		// Invalidate cache BEFORE DB delete: a concurrent reader on cache miss
		// would rebuild from DB; if we deleted the cache after the DB delete,
		// the reader could see a stale cached map for an already-pruned request.
		for _, requestID := range chunk {
			r.signatureMapCache.Delete(requestID)
		}

		if err := r.doBatch(ctx, "PruneSignatureEntitiesForEpoch:batch", func(tx *bolt.Tx) error {
			for _, requestID := range chunk {
				sigPrefix := requestID.Bytes()
				if err := deletePrefixedKeys(tx.Bucket(bucketSignatures), sigPrefix); err != nil {
					return errors.Errorf("failed to delete signatures: %w", err)
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

		if r.pausePrune(ctx, epoch) {
			break
		}
	}

	return nil
}

func (r *Repository) PruneRequestIDEpochIndices(ctx context.Context, epoch symbiotic.Epoch, batchSize int) error {
	requestIDs, err := r.getRequestIDsByEpoch(ctx, epoch)
	if err != nil {
		return errors.Errorf("failed to get request IDs for epoch %d: %w", epoch, err)
	}
	slog.DebugContext(ctx, "Pruning request ID epoch indices", "epoch", epoch, "requestCount", len(requestIDs))

	for _, chunk := range lo.Chunk(requestIDs, pruneBatchSize(batchSize, len(requestIDs))) {
		if err := r.doBatch(ctx, "PruneRequestIDEpochIndices:batch", func(tx *bolt.Tx) error {
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

		if r.pausePrune(ctx, epoch) {
			break
		}
	}

	return nil
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

// deletePrefixedKeys buffers all matching keys during a single forward scan, then deletes
// them. The naive c.Delete()+re-Seek loop is O(N·log N) on B+tree height; this is O(N).
func deletePrefixedKeys(b *bolt.Bucket, prefix []byte) error {
	c := b.Cursor()
	var keys [][]byte
	for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
		keys = append(keys, append([]byte(nil), k...))
	}
	for _, k := range keys {
		if err := b.Delete(k); err != nil {
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

// pausePrune sleeps for the configured prune pause window or returns early if ctx is cancelled.
// Returns true if pruning should stop (ctx cancelled), false otherwise.
func (r *Repository) pausePrune(ctx context.Context, epoch symbiotic.Epoch) bool {
	if r.prunePause <= 0 {
		if ctx.Err() != nil {
			slog.DebugContext(ctx, "Pruning interrupted by context cancellation", "epoch", epoch)
			return true
		}
		return false
	}
	t := time.NewTimer(r.prunePause)
	defer t.Stop()
	select {
	case <-t.C:
		return false
	case <-ctx.Done():
		slog.DebugContext(ctx, "Pruning interrupted by context cancellation", "epoch", epoch)
		return true
	}
}
