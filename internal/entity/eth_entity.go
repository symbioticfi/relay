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
	VotingPowerProviders    []CrossChainAddress `json:"votingPowerProviders"`
	KeysProvider            CrossChainAddress   `json:"keysProvider"`
	Replicas                []CrossChainAddress `json:"replicas"`
	VerificationType        uint32              `json:"verificationType"`
	MaxVotingPower          *big.Int            `json:"maxVotingPower"`
	MinInclusionVotingPower *big.Int            `json:"minInclusionVotingPower"`
	MaxValidatorsCount      *big.Int            `json:"maxValidatorsCount"`
	RequiredKeyTags         []KeyTag            `json:"requiredKeyTags"`
}

type NetworkData struct {
	Address    common.Address `json:"address"`
	Subnetwork [32]byte       `json:"subnetwork"`
	Eip712Data Eip712Domain   `json:"eip712Data"`
}

type VaultVotingPower struct {
	Vault       common.Address `json:"vault"`
	VotingPower *big.Int       `json:"votingPower"`
}

type OperatorVotingPower struct {
	Operator common.Address     `json:"operator"`
	Vaults   []VaultVotingPower `json:"vaults"`
}

type OperatorWithKeys struct {
	Operator common.Address `json:"operator"`
	Keys     []Key          `json:"keys"`
}

type Eip712Domain struct {
	Fields            [1]byte        `json:"fields"`
	Name              string         `json:"name"`
	Version           string         `json:"version"`
	ChainId           *big.Int       `json:"chainId"`
	VerifyingContract common.Address `json:"verifyingContract"`
	Salt              *big.Int       `json:"salt"`
	Extensions        []*big.Int     `json:"extensions"`
}
