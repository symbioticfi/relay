package entity

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"slices"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
)

type Key struct {
	Tag     uint8
	Payload []byte
}

func (k Key) PayloadHash() [32]byte {
	return crypto.Keccak256Hash(k.Payload)
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
	ExtraData              []byte
	Epoch                  *big.Int
	DomainEip712           Eip712Domain
	Subnetwork             []byte
	Timestamp              *big.Int
}

func (v ValidatorSetHeader) Hash() ([]byte, error) {
	bytes, err := v.Encode()
	if err != nil {
		return nil, errors.Errorf("failed to hash validator set header: %w", err)
	}

	return crypto.Keccak256(bytes), nil
}

func (v ValidatorSetHeader) Encode() ([]byte, error) {
	arguments := abi.Arguments{
		{
			Name: "version",
			Type: abi.Type{T: abi.UintTy, Size: 8},
		},
		{
			Name: "activeAggregatedKeys",
			Type: abi.Type{
				T: abi.SliceTy,
				Elem: &abi.Type{
					T: abi.TupleTy,
					TupleElems: []*abi.Type{
						{T: abi.UintTy, Size: 8},
						{T: abi.BytesTy},
					},
					TupleRawNames: []string{"tag", "payload"},
					TupleType:     reflect.TypeOf(Key{}),
				},
			},
		},
		{
			Name: "totalActiveVotingPower",
			Type: abi.Type{T: abi.UintTy, Size: 256},
		},
		{
			Name: "validatorsSszMRoot",
			Type: abi.Type{T: abi.FixedBytesTy, Size: 32},
		},
		{
			Name: "extraData",
			Type: abi.Type{T: abi.BytesTy},
		},
	}

	// Prepend the initial 32-byte offset (value 32 = 0x20)
	initialOffset := make([]byte, 32)
	offsetValue := big.NewInt(32)
	// FillBytes puts the big.Int's value into the byte slice, padded left with zeros
	offsetBytes := offsetValue.FillBytes(make([]byte, 32))
	copy(initialOffset, offsetBytes) // Copy the padded value into our prefix slice

	pack, err := arguments.Pack(v.Version, v.ActiveAggregatedKeys, v.TotalActiveVotingPower, v.ValidatorsSszMRoot, v.ExtraData)
	if err != nil {
		return nil, errors.Errorf("failed to pack arguments: %w", err)
	}

	return append(initialOffset, pack...), err //nolint:makezero // intentionally appending to the initial offset
}

func (v ValidatorSetHeader) EncodeJSON() ([]byte, error) {
	// Convert byte arrays to hex strings before JSON marshaling
	type key struct {
		Tag     uint8  `json:"tag"`
		Payload string `json:"payload"` // hex string
	}
	type jsonHeader struct {
		Version                uint8    `json:"version"`
		ActiveAggregatedKeys   []key    `json:"activeAggregatedKeys"`
		ValidatorsSszMRoot     string   `json:"validatorsSszMRoot"` // hex string
		ExtraData              string   `json:"extraData"`
		TotalActiveVotingPower *big.Int `json:"totalActiveVotingPower"`
	}

	jsonHeaderData := jsonHeader{
		Version:                v.Version,
		ActiveAggregatedKeys:   make([]key, len(v.ActiveAggregatedKeys)),
		ValidatorsSszMRoot:     fmt.Sprintf("0x%064x", v.ValidatorsSszMRoot),
		ExtraData:              fmt.Sprintf("0x%064x", v.ExtraData),
		TotalActiveVotingPower: v.TotalActiveVotingPower,
	}

	for i, key := range v.ActiveAggregatedKeys {
		jsonHeaderData.ActiveAggregatedKeys[i].Tag = key.Tag
		jsonHeaderData.ActiveAggregatedKeys[i].Payload = fmt.Sprintf("0x%0128x", key.Payload)
	}

	jsonData, err := json.MarshalIndent(jsonHeaderData, "", "  ")
	if err != nil {
		return nil, errors.Errorf("failed to marshal header to JSON: %w", err)
	}

	return jsonData, nil
}

type CommitValsetHeaderResult struct {
	TxHash common.Hash
}
