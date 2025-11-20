package badger

import (
	"context"
	"log/slog"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// PruneValsetEntities deletes validator sets and network configs for a specific epoch.
func (r *Repository) PruneValsetEntities(ctx context.Context, epoch symbiotic.Epoch) error {
	if err := r.pruneNetworkConfigs(ctx, epoch); err != nil {
		return errors.Errorf("failed to prune network configs: %w", err)
	}

	if err := r.pruneValidatorSets(ctx, epoch); err != nil {
		return errors.Errorf("failed to prune validator sets: %w", err)
	}

	return nil
}

// PruneProofEntities deletes aggregation proofs and proof commits for a specific epoch.
func (r *Repository) PruneProofEntities(ctx context.Context, epoch symbiotic.Epoch) error {
	if err := r.pruneProofCommits(ctx, epoch); err != nil {
		return errors.Errorf("failed to prune proof commits: %w", err)
	}

	requestIDs, err := r.getRequestIDsByEpoch(ctx, epoch)
	if err != nil {
		return errors.Errorf("failed to get request IDs: %w", err)
	}

	for _, requestID := range requestIDs {
		if err := r.pruneAggregationProof(ctx, epoch, requestID); err != nil {
			return errors.Errorf("failed to prune aggregation proof for request %s: %w", requestID.Hex(), err)
		}

		r.proofsMutexMap.Delete(requestID)
	}

	return nil
}

// PruneSignatureEntitiesForEpoch deletes all signature-related entities for a specific epoch.
func (r *Repository) PruneSignatureEntitiesForEpoch(ctx context.Context, epoch symbiotic.Epoch) error {
	requestIDs, err := r.getRequestIDsByEpoch(ctx, epoch)
	if err != nil {
		return errors.Errorf("failed to get request IDs: %w", err)
	}

	slog.DebugContext(ctx, "Pruning signature entities", "requestCount", len(requestIDs))

	for _, requestID := range requestIDs {
		if err := r.pruneSignatureEntities(ctx, epoch, requestID); err != nil {
			return errors.Errorf("failed to prune signature entities for request %s: %w", requestID.Hex(), err)
		}

		r.signatureMutexMap.Delete(requestID)
	}

	return nil
}

// pruneProofCommits deletes proof commit markers for the epoch.
func (r *Repository) pruneProofCommits(ctx context.Context, epoch symbiotic.Epoch) error {
	return r.doUpdateInTx(ctx, "pruneProofCommits", func(ctx context.Context) error {
		txn := getTxn(ctx)
		if err := txn.Delete(keyAggregationProofCommited(epoch)); err != nil {
			return errors.Errorf("failed to delete proof commit: %w", err)
		}
		return nil
	})
}

// pruneNetworkConfigs deletes network configuration for the epoch.
func (r *Repository) pruneNetworkConfigs(ctx context.Context, epoch symbiotic.Epoch) error {
	return r.doUpdateInTx(ctx, "pruneNetworkConfigs", func(ctx context.Context) error {
		txn := getTxn(ctx)
		if err := txn.Delete(keyNetworkConfig(epoch)); err != nil {
			return errors.Errorf("failed to delete network config: %w", err)
		}
		return nil
	})
}

// pruneValidatorSets deletes all validator set data for the epoch.
func (r *Repository) pruneValidatorSets(ctx context.Context, epoch symbiotic.Epoch) error {
	err := r.doUpdateInTxWithLock(ctx, "pruneValidatorSets", func(ctx context.Context) error {
		txn := getTxn(ctx)

		// Delete validator set header and metadata keys
		staticKeys := [][]byte{
			keyValidatorSetHeader(epoch),
			keyValidatorSetStatus(epoch),
			keyValidatorSetMetadata(epoch),
			keyActiveValidatorCount(epoch),
		}

		for _, key := range staticKeys {
			if err := txn.Delete(key); err != nil {
				return errors.Errorf("failed to delete static key: %w", err)
			}
		}

		// Delete all validators for this epoch
		validatorPrefix := keyValidatorPrefix(epoch)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = validatorPrefix
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(validatorPrefix); it.ValidForPrefix(validatorPrefix); it.Next() {
			key := it.Item().KeyCopy(nil)
			if err := txn.Delete(key); err != nil {
				return errors.Errorf("failed to delete validator key: %w", err)
			}
		}
		it.Close() // Close before opening another iterator

		// Delete all validator key lookups for this epoch
		keyLookupPrefix := keyValidatorKeyLookupPrefix(epoch)
		opts.Prefix = keyLookupPrefix
		it = txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(keyLookupPrefix); it.ValidForPrefix(keyLookupPrefix); it.Next() {
			key := it.Item().KeyCopy(nil)
			if err := txn.Delete(key); err != nil {
				return errors.Errorf("failed to delete key lookup: %w", err)
			}
		}

		return nil
	}, &r.valsetMutexMap, epoch)

	if err != nil {
		return errors.Errorf("failed to prune validator sets: %w", err)
	}

	return nil
}

// getRequestIDsByEpoch extracts all requestIDs associated with an epoch from the index.
func (r *Repository) getRequestIDsByEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]common.Hash, error) {
	var requestIDs []common.Hash

	err := r.doViewInTx(ctx, "getRequestIDsByEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)

		prefix := keyRequestIDEpochPrefix(epoch)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := it.Item().Key()
			requestID, err := extractRequestIDFromEpochKey(key)
			if err != nil {
				slog.WarnContext(ctx, "Failed to extract requestID from key",
					"key", string(key),
					"error", err,
				)
				continue
			}
			requestIDs = append(requestIDs, requestID)
		}

		return nil
	})

	return requestIDs, err
}

