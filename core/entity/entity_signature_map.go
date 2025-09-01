package entity

import (
	"math/big"

	"github.com/RoaringBitmap/roaring/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
)

type SignatureMap struct {
	RequestHash            common.Hash
	Epoch                  uint64
	SignedValidatorsBitmap *roaring.Bitmap
	QuorumThreshold        VotingPower
	TotalVotingPower       VotingPower
	CurrentVotingPower     VotingPower
}

func NewSignatureMap(requestHash common.Hash, vs ValidatorSet) SignatureMap {
	activeValidators := vs.Validators.GetActiveValidators()
	totalVotingPower := big.NewInt(0)
	for _, validator := range activeValidators {
		totalVotingPower = new(big.Int).Add(totalVotingPower, validator.VotingPower.Int)
	}

	return SignatureMap{
		RequestHash:            requestHash,
		Epoch:                  vs.Epoch,
		SignedValidatorsBitmap: roaring.New(),
		QuorumThreshold:        vs.QuorumThreshold,
		TotalVotingPower:       ToVotingPower(totalVotingPower),
		CurrentVotingPower:     ToVotingPower(big.NewInt(0)),
	}
}

func (vm *SignatureMap) SetValidatorPresent(activeIndex int, votingPower VotingPower) error {
	if vm.SignedValidatorsBitmap.Contains(uint32(activeIndex)) {
		return errors.New(ErrEntityAlreadyExist)
	}
	vm.SignedValidatorsBitmap.Add(uint32(activeIndex))
	vm.CurrentVotingPower = ToVotingPower(new(big.Int).Add(vm.CurrentVotingPower.Int, votingPower.Int))
	return nil
}

func (vm *SignatureMap) ThresholdReached() bool {
	return vm.CurrentVotingPower.Cmp(vm.QuorumThreshold.Int) >= 0
}
