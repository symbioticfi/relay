package entity

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestValidatorSet_FindValidatorsBySignatures(t *testing.T) {
	// Test data setup
	keyTag := KeyTag(1)
	operator1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	operator2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	operator3 := common.HexToAddress("0x3333333333333333333333333333333333333333")

	publicKey1 := []byte("publickey1")
	publicKey2 := []byte("publickey2")
	publicKey3 := []byte("publickey3")

	// Create validators
	validators := Validators{
		{
			Operator:    operator1,
			VotingPower: VotingPower{big.NewInt(100)},
			IsActive:    true,
			Keys: []ValidatorKey{
				{Tag: keyTag, Payload: publicKey1},
			},
		},
		{
			Operator:    operator2,
			VotingPower: VotingPower{big.NewInt(200)},
			IsActive:    true,
			Keys: []ValidatorKey{
				{Tag: keyTag, Payload: publicKey2},
			},
		},
		{
			Operator:    operator3,
			VotingPower: VotingPower{big.NewInt(150)},
			IsActive:    true,
			Keys: []ValidatorKey{
				{Tag: KeyTag(2), Payload: publicKey3}, // Different key tag
			},
		},
	}

	validatorSet := ValidatorSet{
		Validators: validators,
	}

	t.Run("finds validators matching signatures", func(t *testing.T) {
		result := validatorSet.FindValidatorsBySignatures(keyTag, []CompactPublicKey{publicKey1, publicKey2})

		require.Len(t, result, 2)

		// Check that we found the right validators
		operatorAddrs := make(map[common.Address]bool)
		for _, v := range result {
			operatorAddrs[v.Operator] = true
		}

		require.True(t, operatorAddrs[operator1])
		require.True(t, operatorAddrs[operator2])
		require.False(t, operatorAddrs[operator3])
	})

	t.Run("handles duplicate public keys from same validator", func(t *testing.T) {
		result := validatorSet.FindValidatorsBySignatures(keyTag, []CompactPublicKey{publicKey1, publicKey1})

		require.Len(t, result, 1)
		require.Equal(t, operator1, result[0].Operator)
	})

	t.Run("returns empty slice when no validators match", func(t *testing.T) {
		result := validatorSet.FindValidatorsBySignatures(keyTag, []CompactPublicKey{[]byte("unknown_public_key")})

		require.Empty(t, result)
	})

	t.Run("ignores validators with wrong key tag", func(t *testing.T) {
		// Validator3 has publicKey3 but for different tag
		result := validatorSet.FindValidatorsBySignatures(keyTag, []CompactPublicKey{publicKey3})

		require.Empty(t, result)
	})

	t.Run("handles empty public keys", func(t *testing.T) {
		result := validatorSet.FindValidatorsBySignatures(keyTag, []CompactPublicKey{})

		require.Empty(t, result)
	})
}
