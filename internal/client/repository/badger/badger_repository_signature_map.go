package badger

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

func keySignatureMap(signatureTargetId common.Hash) []byte {
	return []byte("signature_map:" + signatureTargetId.Hex())
}

func (r *Repository) UpdateSignatureMap(ctx context.Context, vm entity.SignatureMap) error {
	bytes, err := signatureMapToBytes(vm)
	if err != nil {
		return errors.Errorf("failed to marshal valset signature map: %w", err)
	}

	return r.DoUpdateInTx(ctx, func(ctx context.Context) error {
		key := keySignatureMap(vm.SignatureTargetID)

		err = getTxn(ctx).Set(key, bytes)
		if err != nil {
			return errors.Errorf("failed to store valset signature map: %w", err)
		}
		return nil
	})
}

func (r *Repository) GetSignatureMap(ctx context.Context, signatureTargetID common.Hash) (entity.SignatureMap, error) {
	var vm entity.SignatureMap

	// Create a new read-only transaction
	return vm, r.DoViewInTx(ctx, func(ctx context.Context) error {
		var err error
		vm, err = r.getSignatureMap(ctx, signatureTargetID)
		return err
	})
}

func (r *Repository) getSignatureMap(ctx context.Context, signatureTargetId common.Hash) (entity.SignatureMap, error) {
	key := keySignatureMap(signatureTargetId)

	txn := getTxn(ctx)
	item, err := txn.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return entity.SignatureMap{}, errors.Errorf("no signature map found for signature target %s: %w", signatureTargetId.Hex(), entity.ErrEntityNotFound)
		}
		return entity.SignatureMap{}, errors.Errorf("failed to get signature map: %w", err)
	}

	value, err := item.ValueCopy(nil)
	if err != nil {
		return entity.SignatureMap{}, errors.Errorf("failed to copy signature map value: %w", err)
	}

	vm, err := bytesToSignatureMap(value)
	if err != nil {
		return entity.SignatureMap{}, errors.Errorf("failed to unmarshal signature map: %w", err)
	}

	return vm, nil
}

type signatureMapDTO struct {
	SignatureTargetID          string   `json:"signature_target_id"`
	Epoch                      uint64   `json:"epoch"`
	SignedValidatorsBitmapData []byte   `json:"signed_validators_bitmap"`
	CurrentVotingPower         *big.Int `json:"current_voting_power"`
	TotalValidators            uint32   `json:"total_validators"`
}

func signatureMapToBytes(vm entity.SignatureMap) ([]byte, error) {
	bitmapBytes, err := vm.SignedValidatorsBitmap.ToBytes()
	if err != nil {
		return nil, errors.Errorf("failed to serialize roaring bitmap: %w", err)
	}

	dto := signatureMapDTO{
		SignatureTargetID:          vm.SignatureTargetID.Hex(),
		Epoch:                      uint64(vm.Epoch),
		SignedValidatorsBitmapData: bitmapBytes,
		CurrentVotingPower:         vm.CurrentVotingPower.Int,
		TotalValidators:            vm.TotalValidators,
	}

	data, err := json.Marshal(dto)
	if err != nil {
		return nil, errors.Errorf("failed to marshal valset signature map: %w", err)
	}
	return data, nil
}

func bytesToSignatureMap(data []byte) (entity.SignatureMap, error) {
	var dto signatureMapDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return entity.SignatureMap{}, errors.Errorf("failed to unmarshal signature map: %w", err)
	}

	signatureTargetId := common.HexToHash(dto.SignatureTargetID)

	bitmap, err := entity.BitmapFromBytes(dto.SignedValidatorsBitmapData)
	if err != nil {
		return entity.SignatureMap{}, errors.Errorf("failed to deserialize bitmap: %w", err)
	}

	return entity.SignatureMap{
		SignatureTargetID:      signatureTargetId,
		Epoch:                  entity.Epoch(dto.Epoch),
		SignedValidatorsBitmap: bitmap,
		CurrentVotingPower:     entity.ToVotingPower(dto.CurrentVotingPower),
		TotalValidators:        dto.TotalValidators,
	}, nil
}
