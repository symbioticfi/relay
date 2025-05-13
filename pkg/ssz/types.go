package ssz

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type key struct {
	Tag         uint8 `ssz-size:"1"`
	Payload     []byte
	PayloadHash [32]byte `ssz-size:"32"`
}

type vault struct {
	ChainId     uint64         `ssz-size:"8"`
	Vault       common.Address `ssz-size:"20"`
	VotingPower *big.Int       `ssz-size:"32"`
}

type validator struct {
	Operator    common.Address `ssz-size:"20"`
	VotingPower *big.Int       `ssz-size:"32"`
	IsActive    bool           `ssz-size:"1"`
	Keys        []*key         `ssz-max:"128"`
	Vaults      []*vault       `ssz-max:"32"`
}

type validatorSet struct {
	Version    uint8
	Validators []*validator `ssz-max:"1048576"`
}
