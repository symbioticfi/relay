package entity

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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
	Address common.Address
	ChainId uint64
}

//	VotingPowerProviders []CrossChainAddress
//	KeysProvider         CrossChainAddress
//	Replicas             []CrossChainAddress
//	VerificationType     uint32

// MaxVotingPower          *big.Int
// MinInclusionVotingPower *big.Int
// MaxValidatorsCount      *big.Int
// RequiredKeyTags         []uint8

// MasterConfig
// https://github.com/symbioticfi/middleware-sdk-mirror/blob/change-header/src/interfaces/implementations/settlement/IMasterConfigProvider.sol
type MasterConfig struct {
	VotingPowerProviders []CrossChainAddress
	KeysProvider         CrossChainAddress
	Replicas             []CrossChainAddress
	VerificationType     uint32
}

// VotingPowerConfig
// https://github.com/symbioticfi/middleware-sdk-mirror/blob/change-header/src/interfaces/implementations/settlement/IValSetConfigProvider.sol
type VotingPowerConfig struct {
	MaxVotingPower          *big.Int
	MinInclusionVotingPower *big.Int
	MaxValidatorsCount      *big.Int
	RequiredKeyTags         []uint8
}

type Config struct {
	MasterConfig

	VotingPowerConfig
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
