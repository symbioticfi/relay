package badger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

func keySignatureRequest(epoch entity.Epoch, reqHash common.Hash) []byte {
	return []byte(fmt.Sprintf("signature_request:%d:%s", epoch, reqHash.Hex()))
}

func keySignatureRequestEpochPrefix(epoch entity.Epoch) []byte {
	return []byte(fmt.Sprintf("signature_request:%d:", epoch))
}

func keySignatureRequestHashIndex(reqHash common.Hash) []byte {
	return []byte(fmt.Sprintf("signature_request_hash:%s", reqHash.Hex()))
}

func (r *Repository) SaveSignatureRequest(_ context.Context, req entity.SignatureRequest) error {
	bytes, err := signatureRequestToBytes(req)
	if err != nil {
		return errors.Errorf("failed to marshal signature request: %w", err)
	}

	return r.db.Update(func(txn *badger.Txn) error {
		primaryKey := keySignatureRequest(req.RequiredEpoch, req.Hash())
		hashIndexKey := keySignatureRequestHashIndex(req.Hash())

		// Check if already exists via hash index
		_, err := txn.Get(hashIndexKey)
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get signature request hash index: %w", err)
		}
		if err == nil {
			return errors.Errorf("signature request already exists: %w", entity.ErrEntityAlreadyExist)
		}

		// Store primary record
		err = txn.Set(primaryKey, bytes)
		if err != nil {
			return errors.Errorf("failed to store signature request: %w", err)
		}

		// Store hash index pointing to primary key
		err = txn.Set(hashIndexKey, primaryKey)
		if err != nil {
			return errors.Errorf("failed to store signature request hash index: %w", err)
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

func (r *Repository) GetSignatureRequest(_ context.Context, reqHash common.Hash) (entity.SignatureRequest, error) {
	var req entity.SignatureRequest

	return req, r.db.View(func(txn *badger.Txn) error {
		// Get primary key from hash index
		hashIndexItem, err := txn.Get(keySignatureRequestHashIndex(reqHash))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no signature request found for hash %s: %w", reqHash.String(), entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get signature request hash index: %w", err)
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

func (r *Repository) GetSignatureRequestsByEpoch(_ context.Context, epoch entity.Epoch, limit int, lastHash common.Hash) ([]entity.SignatureRequest, error) {
	var requests []entity.SignatureRequest

	return requests, r.db.View(func(txn *badger.Txn) error {
		prefix := keySignatureRequestEpochPrefix(epoch)
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		seekKey := prefix // First page: seek to epoch prefix
		if lastHash != (common.Hash{}) {
			// Subsequent pages: seek to the record after lastHash
			seekKey = keySignatureRequest(epoch, lastHash)
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

type signatureRequestDTO struct {
	KeyTag        uint8  `json:"key_tag"`
	RequiredEpoch uint64 `json:"required_epoch"`
	Message       []byte `json:"message"`
}
