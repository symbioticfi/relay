package entity

import (
	"encoding/json"
	"fmt"
	"github.com/samber/lo"
	"math/big"
	"middleware-offchain/pkg/ssz"
	"slices"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
)

type Key struct {
	Tag     KeyTag
	Payload []byte
}

type ValidatorVault struct {
	ChainID     uint64         `json:"chainId"`
	Vault       common.Address `json:"vault"`
	VotingPower *big.Int       `json:"votingPower"`
}

type Validator struct {
	Operator    common.Address   `json:"operator"`
	VotingPower *big.Int         `json:"votingPower"`
	IsActive    bool             `json:"isActive"`
	Keys        []Key            `json:"keys"`
	Vaults      []ValidatorVault `json:"vaults"`
}

type ValidatorSet struct {
	Version                uint8
	RequiredKeyTag         uint8    // key tag required to commit next valset
	Epoch                  uint64   // valset epoch
	CaptureTimestamp       uint64   // epoch capture timestamp
	QuorumThreshold        *big.Int // absolute number now, not a percent
	PreviousHeaderHash     [32]byte // previous valset header hash
	Validators             []Validator
	TotalActiveVotingPower *big.Int // todo ilya maybe remove
}

// Signature signer.sign() -> Signature
type Signature struct {
	MessageHash []byte // scheme depends on KeyTag
	Signature   []byte // parse based on KeyTag
	PublicKey   []byte // parse based on KeyTag
}

func (v ValidatorSet) FindValidatorByKey(keyTag KeyTag, g1 []byte) (Validator, bool) {
	for _, validator := range v.Validators {
		for _, key := range validator.Keys {
			if slices.Equal(key.Payload, g1) {
				return validator, true
			}
		}
	}
	return Validator{}, false
}

type ValidatorSetHash struct {
	KeyTag KeyTag
	Hash   [32]byte
}

// ValidatorSetHeader represents the input for validator set header
type ValidatorSetHeader struct {
	Version            uint8
	RequiredKeyTag     KeyTag
	Epoch              uint64
	CaptureTimestamp   uint64
	QuorumThreshold    *big.Int
	ValidatorsSszMRoot [32]byte
	PreviousHeaderHash [32]byte
}

type ExtraData struct {
	Key   [32]byte
	Value [32]byte
}

type ExtraDataList []ExtraData

func (e ExtraDataList) Hash() ([]byte, error) {
	bytes, err := e.AbiEncode()
	if err != nil {
		return nil, errors.Errorf("failed to hash extra data: %w", err)
	}
	return crypto.Keccak256(bytes), nil
}

func (e ExtraDataList) AbiEncode() ([]byte, error) {
	tupleArr, err := abi.NewType(
		"tuple[]",
		"",
		[]abi.ArgumentMarshaling{
			{Name: "key", Type: "bytes32"},
			{Name: "value", Type: "bytes32"},
		},
	)
	if err != nil {
		return nil, err
	}

	args := abi.Arguments{
		{Type: tupleArr},
	}

	return args.Pack(e)
}

func (v ValidatorSet) GetTotalActiveVotingPower() *big.Int {
	totalVotingPower := big.NewInt(0)
	for _, validator := range v.Validators {
		if validator.IsActive {
			totalVotingPower = totalVotingPower.Add(totalVotingPower, validator.VotingPower)
		}
	}
	return totalVotingPower
}

func (v ValidatorSet) GetTotalActiveValidators() int64 {
	totalActive := int64(0)
	for _, validator := range v.Validators {
		if validator.IsActive {
			totalActive++
		}
	}
	return totalActive
}

func (v ValidatorSet) GetHeader() (ValidatorSetHeader, error) {
	sszMroot, err := sszTreeRoot(&v)
	if err != nil {
		return ValidatorSetHeader{}, fmt.Errorf("failed to get hash tree root: %w", err)
	}

	return ValidatorSetHeader{
		Version:            v.Version,
		RequiredKeyTag:     v.RequiredKeyTag,
		Epoch:              v.Epoch,
		CaptureTimestamp:   v.CaptureTimestamp,
		QuorumThreshold:    v.QuorumThreshold,
		PreviousHeaderHash: v.PreviousHeaderHash,
		ValidatorsSszMRoot: sszMroot,
	}, nil
}

