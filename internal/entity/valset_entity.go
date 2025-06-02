package entity

import (
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum/common"
)

type Key struct {
	Tag     uint8
	Payload []byte
}

type Vault struct {
	ChainID     uint64
	Vault       common.Address
	VotingPower *big.Int
}

type Validator struct {
	Operator    common.Address
	VotingPower *big.Int
	IsActive    bool
	Keys        []Key
	Vaults      []Vault
}

type ValidatorSet struct {
	Version                uint8
	Validators             []Validator
	TotalActiveVotingPower *big.Int
}

// Signature signer.sign() -> Signature
type Signature struct {
	MessageHash []byte // scheme depends on KeyTag
	Signature   []byte // parse based on KeyTag
	PublicKey   []byte // parse based on KeyTag
}

// todo ilya make g1 G1 not bytes
func (v ValidatorSet) FindValidatorByKey(g1 []byte) (Validator, bool) {
	for _, validator := range v.Validators {
		for _, key := range validator.Keys {
			if slices.Equal(key.Payload, g1) {
				return validator, true
			}
		}
	}
	return Validator{}, false
}

// ValidatorSetHeader represents the input for validator set header
type ValidatorSetHeader struct {
	Version                uint8
	ActiveAggregatedKeys   []Key
	TotalActiveVotingPower *big.Int
	ValidatorsSszMRoot     [32]byte
	ExtraData              []byte // todo ilya move out from header, Make method MakeValsetHeader of ValidatorSet
}

type CommitValsetHeaderResult struct {
	TxHash common.Hash
}
