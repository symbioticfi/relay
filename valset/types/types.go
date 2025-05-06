package types

import (
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"middleware-offchain/bls"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Key struct {
	Tag     uint8  `ssz-size:"8"`
	Payload []byte `ssz-max:"64"`
}

type Vault struct {
	ChainId     uint64         `ssz-size:"8"`
	Vault       common.Address `ssz-size:"20"`
	VotingPower *big.Int       `ssz-size:"32"`
}

type Validator struct {
	// Version     uint8          `ssz-size:"1"` TODO: do we need this?
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

// ValidatorSetHeader represents the input for validator set header
type ValidatorSetHeader struct {
	Version                uint8
	ActiveAggregatedKeys   []Key
	TotalActiveVotingPower *big.Int
	ValidatorsSszMRoot     [32]byte
	ExtraData              []byte
}

type G1 [2]*big.Int

func FormatG1(g1 *bls.G1) G1 {
	G1 := G1{new(big.Int), new(big.Int)}

	g1.G1Affine.X.BigInt(G1[0])
	g1.G1Affine.Y.BigInt(G1[1])
	return G1
}

func Hash(v ValidatorSetHeader) ([]byte, error) {
	bytes, err := v.Encode()
	if err != nil {
		return nil, errors.Errorf("failed to hash validator set header: %w", err)
	}

	return crypto.Keccak256(bytes), nil
}

func (v ValidatorSetHeader) Encode() ([]byte, error) {
	arguments := abi.Arguments{
		{
			Type: abi.Type{
				T:    abi.UintTy,
				Size: 8,
			},
		},
		{
			Type: abi.Type{
				T: abi.SliceTy,
				Elem: &abi.Type{
					T:    abi.ArrayTy,
					Size: 2,
					Elem: &abi.Type{T: abi.UintTy, Size: 256}, // G1 points as array of two uint256
				},
			},
		},
		{
			Type: abi.Type{
				T:    abi.UintTy,
				Size: 256,
			},
		},
		{
			Type: abi.Type{
				T:    abi.FixedBytesTy,
				Size: 32,
			},
		},
		{
			Type: abi.Type{
				T: abi.BytesTy,
			},
		},
	}

	return arguments.Pack(v.Version, v.ActiveAggregatedKeys, v.TotalActiveVotingPower, v.ValidatorsSszMRoot, v.ExtraData)
}

type jsonHexBytes []byte

func (j jsonHexBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(j))
}

func (v ValidatorSetHeader) EncodeJSON() ([]byte, error) {
	type keyDTO struct {
		Tag     uint8        `json:"tag"`
		Payload jsonHexBytes `json:"payload"`
	}
	type valSetHeaderDTO struct {
		Version                uint8        `json:"version"`
		ActiveAggregatedKeys   []keyDTO     `json:"activeAggregatedKeys"`
		TotalActiveVotingPower *big.Int     `json:"total_active_voting_power"`
		ValidatorsSszMRoot     jsonHexBytes `json:"validatorsSszMRoot"`
		ExtraData              jsonHexBytes `json:"extraData"`
	}

	valSetHeader := valSetHeaderDTO{
		Version:                v.Version,
		ActiveAggregatedKeys:   lo.Map(v.ActiveAggregatedKeys, func(k Key, _ int) keyDTO { return keyDTO{Tag: k.Tag, Payload: k.Payload} }),
		TotalActiveVotingPower: v.TotalActiveVotingPower,
		ValidatorsSszMRoot:     v.ValidatorsSszMRoot[:],
		ExtraData:              v.ExtraData,
	}

	data, err := json.Marshal(&valSetHeader)
	if err != nil {
		return nil, errors.Errorf("failed to marshal validator set header: %w", err)
	}
	return data, nil
}