func sszTreeRoot(v *ValidatorSet) ([32]byte, error) {
	sszType := validatorSetToSszValidators(v)
	return sszType.HashTreeRoot()
}

func keyPayloadHash(k Key) common.Hash {
	return crypto.Keccak256Hash(k.Payload)
}

func validatorSetToSszValidators(v *ValidatorSet) ssz.SszValidatorSet {
	return ssz.SszValidatorSet{
		Validators: lo.Map(v.Validators, func(v Validator, _ int) *ssz.SszValidator {
			return &ssz.SszValidator{
				Operator:    v.Operator,
				VotingPower: v.VotingPower,
				IsActive:    v.IsActive,
				Keys: lo.Map(v.Keys, func(k Key, _ int) *ssz.SszKey {
					return &ssz.SszKey{
						Tag:         k.Tag,
						PayloadHash: keyPayloadHash(k),
					}
				}),
				Vaults: lo.Map(v.Vaults, func(v ValidatorVault, _ int) *ssz.SszVault {
					return &ssz.SszVault{
						ChainId:     v.ChainID,
						Vault:       v.Vault,
						VotingPower: v.VotingPower,
					}
				}),
			}
		}),
		Version: v.Version,
	}
}

func (v ValidatorSetHeader) AbiEncode() ([]byte, error) {
	arguments := abi.Arguments{
		{
			Name: "version",
			Type: abi.Type{T: abi.UintTy, Size: 8},
		},
		{
			Name: "requiredKeyTag",
			Type: abi.Type{T: abi.UintTy, Size: 8},
		},
		{
			Name: "epoch",
			Type: abi.Type{T: abi.UintTy, Size: 48},
		},
		{
			Name: "captureTimestamp",
			Type: abi.Type{T: abi.UintTy, Size: 48},
		},
		{
			Name: "verificationType",
			Type: abi.Type{T: abi.UintTy, Size: 32},
		},
		{
			Name: "quorumThreshold",
			Type: abi.Type{T: abi.UintTy, Size: 256},
		},
		{
			Name: "validatorsSszMRoot",
			Type: abi.Type{T: abi.FixedBytesTy, Size: 32},
		},
		{
			Name: "previousHeaderHash",
			Type: abi.Type{T: abi.FixedBytesTy, Size: 32},
		},
	}

	// Prepend the initial 32-byte offset (value 32 = 0x20)
	initialOffset := make([]byte, 32)
	offsetValue := big.NewInt(32)
	// FillBytes puts the big.Int's value into the byte slice, padded left with zeros
	offsetBytes := offsetValue.FillBytes(make([]byte, 32))
	copy(initialOffset, offsetBytes) // Copy the padded value into our prefix slice

	pack, err := arguments.Pack(v.Version, v.RequiredKeyTag, v.Epoch, v.CaptureTimestamp, v.QuorumThreshold, v.ValidatorsSszMRoot, v.PreviousHeaderHash)
	if err != nil {
		return nil, errors.Errorf("failed to pack arguments: %w", err)
	}

	return append(initialOffset, pack...), err
}

func (v ValidatorSetHeader) Hash() ([32]byte, error) {
	abiEncoded, err := v.AbiEncode()
	if err != nil {
		return [32]byte{}, errors.Errorf("failed to hash validator set header: %w", err)
	}

	return [32]byte(crypto.Keccak256(abiEncoded)), nil
}

