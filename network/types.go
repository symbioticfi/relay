package network

import (
	"math/big"
	"offchain-middleware/bls"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// ValidatorSetHeaderInput represents the input for validator set header
type ValidatorSetHeader struct {
	ActiveAggregatedKeys   []G1
	TotalActiveVotingPower *big.Int
	ValidatorsSszMRoot     [32]byte
	ExtraData              []byte
}

type G1 [2]*big.Int

func FormatG1(g1 *bls.G1) G1 {
	G1 := G1{new(big.Int), new(big.Int)}

	g1.G1Affine.X.BigInt(G1[0])
	g1.G1Affine.Y.BigInt(G1[1])
	return G1
}

func (v *ValidatorSetHeader) Encode() ([]byte, error) {
	arguments := abi.Arguments{
		{
			Type: abi.Type{
				T: abi.SliceTy,
				Elem: &abi.Type{
					T:    abi.ArrayTy,
					Size: 2,
					Elem: &abi.Type{T: abi.UintTy, Size: 256}, // G1 points as array of two uint256
				},
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
