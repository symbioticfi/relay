package entity

import (
	"fmt"
	"math/big"
	"slices"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"middleware-offchain/pkg/ssz"
)

type VerificationType uint32

const (
	VerificationTypeZK     VerificationType = 0
	VerificationTypeSimple VerificationType = 1
)

var (
	ExtraDataGlobalKeyPrefixHash = crypto.Keccak256Hash([]byte("symbiotic.Settlement.extraData."))
	ExtraDataKeyTagPrefixHash    = crypto.Keccak256Hash([]byte("keyTag."))
)

var (
	ZkVerificationTotalActiveValidatorsHash = crypto.Keccak256Hash([]byte("totalActiveValidators"))
	ZkVerificationValidatorSetHashMimcHash  = crypto.Keccak256Hash([]byte("validatorSetHashMimc"))
)

var (
	SimpleVerificationValidatorSetHashKeccak256Hash = crypto.Keccak256Hash([]byte("validatorSetHashKeccak256"))
	SimpleVerificationTotalVotingPowerHash          = crypto.Keccak256Hash([]byte("totalVotingPower"))
	SimpleVerificationAggPublicKeyG1Hash            = crypto.Keccak256Hash([]byte("aggPublicKeyG1"))
)

const ValsetHeaderKeyTag = KeyTag(15)

type RawSignature []byte
type RawMessageHash []byte
type RawPublicKey []byte
type CompactPublicKey []byte
type RawMessage []byte
type RawProof []byte
type VotingPower struct {
	*big.Int
}
type QuorumThresholdPct struct {
	*big.Int
}

func ToVotingPower(val *big.Int) VotingPower {
	return VotingPower{Int: val}
}

func ToQuorumThresholdPct(val *big.Int) QuorumThresholdPct {
	return QuorumThresholdPct{Int: val}
}

type Epoch uint64
type Timestamp uint64

func (raw RawSignature) MarshalText() ([]byte, error) {
	return []byte(hexutil.Encode(raw)), nil
}

func (raw RawMessageHash) MarshalText() ([]byte, error) {
	return []byte(hexutil.Encode(raw)), nil
}

func (raw RawPublicKey) MarshalText() ([]byte, error) {
	return []byte(hexutil.Encode(raw)), nil
}

func (raw CompactPublicKey) MarshalText() ([]byte, error) {
	return []byte(hexutil.Encode(raw)), nil
}

func (raw RawProof) MarshalText() ([]byte, error) {
	return []byte(hexutil.Encode(raw)), nil
}

func (vp VotingPower) MarshalJSON() ([]byte, error) {
	// dirty hack to force using string instead of float in json
	return []byte(fmt.Sprintf("\"%s\"", vp.String())), nil
}

func (e Epoch) MarshalJSON() ([]byte, error) {
	// dirty hack to force using string instead of float in json
	return []byte(fmt.Sprintf("\"%d\"", e)), nil
}

func (e Timestamp) MarshalJSON() ([]byte, error) {
	// dirty hack to force using string instead of float in json
	return []byte(fmt.Sprintf("\"%d\"", e)), nil
}

func (q QuorumThresholdPct) MarshalJSON() ([]byte, error) {
	maxQ := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	share := new(big.Float).Quo(new(big.Float).SetInt(q.Int), new(big.Float).SetInt(maxQ))
	pct := new(big.Float).Mul(share, big.NewFloat(100.0))
	return []byte(fmt.Sprintf("\"%s %%\"", pct.Text('f', 5))), nil
}

// SignatureRequest signature request message
// RequestHash = sha256(SignatureRequest) (use as identifier later)
type SignatureRequest struct {
	KeyTag        KeyTag
	RequiredEpoch Epoch
	Message       RawMessage
}

func (r SignatureRequest) Hash() common.Hash {
	return crypto.Keccak256Hash([]byte{uint8(r.KeyTag)}, new(big.Int).SetInt64(int64(r.RequiredEpoch)).Bytes(), r.Message)
}

// SignatureExtended signer.sign() -> SignatureExtended
type SignatureExtended struct {
	MessageHash RawMessageHash // scheme depends on KeyTag
	Signature   RawSignature   // parse based on KeyTag
	// PublicKey for bls will contain g1+g2
	PublicKey RawPublicKey // parse based on KeyTag
}

type SignatureMessage struct {
	RequestHash common.Hash
	KeyTag      KeyTag
	Epoch       Epoch
	Signature   SignatureExtended // parse based on KeyTag
}

// AggregationProof aggregator.proof(signatures []SignatureExtended) -> AggregationProof
type AggregationProof struct {
	VerificationType VerificationType // proof verification type
	MessageHash      RawMessageHash   // scheme depends on KeyTag
	Proof            RawProof         // parse based on KeyTag & VerificationType
}

type AggregatedSignatureMessage struct {
	RequestHash      common.Hash
	KeyTag           KeyTag
	Epoch            Epoch
	AggregationProof AggregationProof
}

type AggregationState struct {
	SignaturesCnt       uint32
	CurrentVotingPower  VotingPower
	RequiredVotingPower VotingPower
}