// pruneSignatureEntities deletes all signature-related entities for a requestID in a single transaction.
// This includes: signatures, signature map, signature request, and related markers.
func (r *Repository) pruneSignatureEntities(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) error {
	return r.doUpdateInTxWithLock(ctx, "pruneSignatureEntities", func(ctx context.Context) error {
		txn := getTxn(ctx)

		// Delete all individual signatures for this requestID
		signaturePrefix := keySignatureRequestIDPrefix(requestID)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = signaturePrefix
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(signaturePrefix); it.ValidForPrefix(signaturePrefix); it.Next() {
			key := it.Item().KeyCopy(nil)
			if err := txn.Delete(key); err != nil {
				return errors.Errorf("failed to delete signature key: %w", err)
			}
		}

		// Delete signature map
		if err := txn.Delete(keySignatureMap(requestID)); err != nil {
			return errors.Errorf("failed to delete signature map: %w", err)
		}

		// Delete signature request
		if err := txn.Delete(keySignatureRequest(epoch, requestID)); err != nil {
			return errors.Errorf("failed to delete signature request: %w", err)
		}

		// Delete signature pending marker
		if err := txn.Delete(keySignatureRequestPending(epoch, requestID)); err != nil {
			return errors.Errorf("failed to delete signature pending: %w", err)
		}

		// Delete requestID index
		if err := txn.Delete(keyRequestIDIndex(requestID)); err != nil {
			return errors.Errorf("failed to delete request ID index: %w", err)
		}

		return nil
	}, &r.signatureMutexMap, requestID)
}

// pruneAggregationProof deletes the aggregation proof and pending marker for a requestID.
func (r *Repository) pruneAggregationProof(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) error {
	return r.doUpdateInTxWithLock(ctx, "pruneAggregationProof", func(ctx context.Context) error {
		txn := getTxn(ctx)

		// Delete aggregation proof
		if err := txn.Delete(keyAggregationProof(requestID)); err != nil {
			return errors.Errorf("failed to delete aggregation proof: %w", err)
		}

		// Delete pending marker
		if err := txn.Delete(keyAggregationProofPending(epoch, requestID)); err != nil {
			return errors.Errorf("failed to delete aggregation proof pending: %w", err)
		}

		return nil
	}, &r.proofsMutexMap, requestID)
}
