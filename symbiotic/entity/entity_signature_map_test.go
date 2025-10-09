package entity

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignatureMap_SetValidatorPresent(t *testing.T) {
	t.Parallel()

	// Setup common test data
	setupSignatureMap := func() (*SignatureMap, Validator, Validator) {
		requestID := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		epoch := uint64(5)

		operator1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
		operator2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

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

		vm := NewSignatureMap(requestID, Epoch(epoch), 2) // 2 total validators for setup

		return &vm, activeValidator1, activeValidator2
	}

	t.Run("successfully sets active validator as present", func(t *testing.T) {
		vm, activeValidator1, _ := setupSignatureMap()

		err := vm.SetValidatorPresent(uint32(0), activeValidator1.VotingPower)
		require.NoError(t, err)

		// Verify validator index is marked as present
		assert.True(t, vm.SignedValidatorsBitmap.Contains(uint32(0)))

		// Verify voting power is updated
		expectedVotingPower := ToVotingPower(big.NewInt(100)) // activeValidator1's voting power
		assert.Equal(t, expectedVotingPower, vm.CurrentVotingPower)
	})

	t.Run("successfully sets multiple validators as present", func(t *testing.T) {
		vm, activeValidator1, activeValidator2 := setupSignatureMap()

		// Set first validator present (index 0)
		err := vm.SetValidatorPresent(uint32(0), activeValidator1.VotingPower)
		require.NoError(t, err)

		// Set second validator present (index 1)
		err = vm.SetValidatorPresent(uint32(1), activeValidator2.VotingPower)
		require.NoError(t, err)

		// Verify both validator indexes are marked as present
		assert.True(t, vm.SignedValidatorsBitmap.Contains(uint32(0)))
		assert.True(t, vm.SignedValidatorsBitmap.Contains(uint32(1)))

		// Verify total voting power is cumulative
		expectedVotingPower := ToVotingPower(big.NewInt(300)) // 100 + 200
		assert.Equal(t, expectedVotingPower, vm.CurrentVotingPower)
	})
	t.Run("returns error when validator index is already present", func(t *testing.T) {
		vm, activeValidator1, _ := setupSignatureMap()

		// Set validator index present first time
		err := vm.SetValidatorPresent(uint32(0), activeValidator1.VotingPower)
		require.NoError(t, err)

		// Try to set the same validator index present again
		err = vm.SetValidatorPresent(uint32(0), activeValidator1.VotingPower)
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
		requestID := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
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

		vm := NewSignatureMap(requestID, Epoch(epoch), 4) // 4 total validators

		// Verify initial state
		assert.False(t, vm.ThresholdReached(quorumThreshold))
		assert.Equal(t, ToVotingPower(big.NewInt(0)), vm.CurrentVotingPower)

		// Add first validator (100) - threshold not reached
		err := vm.SetValidatorPresent(uint32(0), validators[0].VotingPower)
		require.NoError(t, err)
		assert.False(t, vm.ThresholdReached(quorumThreshold))
		assert.Equal(t, ToVotingPower(big.NewInt(100)), vm.CurrentVotingPower)

		// Add second validator (100 + 200 = 300) - threshold not reached
		err = vm.SetValidatorPresent(uint32(1), validators[1].VotingPower)
		require.NoError(t, err)
		assert.False(t, vm.ThresholdReached(quorumThreshold))
		assert.Equal(t, ToVotingPower(big.NewInt(300)), vm.CurrentVotingPower)

		// Add third validator (300 + 300 = 600) - threshold reached!
		err = vm.SetValidatorPresent(uint32(2), validators[2].VotingPower)
		require.NoError(t, err)
		assert.True(t, vm.ThresholdReached(quorumThreshold))
		assert.Equal(t, ToVotingPower(big.NewInt(600)), vm.CurrentVotingPower)

		// Add fourth validator (600 + 150 = 750) - threshold still reached
		err = vm.SetValidatorPresent(uint32(3), validators[3].VotingPower)
		require.NoError(t, err)
		assert.True(t, vm.ThresholdReached(quorumThreshold))
		assert.Equal(t, ToVotingPower(big.NewInt(750)), vm.CurrentVotingPower)
	})

	t.Run("realistic quorum scenario - threshold not reached", func(t *testing.T) {
		requestID := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
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

		vm := NewSignatureMap(requestID, Epoch(epoch), 4) // 4 total validators

		// Add all available active validators (first two are active)
		err := vm.SetValidatorPresent(uint32(0), validators[0].VotingPower)
		require.NoError(t, err)

		err = vm.SetValidatorPresent(uint32(1), validators[1].VotingPower)
		require.NoError(t, err)

		// Even with all active validators, threshold should not be reached
		assert.False(t, vm.ThresholdReached(quorumThreshold))
		assert.Equal(t, ToVotingPower(big.NewInt(300)), vm.CurrentVotingPower)
	})

	t.Run("edge case - exactly 100% participation", func(t *testing.T) {
		requestID := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
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

		vm := NewSignatureMap(requestID, Epoch(epoch), 4) // 4 total validators

		// Add first validator - threshold not reached
		err := vm.SetValidatorPresent(uint32(0), validators[0].VotingPower)
		require.NoError(t, err)
		assert.False(t, vm.ThresholdReached(quorumThreshold))

		// Add second validator - threshold exactly reached
		err = vm.SetValidatorPresent(uint32(1), validators[1].VotingPower)
		require.NoError(t, err)
		assert.True(t, vm.ThresholdReached(quorumThreshold))
		assert.Equal(t, ToVotingPower(big.NewInt(500)), vm.CurrentVotingPower)
	})
}

