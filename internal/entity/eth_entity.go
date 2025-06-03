package entity

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Phase represents the different phases of the protocol
type Phase uint64

const (
	IDLE Phase = iota
	COMMIT
	FAIL
)

type CrossChainAddress struct {
	Address common.Address
	ChainId uint64
}

type Config struct {
	VotingPowerProviders    []CrossChainAddress
	KeysProvider            CrossChainAddress
	Replicas                []CrossChainAddress
	VerificationType        uint32
	MaxVotingPower          *big.Int
	MinInclusionVotingPower *big.Int
	MaxValidatorsCount      *big.Int
	RequiredKeyTags         []uint8
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
