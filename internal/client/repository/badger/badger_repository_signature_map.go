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

func keySignatureMap(reqHash common.Hash) []byte {
	return []byte("signature_map:" + reqHash.Hex())
}

func (r *Repository) UpdateSignatureMap(ctx context.Context, vm entity.SignatureMap) error {
	bytes, err := signatureMapToBytes(vm)
	if err != nil {
		return errors.Errorf("failed to marshal valset signature map: %w", err)
	}

	return r.DoUpdateInTx(ctx, func(ctx context.Context) error {
		key := keySignatureMap(vm.RequestHash)

		err = getTxn(ctx).Set(key, bytes)
		if err != nil {
			return errors.Errorf("failed to store valset signature map: %w", err)
		}
		return nil
	})
}

func (r *Repository) GetSignatureMap(ctx context.Context, reqHash common.Hash) (entity.SignatureMap, error) {
	var vm entity.SignatureMap

	// Create a new read-only transaction
	return vm, r.DoViewInTx(ctx, func(ctx context.Context) error {
		return r.getSignatureMapWithTxn(getTxn(ctx), reqHash, &vm)
	})
}

func (r *Repository) getSignatureMapWithTxn(txn *badger.Txn, reqHash common.Hash, vm *entity.SignatureMap) error {
	key := keySignatureMap(reqHash)

	item, err := txn.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("no signature map found for request %s: %w", reqHash.Hex(), entity.ErrEntityNotFound)
		}
		return errors.Errorf("failed to get signature map: %w", err)
	}

	value, err := item.ValueCopy(nil)
	if err != nil {
		return errors.Errorf("failed to copy signature map value: %w", err)
	}

	result, err := bytesToSignatureMap(value)
	if err != nil {
		return errors.Errorf("failed to unmarshal signature map: %w", err)
	}

	*vm = result
	return nil
}

type signatureMapDTO struct {
	RequestHash        string   `json:"request_hash"`
	Epoch              uint64   `json:"epoch"`
	ActiveValidators   []string `json:"active_validators"`
	PresentValidators  []string `json:"present_validators"`
	QuorumThreshold    *big.Int `json:"quorum_threshold"`
	TotalVotingPower   *big.Int `json:"total_voting_power"`
	CurrentVotingPower *big.Int `json:"current_voting_power"`
}

func signatureMapToBytes(vm entity.SignatureMap) ([]byte, error) {
	activeValidators := make([]string, 0, len(vm.ActiveValidatorsMap))
	for addr := range vm.ActiveValidatorsMap {
		activeValidators = append(activeValidators, addr.Hex())
	}

	presentValidators := make([]string, 0, len(vm.SignedValidatorIndexes))
	for addr := range vm.SignedValidatorIndexes {
		presentValidators = append(presentValidators, addr.Hex())
	}

	dto := signatureMapDTO{
		RequestHash:        vm.RequestHash.Hex(),
		Epoch:              vm.Epoch,
		ActiveValidators:   activeValidators,
		PresentValidators:  presentValidators,
		QuorumThreshold:    vm.QuorumThreshold.Int,
		TotalVotingPower:   vm.TotalVotingPower.Int,
		CurrentVotingPower: vm.CurrentVotingPower.Int,
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

	requestHash := common.HexToHash(dto.RequestHash)

	activeValidators := make(map[common.Address]struct{})
	for _, addrHex := range dto.ActiveValidators {
		activeValidators[common.HexToAddress(addrHex)] = struct{}{}
	}

	presentValidators := make(map[common.Address]struct{})
	for _, addrHex := range dto.PresentValidators {
		presentValidators[common.HexToAddress(addrHex)] = struct{}{}
	}

	return entity.SignatureMap{
		RequestHash:            requestHash,
		Epoch:                  dto.Epoch,
		ActiveValidatorsMap:    activeValidators,
		SignedValidatorIndexes: presentValidators,
		QuorumThreshold:        entity.ToVotingPower(dto.QuorumThreshold),
		TotalVotingPower:       entity.ToVotingPower(dto.TotalVotingPower),
		CurrentVotingPower:     entity.ToVotingPower(dto.CurrentVotingPower),
	}, nil
}
