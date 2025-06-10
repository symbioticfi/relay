package entity

import (
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"middleware-offchain/pkg/ssz"
)

const ValsetHeaderKeyTag = KeyTag(15)

// SignatureRequest signature request message
// RequestHash = sha256(SignatureRequest) (use as identifier later)
type SignatureRequest struct {
	KeyTag        KeyTag
	RequiredEpoch uint64
	Message       []byte
}

func (r SignatureRequest) Hash() common.Hash {
	return crypto.Keccak256Hash([]byte{uint8(r.KeyTag)}, new(big.Int).SetInt64(int64(r.RequiredEpoch)).Bytes(), r.Message)
}

type SignatureMessage struct {
	RequestHash common.Hash
	KeyTag      KeyTag
	Epoch       uint64
	Signature   Signature // parse based on KeyTag
}

type AggregationState struct {
	SignaturesCnt       uint32
	CurrentVotingPower  *big.Int
	RequiredVotingPower *big.Int
}

// AggregationProof aggregator.proof(signatures []Signature) -> AggregationProof
type AggregationProof struct {
	VerificationType VerificationType // proof verification type
	MessageHash      []byte           // scheme depends on KeyTag
	Proof            []byte           // parse based on KeyTag & VerificationType
}

type AggregatedSignatureMessage struct {
	RequestHash      common.Hash
	KeyTag           KeyTag
	Epoch            uint64
	AggregationProof AggregationProof
}

type VerificationType uint32

const (
	VerificationTypeZK     VerificationType = 0
	VerificationTypeSimple VerificationType = 1
)

const (
	ExtraDataGlobalKeyPrefix = "symbiotic.Settlement.extraData."
	ExtraDataKeyTagPrefix    = "keyTag."
)

const (
	ZkVerificationTotalActiveValidators = "totalActiveValidators"
	ZkVerificationValidatorSetHashMimc  = "validatorSetHashMimc"
)

const (
	SimpleVerificationValidatorSetHashKeccak256 = "validatorSetHashKeccak256"
	SimpleVerificationTotalVotingPower          = "totalVotingPower"
	SimpleVerificationAggPublicKeyG1            = "aggPublicKeyG1"
)

// Phase represents the different phases of the protocol
type Phase uint64

const (
	IDLE    Phase = 0
	COMMIT  Phase = 1
	PROLONG Phase = 2
	FAIL    Phase = 3
)

type CrossChainAddress struct {
	Address common.Address `json:"addr"`
	ChainId uint64         `json:"chainId"`
}

type NetworkConfig struct {
	VotingPowerProviders    []CrossChainAddress
	KeysProvider            CrossChainAddress
	Replicas                []CrossChainAddress
	VerificationType        VerificationType
	MaxVotingPower          *big.Int
	MinInclusionVotingPower *big.Int
	MaxValidatorsCount      *big.Int
	RequiredKeyTags         []KeyTag
}

type NetworkData struct {
	Address    common.Address
	Subnetwork [32]byte
	Eip712Data Eip712Domain
}

type VaultVotingPower struct {
	Vault       common.Address
	VotingPower *big.Int
}

type OperatorVotingPower struct {
	Operator common.Address
	Vaults   []VaultVotingPower
}

type OperatorWithKeys struct {
	Operator common.Address
	Keys     []Key
}

type Eip712Domain struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              *big.Int
	Extensions        []*big.Int
}

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

func (v Validator) FindKeyByKeyTag(keyTag KeyTag) ([]byte, bool) {
	for _, key := range v.Keys {
		if key.Tag == keyTag {
			return key.Payload, true
		}
	}
	return nil, false
}

type ValidatorSetStatus int

const (
	HeaderPending ValidatorSetStatus = iota
	HeaderMissed
	HeaderCommitted
)

type ValidatorSet struct {
	Version            uint8
	RequiredKeyTag     KeyTag      // key tag required to commit next valset
	Epoch              uint64      // valset epoch
	CaptureTimestamp   uint64      // epoch capture timestamp
	QuorumThreshold    *big.Int    // absolute number now, not a percent
	PreviousHeaderHash common.Hash // previous valset header hash
	Validators         []Validator

	// internal usage only
	Status ValidatorSetStatus
}

// Signature signer.sign() -> Signature
type Signature struct {
	MessageHash []byte // scheme depends on KeyTag
	Signature   []byte // parse based on KeyTag
	// PublicKey for bls will contain g1+g2
	PublicKey []byte // parse based on KeyTag
}

func (v ValidatorSet) FindValidatorByKey(keyTag KeyTag, publicKey []byte) (Validator, bool) {
	for _, validator := range v.Validators {
		for _, key := range validator.Keys {
			if key.Tag == keyTag && slices.Equal(key.Payload, publicKey) {
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
	ValidatorsSszMRoot common.Hash
	PreviousHeaderHash common.Hash
}

type ExtraData struct {
	Key   common.Hash
	Value common.Hash
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
		return ValidatorSetHeader{}, errors.Errorf("failed to get hash tree root: %w", err)
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
						Tag:         uint8(k.Tag),
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

	pack, err := arguments.Pack(v.Version, v.RequiredKeyTag, new(big.Int).SetUint64(v.Epoch), new(big.Int).SetUint64(v.CaptureTimestamp), v.QuorumThreshold, v.ValidatorsSszMRoot, v.PreviousHeaderHash)
	if err != nil {
		return nil, errors.Errorf("failed to pack arguments: %w", err)
	}

	return pack, nil
}

func (v ValidatorSetHeader) Hash() ([32]byte, error) {
	abiEncoded, err := v.AbiEncode()
	if err != nil {
		return [32]byte{}, errors.Errorf("failed to hash validator set header: %w", err)
	}

	return [32]byte(crypto.Keccak256(abiEncoded)), nil
}

type CommitValsetHeaderResult struct {
	TxHash common.Hash
}
