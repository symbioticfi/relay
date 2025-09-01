package entity

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
)

type SignatureMap struct {
	RequestHash         common.Hash
	Epoch               uint64
	ActiveValidatorsMap map[common.Address]struct{}
	IsPresent           map[common.Address]struct{}
	QuorumThreshold     VotingPower
	TotalVotingPower    VotingPower
	CurrentVotingPower  VotingPower
}

func NewSignatureMap(requestHash common.Hash, vs ValidatorSet) SignatureMap {
	activeValidators := vs.Validators.GetActiveValidators()
	m := make(map[common.Address]struct{}, len(activeValidators))
	totalVotingPower := big.NewInt(0)
	for _, validator := range activeValidators {
		m[validator.Operator] = struct{}{}
		totalVotingPower = new(big.Int).Add(totalVotingPower, validator.VotingPower.Int)
	}

	return SignatureMap{
		RequestHash:         requestHash,
		Epoch:               vs.Epoch,
		ActiveValidatorsMap: m,
		IsPresent:           make(map[common.Address]struct{}),
		QuorumThreshold:     vs.QuorumThreshold,
		TotalVotingPower:    ToVotingPower(totalVotingPower),
		CurrentVotingPower:  ToVotingPower(big.NewInt(0)),
	}
}

func (vm *SignatureMap) SetValidatorPresent(v Validator) error {
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

func (vm *SignatureMap) ThresholdReached() bool {
	return vm.CurrentVotingPower.Cmp(vm.QuorumThreshold.Int) >= 0
}
