package badger

import (
	"context"

	"github.com/dgraph-io/badger/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/client/repository/codec"
	"github.com/symbioticfi/relay/internal/entity"
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

var (
	signatureMapToBytes = codec.SignatureMapToBytes
	bytesToSignatureMap = codec.BytesToSignatureMap
)
