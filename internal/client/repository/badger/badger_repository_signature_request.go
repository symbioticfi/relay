package badger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

const (
	keySignatureRequestPendingPrefix = "signature_pending:"
)

func keySignatureRequest(epoch entity.Epoch, requestID common.Hash) []byte {
	return []byte(fmt.Sprintf("signature_request:%d:%s", epoch, requestID.Hex()))
}

func keySignatureRequestEpochPrefix(epoch entity.Epoch) []byte {
	return []byte(fmt.Sprintf("signature_request:%d:", epoch))
}

func keyRequestIDIndex(requestID common.Hash) []byte {
	return []byte(fmt.Sprintf("request_id:%s", requestID.Hex()))
}

func keySignatureRequestPending(epoch entity.Epoch, requestID common.Hash) []byte {
	return []byte(fmt.Sprintf("%v%d:%s", keySignatureRequestPendingPrefix, epoch, requestID.Hex()))
}

// saveSignatureRequestToKey saves a signature request to a specific key
func (r *Repository) saveSignatureRequestToKey(ctx context.Context, req entity.SignatureRequest, key []byte) error {
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

func (r *Repository) SaveSignatureRequest(ctx context.Context, requestID common.Hash, req entity.SignatureRequest) error {
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

func (r *Repository) saveSignatureRequest(ctx context.Context, requestID common.Hash, req entity.SignatureRequest) error {
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

func (r *Repository) saveSignaturePending(ctx context.Context, requestID common.Hash, req entity.SignatureRequest) error {
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

func (r *Repository) RemoveSignaturePending(ctx context.Context, epoch entity.Epoch, requestID common.Hash) error {
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

func signatureRequestToBytes(req entity.SignatureRequest) ([]byte, error) {
	dto := signatureRequestDTO{
		KeyTag:        uint8(req.KeyTag),
		RequiredEpoch: uint64(req.RequiredEpoch),
		Message:       req.Message,
	}
	return json.Marshal(dto)
}

func bytesToSignatureRequest(data []byte) (entity.SignatureRequest, error) {
	var dto signatureRequestDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return entity.SignatureRequest{}, errors.Errorf("failed to unmarshal signature request: %w", err)
	}

	return entity.SignatureRequest{
		KeyTag:        entity.KeyTag(dto.KeyTag),
		RequiredEpoch: entity.Epoch(dto.RequiredEpoch),
		Message:       dto.Message,
	}, nil
}

func (r *Repository) GetSignatureRequest(ctx context.Context, requestID common.Hash) (entity.SignatureRequest, error) {
	var req entity.SignatureRequest

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

// getSignatureRequestsByEpochWithKeys is a generic method for retrieving signature requests by epoch
// using provided prefix and key generation function
func (r *Repository) getSignatureRequestsByEpochWithKeys(
	ctx context.Context,
	epoch entity.Epoch,
	limit int,
	lastHash common.Hash,
	prefix []byte,
	keyFunc func(entity.Epoch, common.Hash) []byte,
) ([]entity.SignatureRequest, error) {
	var requests []entity.SignatureRequest

	return requests, r.doViewInTx(ctx, "getSignatureRequestsByEpochWithKeys", func(ctx context.Context) error {
		txn := getTxn(ctx)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		seekKey := prefix
		if lastHash != (common.Hash{}) {
			// Subsequent pages: seek to the record after lastHash
			seekKey = keyFunc(epoch, lastHash)
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

			item := it.Item()
			value, err := item.ValueCopy(nil)
			if err != nil {
				return errors.Errorf("failed to copy signature request value: %w", err)
			}

			req, err := bytesToSignatureRequest(value)
			if err != nil {
				return errors.Errorf("failed to unmarshal signature request: %w", err)
			}

			requests = append(requests, req)
			count++
		}

		return nil
	})
}

func (r *Repository) GetSignatureRequestsByEpoch(ctx context.Context, epoch entity.Epoch, limit int, lastHash common.Hash) ([]entity.SignatureRequest, error) {
	return r.getSignatureRequestsByEpochWithKeys(
		ctx,
		epoch,
		limit,
		lastHash,
		keySignatureRequestEpochPrefix(epoch),
		keySignatureRequest,
	)
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

			// Extract request id from the pending key: "signature_pending:epoch:hash"
			item := it.Item()
			key := string(item.Key())

			// Find the hash part after the second colon
			parts := strings.Split(key, ":")
			if len(parts) != 3 {
				return errors.Errorf("invalid pending signature key format: %s", key)
			}

			requestIDStr := parts[2]
			requestID := common.HexToHash(requestIDStr)

			requests = append(requests, requestID)
			count++
		}

		return nil
	})
}

type signatureRequestDTO struct {
	KeyTag        uint8  `json:"key_tag"`
	RequiredEpoch uint64 `json:"required_epoch"`
	Message       []byte `json:"message"`
}
