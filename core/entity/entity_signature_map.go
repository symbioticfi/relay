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
	CurrentVotingPower     VotingPower
}

func NewSignatureMap(requestHash common.Hash, vs ValidatorSet) SignatureMap {
	return SignatureMap{
		RequestHash:            requestHash,
		Epoch:                  vs.Epoch,
		SignedValidatorsBitmap: roaring.New(),
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

func (vm *SignatureMap) ThresholdReached(quorumThreshold VotingPower) bool {
	return vm.CurrentVotingPower.Cmp(quorumThreshold.Int) >= 0
}
