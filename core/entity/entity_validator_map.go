package entity

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
)

type ValidatorMap struct {
	RequestHash         common.Hash
	Epoch               uint64
	ActiveValidatorsMap map[common.Address]struct{}
	IsPresent           map[common.Address]struct{}
	QuorumThreshold     VotingPower
	TotalVotingPower    VotingPower
	CurrentVotingPower  VotingPower
}

func NewValidatorMap(requestHash common.Hash, vs ValidatorSet) ValidatorMap {
	activeValidators := vs.Validators.GetActiveValidators()
	m := make(map[common.Address]struct{}, len(activeValidators))
	totalVotingPower := big.NewInt(0)
	for _, validator := range activeValidators {
		m[validator.Operator] = struct{}{}
		totalVotingPower = new(big.Int).Add(totalVotingPower, validator.VotingPower.Int)
	}

	return ValidatorMap{
		RequestHash:         requestHash,
		Epoch:               vs.Epoch,
		ActiveValidatorsMap: m,
		IsPresent:           make(map[common.Address]struct{}),
		QuorumThreshold:     vs.QuorumThreshold,
		TotalVotingPower:    ToVotingPower(totalVotingPower),
		CurrentVotingPower:  ToVotingPower(big.NewInt(0)),
	}
}

func (vm *ValidatorMap) SetValidatorPresent(v Validator) error {
	if _, ok := vm.ActiveValidatorsMap[v.Operator]; !ok {
		return errors.New(ErrValidatorNotFound)
	}
	if _, ok := vm.IsPresent[v.Operator]; ok {
		return errors.New(ErrEntityAlreadyExist)
	}
	vm.IsPresent[v.Operator] = struct{}{}
	vm.CurrentVotingPower = ToVotingPower(new(big.Int).Add(vm.CurrentVotingPower.Int, v.VotingPower.Int))
	return nil
}

func (vm *ValidatorMap) ThresholdReached() bool {
	return vm.CurrentVotingPower.Cmp(vm.QuorumThreshold.Int) >= 0
}