//func (v ValidatorSetHeader) EncodeJSONFull() ([]byte, error) {
//	// Convert byte arrays to hex strings before JSON marshaling
//	type key struct {
//		Tag     uint8  `json:"tag"`
//		Payload string `json:"payload"` // hex string
//	}
//
//	type eip712Domain struct {
//		Fields            string // hex string
//		Name              string
//		Version           string
//		ChainId           *big.Int
//		VerifyingContract string // hex string
//		Salt              *big.Int
//		Extensions        []*big.Int
//	}
//
//	type validatorSetHash struct {
//		KeyTag uint8  `json:"keyTag"`
//		Hash   string `json:"hash"` // hex string
//	}
//
//	type extraData struct {
//		Key   string `json:"key"`
//		Value string `json:"value"`
//	}
//
//	type jsonHeader struct {
//		Version                     uint8              `json:"version"`
//		TotalActiveValidators       *big.Int           `json:"totalActiveValidators"`
//		ActiveAggregatedKeys        []key              `json:"activeAggregatedKeys"`
//		TotalActiveVotingPower      *big.Int           `json:"totalActiveVotingPower"`
//		ValidatorsSszMRoot          string             `json:"validatorsSszMRoot"` // hex string
//		Epoch                       *big.Int           `json:"epoch"`
//		DomainEip712                eip712Domain       `json:"domainEip712"`
//		Subnetwork                  string             `json:"subnetwork"` // hex string
//		ValidatorSetHashesMimc      []validatorSetHash `json:"validatorSetHashesMimc"`
//		ValidatorSetHashesKeccak256 []validatorSetHash `json:"validatorSetHashesKeccak256"`
//		RequiredKeyTag              uint8              `json:"requiredKeyTag"`
//		CaptureTimestamp            *big.Int           `json:"captureTimestamp"`
//		QuorumThreshold             *big.Int           `json:"quorumThreshold"`
//		PreviousHeaderHash          string             `json:"previousHeaderHash"` // hex string
//	}
//
//	jsonHeaderData := jsonHeader{
//		Version:                v.Version,
//		TotalActiveValidators:  v.TotalActiveValidators,
//		TotalActiveVotingPower: v.TotalActiveVotingPower,
//		ValidatorsSszMRoot:     fmt.Sprintf("0x%064x", v.ValidatorsSszMRoot),
//		Epoch:                  v.Epoch,
//		DomainEip712: eip712Domain{
//			Fields:            fmt.Sprintf("0x%02x", v.DomainEip712.Fields),
//			Name:              v.DomainEip712.Name,
//			Version:           v.DomainEip712.Version,
//			ChainId:           v.DomainEip712.ChainId,
//			VerifyingContract: v.DomainEip712.VerifyingContract.Hex(),
//			Salt:              v.DomainEip712.Salt,
//			Extensions:        v.DomainEip712.Extensions,
//		},
//		Subnetwork:         fmt.Sprintf("0x%40x", v.Subnetwork),
//		RequiredKeyTag:     v.RequiredKeyTag,
//		CaptureTimestamp:   v.CaptureTimestamp,
//		QuorumThreshold:    v.QuorumThreshold,
//		PreviousHeaderHash: fmt.Sprintf("0x%064x", v.PreviousHeaderHash),
//	}
//
//	for i, key := range v.ActiveAggregatedKeys {
//		jsonHeaderData.ActiveAggregatedKeys[i].Tag = key.Tag
//		jsonHeaderData.ActiveAggregatedKeys[i].Payload = fmt.Sprintf("0x%0128x", key.Payload)
//	}
//
//	for i, hash := range v.ValidatorSetHashesMimc {
//		jsonHeaderData.ValidatorSetHashesMimc[i].KeyTag = hash.KeyTag
//		jsonHeaderData.ValidatorSetHashesMimc[i].Hash = fmt.Sprintf("0x%064x", hash.Hash)
//	}
//
//	for i, hash := range v.ValidatorSetHashesKeccak256 {
//		jsonHeaderData.ValidatorSetHashesKeccak256[i].KeyTag = hash.KeyTag
//		jsonHeaderData.ValidatorSetHashesKeccak256[i].Hash = fmt.Sprintf("0x%064x", hash.Hash)
//	}
//
//	jsonData, err := json.MarshalIndent(jsonHeaderData, "", "  ")
//	if err != nil {
//		return nil, errors.Errorf("failed to marshal header to JSON: %w", err)
//	}
//
//	return jsonData, nil
//}

