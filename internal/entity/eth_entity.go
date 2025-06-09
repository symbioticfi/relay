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
