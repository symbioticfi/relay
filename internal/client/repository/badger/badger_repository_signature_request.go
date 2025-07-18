package badger

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

func keySignatureRequest(reqHash common.Hash) []byte {
	return []byte(fmt.Sprintf("signature_request:%s", reqHash.Hex()))
}

func (r *Repository) SaveSignatureRequest(_ context.Context, req entity.SignatureRequest) error {
	bytes, err := signatureRequestToBytes(req)
	if err != nil {
		return errors.Errorf("failed to marshal signature request: %w", err)
	}

	return r.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(keySignatureRequest(req.Hash()))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get signature request: %w", err)
		}
		if err == nil {
			return errors.Errorf("signature request already exists: %w", entity.ErrEntityAlreadyExist)
		}

		err = txn.Set(keySignatureRequest(req.Hash()), bytes)
		if err != nil {
			return errors.Errorf("failed to store signature request: %w", err)
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
		return entity.SignatureRequest{}, fmt.Errorf("failed to unmarshal signature request: %w", err)
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
		item, err := txn.Get(keySignatureRequest(reqHash))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no signature request found for hash %s: %w", reqHash.String(), entity.ErrEntityNotFound)
			}
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

type signatureRequestDTO struct {
	KeyTag        uint8  `json:"key_tag"`
	RequiredEpoch uint64 `json:"required_epoch"`
	Message       []byte `json:"message"`
}
