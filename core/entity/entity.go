package entity

import (
	"fmt"
	"math/big"
	"slices"
	"time"

	"github.com/symbioticfi/relay/core/usecase/ssz"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/samber/lo"
)

type VerificationType uint32
type AggregationPolicyType uint32

const (
	VerificationTypeBlsBn254ZK     VerificationType = 0
	VerificationTypeBlsBn254Simple VerificationType = 1

	AggregationPolicyLowLatency AggregationPolicyType = 0
	AggregationPolicyLowCost    AggregationPolicyType = 1
)

var (
	ExtraDataKeyTagPrefixHash = crypto.Keccak256Hash([]byte("keyTag."))
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

type ValidatorSetStatus uint8

const (
	HeaderDerived ValidatorSetStatus = iota
	HeaderAggregated
	HeaderCommitted
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

func (s ValidatorSetStatus) MarshalJSON() ([]byte, error) {
	switch s {
	case HeaderDerived:
		return []byte("\"Derived\""), nil
	case HeaderAggregated:
		return []byte("\"Aggregated\""), nil
	case HeaderCommitted:
		return []byte("\"Committed\""), nil
	default:
		return []byte("\"Unknown\""), nil
	}
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
	return []byte(vt.String()), nil
}

func (vt VerificationType) String() string {
	switch vt {
	case VerificationTypeBlsBn254ZK:
		return fmt.Sprintf("%d (BLS-BN254-ZK)", uint32(vt))
	case VerificationTypeBlsBn254Simple:
		return fmt.Sprintf("%d (BLS-BN254-SIMPLE)", uint32(vt))
	}
	return fmt.Sprintf("%d (UNKNOWN)", uint32(vt))
}

func (ap AggregationPolicyType) MarshalText() (text []byte, err error) {
	return []byte(ap.String()), nil
}

func (ap AggregationPolicyType) String() string {
	switch ap {
	case AggregationPolicyLowLatency:
		return fmt.Sprintf("%d AGGREGATION-POLICY-LOW-LATENCY", uint32(ap))
	case AggregationPolicyLowCost:
		return fmt.Sprintf("%d AGGREGATION-POLICY-LOW-COST", uint32(ap))
	}
	return fmt.Sprintf("%d (UNKNOWN)", uint32(ap))
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
	Settlements             []CrossChainAddress
	VerificationType        VerificationType
	MaxVotingPower          VotingPower
	MinInclusionVotingPower VotingPower
	MaxValidatorsCount      VotingPower
	RequiredKeyTags         []KeyTag
	RequiredHeaderKeyTag    KeyTag
	QuorumThresholds        []QuorumThreshold

	// scheduler config
	NumAggregators uint64
	NumCommitters  uint64
}

func maxThreshold() *big.Int {
	// 10^18 is the maximum threshold value
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
}

func (nc NetworkConfig) CalcQuorumThreshold(totalVP VotingPower) (VotingPower, error) {
	quorumThresholdPercent := big.NewInt(0)
	for _, quorumThreshold := range nc.QuorumThresholds {
		if quorumThreshold.KeyTag == nc.RequiredHeaderKeyTag {
			quorumThresholdPercent = quorumThreshold.QuorumThreshold.Int
		}
	}
	if quorumThresholdPercent.Cmp(big.NewInt(0)) == 0 {
		return VotingPower{}, errors.Errorf("quorum threshold is zero")
	}

	mul := new(big.Int).Mul(totalVP.Int, quorumThresholdPercent)
	div := new(big.Int).Div(mul, maxThreshold())
	// add 1 to apply up rounding
	return ToVotingPower(new(big.Int).Add(div, big.NewInt(1))), nil
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

type Validators []Validator

func (va Validators) SortByVotingPowerDescAndOperatorAddressAsc() {
	slices.SortFunc(va, func(a, b Validator) int {
		if cmp := -a.VotingPower.Cmp(b.VotingPower.Int); cmp != 0 {
			return cmp
		}
		return a.Operator.Cmp(b.Operator)
	})
}

func (va Validators) SortByOperatorAddressAsc() {
	slices.SortFunc(va, func(a, b Validator) int {
		return a.Operator.Cmp(b.Operator)
	})
}

func (va Validators) CheckIsSortedByOperatorAddressAsc() error {
	if !slices.IsSortedFunc(va, func(a, b Validator) int {
		return a.Operator.Cmp(b.Operator)
	}) {
		return errors.New("validators are not sorted by operator address ascending")
	}
	return nil
}

type Vaults []ValidatorVault

func (v Vaults) SortByAddressAsc() {
	slices.SortFunc(v, func(a, b ValidatorVault) int {
		return a.Vault.Cmp(b.Vault)
	})
}

func (v Vaults) SortVaultsByVotingPowerDescAndAddressAsc() {
	slices.SortFunc(v, func(a, b ValidatorVault) int {
		if cmp := -a.VotingPower.Cmp(b.VotingPower.Int); cmp != 0 {
			return cmp
		}
		return a.Vault.Cmp(b.Vault)
	})
}

func (va Validators) GetTotalActiveVotingPower() VotingPower {
	totalVotingPower := big.NewInt(0)
	for _, validator := range va {
		if validator.IsActive {
			totalVotingPower = totalVotingPower.Add(totalVotingPower, validator.VotingPower.Int)
		}
	}
	return VotingPower{totalVotingPower}
}

func (va Validators) GetActiveValidators() Validators {
	var activeValidators Validators
	for _, validator := range va {
		if validator.IsActive {
			activeValidators = append(activeValidators, validator)
		}
	}
	activeValidators.SortByOperatorAddressAsc()
	return activeValidators
}

type Validator struct {
	Operator    common.Address `json:"operator"`
	VotingPower VotingPower    `json:"votingPower"`
	IsActive    bool           `json:"isActive"`
	Keys        []ValidatorKey `json:"keys"`
	Vaults      Vaults         `json:"vaults"`
}

func (v Validator) FindKeyByKeyTag(keyTag KeyTag) ([]byte, bool) {
	for _, key := range v.Keys {
		if key.Tag == keyTag {
			return key.Payload, true
		}
	}
	return nil, false
}

type ValidatorSet struct {
	Version          uint8
	RequiredKeyTag   KeyTag      // key tag required to commit next valset
	Epoch            uint64      // valset epoch
	CaptureTimestamp uint64      // epoch capture timestamp
	QuorumThreshold  VotingPower // absolute number now, not a percent
	Validators       Validators
	Status           ValidatorSetStatus

	// Scheduler info for current validator set, completely offchain not included in header
	AggregatorIndices Bitmap
	CommitterIndices  Bitmap
}

func (v ValidatorSet) IsAggregator(requiredKey CompactPublicKey) bool {
	return v.findMembership(v.AggregatorIndices, requiredKey)
}

func (v ValidatorSet) IsCommitter(requiredKey CompactPublicKey) bool {
	return v.findMembership(v.CommitterIndices, requiredKey)
}

func (v ValidatorSet) IsSigner(requiredKey CompactPublicKey) bool {
	for _, validator := range v.Validators {
		for _, key := range validator.Keys {
			if key.Tag == v.RequiredKeyTag && slices.Equal(key.Payload, requiredKey) {
				return true
			}
		}
	}
	return false
}

func (v ValidatorSet) findMembership(bitmap Bitmap, requiredKey CompactPublicKey) bool {
	iterator := bitmap.Iterator()
	for iterator.HasNext() {
		index := iterator.Next()
		for _, key := range v.Validators[index].Keys {
			if key.Tag == v.RequiredKeyTag && slices.Equal(key.Payload, requiredKey) {
				return true
			}
		}
	}
	return false
}

func (v ValidatorSet) FindValidatorByKey(keyTag KeyTag, publicKey []byte) (Validator, bool) { // DON'T USE INSIDE LOOPS
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
	TotalVotingPower   VotingPower
	ValidatorsSszMRoot common.Hash
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

func (v ValidatorSet) GetTotalActiveVotingPower() VotingPower {
	return v.Validators.GetTotalActiveVotingPower()
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
		TotalVotingPower:   v.GetTotalActiveVotingPower(),
		ValidatorsSszMRoot: sszMroot,
	}, nil
}

func (v ValidatorSet) FindValidatorsByKeys(keyTag KeyTag, publicKeys []CompactPublicKey) (Validators, error) {
	// Build lookup map: publicKey -> validator
	publicKeyToValidator := make(map[string]Validator)
	for _, validator := range v.Validators {
		if publicKey, found := validator.FindKeyByKeyTag(keyTag); found {
			publicKeyToValidator[string(publicKey)] = validator
		}
	}

	// Find validators for each public key
	result := make(Validators, 0, len(publicKeys))
	for _, publicKey := range publicKeys {
		validator, found := publicKeyToValidator[string(publicKey)]
		if !found {
			return nil, errors.Errorf("validator not found for public key %x with key tag %d", publicKey, keyTag)
		}
		result = append(result, validator)
	}

	return result, nil
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
			Name: "totalVotingPower",
			Type: abi.Type{T: abi.UintTy, Size: 256},
		},
		{
			Name: "validatorsSszMRoot",
			Type: abi.Type{T: abi.FixedBytesTy, Size: 32},
		},
	}

	if v.QuorumThreshold.Int == nil {
		v.QuorumThreshold.Int = big.NewInt(0)
	}
	if v.TotalVotingPower.Int == nil {
		v.TotalVotingPower.Int = big.NewInt(0)
	}

	pack, err := arguments.Pack(v.Version, v.RequiredKeyTag, new(big.Int).SetUint64(v.Epoch), new(big.Int).SetUint64(v.CaptureTimestamp), v.QuorumThreshold.Int, v.TotalVotingPower.Int, v.ValidatorsSszMRoot)
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

type ChainURL struct {
	ChainID uint64
	RPCURL  string
}

type SignatureStatStage string

const (
	SignatureStatStageUnknown             = "Unknown"
	SignatureStatStageSignRequestReceived = "SignRequestReceived"
	SignatureStatStageSignCompleted       = "SignCompleted"
	SignatureStatStageAggQuorumReached    = "AggQuorumReached"
	SignatureStatStageAggCompleted        = "AggCompleted"
	SignatureStatStageAggProofReceived    = "AggProofReceived"

	SignatureStatStageAggregationSkipped = "AggSkipped"
)

type SignatureStat struct {
	ReqHash common.Hash
	StatMap map[SignatureStatStage]time.Time
}

type AggregationStatus struct {
	VotingPower VotingPower
	Validators  []Validator
}
