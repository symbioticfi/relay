package types

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// signature request message
// RequestHash = sha256(SignatureRequest) (use as identifier later)
type SignatureRequest struct {
	KeyTag        uint8
	RequiredEpoch uint64 // NEED TO BE CAREFUL HERE
	Message       []byte
}

// signer.sign() -> Signature
type Signature struct {
	MessageHash []byte // scheme depends on KeyTag
	Signature   []byte // parse based on KeyTag
	PublicKey   []byte // parse based on KeyTag
}

// message about signature
type SignatureMessage struct {
	RequestHash [32]byte
	KeyTag      uint8
	Epoch       uint64
	Signature   Signature // parse based on KeyTag
}

type AggregationState struct {
	SignaturesCnt       uint32
	CurrentVotingPower  *big.Int
	RequiredVotingPower *big.Int
}

// aggregator.proof(signatures []Signature) -> AggregationProof
type AggregationProof struct {
	VerificationType uint32 // proof verification type
	MessageHash      []byte // scheme depends on KeyTag
	Proof            []byte // parse based on KeyTag & VerificationType
}

// aggregated signature message
type AggregatedSignatureMessage struct {
	RequestHash      [32]byte
	KeyTag           uint8
	Epoch            uint64
	AggregationProof AggregationProof
}

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

// validator set structure
type ValSet struct {
	Version            uint8       // valset type version
	RequiredKeyTag     uint8       // key tag requred to commit next valset
	Epoch              uint64      // valset epoch
	CaptureTimestamp   uint64      // epoch capture timestamp
	QuorumThreshold    *big.Int    // absolute number now, not a percent
	PreviousHeaderHash [32]byte    // previous valset header hash
	Validators         []Validator // validators
}

type CrossChainAddress struct {
	Address common.Address
	ChainId uint64
}

// https://github.com/symbioticfi/middleware-sdk-mirror/blob/change-header/src/interfaces/implementations/settlement/IMasterConfigProvider.sol
type MasterConfig struct {
	VotingPowerProviders []CrossChainAddress
	KeysProvider         CrossChainAddress
	Replicas             []CrossChainAddress
	VerificationType     uint32
}

// https://github.com/symbioticfi/middleware-sdk-mirror/blob/change-header/src/interfaces/implementations/settlement/IValSetConfigProvider.sol
type VotingPowerConfig struct {
	MaxVotingPower          *big.Int
	MinInclusionVotingPower *big.Int
	MaxValidatorsCount      *big.Int
	RequiredKeyTags         []uint8
}

type NetworkConfig struct {
	MasterConfig      MasterConfig
	VotingPowerConfig VotingPowerConfig
}

// abstract storage, should be persistent, and prunable by epoch
type Storage struct {
	ValSets        map[uint64]ValSet
	NetworkConfigs map[uint64]NetworkConfig // it's required for derivation and new valset header extra data generation

	SignatureRequests map[[32]byte]SignatureRequest
	AggregationProofs map[[32]byte]AggregationProof
	AggregationState  map[[32]byte]AggregationState
	Signatures        map[[32]byte][]Signature
}
