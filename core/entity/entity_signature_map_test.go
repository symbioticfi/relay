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
		assert.Equal(t, ToVotingPower(big.NewInt(0)), vm.CurrentVotingPower)

		// Verify SignedValidatorsBitmap is empty initially
		assert.True(t, vm.SignedValidatorsBitmap.IsEmpty())
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
		assert.Equal(t, ToVotingPower(big.NewInt(0)), vm.CurrentVotingPower)
		assert.True(t, vm.SignedValidatorsBitmap.IsEmpty())
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
		assert.Equal(t, ToVotingPower(big.NewInt(0)), vm.CurrentVotingPower)
		assert.True(t, vm.SignedValidatorsBitmap.IsEmpty())
	})
}

func TestSignatureMap_SetValidatorPresent(t *testing.T) {
	t.Parallel()

	// Setup common test data
	setupSignatureMap := func() (*SignatureMap, Validator, Validator) {
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

		return &vm, activeValidator1, activeValidator2
	}

	t.Run("successfully sets active validator as present", func(t *testing.T) {
		vm, activeValidator1, _ := setupSignatureMap()

		err := vm.SetValidatorPresent(0, activeValidator1.VotingPower)
		require.NoError(t, err)

		// Verify validator index is marked as present
		assert.True(t, vm.SignedValidatorsBitmap.Contains(0))

		// Verify voting power is updated
		expectedVotingPower := ToVotingPower(big.NewInt(100)) // activeValidator1's voting power
		assert.Equal(t, expectedVotingPower, vm.CurrentVotingPower)
	})

	t.Run("successfully sets multiple validators as present", func(t *testing.T) {
		vm, activeValidator1, activeValidator2 := setupSignatureMap()

		// Set first validator present (index 0)
		err := vm.SetValidatorPresent(0, activeValidator1.VotingPower)
		require.NoError(t, err)

		// Set second validator present (index 1)
		err = vm.SetValidatorPresent(1, activeValidator2.VotingPower)
		require.NoError(t, err)

		// Verify both validator indexes are marked as present
		assert.True(t, vm.SignedValidatorsBitmap.Contains(0))
		assert.True(t, vm.SignedValidatorsBitmap.Contains(1))

		// Verify total voting power is cumulative
		expectedVotingPower := ToVotingPower(big.NewInt(300)) // 100 + 200
		assert.Equal(t, expectedVotingPower, vm.CurrentVotingPower)
	})
	t.Run("returns error when validator index is already present", func(t *testing.T) {
		vm, activeValidator1, _ := setupSignatureMap()

		// Set validator index present first time
		err := vm.SetValidatorPresent(0, activeValidator1.VotingPower)
		require.NoError(t, err)

		// Try to set the same validator index present again
		err = vm.SetValidatorPresent(0, activeValidator1.VotingPower)
		require.Error(t, err)
		assert.True(t, errors.Is(err, errors.New(ErrEntityAlreadyExist)))

		// Verify voting power is not double-counted
		expectedVotingPower := ToVotingPower(big.NewInt(100)) // Should still be 100, not 200
		assert.Equal(t, expectedVotingPower, vm.CurrentVotingPower)
	})
}

