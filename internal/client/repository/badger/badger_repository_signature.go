package badger

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

func keySignature(reqHash common.Hash, key []byte) []byte {
	keyHash := crypto.Keccak256Hash(key)
	return []byte("signature:" + reqHash.Hex() + ":" + keyHash.Hex())
}

// keySignaturePrefix returns prefix for all signatures of a request
func keySignaturePrefix(reqHash common.Hash) []byte {
	return []byte("signature:" + reqHash.Hex() + ":")
}

func (r *Repository) SaveSignature(_ context.Context, reqHash common.Hash, inKey []byte, sig entity.SignatureExtended) error {
	bytes, err := signatureToBytes(sig)
	if err != nil {
		return errors.Errorf("failed to marshal signature: %w", err)
	}

	return r.db.Update(func(txn *badger.Txn) error {
		key := keySignature(reqHash, inKey)
		err = txn.Set(key, bytes)
		if err != nil {
			return errors.Errorf("failed to store signature: %w", err)
		}
		return nil
	})
}

func (r *Repository) GetAllSignatures(_ context.Context, reqHash common.Hash) ([]entity.SignatureExtended, error) {
	var signatures []entity.SignatureExtended

	return signatures, r.db.View(func(txn *badger.Txn) error {
		prefix := keySignaturePrefix(reqHash)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			value, err := item.ValueCopy(nil)
			if err != nil {
				return errors.Errorf("failed to copy signature value: %w", err)
			}

			sig, err := bytesToSignature(value)
			if err != nil {
				return errors.Errorf("failed to unmarshal signature: %w", err)
			}

			signatures = append(signatures, sig)
		}

		return nil
	})
}

type signatureDTO struct {
	MessageHash []byte `json:"message_hash"`
	Signature   []byte `json:"signature"`
	PublicKey   []byte `json:"public_key"`
}

func signatureToBytes(sig entity.SignatureExtended) ([]byte, error) {
	dto := signatureDTO{
		MessageHash: sig.MessageHash,
		Signature:   sig.Signature,
		PublicKey:   sig.PublicKey,
	}
	data, err := json.Marshal(dto)
	if err != nil {
		return nil, errors.Errorf("failed to marshal signature: %w", err)
	}
	return data, nil
}

func bytesToSignature(value []byte) (entity.SignatureExtended, error) {
	var dto signatureDTO
	if err := json.Unmarshal(value, &dto); err != nil {
		return entity.SignatureExtended{}, fmt.Errorf("failed to unmarshal signature: %w", err)
	}

	return entity.SignatureExtended{
		MessageHash: dto.MessageHash,
		Signature:   dto.Signature,
		PublicKey:   dto.PublicKey,
	}, nil
}