func (v ValidatorSetHeader) EncodeJSON() ([]byte, error) {
	type jsonHeader struct {
		Version            uint8    `json:"version"`
		ValidatorsSszMRoot string   `json:"validatorsSszMRoot"` // hex string
		Epoch              *big.Int `json:"epoch"`
		RequiredKeyTag     KeyTag   `json:"requiredKeyTag"`
		CaptureTimestamp   *big.Int `json:"captureTimestamp"`
		QuorumThreshold    *big.Int `json:"quorumThreshold"`
		PreviousHeaderHash string   `json:"previousHeaderHash"` // hex string
	}

	jsonHeaderData := jsonHeader{
		Version:            v.Version,
		ValidatorsSszMRoot: fmt.Sprintf("0x%064x", v.ValidatorsSszMRoot),
		Epoch:              v.Epoch,
		RequiredKeyTag:     v.RequiredKeyTag,
		CaptureTimestamp:   v.CaptureTimestamp,
		QuorumThreshold:    v.QuorumThreshold,
		PreviousHeaderHash: fmt.Sprintf("0x%064x", v.PreviousHeaderHash),
	}

	jsonData, err := json.MarshalIndent(jsonHeaderData, "", "  ")
	if err != nil {
		return nil, errors.Errorf("failed to marshal header to JSON: %w", err)
	}

	return jsonData, nil
}

func (e ExtraDataList) EncodeJSON() ([]byte, error) {
	type jsonExtraData struct {
		Key   string `json:"key"`   // hex string
		Value string `json:"value"` // hex string
	}

	jsonExtraDataList := make([]jsonExtraData, len(e))
	for i, extraData := range e {
		jsonExtraDataList[i].Key = fmt.Sprintf("0x%064x", extraData.Key)
		jsonExtraDataList[i].Value = fmt.Sprintf("0x%064x", extraData.Value)
	}

	jsonData, err := json.MarshalIndent(jsonExtraDataList, "", "  ")
	if err != nil {
		return nil, errors.Errorf("failed to marshal extra data to JSON: %w", err)
	}

	return jsonData, nil
}

func (v ValidatorSetHeaderWithExtraData) EncodeJSON() ([]byte, error) {
	type jsonHeader struct {
		Version            uint8    `json:"version"`
		ValidatorsSszMRoot string   `json:"validatorsSszMRoot"` // hex string
		Epoch              uint64   `json:"epoch"`
		RequiredKeyTag     KeyTag   `json:"requiredKeyTag"`
		CaptureTimestamp   uint64   `json:"captureTimestamp"`
		QuorumThreshold    *big.Int `json:"quorumThreshold"`
		PreviousHeaderHash string   `json:"previousHeaderHash"` // hex string
	}

	type jsonExtraData struct {
		Key   string `json:"key"`   // hex string
		Value string `json:"value"` // hex string
	}

	type jsonValidatorSetHeaderWithExtraData struct {
		Header        jsonHeader      `json:"header"`
		ExtraDataList []jsonExtraData `json:"extraData"`
	}

	jsonHeaderData := jsonHeader{
		Version:            v.ValidatorSetHeader.Version,
		ValidatorsSszMRoot: fmt.Sprintf("0x%064x", v.ValidatorSetHeader.ValidatorsSszMRoot),
		Epoch:              v.ValidatorSetHeader.Epoch,
		RequiredKeyTag:     v.ValidatorSetHeader.RequiredKeyTag,
		CaptureTimestamp:   v.ValidatorSetHeader.CaptureTimestamp,
		QuorumThreshold:    v.ValidatorSetHeader.QuorumThreshold,
		PreviousHeaderHash: fmt.Sprintf("0x%064x", v.ValidatorSetHeader.PreviousHeaderHash),
	}

	jsonExtraDataList := make([]jsonExtraData, len(v.ExtraDataList))
	for i, extraData := range v.ExtraDataList {
		jsonExtraDataList[i].Key = fmt.Sprintf("0x%064x", extraData.Key)
		jsonExtraDataList[i].Value = fmt.Sprintf("0x%064x", extraData.Value)
	}

	jsonValidatorSetHeaderWithExtraDataData := jsonValidatorSetHeaderWithExtraData{
		Header:        jsonHeaderData,
		ExtraDataList: jsonExtraDataList,
	}

	jsonData, err := json.MarshalIndent(jsonValidatorSetHeaderWithExtraDataData, "", "  ")
	if err != nil {
		return nil, errors.Errorf("failed to marshal extra data to JSON: %w", err)
	}

	return jsonData, nil
}

type ValidatorSetHeaderWithExtraData struct {
	ValidatorSetHeader ValidatorSetHeader
	ExtraDataList      ExtraDataList
}

type CommitValsetHeaderResult struct {
	TxHash common.Hash
}
