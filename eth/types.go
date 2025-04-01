package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Phase represents the different phases of the protocol
type Phase uint64

const (
	IDLE Phase = iota
	COMMIT
	ACCEPT
	FAIL
)

type Key struct {
	Tag     uint8  `ssz-size:"8"`
	Payload []byte `ssz-max:"64"`
}

type Vault struct {
	VaultAddress common.Address `ssz-size:"20"`
	VotingPower  *big.Int       `ssz-size:"32"`
}

type Validator struct {
	Version     uint8          `ssz-size:"8"`
	Operator    common.Address `ssz-size:"20"`
	VotingPower *big.Int       `ssz-size:"32"`
	IsActive    bool           `ssz-size:"1"`
	Keys        []*Key         `ssz-max:"128"`
	Vaults      []*Vault       `ssz-max:"10"`
}

type ValidatorSet struct {
	Version                uint8
	TotalActiveVotingPower *big.Int
	Validators             []*Validator
}
