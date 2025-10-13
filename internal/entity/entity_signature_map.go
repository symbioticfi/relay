package entity

import (
	"math/big"

	"github.com/RoaringBitmap/roaring/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type SignatureMap struct {
	RequestID              common.Hash
	Epoch                  symbiotic.Epoch
	SignedValidatorsBitmap Bitmap
	CurrentVotingPower     symbiotic.VotingPower
	TotalValidators        uint32
}

func NewSignatureMap(requestID common.Hash, epoch symbiotic.Epoch, totalValidators uint32) SignatureMap {
	return SignatureMap{
		RequestID:              requestID,
		Epoch:                  epoch,
		SignedValidatorsBitmap: NewBitmap(),
		CurrentVotingPower:     symbiotic.ToVotingPower(big.NewInt(0)),
		TotalValidators:        totalValidators,
	}
}

func (vm *SignatureMap) SetValidatorPresent(activeIndex uint32, votingPower symbiotic.VotingPower) error {
	if activeIndex >= vm.TotalValidators {
		return errors.New("invalid active index")
	}
	if vm.SignedValidatorsBitmap.Contains(activeIndex) {
		return errors.New(ErrEntityAlreadyExist)
	}

	vm.SignedValidatorsBitmap.Add(activeIndex)
	vm.CurrentVotingPower = symbiotic.ToVotingPower(new(big.Int).Add(vm.CurrentVotingPower.Int, votingPower.Int))

	return nil
}

func (vm *SignatureMap) ThresholdReached(quorumThreshold symbiotic.VotingPower) bool {
	return vm.CurrentVotingPower.Cmp(quorumThreshold.Int) >= 0
}

func (vm *SignatureMap) GetMissingValidators() Bitmap {
	missing := vm.SignedValidatorsBitmap.Clone()
	missing.FlipInt(0, int(vm.TotalValidators))
	return Bitmap{Bitmap: missing}
}

type Bitmap struct {
	*roaring.Bitmap
}

func NewBitmap() Bitmap {
	return Bitmap{Bitmap: roaring.New()}
}

func NewBitmapOf(dat ...uint32) Bitmap {
	return Bitmap{Bitmap: roaring.BitmapOf(dat...)}
}

func BitmapFromBytes(b []byte) (Bitmap, error) {
	bitmap := roaring.New()
	if _, err := bitmap.FromBuffer(b); err != nil {
		return Bitmap{}, err
	}

	return Bitmap{Bitmap: bitmap}, nil
}
