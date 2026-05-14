package badger

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/client/repository/codec"
	"github.com/symbioticfi/relay/internal/client/repository/repoutil"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func keyAggregationProof(requestID common.Hash) []byte {
	return fmt.Appendf(nil, "aggregation_proof:%s", requestID.Hex())
}

const aggregationProofPendingPrefix = "aggregation_proof_pending:"

func keyAggregationProofPending(epoch symbiotic.Epoch, requestID common.Hash) []byte {
	key := epochKeyWithColon(aggregationProofPendingPrefix, epoch)
	return append(key, []byte(requestID.Hex())...)
}

func keyAggregationProofPendingEpochPrefix(epoch symbiotic.Epoch) []byte {
	return epochKeyWithColon(aggregationProofPendingPrefix, epoch)
}

func (r *Repository) saveAggregationProof(ctx context.Context, requestID common.Hash, ap symbiotic.AggregationProof) error {
	return r.doUpdateInTxWithLock(ctx, "saveAggregationProof", func(ctx context.Context) error {
		return r.writeAggregationProof(ctx, requestID, ap)
	}, &r.requestIDMutexMap, requestID)
}

func (r *Repository) writeAggregationProof(ctx context.Context, requestID common.Hash, ap symbiotic.AggregationProof) error {
	proofBytes, err := aggregationProofToBytes(ap)
	if err != nil {
		return errors.Errorf("failed to marshal aggregation proof: %w", err)
	}
	return r.doUpdateInTx(ctx, "writeAggregationProof", func(ctx context.Context) error {
		txn := getTxn(ctx)

		valueKey := keyAggregationProof(requestID)

		_, err := txn.Get(valueKey)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get aggregation proof: %w", err)
		}
		if err == nil {
			return errors.Errorf("aggregation proof already exists: %w", entity.ErrEntityAlreadyExist)
		}

		if err = txn.Set(valueKey, proofBytes); err != nil {
			return errors.Errorf("failed to store aggregation proof: %w", err)
		}

		reqIDEpochKey := keyRequestIDEpoch(ap.Epoch, requestID)

		_, err = txn.Get(reqIDEpochKey)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get request id epoch link: %w", err)
		}
		if err == nil {
			return nil
		}

		if err = txn.Set(reqIDEpochKey, []byte{}); err != nil {
			return errors.Errorf("failed to store request id epoch link: %w", err)
		}

		return nil
	})
}

func (r *Repository) GetAggregationProof(ctx context.Context, requestID common.Hash) (symbiotic.AggregationProof, error) {
	var ap symbiotic.AggregationProof

	return ap, r.doViewInTx(ctx, "GetAggregationProof", func(ctx context.Context) error {
		txn := getTxn(ctx)
		key := keyAggregationProof(requestID)
		item, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no aggregation proof found for request id %s: %w", requestID.Hex(), entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get aggregation proof: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy aggregation proof value: %w", err)
		}

		ap, err = bytesToAggregationProof(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal aggregation proof: %w", err)
		}

		return nil
	})
}