func TestSignatureMap_GetMissingValidators(t *testing.T) {
	t.Parallel()

	t.Run("returns all validators when none are present", func(t *testing.T) {
		requestID := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		vm := NewSignatureMap(requestID, Epoch(1), 4)

		missing := vm.GetMissingValidators()

		// All validators should be missing
		assert.True(t, missing.Contains(0))
		assert.True(t, missing.Contains(1))
		assert.True(t, missing.Contains(2))
		assert.True(t, missing.Contains(3))
		assert.Equal(t, uint64(4), missing.GetCardinality())
	})

	t.Run("returns subset when some validators are present", func(t *testing.T) {
		requestID := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		vm := NewSignatureMap(requestID, Epoch(1), 4)

		// Set validators 0 and 2 as present
		err := vm.SetValidatorPresent(0, ToVotingPower(big.NewInt(100)))
		require.NoError(t, err)
		err = vm.SetValidatorPresent(2, ToVotingPower(big.NewInt(200)))
		require.NoError(t, err)

		missing := vm.GetMissingValidators()

		// Only validators 1 and 3 should be missing
		assert.False(t, missing.Contains(0))
		assert.True(t, missing.Contains(1))
		assert.False(t, missing.Contains(2))
		assert.True(t, missing.Contains(3))
		assert.Equal(t, uint64(2), missing.GetCardinality())
	})

	t.Run("returns empty when all validators are present", func(t *testing.T) {
		requestID := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		vm := NewSignatureMap(requestID, Epoch(1), 3)

		// Set all validators as present
		for i := uint32(0); i < 3; i++ {
			err := vm.SetValidatorPresent(i, ToVotingPower(big.NewInt(100)))
			require.NoError(t, err)
		}

		missing := vm.GetMissingValidators()

		// No validators should be missing
		assert.False(t, missing.Contains(0))
		assert.False(t, missing.Contains(1))
		assert.False(t, missing.Contains(2))
		assert.Equal(t, uint64(0), missing.GetCardinality())
	})

	t.Run("handles single validator", func(t *testing.T) {
		requestID := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		vm := NewSignatureMap(requestID, Epoch(1), 1)

		missing := vm.GetMissingValidators()

		// Single validator should be missing
		assert.True(t, missing.Contains(0))
		assert.Equal(t, uint64(1), missing.GetCardinality())

		// After setting present
		err := vm.SetValidatorPresent(0, ToVotingPower(big.NewInt(100)))
		require.NoError(t, err)

		missing = vm.GetMissingValidators()
		assert.False(t, missing.Contains(0))
		assert.Equal(t, uint64(0), missing.GetCardinality())
	})
}

func TestBitmapFromBytes(t *testing.T) {
	t.Parallel()

	t.Run("creates bitmap from valid bytes", func(t *testing.T) {
		// Create a bitmap with some values
		original := NewBitmapOf(0, 2, 5, 10)

		// Serialize to bytes
		buf, err := original.ToBytes()
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		// Deserialize from bytes
		restored, err := BitmapFromBytes(buf)
		require.NoError(t, err)

		// Verify the restored bitmap has the same values
		assert.True(t, restored.Contains(0))
		assert.False(t, restored.Contains(1))
		assert.True(t, restored.Contains(2))
		assert.False(t, restored.Contains(3))
		assert.False(t, restored.Contains(4))
		assert.True(t, restored.Contains(5))
		assert.True(t, restored.Contains(10))
		assert.Equal(t, original.GetCardinality(), restored.GetCardinality())
	})

	t.Run("creates empty bitmap from empty bytes", func(t *testing.T) {
		// Create empty bitmap
		original := NewBitmap()

		// Serialize to bytes
		buf, err := original.ToBytes()
		require.NoError(t, err)

		// Deserialize from bytes
		restored, err := BitmapFromBytes(buf)
		require.NoError(t, err)

		// Verify the restored bitmap is empty
		assert.Equal(t, uint64(0), restored.GetCardinality())
	})

	t.Run("returns error for invalid bytes", func(t *testing.T) {
		// Test with invalid byte sequence
		invalidBytes := []byte{0xFF, 0xFF, 0xFF, 0xFF}

		_, err := BitmapFromBytes(invalidBytes)
		assert.Error(t, err)
	})

	t.Run("handles large bitmap", func(t *testing.T) {
		// Create bitmap with large indices
		original := NewBitmapOf(0, 100, 1000, 10000, 65536)

		// Serialize to bytes
		buf, err := original.ToBytes()
		require.NoError(t, err)

		// Deserialize from bytes
		restored, err := BitmapFromBytes(buf)
		require.NoError(t, err)

		// Verify all large indices are preserved
		assert.True(t, restored.Contains(0))
		assert.True(t, restored.Contains(100))
		assert.True(t, restored.Contains(1000))
		assert.True(t, restored.Contains(10000))
		assert.True(t, restored.Contains(65536))
		assert.Equal(t, original.GetCardinality(), restored.GetCardinality())
	})

	t.Run("roundtrip consistency", func(t *testing.T) {
		// Test multiple roundtrips to ensure consistency
		testValues := []uint32{1, 3, 7, 15, 31, 63, 127, 255, 511, 1023}
		original := NewBitmapOf(testValues...)

		for i := 0; i < 3; i++ {
			// Serialize
			buf, err := original.ToBytes()
			require.NoError(t, err)

			// Deserialize
			restored, err := BitmapFromBytes(buf)
			require.NoError(t, err)

			// Verify equality
			assert.Equal(t, original.GetCardinality(), restored.GetCardinality())
			for _, val := range testValues {
				assert.True(t, restored.Contains(val), "value %d should be present after roundtrip %d", val, i+1)
			}

			// Use restored for next iteration
			original = restored
		}
	})
}