func TestSignatureMap_ThresholdReached(t *testing.T) {
	t.Parallel()

	t.Run("returns false when current voting power is below threshold", func(t *testing.T) {
		vm := &SignatureMap{
			CurrentVotingPower: ToVotingPower(big.NewInt(500)),
		}
		quorumThreshold := ToVotingPower(big.NewInt(1000))

		assert.False(t, vm.ThresholdReached(quorumThreshold))
	})

	t.Run("returns true when current voting power equals threshold", func(t *testing.T) {
		vm := &SignatureMap{
			CurrentVotingPower: ToVotingPower(big.NewInt(1000)),
		}
		quorumThreshold := ToVotingPower(big.NewInt(1000))

		assert.True(t, vm.ThresholdReached(quorumThreshold))
	})

	t.Run("returns true when current voting power exceeds threshold", func(t *testing.T) {
		vm := &SignatureMap{
			CurrentVotingPower: ToVotingPower(big.NewInt(1500)),
		}
		quorumThreshold := ToVotingPower(big.NewInt(1000))

		assert.True(t, vm.ThresholdReached(quorumThreshold))
	})

	t.Run("handles zero threshold", func(t *testing.T) {
		vm := &SignatureMap{
			CurrentVotingPower: ToVotingPower(big.NewInt(0)),
		}
		quorumThreshold := ToVotingPower(big.NewInt(0))

		assert.True(t, vm.ThresholdReached(quorumThreshold))
	})

	t.Run("handles zero current voting power", func(t *testing.T) {
		vm := &SignatureMap{
			CurrentVotingPower: ToVotingPower(big.NewInt(0)),
		}
		quorumThreshold := ToVotingPower(big.NewInt(100))

		assert.False(t, vm.ThresholdReached(quorumThreshold))
	})

	t.Run("handles large numbers", func(t *testing.T) {
		largeThreshold := new(big.Int).Exp(big.NewInt(10), big.NewInt(30), nil)
		largeVotingPower := new(big.Int).Add(largeThreshold, big.NewInt(1))

		vm := &SignatureMap{
			CurrentVotingPower: ToVotingPower(largeVotingPower),
		}
		quorumThreshold := ToVotingPower(largeThreshold)

		assert.True(t, vm.ThresholdReached(quorumThreshold))
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
		assert.False(t, vm.ThresholdReached(vs.QuorumThreshold))
		assert.Equal(t, ToVotingPower(big.NewInt(0)), vm.CurrentVotingPower)

		// Add first validator (100) - threshold not reached
		err := vm.SetValidatorPresent(0, validators[0].VotingPower)
		require.NoError(t, err)
		assert.False(t, vm.ThresholdReached(vs.QuorumThreshold))
		assert.Equal(t, ToVotingPower(big.NewInt(100)), vm.CurrentVotingPower)

		// Add second validator (100 + 200 = 300) - threshold not reached
		err = vm.SetValidatorPresent(1, validators[1].VotingPower)
		require.NoError(t, err)
		assert.False(t, vm.ThresholdReached(vs.QuorumThreshold))
		assert.Equal(t, ToVotingPower(big.NewInt(300)), vm.CurrentVotingPower)

		// Add third validator (300 + 300 = 600) - threshold reached!
		err = vm.SetValidatorPresent(2, validators[2].VotingPower)
		require.NoError(t, err)
		assert.True(t, vm.ThresholdReached(vs.QuorumThreshold))
		assert.Equal(t, ToVotingPower(big.NewInt(600)), vm.CurrentVotingPower)

		// Add fourth validator (600 + 150 = 750) - threshold still reached
		err = vm.SetValidatorPresent(3, validators[3].VotingPower)
		require.NoError(t, err)
		assert.True(t, vm.ThresholdReached(vs.QuorumThreshold))
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

		// Add all available active validators (first two are active)
		err := vm.SetValidatorPresent(0, validators[0].VotingPower)
		require.NoError(t, err)

		err = vm.SetValidatorPresent(1, validators[1].VotingPower)
		require.NoError(t, err)

		// Even with all active validators, threshold should not be reached
		assert.False(t, vm.ThresholdReached(vs.QuorumThreshold))
		assert.Equal(t, ToVotingPower(big.NewInt(300)), vm.CurrentVotingPower)
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
		err := vm.SetValidatorPresent(0, validators[0].VotingPower)
		require.NoError(t, err)
		assert.False(t, vm.ThresholdReached(vs.QuorumThreshold))

		// Add second validator - threshold exactly reached
		err = vm.SetValidatorPresent(1, validators[1].VotingPower)
		require.NoError(t, err)
		assert.True(t, vm.ThresholdReached(vs.QuorumThreshold))
		assert.Equal(t, ToVotingPower(big.NewInt(500)), vm.CurrentVotingPower)
	})
}
