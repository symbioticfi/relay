package network

import (
	"math/big"
	"offchain-middleware/bls"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// ValidatorSetHeaderInput represents the input for validator set header
type ValidatorSetHeader struct {
	ActiveAggregatedKeys   []bls.G1
	TotalActiveVotingPower *big.Int
	ValidatorsSszMRoot     [32]byte
	ExtraData              []byte
}

func (v *ValidatorSetHeader) Encode() ([]byte, error) {
	arguments := abi.Arguments{
		{
			Type: abi.Type{
				T:    abi.SliceTy,
				Elem: &abi.Type{T: abi.UintTy, Size: 256}, // G1 points as uint256
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

	return arguments.Pack(v.ActiveAggregatedKeys, v.TotalActiveVotingPower, v.ValidatorsSszMRoot, v.ExtraData)
}
