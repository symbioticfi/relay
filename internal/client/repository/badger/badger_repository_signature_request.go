package badger

import (
	"bytes"
	"context"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/client/repository/codec"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

const (
	keySignatureRequestPrefix        = "signature_request:"
	keySignatureRequestPendingPrefix = "signature_pending:"
)

func keySignatureRequest(epoch symbiotic.Epoch, requestID common.Hash) []byte {
	key := epochKeyWithColon(keySignatureRequestPrefix, epoch)
	return append(key, []byte(requestID.Hex())...)
}

func keySignatureRequestEpochPrefix(epoch symbiotic.Epoch) []byte {
	return epochKeyWithColon(keySignatureRequestPrefix, epoch)
}

func keyRequestIDIndex(requestID common.Hash) []byte {
	key := []byte("request_id:")
	return append(key, []byte(requestID.Hex())...)
}

func keySignatureRequestPending(epoch symbiotic.Epoch, requestID common.Hash) []byte {
	key := epochKeyWithColon(keySignatureRequestPendingPrefix, epoch)
	return append(key, []byte(requestID.Hex())...)
}

// saveSignatureRequestToKey saves a signature request to a specific key
func (r *Repository) saveSignatureRequestToKey(ctx context.Context, req symbiotic.SignatureRequest, key []byte) error {
	requestBytes, err := signatureRequestToBytes(req)
	if err != nil {
		return errors.Errorf("failed to marshal signature request: %w", err)
	}

	txn := getTxn(ctx)

	_, err = txn.Get(key)
	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
		return errors.Errorf("failed to check signature request: %w", err)
	}
	if err == nil {
		return errors.Errorf("signature request already exists: %w", entity.ErrEntityAlreadyExist)
	}

	// Store the record
	err = txn.Set(key, requestBytes)
	if err != nil {
		return errors.Errorf("failed to store signature request: %w", err)
	}

	return nil
}

func (r *Repository) SaveSignatureRequest(ctx context.Context, requestID common.Hash, req symbiotic.SignatureRequest) error {
	return r.doUpdateInTx(ctx, "SaveSignatureRequest", func(ctx context.Context) error {
		if err := r.saveSignatureRequest(ctx, requestID, req); err != nil {
			return err
		}

		// Save pending signature for all key tags because we should attempt
		// to sync signatures from all signers even when keytag is non aggregation
		if err := r.saveSignaturePending(ctx, requestID, req); err != nil {
			return errors.Errorf("failed to save signature request to pending collection: %v", err)
		}
		return nil
	})
}

func (r *Repository) saveSignatureRequest(ctx context.Context, requestID common.Hash, req symbiotic.SignatureRequest) error {
	return r.doUpdateInTx(ctx, "saveSignatureRequest", func(ctx context.Context) error {
		primaryKey := keySignatureRequest(req.RequiredEpoch, requestID)
		requestIDIndexKey := keyRequestIDIndex(requestID)

		if err := r.saveSignatureRequestToKey(ctx, req, primaryKey); err != nil {
			return err
		}

		if err := getTxn(ctx).Set(requestIDIndexKey, primaryKey); err != nil {
			return errors.Errorf("failed to store signature request id index: %w", err)
		}

		return nil
	})
}

func (r *Repository) saveSignaturePending(ctx context.Context, requestID common.Hash, req symbiotic.SignatureRequest) error {
	return r.doUpdateInTx(ctx, "saveSignaturePending", func(ctx context.Context) error {
		txn := getTxn(ctx)
		pendingKey := keySignatureRequestPending(req.RequiredEpoch, requestID)

		_, err := txn.Get(pendingKey)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to check pending signature: %w", err)
		}
		if err == nil {
			return errors.Errorf("pending signature already exists: %w", entity.ErrEntityAlreadyExist)
		}

		// Store just a marker (empty value) - we don't need the full request data here
		err = txn.Set(pendingKey, []byte{})
		if err != nil {
			return errors.Errorf("failed to store pending signature: %w", err)
		}
		return nil
	})
}

func (r *Repository) RemoveSignaturePending(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) error {
	return r.doUpdateInTx(ctx, "RemoveSignaturePending", func(ctx context.Context) error {
		txn := getTxn(ctx)
		pendingKey := keySignatureRequestPending(epoch, requestID)

		// Remove from pending collection
		if err := txn.Delete(pendingKey); err != nil {
			return errors.Errorf("failed to remove pending signature: %w", err)
		}

		return nil
	})
}

var (
	signatureRequestToBytes = codec.SignatureRequestToBytes
	bytesToSignatureRequest = codec.BytesToSignatureRequest
)

func (r *Repository) GetSignatureRequest(ctx context.Context, requestID common.Hash) (symbiotic.SignatureRequest, error) {
	var req symbiotic.SignatureRequest

	return req, r.doViewInTx(ctx, "GetSignatureRequest", func(ctx context.Context) error {
		txn := getTxn(ctx)
		// Get primary key from hash index
		hashIndexItem, err := txn.Get(keyRequestIDIndex(requestID))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no signature request found for request id %s: %w", requestID.String(), entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get request id index: %w", err)
		}

		primaryKey, err := hashIndexItem.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy hash index value: %w", err)
		}

		// Get actual data using primary key
		item, err := txn.Get(primaryKey)
		if err != nil {
			return errors.Errorf("failed to get signature request: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy signature request value: %w", err)
		}

		req, err = bytesToSignatureRequest(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal signature request: %w", err)
		}

		return nil
	})
}

