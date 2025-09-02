package entity

import (
	"math/big"

	"github.com/RoaringBitmap/roaring/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
)

type SignatureMap struct {
	RequestHash            common.Hash
	Epoch                  Epoch
	SignedValidatorsBitmap *roaring.Bitmap
	CurrentVotingPower     VotingPower
	TotalValidators        uint32
}

func NewSignatureMap(requestHash common.Hash, epoch Epoch, totalValidators uint32) SignatureMap {
	return SignatureMap{
		RequestHash:            requestHash,
		Epoch:                  epoch,
		SignedValidatorsBitmap: roaring.New(),
		CurrentVotingPower:     ToVotingPower(big.NewInt(0)),
		TotalValidators:        totalValidators,
	}
}

func (vm *SignatureMap) SetValidatorPresent(activeIndex uint32, votingPower VotingPower) error {
	if activeIndex >= vm.TotalValidators {
		return errors.New("invalid active index")
	}
	if vm.SignedValidatorsBitmap.Contains(activeIndex) {
		return errors.New(ErrEntityAlreadyExist)
	}

	vm.SignedValidatorsBitmap.Add(activeIndex)
	vm.CurrentVotingPower = ToVotingPower(new(big.Int).Add(vm.CurrentVotingPower.Int, votingPower.Int))

	return nil
}

func (vm *SignatureMap) ThresholdReached(quorumThreshold VotingPower) bool {
	return vm.CurrentVotingPower.Cmp(quorumThreshold.Int) >= 0
}

func (vm *SignatureMap) GetMissingValidators() *roaring.Bitmap {
	missing := vm.SignedValidatorsBitmap.Clone()
	missing.FlipInt(0, int(vm.TotalValidators))
	return missing
}

// SaveSignatureParam bundles parameters needed for signature processing with SignatureMap operations
type SaveSignatureParam struct {
	RequestHash      common.Hash
	Key              RawPublicKey
	Signature        SignatureExtended
	ActiveIndex      uint32
	VotingPower      VotingPower
	Epoch            Epoch
	SignatureRequest *SignatureRequest // Optional - used by signer-app, nil for signature-listener
}
