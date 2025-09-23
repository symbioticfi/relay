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
	SignedValidatorsBitmap SignatureBitmap
	CurrentVotingPower     VotingPower
	TotalValidators        uint32
}

func NewSignatureMap(requestHash common.Hash, epoch Epoch, totalValidators uint32) SignatureMap {
	return SignatureMap{
		RequestHash:            requestHash,
		Epoch:                  epoch,
		SignedValidatorsBitmap: NewSignatureBitmap(),
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

func (vm *SignatureMap) GetMissingValidators() SignatureBitmap {
	missing := vm.SignedValidatorsBitmap.Clone()
	missing.FlipInt(0, int(vm.TotalValidators))
	return SignatureBitmap{Bitmap: missing}
}

// SaveSignatureParam bundles parameters needed for signature processing with SignatureMap operations
type SaveSignatureParam struct {
	KeyTag           KeyTag
	RequestHash      common.Hash
	Signature        SignatureExtended
	Epoch            Epoch
	SignatureRequest *SignatureRequest // Optional
}

type SignatureBitmap struct {
	*roaring.Bitmap
}

func NewSignatureBitmap() SignatureBitmap {
	return SignatureBitmap{Bitmap: roaring.New()}
}

func NewSignatureBitmapOf(dat ...uint32) SignatureBitmap {
	return SignatureBitmap{Bitmap: roaring.BitmapOf(dat...)}
}

func SignatureBitmapFromBytes(b []byte) (SignatureBitmap, error) {
	bitmap := roaring.New()
	if _, err := bitmap.FromBuffer(b); err != nil {
		return SignatureBitmap{}, err
	}

	return SignatureBitmap{Bitmap: bitmap}, nil
}
