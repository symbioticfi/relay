package entity

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSignatureMap(t *testing.T) {
	t.Parallel()

	t.Run("creates signature map with active validators", func(t *testing.T) {
		// Setup test data
		requestHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		epoch := uint64(5)
		quorumThreshold := ToVotingPower(big.NewInt(1000))

		operator1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
		operator2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
		operator3 := common.HexToAddress("0x3333333333333333333333333333333333333333")

		validators := Validators{
			{
				Operator:    operator1,
				VotingPower: ToVotingPower(big.NewInt(100)),
				IsActive:    true,
			},
			{
				Operator:    operator2,
				VotingPower: ToVotingPower(big.NewInt(200)),
				IsActive:    true,
			},
			{
				Operator:    operator3,
				VotingPower: ToVotingPower(big.NewInt(150)),
				IsActive:    false, // Inactive validator
			},
		}

		vs := ValidatorSet{
			Epoch:           epoch,
			QuorumThreshold: quorumThreshold,
			Validators:      validators,
		}

		// Create signatures map
		vm := NewSignatureMap(requestHash, vs)

		// Verify basic properties
		assert.Equal(t, requestHash, vm.RequestHash)
		assert.Equal(t, epoch, vm.Epoch)
		assert.Equal(t, quorumThreshold, vm.QuorumThreshold)
		assert.Equal(t, ToVotingPower(big.NewInt(300)), vm.TotalVotingPower) // Only active validators: 100 + 200
		assert.Equal(t, ToVotingPower(big.NewInt(0)), vm.CurrentVotingPower)

		// Verify active validators map contains only active validators
		assert.Len(t, vm.ActiveValidatorsMap, 2)
		assert.Contains(t, vm.ActiveValidatorsMap, operator1)
		assert.Contains(t, vm.ActiveValidatorsMap, operator2)
		assert.NotContains(t, vm.ActiveValidatorsMap, operator3) // Inactive should not be included

		// Verify SignedValidatorIndexes map is empty initially
		assert.Empty(t, vm.SignedValidatorIndexes)
	})

	t.Run("creates signature map with no active validators", func(t *testing.T) {
		requestHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		epoch := uint64(1)
		quorumThreshold := ToVotingPower(big.NewInt(500))

		validators := Validators{
			{
				Operator:    common.HexToAddress("0x1111111111111111111111111111111111111111"),
				VotingPower: ToVotingPower(big.NewInt(100)),
				IsActive:    false,
			},
			{
				Operator:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
				VotingPower: ToVotingPower(big.NewInt(200)),
				IsActive:    false,
			},
		}

		vs := ValidatorSet{
			Epoch:           epoch,
			QuorumThreshold: quorumThreshold,
			Validators:      validators,
		}

		vm := NewSignatureMap(requestHash, vs)

		assert.Equal(t, requestHash, vm.RequestHash)
		assert.Equal(t, epoch, vm.Epoch)
		assert.Equal(t, quorumThreshold, vm.QuorumThreshold)
		assert.Equal(t, ToVotingPower(big.NewInt(0)), vm.TotalVotingPower) // No active validators
		assert.Equal(t, ToVotingPower(big.NewInt(0)), vm.CurrentVotingPower)
		assert.Empty(t, vm.ActiveValidatorsMap)
		assert.Empty(t, vm.SignedValidatorIndexes)
	})

	t.Run("creates signature map with empty validator set", func(t *testing.T) {
		requestHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		epoch := uint64(0)
		quorumThreshold := ToVotingPower(big.NewInt(0))

		vs := ValidatorSet{
			Epoch:           epoch,
			QuorumThreshold: quorumThreshold,
			Validators:      Validators{},
		}

		vm := NewSignatureMap(requestHash, vs)

		assert.Equal(t, requestHash, vm.RequestHash)
		assert.Equal(t, epoch, vm.Epoch)
		assert.Equal(t, quorumThreshold, vm.QuorumThreshold)
		assert.Equal(t, ToVotingPower(big.NewInt(0)), vm.TotalVotingPower)
		assert.Equal(t, ToVotingPower(big.NewInt(0)), vm.CurrentVotingPower)
		assert.Empty(t, vm.ActiveValidatorsMap)
		assert.Empty(t, vm.SignedValidatorIndexes)
	})
}

