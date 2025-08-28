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

func keyValidatorMap(reqHash common.Hash) []byte {
	return []byte("validator_map:" + reqHash.Hex())
}

func (r *Repository) UpdateValidatorMap(ctx context.Context, vm entity.ValidatorMap) error {
	bytes, err := validatorMapToBytes(vm)
	if err != nil {
		return errors.Errorf("failed to marshal valset validator map: %w", err)
	}

	return r.DoUpdateInTx(ctx, func(ctx context.Context) error {
		key := keyValidatorMap(vm.RequestHash)

		err = getTxn(ctx).Set(key, bytes)
		if err != nil {
			return errors.Errorf("failed to store valset validator map: %w", err)
		}
		return nil
	})
}

func (r *Repository) GetValidatorMap(ctx context.Context, reqHash common.Hash) (entity.ValidatorMap, error) {
	var vm entity.ValidatorMap

	// Create a new read-only transaction
	return vm, r.DoViewInTx(ctx, func(ctx context.Context) error {
		return r.getValidatorMapWithTxn(getTxn(ctx), reqHash, &vm)
	})
}

func (r *Repository) getValidatorMapWithTxn(txn *badger.Txn, reqHash common.Hash, vm *entity.ValidatorMap) error {
	key := keyValidatorMap(reqHash)

	item, err := txn.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return errors.Errorf("no validator map found for request %s: %w", reqHash.Hex(), entity.ErrEntityNotFound)
		}
		return errors.Errorf("failed to get validator map: %w", err)
	}

	value, err := item.ValueCopy(nil)
	if err != nil {
		return errors.Errorf("failed to copy validator map value: %w", err)
	}

	result, err := bytesToValidatorMap(value)
	if err != nil {
		return errors.Errorf("failed to unmarshal validator map: %w", err)
	}

	*vm = result
	return nil
}

type validatorMapDTO struct {
	RequestHash        string   `json:"request_hash"`
	Epoch              uint64   `json:"epoch"`
	ActiveValidators   []string `json:"active_validators"`
	PresentValidators  []string `json:"present_validators"`
	QuorumThreshold    *big.Int `json:"quorum_threshold"`
	TotalVotingPower   *big.Int `json:"total_voting_power"`
	CurrentVotingPower *big.Int `json:"current_voting_power"`
}

func validatorMapToBytes(vm entity.ValidatorMap) ([]byte, error) {
	activeValidators := make([]string, 0, len(vm.ActiveValidatorsMap))
	for addr := range vm.ActiveValidatorsMap {
		activeValidators = append(activeValidators, addr.Hex())
	}

	presentValidators := make([]string, 0, len(vm.IsPresent))
	for addr := range vm.IsPresent {
		presentValidators = append(presentValidators, addr.Hex())
	}

	dto := validatorMapDTO{
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
		return nil, errors.Errorf("failed to marshal valset validator map: %w", err)
	}
	return data, nil
}

func bytesToValidatorMap(data []byte) (entity.ValidatorMap, error) {
	var dto validatorMapDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return entity.ValidatorMap{}, errors.Errorf("failed to unmarshal validator map: %w", err)
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

	return entity.ValidatorMap{
		RequestHash:         requestHash,
		Epoch:               dto.Epoch,
		ActiveValidatorsMap: activeValidators,
		IsPresent:           presentValidators,
		QuorumThreshold:     entity.ToVotingPower(dto.QuorumThreshold),
		TotalVotingPower:    entity.ToVotingPower(dto.TotalVotingPower),
		CurrentVotingPower:  entity.ToVotingPower(dto.CurrentVotingPower),
	}, nil
}
