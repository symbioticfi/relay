package badger

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

const (
	aggregationProofCommitPrefix = "aggregation_proof_commit:"
)

func keyAggregationProofCommited(epoch entity.Epoch, requestID common.Hash) []byte {
	return []byte(fmt.Sprintf("%v%d:%s", aggregationProofCommitPrefix, epoch, requestID.Hex()))
}

func (r *Repository) SaveProofCommitPending(ctx context.Context, epoch entity.Epoch, requestID common.Hash) error {
	return r.DoUpdateInTx(ctx, func(ctx context.Context) error {
		txn := getTxn(ctx)
		_, err := txn.Get(keyAggregationProofCommited(epoch, requestID))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get proof commit pending: %w", err)
		}
		if err == nil {
			return errors.Errorf("proof commit pending already exists: %w", entity.ErrEntityAlreadyExist)
		}

		// Adding TTL ensure that we don't access the old proofs when querying for pending proofs
		// DEV: setting to 24hours, if the proof doesn't get committed in that time frame manual intervention is expected
		err = txn.SetEntry(badger.NewEntry(keyAggregationProofCommited(epoch, requestID), []byte{0x01}).WithTTL(time.Hour * 24))
		if err != nil {
			return errors.Errorf("failed to store proof commit pending: %w", err)
		}
		return nil
	})
}

func (r *Repository) RemoveProofCommitPending(ctx context.Context, epoch entity.Epoch, requestID common.Hash) error {
	return r.DoUpdateInTx(ctx, func(ctx context.Context) error {
		txn := getTxn(ctx)
		pendingKey := keyAggregationProofCommited(epoch, requestID)

		// Check if exists before removing
		_, err := txn.Get(pendingKey)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("proof commit pending not found for epoch %d and hash %s: %w", epoch, requestID.Hex(), entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to check proof commit pending: %w", err)
		}

		err = txn.Delete(pendingKey)
		if err != nil {
			return errors.Errorf("failed to delete proof commit pending: %w", err)
		}

		return nil
	})
}

func (r *Repository) GetPendingProofCommitsSinceEpoch(ctx context.Context, epoch entity.Epoch, limit int) ([]entity.ProofCommitKey, error) {
	var requests []entity.ProofCommitKey

	if err := r.DoViewInTx(ctx, func(ctx context.Context) error {
		txn := getTxn(ctx)

		// Step 1: Collect all keys with their parsed epochs and hashes
		var keys []entity.ProofCommitKey

		// Use broader prefix to capture all epochs
		// Key format: "aggregation_proof_commit:epoch:hash"
		basePrefix := []byte(aggregationProofCommitPrefix)
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // We only need keys, not values
		it := txn.NewIterator(opts)
		defer it.Close()

		// Iterate through all aggregation proof commit keys
		it.Seek(basePrefix)
		for it.Valid() && bytes.HasPrefix(it.Item().Key(), basePrefix) {
			item := it.Item()
			key := string(item.Key())

			// Parse the key: "aggregation_proof_commit:epoch:hash"
			parts := strings.Split(key, ":")
			if len(parts) != 3 {
				it.Next()
				continue // Skip invalid keys
			}

			// Parse the epoch
			keyEpochInt, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				it.Next()
				continue // Skip invalid epoch
			}
			keyEpoch := entity.Epoch(keyEpochInt)

			// Skip if this epoch is less than our target epoch
			if keyEpoch < epoch {
				it.Next()
				continue
			}

			requestIDStr := parts[2]
			requestID := common.HexToHash(requestIDStr)

			keys = append(keys, entity.ProofCommitKey{
				Epoch:     keyEpoch,
				RequestID: requestID,
			})

			it.Next()
		}

		if len(keys) == 0 {
			return nil // No keys found
		}

		// Step 2: Sort keys by epoch (ascending) and then by hash for deterministic ordering
		sort.Slice(keys, func(i, j int) bool {
			if keys[i].Epoch != keys[j].Epoch {
				return keys[i].Epoch < keys[j].Epoch
			}
			return bytes.Compare(keys[i].RequestID[:], keys[j].RequestID[:]) < 0
		})

		// Step 3: limit response
		if limit > 0 {
			limit = min(limit, len(keys))
		} else {
			limit = len(keys)
		}
		requests = keys[0:limit]

		return nil
	}); err != nil {
		return nil, err
	}

	return requests, nil
}