// GetAggregationProofsByEpoch returns one page of aggregation proofs for the
// given epoch, paginated via opaque cursor `from` (32-byte requestID raw).
// Iterates the request_id_epoch index (sorted by requestID within epoch);
// nextFrom == nil signals the last page.
func (r *Repository) GetAggregationProofsByEpoch(
	ctx context.Context,
	epoch symbiotic.Epoch,
	pageSize int,
	from []byte,
) ([]symbiotic.AggregationProof, []byte, error) {
	fromHash, err := repoutil.DecodeHashCursor(from)
	if err != nil {
		return nil, nil, err
	}

	var (
		proofs   []symbiotic.AggregationProof
		lastID   common.Hash
		moreLeft bool
	)

	err = r.doViewInTx(ctx, "GetAggregationProofsByEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)
		prefix := keyRequestIDEpochPrefix(epoch)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = keyRequestIDEpochAll()
		opts.PrefetchValues = false // index entries are key-only; the proof value comes from a separate Get.
		it := txn.NewIterator(opts)
		defer it.Close()

		seekKey := prefix
		if fromHash != (common.Hash{}) {
			seekKey = keyRequestIDEpoch(epoch, fromHash)
		}

		it.Seek(seekKey)
		if fromHash != (common.Hash{}) && it.ValidForPrefix(prefix) && bytes.Equal(it.Item().Key(), seekKey) {
			it.Next()
		}

		for ; it.ValidForPrefix(prefix); it.Next() {
			if pageSize > 0 && len(proofs) >= pageSize {
				moreLeft = true
				return nil
			}
			id, err := extractRequestIDFromEpochKey(it.Item().Key())
			if err != nil {
				slog.ErrorContext(ctx, errCorruptedRequestIDEpochLink.Error(), "key", string(it.Item().Key()))
				continue
			}
			proof, err := getAggregationProofByEpochFromItem(txn, it)
			if err != nil {
				if errors.Is(err, errCorruptedRequestIDEpochLink) {
					slog.ErrorContext(ctx, errCorruptedRequestIDEpochLink.Error(), "key", string(it.Item().Key()))
					continue
				}
				return err
			}
			proofs = append(proofs, proof)
			lastID = id
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if !moreLeft {
		return proofs, nil, nil
	}
	return proofs, repoutil.EncodeHashCursor(lastID), nil
}

func getAggregationProofByEpochFromItem(txn *badger.Txn, it *badger.Iterator) (symbiotic.AggregationProof, error) {
	key := it.Item().Key()

	requestID, err := extractRequestIDFromEpochKey(key)
	if err != nil {
		return symbiotic.AggregationProof{}, errors.Join(errCorruptedRequestIDEpochLink, err)
	}

	item, err := txn.Get(keyAggregationProof(requestID))
	if err != nil {
		return symbiotic.AggregationProof{}, errors.Errorf("failed to get aggregation proof: %w", err)
	}

	value, err := item.ValueCopy(nil)
	if err != nil {
		return symbiotic.AggregationProof{}, errors.Errorf("failed to copy aggregation proof: %w", err)
	}

	proof, err := bytesToAggregationProof(value)
	if err != nil {
		return symbiotic.AggregationProof{}, errors.Errorf("failed to unmarshal aggregation proof: %w", err)
	}

	return proof, nil
}

var (
	aggregationProofToBytes = codec.AggregationProofToBytes
	bytesToAggregationProof = codec.BytesToAggregationProof
)

func (r *Repository) saveAggregationProofPending(ctx context.Context, requestID common.Hash, epoch symbiotic.Epoch) error {
	return r.doUpdateInTx(ctx, "saveAggregationProofPending", func(ctx context.Context) error {
		txn := getTxn(ctx)
		pendingKey := keyAggregationProofPending(epoch, requestID)

		_, err := txn.Get(pendingKey)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to check pending aggregation proof: %w", err)
		}
		if err == nil {
			return errors.Errorf("pending aggregation proof already exists: %w", entity.ErrEntityAlreadyExist)
		}

		if err := txn.Set(pendingKey, []byte{}); err != nil {
			return errors.Errorf("failed to store pending aggregation proof: %w", err)
		}
		return nil
	})
}

func (r *Repository) removeAggregationProofPending(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) error {
	return r.doUpdateInTx(ctx, "removeAggregationProofPending", func(ctx context.Context) error {
		txn := getTxn(ctx)
		pendingKey := keyAggregationProofPending(epoch, requestID)

		_, err := txn.Get(pendingKey)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("pending aggregation proof not found for epoch %d and request id %s: %w", epoch, requestID.Hex(), entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to check pending aggregation proof: %w", err)
		}

		if err := txn.Delete(pendingKey); err != nil {
			return errors.Errorf("failed to delete pending aggregation proof: %w", err)
		}
		return nil
	})
}

func (r *Repository) RemoveAggregationProofPending(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) error {
	return r.doUpdateInTxWithLock(ctx, "RemoveAggregationProofPending", func(ctx context.Context) error {
		return r.removeAggregationProofPending(ctx, epoch, requestID)
	}, &r.requestIDMutexMap, requestID)
}

func (r *Repository) GetSignatureRequestsWithoutAggregationProof(ctx context.Context, epoch symbiotic.Epoch, limit int, lastHash common.Hash) ([]symbiotic.SignatureRequestWithID, error) {
	var requests []symbiotic.SignatureRequestWithID

	return requests, r.doViewInTx(ctx, "GetSignatureRequestsWithoutAggregationProof", func(ctx context.Context) error {
		txn := getTxn(ctx)

		// Iterate through pending aggregation proof markers
		prefix := keyAggregationProofPendingEpochPrefix(epoch)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		opts.PrefetchValues = false // We don't need the values, just the keys
		it := txn.NewIterator(opts)
		defer it.Close()

		seekKey := prefix
		if lastHash != (common.Hash{}) {
			// Subsequent pages: seek to the record after lastHash
			seekKey = keyAggregationProofPending(epoch, lastHash)
		}

		count := 0
		it.Seek(seekKey)
		// If we're seeking from a specific hash and positioned exactly on that key, skip it (already returned in previous page)
		if lastHash != (common.Hash{}) && it.ValidForPrefix(prefix) && bytes.Equal(it.Item().Key(), seekKey) {
			it.Next()
		}

		for ; it.ValidForPrefix(prefix); it.Next() {
			// Stop if we've reached the limit
			if limit > 0 && count >= limit {
				break
			}

			requestID, err := extractRequestIDFromEpochDelimitedKey(it.Item().Key(), aggregationProofPendingPrefix)
			if err != nil {
				return err
			}

			// Get the actual signature request
			sigReqKey := keySignatureRequest(epoch, requestID)
			sigReqItem, err := txn.Get(sigReqKey)
			if err != nil {
				if errors.Is(err, badger.ErrKeyNotFound) {
					// This shouldn't happen - pending marker exists but signature request doesn't
					// Skip this entry and continue
					continue
				}
				return errors.Errorf("failed to get signature request for hash %s: %w", requestID.Hex(), err)
			}

			value, err := sigReqItem.ValueCopy(nil)
			if err != nil {
				return errors.Errorf("failed to copy signature request value: %w", err)
			}

			req, err := bytesToSignatureRequest(value)
			if err != nil {
				return errors.Errorf("failed to unmarshal signature request: %w", err)
			}

			requests = append(requests, symbiotic.SignatureRequestWithID{
				SignatureRequest: req,
				RequestID:        requestID,
			})
			count++
		}

		return nil
	})
}
