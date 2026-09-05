package ssz

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestVotingPowerProofLeavesPreserveVersionOneConsensus(t *testing.T) {
	expected := [32]byte{1}

	validator := &SszValidator{
		Operator:    common.HexToAddress("0x1111111111111111111111111111111111111111"),
		VotingPower: big.NewInt(1),
	}
	validatorProof, err := validator.ProveValidatorVotingPower()
	require.NoError(t, err)
	require.Equal(t, expected[:], validatorProof.Leaf)

	vault := &SszVault{
		ChainId:     1,
		Vault:       common.HexToAddress("0x2222222222222222222222222222222222222222"),
		VotingPower: big.NewInt(1),
	}
	vaultProof, err := vault.ProveVaultVotingPower()
	require.NoError(t, err)
	require.Equal(t, expected[:], vaultProof.Leaf)
}

func TestInvalidVotingPowerReturnsError(t *testing.T) {
	for _, power := range []*big.Int{nil, big.NewInt(-1), new(big.Int).Lsh(big.NewInt(1), 256)} {
		require.NotPanics(t, func() {
			validator := &SszValidator{VotingPower: power}
			_, err := validator.MarshalSSZ()
			require.Error(t, err)
			_, err = validator.HashTreeRoot()
			require.Error(t, err)
			vault := &SszVault{VotingPower: power}
			_, err = vault.MarshalSSZ()
			require.Error(t, err)
			_, err = vault.HashTreeRoot()
			require.Error(t, err)
		})
	}
}
