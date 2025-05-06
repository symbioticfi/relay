package types

import (
	"github.com/karalabe/ssz"
)

func (key *Key) SizeSSZ(_ *ssz.Sizer) uint32 {
	return 1 + 32
}

func (key *Key) DefineSSZ(codec *ssz.Codec) {
	ssz.DefineUint8(codec, &key.Tag)
	ssz.DefineStaticBytes(codec, &key.PayloadHash)
}

func (vault *Vault) SizeSSZ(_ *ssz.Sizer) uint32 {
	return 8 + 20 + 32
}

func (vault *Vault) DefineSSZ(codec *ssz.Codec) {
	ssz.DefineUint64(codec, &vault.ChainId)
	ssz.DefineStaticBytes(codec, &vault.Vault)
	ssz.DefineUint256BigInt(codec, &vault.VotingPower)
}

func (validator *Validator) SizeSSZ(siz *ssz.Sizer, fixed bool) (size uint32) {
	size = 20 + 32 + 1 + 4 + 4
	if fixed {
		return size
	}
	size += ssz.SizeSliceOfStaticObjects(siz, validator.Keys)
	size += ssz.SizeSliceOfStaticObjects(siz, validator.Vaults)
	return size
}

func (validator *Validator) DefineSSZ(codec *ssz.Codec) {
	// ssz.DefineUint8(codec, &validator.Version)
	ssz.DefineStaticBytes(codec, &validator.Operator)
	ssz.DefineUint256BigInt(codec, &validator.VotingPower)
	ssz.DefineBool(codec, &validator.IsActive)
	ssz.DefineSliceOfStaticObjectsOffset(codec, &validator.Keys, 128)
	ssz.DefineSliceOfStaticObjectsOffset(codec, &validator.Vaults, 32)
	ssz.DefineSliceOfStaticObjectsOffset(codec, &validator.Keys, 128)
	ssz.DefineSliceOfStaticObjectsContent(codec, &validator.Vaults, 32)
}

func (valSet *ValidatorSet) SizeSSZ(siz *ssz.Sizer, fixed bool) (size uint32) {
	size = 4
	if fixed {
		return size
	}
	size += ssz.SizeSliceOfDynamicObjects(siz, valSet.Validators)
	return size
}

func (valSet *ValidatorSet) DefineSSZ(codec *ssz.Codec) {
	ssz.DefineSliceOfDynamicObjectsOffset(codec, &valSet.Validators, 1048576)
	ssz.DefineSliceOfDynamicObjectsContent(codec, &valSet.Validators, 1048576)
}
