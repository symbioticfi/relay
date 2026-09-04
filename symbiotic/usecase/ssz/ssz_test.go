package ssz

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestVotingPowerProofLeavesUseFixedWidthEncoding(t *testing.T) {
	expected := [32]byte{}
	big.NewInt(1).FillBytes(expected[:])

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
