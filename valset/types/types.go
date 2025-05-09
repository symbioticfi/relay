package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"

	"middleware-offchain/bls"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Key struct {
	Tag         uint8 `ssz-size:"1"`
	Payload     []byte
	PayloadHash [32]byte `ssz-size:"32"`
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
	Vaults      []*Vault       `ssz-max:"32"`
}

type ValidatorSet struct {
	Version                uint8
	TotalActiveVotingPower *big.Int
	Validators             []*Validator `ssz-max:"1048576"`
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
	newG1 := G1{new(big.Int), new(big.Int)}

	g1.X.BigInt(newG1[0])
	g1.Y.BigInt(newG1[1])
	return newG1
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
		ExtraData:              formatPayload(v.ExtraData),
		TotalActiveVotingPower: v.TotalActiveVotingPower,
	}

	for i, key := range v.ActiveAggregatedKeys {
		jsonHeaderData.ActiveAggregatedKeys[i].Tag = key.Tag
		jsonHeaderData.ActiveAggregatedKeys[i].Payload = formatPayload(key.Payload)
	}

	jsonData, err := json.MarshalIndent(jsonHeaderData, "", "  ")
	if err != nil {
		return nil, errors.Errorf("failed to marshal header to JSON: %w", err)
	}

	return jsonData, nil
}

func formatPayload(payload []byte) string {
	lengthHex := fmt.Sprintf("%064x", len(payload)) // 64 hex digits (32 bytes) for length
	payloadHex := hex.EncodeToString(payload)       // raw bytes → hex

	return "0x" + lengthHex + payloadHex
}