func TestSignatureMap_SetValidatorPresent(t *testing.T) {
	t.Parallel()

	// Setup common test data
	setupSignatureMap := func() (*SignatureMap, Validator, Validator, Validator) {
		requestHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		epoch := uint64(5)
		quorumThreshold := ToVotingPower(big.NewInt(250))

		operator1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
		operator2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
		operator3 := common.HexToAddress("0x3333333333333333333333333333333333333333")

		activeValidator1 := Validator{
			Operator:    operator1,
			VotingPower: ToVotingPower(big.NewInt(100)),
			IsActive:    true,
		}

		activeValidator2 := Validator{
			Operator:    operator2,
			VotingPower: ToVotingPower(big.NewInt(200)),
			IsActive:    true,
		}

		inactiveValidator := Validator{
			Operator:    operator3,
			VotingPower: ToVotingPower(big.NewInt(150)),
			IsActive:    false,
		}

		validators := Validators{activeValidator1, activeValidator2, inactiveValidator}

		vs := ValidatorSet{
			Epoch:           epoch,
			QuorumThreshold: quorumThreshold,
			Validators:      validators,
		}

		vm := NewSignatureMap(requestHash, vs)

		return &vm, activeValidator1, activeValidator2, inactiveValidator
	}

	t.Run("successfully sets active validator as present", func(t *testing.T) {
		vm, activeValidator1, _, _ := setupSignatureMap()

		err := vm.SetValidatorPresent(activeValidator1, 0)
		require.NoError(t, err)

		// Verify validator is marked as present
		assert.Contains(t, vm.SignedValidatorIndexes, activeValidator1.Operator)

		// Verify voting power is updated
		expectedVotingPower := ToVotingPower(big.NewInt(100)) // activeValidator1's voting power
		assert.Equal(t, expectedVotingPower, vm.CurrentVotingPower)
	})

	t.Run("successfully sets multiple validators as present", func(t *testing.T) {
		vm, activeValidator1, activeValidator2, _ := setupSignatureMap()

		// Set first validator present
		err := vm.SetValidatorPresent(activeValidator1, 0)
		require.NoError(t, err)

		// Set second validator present
		err = vm.SetValidatorPresent(activeValidator2, 0)
		require.NoError(t, err)

		// Verify both validators are marked as present
		assert.Contains(t, vm.SignedValidatorIndexes, activeValidator1.Operator)
		assert.Contains(t, vm.SignedValidatorIndexes, activeValidator2.Operator)

		// Verify total voting power is cumulative
		expectedVotingPower := ToVotingPower(big.NewInt(300)) // 100 + 200
		assert.Equal(t, expectedVotingPower, vm.CurrentVotingPower)
	})

	t.Run("returns error when validator is not in active validators map", func(t *testing.T) {
		vm, _, _, inactiveValidator := setupSignatureMap()

		err := vm.SetValidatorPresent(inactiveValidator, 0)
		require.Error(t, err)
		assert.True(t, errors.Is(err, errors.New(ErrValidatorNotFound)))

		// Verify validator is not marked as present
		assert.NotContains(t, vm.SignedValidatorIndexes, inactiveValidator.Operator)

		// Verify voting power remains unchanged
		assert.Equal(t, ToVotingPower(big.NewInt(0)), vm.CurrentVotingPower)
	})

	t.Run("returns error when validator is already present", func(t *testing.T) {
		vm, activeValidator1, _, _ := setupSignatureMap()

		// Set validator present first time
		err := vm.SetValidatorPresent(activeValidator1, 0)
		require.NoError(t, err)

		// Try to set the same validator present again
		err = vm.SetValidatorPresent(activeValidator1, 0)
		require.Error(t, err)
		assert.True(t, errors.Is(err, errors.New(ErrEntityAlreadyExist)))

		// Verify voting power is not double-counted
		expectedVotingPower := ToVotingPower(big.NewInt(100)) // Should still be 100, not 200
		assert.Equal(t, expectedVotingPower, vm.CurrentVotingPower)
	})

	t.Run("returns error for unknown validator", func(t *testing.T) {
		vm, _, _, _ := setupSignatureMap()

		unknownValidator := Validator{
			Operator:    common.HexToAddress("0x9999999999999999999999999999999999999999"),
			VotingPower: ToVotingPower(big.NewInt(50)),
			IsActive:    true,
		}

		err := vm.SetValidatorPresent(unknownValidator, 0)
		require.Error(t, err)
		assert.True(t, errors.Is(err, errors.New(ErrValidatorNotFound)))

		// Verify validator is not marked as present
		assert.NotContains(t, vm.SignedValidatorIndexes, unknownValidator.Operator)

		// Verify voting power remains unchanged
		assert.Equal(t, ToVotingPower(big.NewInt(0)), vm.CurrentVotingPower)
	})
}

