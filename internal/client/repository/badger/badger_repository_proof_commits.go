package badger

import (
	"bytes"
	"context"
	"math/big"
	"sort"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

const (
	aggregationProofCommitPrefix = "aggregation_proof_commit:"
)

func keyAggregationProofCommited(epoch symbiotic.Epoch) []byte {
	return append([]byte(aggregationProofCommitPrefix), epoch.Bytes()...)
}

func (r *Repository) SaveProofCommitPending(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) error {
	return r.doUpdateInTx(ctx, "SaveProofCommitPending", func(ctx context.Context) error {
		txn := getTxn(ctx)
		_, err := txn.Get(keyAggregationProofCommited(epoch))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get proof commit pending: %w", err)
		}
		if err == nil {
			return errors.Errorf("proof commit pending already exists: %w", entity.ErrEntityAlreadyExist)
		}

		err = txn.SetEntry(badger.NewEntry(keyAggregationProofCommited(epoch), requestID.Bytes()))
		if err != nil {
			return errors.Errorf("failed to store proof commit pending: %w", err)
		}
		return nil
	})
}

func (r *Repository) RemoveProofCommitPending(ctx context.Context, epoch symbiotic.Epoch) error {
	return r.doUpdateInTx(ctx, "RemoveProofCommitPending", func(ctx context.Context) error {
		txn := getTxn(ctx)
		pendingKey := keyAggregationProofCommited(epoch)

		// Check if exists before removing
		_, err := txn.Get(pendingKey)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("proof commit pending not found for epoch %d: %w", epoch, entity.ErrEntityNotFound)
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

func (r *Repository) GetPendingProofCommitsSinceEpoch(ctx context.Context, epoch symbiotic.Epoch, limit int) ([]symbiotic.ProofCommitKey, error) {
	var requests []symbiotic.ProofCommitKey

	if err := r.doViewInTx(ctx, "GetPendingProofCommitsSinceEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)

		// Step 1: Collect all keys with their parsed epochs and hashes
		var keys []symbiotic.ProofCommitKey

		// Use broader prefix to capture all epochs
		// Key format: "aggregation_proof_commit:epoch:hash"
		basePrefix := []byte(aggregationProofCommitPrefix)
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		// Iterate through all aggregation proof commit keys
		it.Seek(basePrefix)
		for it.Valid() && bytes.HasPrefix(it.Item().Key(), basePrefix) {
			item := it.Item()
			keyBytes := item.Key()
			if !bytes.HasPrefix(keyBytes, basePrefix) {
				it.Next()
				continue // Skip invalid keys
			}

			// Parse the epoch
			epochBytes := keyBytes[len(basePrefix):]
			keyEpochInt := big.NewInt(0).SetBytes(epochBytes).Uint64()
			keyEpoch := symbiotic.Epoch(keyEpochInt)

			// Skip if this epoch is less than our target epoch
			if keyEpoch < epoch {
				it.Next()
				continue
			}

			requestIDBytes, err := it.Item().ValueCopy(nil)
			if err != nil || len(requestIDBytes) != 32 {
				it.Next()
				continue // Skip invalid request IDs
			}
			requestID := common.BytesToHash(requestIDBytes)

			keys = append(keys, symbiotic.ProofCommitKey{
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
