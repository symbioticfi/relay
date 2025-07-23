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

func keyAggregationProof(reqHash common.Hash) []byte {
	return []byte(fmt.Sprintf("aggregation_proof:%s", reqHash.Hex()))
}

func (r *Repository) SaveAggregationProof(_ context.Context, reqHash common.Hash, ap entity.AggregationProof) error {
	bytes, err := aggregationProofToBytes(ap)
	if err != nil {
		return errors.Errorf("failed to marshal aggregation proof: %w", err)
	}

	return r.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(keyAggregationProof(reqHash))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("failed to get aggregation proof: %w", err)
		}
		if err == nil {
			return errors.Errorf("aggregation proof already exists: %w", entity.ErrEntityAlreadyExist)
		}

		err = txn.Set(keyAggregationProof(reqHash), bytes)
		if err != nil {
			return errors.Errorf("failed to store aggregation proof: %w", err)
		}
		return nil
	})
}

func (r *Repository) GetAggregationProof(_ context.Context, reqHash common.Hash) (entity.AggregationProof, error) {
	var ap entity.AggregationProof

	return ap, r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(keyAggregationProof(reqHash))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Errorf("no aggregation proof found for hash %s: %w", reqHash.Hex(), entity.ErrEntityNotFound)
			}
			return errors.Errorf("failed to get aggregation proof: %w", err)
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Errorf("failed to copy network config value: %w", err)
		}

		ap, err = bytesToAggregationProof(value)
		if err != nil {
			return errors.Errorf("failed to unmarshal aggregation proof: %w", err)
		}

		return nil
	})
}

type aggregationProofDTO struct {
	VerificationType uint32 `json:"verification_type"`
	MessageHash      []byte `json:"message_hash"`
	Proof            []byte `json:"proof"`
}

func aggregationProofToBytes(ap entity.AggregationProof) ([]byte, error) {
	dto := aggregationProofDTO{
		VerificationType: uint32(ap.VerificationType),
		MessageHash:      ap.MessageHash,
		Proof:            ap.Proof,
	}
	data, err := json.Marshal(dto)
	if err != nil {
		return nil, errors.Errorf("failed to marshal aggregation proof: %w", err)
	}

	return data, nil
}

func bytesToAggregationProof(value []byte) (entity.AggregationProof, error) {
	var dto aggregationProofDTO
	if err := json.Unmarshal(value, &dto); err != nil {
		return entity.AggregationProof{}, fmt.Errorf("failed to unmarshal aggregation proof: %w", err)
	}

	return entity.AggregationProof{
		VerificationType: entity.VerificationType(dto.VerificationType),
		MessageHash:      dto.MessageHash,
		Proof:            dto.Proof,
	}, nil
}