func TestSignatureMap_ThresholdReached(t *testing.T) {
	t.Parallel()

	t.Run("returns false when current voting power is below threshold", func(t *testing.T) {
		vm := &SignatureMap{
			QuorumThreshold:    ToVotingPower(big.NewInt(1000)),
			CurrentVotingPower: ToVotingPower(big.NewInt(500)),
		}

		assert.False(t, vm.ThresholdReached())
	})

	t.Run("returns true when current voting power equals threshold", func(t *testing.T) {
		vm := &SignatureMap{
			QuorumThreshold:    ToVotingPower(big.NewInt(1000)),
			CurrentVotingPower: ToVotingPower(big.NewInt(1000)),
		}

		assert.True(t, vm.ThresholdReached())
	})

	t.Run("returns true when current voting power exceeds threshold", func(t *testing.T) {
		vm := &SignatureMap{
			QuorumThreshold:    ToVotingPower(big.NewInt(1000)),
			CurrentVotingPower: ToVotingPower(big.NewInt(1500)),
		}

		assert.True(t, vm.ThresholdReached())
	})

	t.Run("handles zero threshold", func(t *testing.T) {
		vm := &SignatureMap{
			QuorumThreshold:    ToVotingPower(big.NewInt(0)),
			CurrentVotingPower: ToVotingPower(big.NewInt(0)),
		}

		assert.True(t, vm.ThresholdReached())
	})

	t.Run("handles zero current voting power", func(t *testing.T) {
		vm := &SignatureMap{
			QuorumThreshold:    ToVotingPower(big.NewInt(100)),
			CurrentVotingPower: ToVotingPower(big.NewInt(0)),
		}

		assert.False(t, vm.ThresholdReached())
	})

	t.Run("handles large numbers", func(t *testing.T) {
		largeThreshold := new(big.Int).Exp(big.NewInt(10), big.NewInt(30), nil)
		largeVotingPower := new(big.Int).Add(largeThreshold, big.NewInt(1))

		vm := &SignatureMap{
			QuorumThreshold:    ToVotingPower(largeThreshold),
			CurrentVotingPower: ToVotingPower(largeVotingPower),
		}

		assert.True(t, vm.ThresholdReached())
	})
}