// GetSignatureRequestsWithIDByEpoch returns one page of signature requests for
// the given epoch, paginated via opaque cursor `from` (32-byte requestID raw).
// nextFrom == nil signals the last page; invalid `from` returns entity.ErrInvalidCursor.
func (r *Repository) GetSignatureRequestsWithIDByEpoch(
	ctx context.Context,
	epoch symbiotic.Epoch,
	pageSize int,
	from []byte,
) ([]entity.SignatureRequestWithID, []byte, error) {
	fromHash, err := decodeHashCursor(from)
	if err != nil {
		return nil, nil, err
	}

	var (
		requests   []entity.SignatureRequestWithID
		lastID     common.Hash
		filledFull bool
	)

	err = r.doViewInTx(ctx, "GetSignatureRequestsWithIDByEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)
		prefix := keySignatureRequestEpochPrefix(epoch)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		seekKey := prefix
		if fromHash != (common.Hash{}) {
			seekKey = keySignatureRequest(epoch, fromHash)
		}

		it.Seek(seekKey)
		// Skip exact-match cursor item (already returned in previous page).
		if fromHash != (common.Hash{}) && it.ValidForPrefix(prefix) && bytes.Equal(it.Item().Key(), seekKey) {
			it.Next()
		}

		count := 0
		for ; it.ValidForPrefix(prefix); it.Next() {
			if pageSize > 0 && count >= pageSize {
				filledFull = true
				return nil
			}
			item := it.Item()
			requestID, err := extractRequestIDFromEpochDelimitedKey(item.Key(), keySignatureRequestPrefix)
			if err != nil {
				return err
			}
			value, err := item.ValueCopy(nil)
			if err != nil {
				return errors.Errorf("failed to copy signature request value: %w", err)
			}
			req, err := bytesToSignatureRequest(value)
			if err != nil {
				return errors.Errorf("failed to unmarshal signature request: %w", err)
			}
			requests = append(requests, entity.SignatureRequestWithID{
				RequestID:        requestID,
				SignatureRequest: req,
			})
			lastID = requestID
			count++
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if !filledFull {
		return requests, nil, nil
	}
	return requests, encodeHashCursor(lastID), nil
}

// GetSignatureRequestIDsByEpoch returns one page of request IDs for the given
// epoch (keys only — no value materialization). Cursor format same as
// GetSignatureRequestsWithIDByEpoch.
func (r *Repository) GetSignatureRequestIDsByEpoch(
	ctx context.Context,
	epoch symbiotic.Epoch,
	pageSize int,
	from []byte,
) ([]common.Hash, []byte, error) {
	fromHash, err := decodeHashCursor(from)
	if err != nil {
		return nil, nil, err
	}

	var (
		requestIDs []common.Hash
		lastID     common.Hash
		filledFull bool
	)

	err = r.doViewInTx(ctx, "GetSignatureRequestIDsByEpoch", func(ctx context.Context) error {
		txn := getTxn(ctx)
		prefix := keySignatureRequestEpochPrefix(epoch)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		seekKey := prefix
		if fromHash != (common.Hash{}) {
			seekKey = keySignatureRequest(epoch, fromHash)
		}

		it.Seek(seekKey)
		if fromHash != (common.Hash{}) && it.ValidForPrefix(prefix) && bytes.Equal(it.Item().Key(), seekKey) {
			it.Next()
		}

		count := 0
		for ; it.ValidForPrefix(prefix); it.Next() {
			if pageSize > 0 && count >= pageSize {
				filledFull = true
				return nil
			}
			id, err := extractRequestIDFromEpochDelimitedKey(it.Item().Key(), keySignatureRequestPrefix)
			if err != nil {
				return err
			}
			requestIDs = append(requestIDs, id)
			lastID = id
			count++
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if !filledFull {
		return requestIDs, nil, nil
	}
	return requestIDs, encodeHashCursor(lastID), nil
}

func (r *Repository) GetSignaturePending(ctx context.Context, limit int) ([]common.Hash, error) {
	var requests []common.Hash

	return requests, r.doViewInTx(ctx, "GetSignaturePending", func(ctx context.Context) error {
		txn := getTxn(ctx)

		// Iterate through pending signature markers
		opts := badger.DefaultIteratorOptions
		prefix := []byte(keySignatureRequestPendingPrefix)
		opts.Prefix = prefix
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		count := 0
		it.Seek(prefix)

		for ; it.ValidForPrefix(prefix); it.Next() {
			// Stop if we've reached the limit
			if limit > 0 && count >= limit {
				break
			}

			requestID, err := extractRequestIDFromEpochDelimitedKey(it.Item().Key(), keySignatureRequestPendingPrefix)
			if err != nil {
				return err
			}

			requests = append(requests, requestID)
			count++
		}

		return nil
	})
}
