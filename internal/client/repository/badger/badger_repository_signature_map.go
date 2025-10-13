package badger

import (
	"context"
	"math/big"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/client/repository/badger/proto/v1"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func keySignatureMap(requestID common.Hash) []byte {
	return []byte("signature_map:" + requestID.Hex())
}

func (r *Repository) UpdateSignatureMap(ctx context.Context, vm entity.SignatureMap) error {
	bytes, err := signatureMapToBytes(vm)
	if err != nil {
		return errors.Errorf("failed to marshal valset signature map: %w", err)
	}

	return r.doUpdateInTx(ctx, "UpdateSignatureMap", func(ctx context.Context) error {
		key := keySignatureMap(vm.RequestID)

		err = getTxn(ctx).Set(key, bytes)
		if err != nil {
			return errors.Errorf("failed to store valset signature map: %w", err)
		}
		return nil
	})
}

func (r *Repository) GetSignatureMap(ctx context.Context, requestID common.Hash) (entity.SignatureMap, error) {
	var vm entity.SignatureMap

	// Create a new read-only transaction
	return vm, r.doViewInTx(ctx, "GetSignatureMap", func(ctx context.Context) error {
		var err error
		vm, err = r.getSignatureMap(ctx, requestID)
		return err
	})
}

func (r *Repository) getSignatureMap(ctx context.Context, requestID common.Hash) (entity.SignatureMap, error) {
	key := keySignatureMap(requestID)

	txn := getTxn(ctx)
	item, err := txn.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return entity.SignatureMap{}, errors.Errorf("no signature map found for request id %s: %w", requestID.Hex(), entity.ErrEntityNotFound)
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

func signatureMapToBytes(vm entity.SignatureMap) ([]byte, error) {
	bitmapBytes, err := vm.SignedValidatorsBitmap.ToBytes()
	if err != nil {
		return nil, errors.Errorf("failed to serialize roaring bitmap: %w", err)
	}

	return marshalAndCompress(&v1.SignatureMap{
		RequestId:              vm.RequestID.Bytes(),
		Epoch:                  uint64(vm.Epoch),
		SignedValidatorsBitmap: bitmapBytes,
		CurrentVotingPower:     vm.CurrentVotingPower.String(),
		TotalValidators:        vm.TotalValidators,
	})
}

func bytesToSignatureMap(data []byte) (entity.SignatureMap, error) {
	pb := &v1.SignatureMap{}
	if err := unmarshalAndDecompress(data, pb); err != nil {
		return entity.SignatureMap{}, errors.Errorf("failed to unmarshal signature map: %w", err)
	}

	requestId := common.BytesToHash(pb.GetRequestId())

	bitmap, err := entity.BitmapFromBytes(pb.GetSignedValidatorsBitmap())
	if err != nil {
		return entity.SignatureMap{}, errors.Errorf("failed to deserialize bitmap: %w", err)
	}

	currentVotingPower, ok := new(big.Int).SetString(pb.GetCurrentVotingPower(), 10)
	if !ok {
		return entity.SignatureMap{}, errors.Errorf("failed to parse current voting power: %s", pb.GetCurrentVotingPower())
	}

	return entity.SignatureMap{
		RequestID:              requestId,
		Epoch:                  symbiotic.Epoch(pb.GetEpoch()),
		SignedValidatorsBitmap: bitmap,
		CurrentVotingPower:     symbiotic.ToVotingPower(currentVotingPower),
		TotalValidators:        pb.GetTotalValidators(),
	}, nil
}