func TestSignatureMap_IntegrationScenarios(t *testing.T) {
	t.Parallel()

	t.Run("realistic quorum scenario - threshold reached", func(t *testing.T) {
		requestHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		epoch := uint64(10)

		// Setup validators with different voting powers
		validators := Validators{
			{
				Operator:    common.HexToAddress("0x1111111111111111111111111111111111111111"),
				VotingPower: ToVotingPower(big.NewInt(100)),
				IsActive:    true,
			},
			{
				Operator:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
				VotingPower: ToVotingPower(big.NewInt(200)),
				IsActive:    true,
			},
			{
				Operator:    common.HexToAddress("0x3333333333333333333333333333333333333333"),
				VotingPower: ToVotingPower(big.NewInt(300)),
				IsActive:    true,
			},
			{
				Operator:    common.HexToAddress("0x4444444444444444444444444444444444444444"),
				VotingPower: ToVotingPower(big.NewInt(150)),
				IsActive:    true,
			},
		}

		// Total active voting power: 750
		// Set quorum threshold to 67% (approximately 500)
		quorumThreshold := ToVotingPower(big.NewInt(500))

		vs := ValidatorSet{
			Epoch:           epoch,
			QuorumThreshold: quorumThreshold,
			Validators:      validators,
		}

		vm := NewSignatureMap(requestHash, vs)

		// Verify initial state
		assert.False(t, vm.ThresholdReached())
		assert.Equal(t, ToVotingPower(big.NewInt(0)), vm.CurrentVotingPower)

		// Add first validator (100) - threshold not reached
		err := vm.SetValidatorPresent(validators[0], 0)
		require.NoError(t, err)
		assert.False(t, vm.ThresholdReached())
		assert.Equal(t, ToVotingPower(big.NewInt(100)), vm.CurrentVotingPower)

		// Add second validator (100 + 200 = 300) - threshold not reached
		err = vm.SetValidatorPresent(validators[1], 0)
		require.NoError(t, err)
		assert.False(t, vm.ThresholdReached())
		assert.Equal(t, ToVotingPower(big.NewInt(300)), vm.CurrentVotingPower)

		// Add third validator (300 + 300 = 600) - threshold reached!
		err = vm.SetValidatorPresent(validators[2], 0)
		require.NoError(t, err)
		assert.True(t, vm.ThresholdReached())
		assert.Equal(t, ToVotingPower(big.NewInt(600)), vm.CurrentVotingPower)

		// Add fourth validator (600 + 150 = 750) - threshold still reached
		err = vm.SetValidatorPresent(validators[3], 0)
		require.NoError(t, err)
		assert.True(t, vm.ThresholdReached())
		assert.Equal(t, ToVotingPower(big.NewInt(750)), vm.CurrentVotingPower)
	})

	t.Run("realistic quorum scenario - threshold not reached", func(t *testing.T) {
		requestHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		epoch := uint64(15)

		validators := Validators{
			{
				Operator:    common.HexToAddress("0x1111111111111111111111111111111111111111"),
				VotingPower: ToVotingPower(big.NewInt(100)),
				IsActive:    true,
			},
			{
				Operator:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
				VotingPower: ToVotingPower(big.NewInt(200)),
				IsActive:    true,
			},
			{
				Operator:    common.HexToAddress("0x3333333333333333333333333333333333333333"),
				VotingPower: ToVotingPower(big.NewInt(300)),
				IsActive:    false, // Inactive - not available for signatures
			},
		}

		// Total active voting power: 300 (only first two validators)
		// Set high quorum threshold that can't be reached
		quorumThreshold := ToVotingPower(big.NewInt(500))

		vs := ValidatorSet{
			Epoch:           epoch,
			QuorumThreshold: quorumThreshold,
			Validators:      validators,
		}

		vm := NewSignatureMap(requestHash, vs)

		// Add all available active validators
		err := vm.SetValidatorPresent(validators[0], 0)
		require.NoError(t, err)

		err = vm.SetValidatorPresent(validators[1], 0)
		require.NoError(t, err)

		// Even with all active validators, threshold should not be reached
		assert.False(t, vm.ThresholdReached())
		assert.Equal(t, ToVotingPower(big.NewInt(300)), vm.CurrentVotingPower)

		// Verify we can't add inactive validator
		err = vm.SetValidatorPresent(validators[2], 0)
		require.Error(t, err)
		assert.True(t, errors.Is(err, errors.New(ErrValidatorNotFound)))
	})

	t.Run("edge case - exactly 100% participation", func(t *testing.T) {
		requestHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		epoch := uint64(20)

		validators := Validators{
			{
				Operator:    common.HexToAddress("0x1111111111111111111111111111111111111111"),
				VotingPower: ToVotingPower(big.NewInt(250)),
				IsActive:    true,
			},
			{
				Operator:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
				VotingPower: ToVotingPower(big.NewInt(250)),
				IsActive:    true,
			},
		}

		// Total active voting power: 500
		// Set quorum threshold to 100%
		quorumThreshold := ToVotingPower(big.NewInt(500))

		vs := ValidatorSet{
			Epoch:           epoch,
			QuorumThreshold: quorumThreshold,
			Validators:      validators,
		}

		vm := NewSignatureMap(requestHash, vs)

		// Add first validator - threshold not reached
		err := vm.SetValidatorPresent(validators[0], 0)
		require.NoError(t, err)
		assert.False(t, vm.ThresholdReached())

		// Add second validator - threshold exactly reached
		err = vm.SetValidatorPresent(validators[1], 0)
		require.NoError(t, err)
		assert.True(t, vm.ThresholdReached())
		assert.Equal(t, ToVotingPower(big.NewInt(500)), vm.CurrentVotingPower)
		assert.Equal(t, vm.CurrentVotingPower, vm.TotalVotingPower)
	})
}
