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

func keySignature(signatureTargetId common.Hash, validatorIndex uint32) []byte {
	return []byte(fmt.Sprintf("signature:%s:%010d", signatureTargetId.Hex(), validatorIndex))
}

// keySignaturePrefix returns prefix for all signatures of a signature target
func keySignaturePrefix(signatureTargetId common.Hash) []byte {
	return []byte("signature:" + signatureTargetId.Hex() + ":")
}

func (r *Repository) SaveSignature(ctx context.Context, validatorIndex uint32, sig entity.SignatureExtended) error {
	txn := getTxn(ctx)
	if txn == nil {
		return errors.New("no transaction found in context, use signature processor in order to store signatures")
	}

	bytes, err := signatureToBytes(sig)
	if err != nil {
		return errors.Errorf("failed to marshal signature: %w", err)
	}

	key := keySignature(sig.SignatureTargetID(), validatorIndex)
	err = txn.Set(key, bytes)
	if err != nil {
		return errors.Errorf("failed to store signature: %w", err)
	}
	return nil
}

func (r *Repository) GetAllSignatures(ctx context.Context, signatureTargetID common.Hash) ([]entity.SignatureExtended, error) {
	var signatures []entity.SignatureExtended

	return signatures, r.DoViewInTx(ctx, func(ctx context.Context) error {
		txn := getTxn(ctx)
		prefix := keySignaturePrefix(signatureTargetID)
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

func (r *Repository) GetSignatureByIndex(ctx context.Context, signatureTargetID common.Hash, validatorIndex uint32) (entity.SignatureExtended, error) {
	var signature entity.SignatureExtended

	err := r.DoViewInTx(ctx, func(ctx context.Context) error {
		txn := getTxn(ctx)
		key := keySignature(signatureTargetID, validatorIndex)

		item, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return entity.ErrEntityNotFound
			}
			return errors.Errorf("failed to get signature: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy signature value: %w", err)
		}

		sig, err := bytesToSignature(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal signature: %w", err)
		}

		signature = sig
		return nil
	})

	return signature, err
}

type signatureDTO struct {
	MessageHash []byte `json:"message_hash"`
	KeyTag      uint8  `json:"key_tag"`
	Epoch       uint64 `json:"epoch"`
	Signature   []byte `json:"signature"`
	PublicKey   []byte `json:"public_key"`
}

func signatureToBytes(sig entity.SignatureExtended) ([]byte, error) {
	dto := signatureDTO{
		MessageHash: sig.MessageHash,
		KeyTag:      uint8(sig.KeyTag),
		Epoch:       uint64(sig.Epoch),
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
		return entity.SignatureExtended{}, errors.Errorf("failed to unmarshal signature: %w", err)
	}

	return entity.SignatureExtended{
		MessageHash: dto.MessageHash,
		KeyTag:      entity.KeyTag(dto.KeyTag),
		Epoch:       entity.Epoch(dto.Epoch),
		PublicKey:   dto.PublicKey,
		Signature:   dto.Signature,
	}, nil
}