func (vt VerificationType) MarshalText() (text []byte, err error) {
	switch vt {
	case VerificationTypeZK:
		return []byte(fmt.Sprintf("%d (BLS-BN254-ZK)", uint32(vt))), nil
	case VerificationTypeSimple:
		return []byte(fmt.Sprintf("%d (BLS-BN254-SIMPLE)", uint32(vt))), nil
	}
	return []byte(fmt.Sprintf("%d (UNKNOWN)", uint32(vt))), nil
}

type CrossChainAddress struct {
	ChainId uint64
	Address common.Address
}

type QuorumThreshold struct {
	KeyTag          KeyTag
	QuorumThreshold QuorumThresholdPct
}

type NetworkConfig struct {
	VotingPowerProviders    []CrossChainAddress
	KeysProvider            CrossChainAddress
	Replicas                []CrossChainAddress
	VerificationType        VerificationType
	MaxVotingPower          VotingPower
	MinInclusionVotingPower VotingPower
	MaxValidatorsCount      VotingPower
	RequiredKeyTags         []KeyTag
	RequiredHeaderKeyTag    KeyTag
	QuorumThresholds        []QuorumThreshold
}

type NetworkData struct {
	Address    common.Address
	Subnetwork common.Hash
	Eip712Data Eip712Domain
}

type VaultVotingPower struct {
	Vault       common.Address
	VotingPower VotingPower
}

type OperatorVotingPower struct {
	Operator common.Address
	Vaults   []VaultVotingPower
}

type OperatorWithKeys struct {
	Operator common.Address
	Keys     []ValidatorKey
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

type ValidatorKey struct {
	Tag     KeyTag
	Payload CompactPublicKey
}

type ValidatorVault struct {
	ChainID     uint64         `json:"chainId"`
	Vault       common.Address `json:"vault"`
	VotingPower VotingPower    `json:"votingPower"`
}

type Validator struct {
	Operator    common.Address   `json:"operator"`
	VotingPower VotingPower      `json:"votingPower"`
	IsActive    bool             `json:"isActive"`
	Keys        []ValidatorKey   `json:"keys"`
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
	QuorumThreshold    VotingPower // absolute number now, not a percent
	PreviousHeaderHash common.Hash // previous valset header hash
	Validators         []Validator

	// internal usage only
	Status ValidatorSetStatus
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
	Hash   common.Hash
}

// ValidatorSetHeader represents the input for validator set header
type ValidatorSetHeader struct {
	Version            uint8
	RequiredKeyTag     KeyTag
	Epoch              uint64
	CaptureTimestamp   uint64
	QuorumThreshold    VotingPower
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

func (v ValidatorSet) GetTotalActiveVotingPower() *VotingPower {
	totalVotingPower := big.NewInt(0)
	for _, validator := range v.Validators {
		if validator.IsActive {
			totalVotingPower = totalVotingPower.Add(totalVotingPower, validator.VotingPower.Int)
		}
	}
	return &VotingPower{totalVotingPower}
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

func sszTreeRoot(v *ValidatorSet) (common.Hash, error) {
	sszType := validatorSetToSszValidators(v)
	return sszType.HashTreeRoot()
}

func keyPayloadHash(k ValidatorKey) common.Hash {
	return crypto.Keccak256Hash(k.Payload)
}

func validatorSetToSszValidators(v *ValidatorSet) ssz.SszValidatorSet {
	return ssz.SszValidatorSet{
		Validators: lo.Map(v.Validators, func(v Validator, _ int) *ssz.SszValidator {
			return &ssz.SszValidator{
				Operator:    v.Operator,
				VotingPower: v.VotingPower.Int,
				IsActive:    v.IsActive,
				Keys: lo.Map(v.Keys, func(k ValidatorKey, _ int) *ssz.SszKey {
					return &ssz.SszKey{
						Tag:         uint8(k.Tag),
						PayloadHash: keyPayloadHash(k),
					}
				}),
				Vaults: lo.Map(v.Vaults, func(v ValidatorVault, _ int) *ssz.SszVault {
					return &ssz.SszVault{
						ChainId:     v.ChainID,
						Vault:       v.Vault,
						VotingPower: v.VotingPower.Int,
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

	pack, err := arguments.Pack(v.Version, v.RequiredKeyTag, new(big.Int).SetUint64(v.Epoch), new(big.Int).SetUint64(v.CaptureTimestamp), v.QuorumThreshold.Int, v.ValidatorsSszMRoot, v.PreviousHeaderHash)
	if err != nil {
		return nil, errors.Errorf("failed to pack arguments: %w", err)
	}

	return pack, nil
}

func (v ValidatorSetHeader) Hash() (common.Hash, error) {
	abiEncoded, err := v.AbiEncode()
	if err != nil {
		return common.Hash{}, errors.Errorf("failed to hash validator set header: %w", err)
	}

	return common.Hash(crypto.Keccak256(abiEncoded)), nil
}

type TxResult struct {
	TxHash common.Hash
}
