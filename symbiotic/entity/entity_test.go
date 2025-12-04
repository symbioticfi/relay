package entity

import (
	"context"
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
		result, err := validatorSet.FindValidatorsByKeys(keyTag, []CompactPublicKey{publicKey1, publicKey2})
		require.NoError(t, err)

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

	t.Run("returns error when no validators match", func(t *testing.T) {
		result, err := validatorSet.FindValidatorsByKeys(keyTag, []CompactPublicKey{[]byte("unknown_public_key")})

		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "validator not found for public key")
	})

	t.Run("returns error for validators with wrong key tag", func(t *testing.T) {
		// Validator3 has publicKey3 but for different tag
		result, err := validatorSet.FindValidatorsByKeys(keyTag, []CompactPublicKey{publicKey3})

		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "validator not found for public key")
	})

	t.Run("handles empty public keys", func(t *testing.T) {
		result, err := validatorSet.FindValidatorsByKeys(keyTag, []CompactPublicKey{})
		require.NoError(t, err)

		require.Empty(t, result)
	})
}

func TestValidatorSet_IsActiveCommitter(t *testing.T) {
	ctx := context.Background()
	keyTag := KeyTag(1)

	// Create test validators with keys
	publicKey1 := []byte("committer1_key")
	publicKey2 := []byte("committer2_key")
	publicKey3 := []byte("committer3_key")
	publicKey4 := []byte("non_committer_key")

	validators := Validators{
		{ // Index 0 - committer
			Operator:    common.HexToAddress("0x1111111111111111111111111111111111111111"),
			VotingPower: VotingPower{big.NewInt(100)},
			IsActive:    true,
			Keys: []ValidatorKey{
				{Tag: keyTag, Payload: publicKey1},
			},
		},
		{ // Index 1 - committer
			Operator:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
			VotingPower: VotingPower{big.NewInt(200)},
			IsActive:    true,
			Keys: []ValidatorKey{
				{Tag: keyTag, Payload: publicKey2},
			},
		},
		{ // Index 2 - committer
			Operator:    common.HexToAddress("0x3333333333333333333333333333333333333333"),
			VotingPower: VotingPower{big.NewInt(150)},
			IsActive:    true,
			Keys: []ValidatorKey{
				{Tag: keyTag, Payload: publicKey3},
			},
		},
		{ // Index 3 - not a committer
			Operator:    common.HexToAddress("0x4444444444444444444444444444444444444444"),
			VotingPower: VotingPower{big.NewInt(100)},
			IsActive:    true,
			Keys: []ValidatorKey{
				{Tag: keyTag, Payload: publicKey4},
			},
		},
	}

	t.Run("node is not a committer", func(t *testing.T) {
		validatorSet := ValidatorSet{
			RequiredKeyTag:   keyTag,
			CaptureTimestamp: 1000,
			Validators:       validators,
			CommitterIndices: []uint32{0, 1, 2}, // Only first three are committers
		}

		// Test with non-committer key
		result := validatorSet.IsActiveCommitter(ctx, 100, 1500, 10, publicKey4)
		require.False(t, result, "Non-committer should return false")

		// Test with unknown key
		result = validatorSet.IsActiveCommitter(ctx, 100, 1500, 10, []byte("unknown_key"))
		require.False(t, result, "Unknown key should return false")
	})

	t.Run("zero committer slot duration", func(t *testing.T) {
		validatorSet := ValidatorSet{
			RequiredKeyTag:   keyTag,
			CaptureTimestamp: 1000,
			Validators:       validators,
			CommitterIndices: []uint32{0, 1, 2},
		}

		// When duration is 0, any committer should always be active
		result := validatorSet.IsActiveCommitter(ctx, 0, 1500, 10, publicKey1)
		require.True(t, result, "With zero duration, committer should always be active")

		result = validatorSet.IsActiveCommitter(ctx, 0, 1500, 10, publicKey2)
		require.True(t, result, "With zero duration, any committer should be active")
	})

	t.Run("single committer", func(t *testing.T) {
		validatorSet := ValidatorSet{
			RequiredKeyTag:   keyTag,
			CaptureTimestamp: 1000,
			Validators:       validators,
			CommitterIndices: []uint32{1}, // Only one committer
		}

		// Single committer should always be active after capture time (no time slot rotation needed)
		result := validatorSet.IsActiveCommitter(ctx, 100, 1000, 10, publicKey2)
		require.True(t, result, "Single committer should be active at capture time")

		result = validatorSet.IsActiveCommitter(ctx, 100, 2000, 10, publicKey2)
		require.True(t, result, "Single committer should be active later")

		// but capture time is in the future, so before that it should be false
		result = validatorSet.IsActiveCommitter(ctx, 100, 999, 10, publicKey2)
		require.False(t, result, "Single committer returns true even before capture time due to early return")

		// But non-committer key should still return false
		result = validatorSet.IsActiveCommitter(ctx, 100, 999, 10, publicKey1)
		require.False(t, result, "Non-committer key should return false")
	})

	t.Run("current time before capture timestamp", func(t *testing.T) {
		validatorSet := ValidatorSet{
			RequiredKeyTag:   keyTag,
			CaptureTimestamp: 1000,
			Validators:       validators,
			CommitterIndices: []uint32{0, 1, 2},
		}

		// No committer should be active before capture timestamp
		result := validatorSet.IsActiveCommitter(ctx, 100, 999, 10, publicKey1)
		require.False(t, result, "Should not be active before capture timestamp")

		result = validatorSet.IsActiveCommitter(ctx, 100, 500, 10, publicKey2)
		require.False(t, result, "Should not be active before capture timestamp")
	})

	t.Run("multiple committers with time slots", func(t *testing.T) {
		validatorSet := ValidatorSet{
			RequiredKeyTag:   keyTag,
			CaptureTimestamp: 1000,
			Validators:       validators,
			CommitterIndices: []uint32{0, 1, 2}, // Three committers
		}

		slotDuration := uint64(100)

		// Test first slot (1000-1099): committer 0 should be active
		result := validatorSet.IsActiveCommitter(ctx, slotDuration, 1000, 0, publicKey1)
		require.True(t, result, "Committer 0 should be active in first slot at start")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1050, 0, publicKey1)
		require.True(t, result, "Committer 0 should be active in first slot middle")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1099, 0, publicKey1)
		require.True(t, result, "Committer 0 should be active in first slot end")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1050, 0, publicKey2)
		require.False(t, result, "Committer 1 should not be active in first slot")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1050, 0, publicKey3)
		require.False(t, result, "Committer 2 should not be active in first slot")

		// Test second slot (1100-1199): committer 1 should be active
		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1100, 0, publicKey2)
		require.True(t, result, "Committer 1 should be active in second slot")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1150, 0, publicKey2)
		require.True(t, result, "Committer 1 should be active in second slot middle")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1100, 0, publicKey1)
		require.False(t, result, "Committer 0 should not be active in second slot")

		// Test third slot (1200-1299): committer 2 should be active
		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1200, 0, publicKey3)
		require.True(t, result, "Committer 2 should be active in third slot")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1250, 0, publicKey3)
		require.True(t, result, "Committer 2 should be active in third slot middle")

		// Test fourth slot (1300-1399): wraps back to committer 0
		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1300, 0, publicKey1)
		require.True(t, result, "Committer 0 should be active in fourth slot (wrap-around)")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1350, 0, publicKey1)
		require.True(t, result, "Committer 0 should be active in fourth slot middle")

		// Test fifth slot (1400-1499): committer 1 again
		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1400, 0, publicKey2)
		require.True(t, result, "Committer 1 should be active in fifth slot")

		// Test sixth slot (1500-1599): committer 2 again
		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1500, 0, publicKey3)
		require.True(t, result, "Committer 2 should be active in sixth slot")
	})

	t.Run("grace period functionality", func(t *testing.T) {
		validatorSet := ValidatorSet{
			RequiredKeyTag:   keyTag,
			CaptureTimestamp: 1000,
			Validators:       validators,
			CommitterIndices: []uint32{0, 1, 2}, // Three committers
		}

		slotDuration := uint64(100)
		graceSeconds := uint64(10)

		// Test grace period at end of slot 0 (time 1095, grace brings us to 1105 which is slot 1)
		result := validatorSet.IsActiveCommitter(ctx, slotDuration, 1095, graceSeconds, publicKey1)
		require.True(t, result, "Committer 0 should still be active in their slot")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1095, graceSeconds, publicKey2)
		require.True(t, result, "Committer 1 should be active with grace period (upcoming slot)")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1095, graceSeconds, publicKey3)
		require.False(t, result, "Committer 2 should not be active even with grace period")

		// Test grace period at end of slot 1 (time 1195, grace brings us to 1205 which is slot 2)
		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1195, graceSeconds, publicKey2)
		require.True(t, result, "Committer 1 should still be active in their slot")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1195, graceSeconds, publicKey3)
		require.True(t, result, "Committer 2 should be active with grace period (upcoming slot)")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1195, graceSeconds, publicKey1)
		require.False(t, result, "Committer 0 should not be active even with grace period")

		// Test grace period at end of slot 2 (time 1295, grace brings us to 1305 which is slot 3, wraps to committer 0)
		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1295, graceSeconds, publicKey3)
		require.True(t, result, "Committer 2 should still be active in their slot")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1295, graceSeconds, publicKey1)
		require.True(t, result, "Committer 0 should be active with grace period (upcoming slot after wrap)")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1295, graceSeconds, publicKey2)
		require.False(t, result, "Committer 1 should not be active even with grace period")

		// Test grace period that doesn't cross slot boundary
		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1050, graceSeconds, publicKey1)
		require.True(t, result, "Committer 0 should be active in middle of their slot")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1050, graceSeconds, publicKey2)
		require.False(t, result, "Committer 1 should not be active when grace doesn't reach their slot")
	})

	t.Run("round-robin scheduling", func(t *testing.T) {
		validatorSet := ValidatorSet{
			RequiredKeyTag:   keyTag,
			CaptureTimestamp: 1000,
			Validators:       validators,
			CommitterIndices: []uint32{2, 0, 1}, // Different order to test proper indexing
		}

		slotDuration := uint64(100)

		// First slot should go to validator at index 2 (first in CommitterIndices)
		result := validatorSet.IsActiveCommitter(ctx, slotDuration, 1050, 0, publicKey3)
		require.True(t, result, "First committer in indices should be active in first slot")

		// Second slot should go to validator at index 0 (second in CommitterIndices)
		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1150, 0, publicKey1)
		require.True(t, result, "Second committer in indices should be active in second slot")

		// Third slot should go to validator at index 1 (third in CommitterIndices)
		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1250, 0, publicKey2)
		require.True(t, result, "Third committer in indices should be active in third slot")

		// Fourth slot wraps back to validator at index 2
		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1350, 0, publicKey3)
		require.True(t, result, "Should wrap back to first committer")
	})

	t.Run("large time values", func(t *testing.T) {
		validatorSet := ValidatorSet{
			RequiredKeyTag:   keyTag,
			CaptureTimestamp: 1000000000, // 1 billion
			Validators:       validators,
			CommitterIndices: []uint32{0, 1},
		}

		slotDuration := uint64(1000000) // 1 million seconds per slot

		// Test first slot
		result := validatorSet.IsActiveCommitter(ctx, slotDuration, 1000500000, 0, publicKey1)
		require.True(t, result, "Committer 0 should be active in first slot with large values")

		// Test second slot
		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1001500000, 0, publicKey2)
		require.True(t, result, "Committer 1 should be active in second slot with large values")

		// Test wrap-around with large values
		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1002500000, 0, publicKey1)
		require.True(t, result, "Should handle wrap-around correctly with large values")
	})

	t.Run("edge case - exact slot boundaries", func(t *testing.T) {
		validatorSet := ValidatorSet{
			RequiredKeyTag:   keyTag,
			CaptureTimestamp: 1000,
			Validators:       validators,
			CommitterIndices: []uint32{0, 1, 2},
		}

		slotDuration := uint64(100)

		// Test exact boundary between slot 0 and slot 1
		result := validatorSet.IsActiveCommitter(ctx, slotDuration, 1099, 0, publicKey1)
		require.True(t, result, "Committer 0 should be active at last moment of slot 0")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1100, 0, publicKey1)
		require.False(t, result, "Committer 0 should not be active at start of slot 1")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1100, 0, publicKey2)
		require.True(t, result, "Committer 1 should be active at start of slot 1")

		// Test exact boundary between slot 1 and slot 2
		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1199, 0, publicKey2)
		require.True(t, result, "Committer 1 should be active at last moment of slot 1")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1200, 0, publicKey2)
		require.False(t, result, "Committer 1 should not be active at start of slot 2")

		result = validatorSet.IsActiveCommitter(ctx, slotDuration, 1200, 0, publicKey3)
		require.True(t, result, "Committer 2 should be active at start of slot 2")
	})

	t.Run("empty committer indices", func(t *testing.T) {
		validatorSet := ValidatorSet{
			RequiredKeyTag:   keyTag,
			CaptureTimestamp: 1000,
			Validators:       validators,
			CommitterIndices: []uint32{}, // No committers
		}

		// No one should be active if there are no committers
		result := validatorSet.IsActiveCommitter(ctx, 100, 1500, 10, publicKey1)
		require.False(t, result, "Should return false when no committers are defined")
	})
}

func TestValidatorSetStatus_TurnOn(t *testing.T) {
	t.Run("HeaderDerived only turns on HeaderDerived", func(t *testing.T) {
		var status ValidatorSetStatus
		status.TurnOn(HeaderDerived)

		require.True(t, status.IsOn(HeaderDerived))
		require.False(t, status.IsOn(HeaderSigned))
		require.False(t, status.IsOn(HeaderAggregated))
		require.False(t, status.IsOn(HeaderCommitted))
		require.False(t, status.IsOn(HeaderMissed))
	})

	t.Run("HeaderSigned turns on HeaderDerived and HeaderSigned", func(t *testing.T) {
		var status ValidatorSetStatus
		status.TurnOn(HeaderSigned)

		require.True(t, status.IsOn(HeaderDerived))
		require.True(t, status.IsOn(HeaderSigned))
		require.False(t, status.IsOn(HeaderAggregated))
		require.False(t, status.IsOn(HeaderCommitted))
		require.False(t, status.IsOn(HeaderMissed))
	})

	t.Run("HeaderAggregated turns on HeaderDerived, HeaderSigned, and HeaderAggregated", func(t *testing.T) {
		var status ValidatorSetStatus
		status.TurnOn(HeaderAggregated)

		require.True(t, status.IsOn(HeaderDerived))
		require.True(t, status.IsOn(HeaderSigned))
		require.True(t, status.IsOn(HeaderAggregated))
		require.False(t, status.IsOn(HeaderCommitted))
		require.False(t, status.IsOn(HeaderMissed))
	})

	t.Run("HeaderCommitted turns on all statuses and turns off HeaderMissed", func(t *testing.T) {
		var status ValidatorSetStatus
		status.TurnOn(HeaderCommitted)

		require.True(t, status.IsOn(HeaderDerived))
		require.True(t, status.IsOn(HeaderSigned))
		require.True(t, status.IsOn(HeaderAggregated))
		require.True(t, status.IsOn(HeaderCommitted))
		require.False(t, status.IsOn(HeaderMissed))
	})

	t.Run("HeaderCommitted clears HeaderMissed if it was previously set", func(t *testing.T) {
		var status = ValidatorSetStatus(HeaderMissed)
		require.True(t, status.IsOn(HeaderMissed))

		status.TurnOn(HeaderCommitted)

		require.True(t, status.IsOn(HeaderDerived))
		require.True(t, status.IsOn(HeaderSigned))
		require.True(t, status.IsOn(HeaderAggregated))
		require.True(t, status.IsOn(HeaderCommitted))
		require.False(t, status.IsOn(HeaderMissed))
	})

	t.Run("HeaderMissed clears HeaderCommitted", func(t *testing.T) {
		var status ValidatorSetStatus
		status.TurnOn(HeaderCommitted)
		require.True(t, status.IsOn(HeaderCommitted))

		status.TurnOn(HeaderMissed)

		require.False(t, status.IsOn(HeaderCommitted))
	})

	t.Run("idempotency - turning on already set status has no effect", func(t *testing.T) {
		var status ValidatorSetStatus
		status.TurnOn(HeaderSigned)
		firstState := status

		status.TurnOn(HeaderSigned)
		secondState := status

		require.Equal(t, firstState, secondState)
		require.True(t, status.IsOn(HeaderDerived))
		require.True(t, status.IsOn(HeaderSigned))
	})

	t.Run("progressive state transitions", func(t *testing.T) {
		var status ValidatorSetStatus

		// Start with HeaderDerived
		status.TurnOn(HeaderDerived)
		require.True(t, status.IsOn(HeaderDerived))
		require.False(t, status.IsOn(HeaderSigned))

		// Add HeaderSigned
		status.TurnOn(HeaderSigned)
		require.True(t, status.IsOn(HeaderDerived))
		require.True(t, status.IsOn(HeaderSigned))
		require.False(t, status.IsOn(HeaderAggregated))

		// Add HeaderAggregated
		status.TurnOn(HeaderAggregated)
		require.True(t, status.IsOn(HeaderDerived))
		require.True(t, status.IsOn(HeaderSigned))
		require.True(t, status.IsOn(HeaderAggregated))
		require.False(t, status.IsOn(HeaderCommitted))

		// Add HeaderCommitted
		status.TurnOn(HeaderCommitted)
		require.True(t, status.IsOn(HeaderDerived))
		require.True(t, status.IsOn(HeaderSigned))
		require.True(t, status.IsOn(HeaderAggregated))
		require.True(t, status.IsOn(HeaderCommitted))
	})

	t.Run("jumping to HeaderAggregated fills in prerequisites", func(t *testing.T) {
		var status ValidatorSetStatus

		// Jump directly to HeaderAggregated
		status.TurnOn(HeaderAggregated)

		// Should have filled in HeaderDerived and HeaderSigned
		require.True(t, status.IsOn(HeaderDerived))
		require.True(t, status.IsOn(HeaderSigned))
		require.True(t, status.IsOn(HeaderAggregated))
	})

	t.Run("TurnOn(HeaderMissed) turns on HeaderMissed and clears HeaderCommitted", func(t *testing.T) {
		var status ValidatorSetStatus
		status.TurnOn(HeaderCommitted)
		require.True(t, status.IsOn(HeaderCommitted))

		// Calling TurnOn(HeaderMissed) should set HeaderMissed and clear HeaderCommitted
		status.TurnOn(HeaderMissed)

		require.True(t, status.IsOn(HeaderMissed))
		require.False(t, status.IsOn(HeaderCommitted))
	})

	t.Run("HeaderMissed only affects HeaderCommitted", func(t *testing.T) {
		var status ValidatorSetStatus
		status.TurnOn(HeaderAggregated)

		// Turn on HeaderMissed - should not change other states besides committed flag
		status.TurnOn(HeaderMissed)

		// Other statuses should remain unchanged while committed is cleared and missed is set
		require.True(t, status.IsOn(HeaderDerived))
		require.True(t, status.IsOn(HeaderSigned))
		require.True(t, status.IsOn(HeaderAggregated))
		require.True(t, status.IsOn(HeaderMissed))
		require.False(t, status.IsOn(HeaderCommitted))
	})

	t.Run("zero status has nothing turned on", func(t *testing.T) {
		var status ValidatorSetStatus

		require.False(t, status.IsOn(HeaderDerived))
		require.False(t, status.IsOn(HeaderSigned))
		require.False(t, status.IsOn(HeaderAggregated))
		require.False(t, status.IsOn(HeaderCommitted))
		require.False(t, status.IsOn(HeaderMissed))
	})
}

func TestPaddedUint64(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected []byte
	}{
		{
			name:     "zero value",
			input:    0,
			expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:     "small positive value",
			input:    255,
			expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff},
		},
		{
			name:     "medium value",
			input:    65535, // 0xFFFF
			expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff},
		},
		{
			name:     "large value",
			input:    4294967295, // 0xFFFFFFFF
			expected: []byte{0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name:     "maximum uint64 value",
			input:    18446744073709551615, // 0xFFFFFFFFFFFFFFFF
			expected: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			name:     "epoch-like value",
			input:    1640995200, // Typical epoch timestamp
			expected: []byte{0x00, 0x00, 0x00, 0x00, 0x61, 0xcf, 0x99, 0x80},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := paddedUint64(tt.input)
			require.Equal(t, tt.expected, result, "paddedUint64(%d) should return correct big-endian bytes", tt.input)
			require.Len(t, result, 8, "paddedUint64 should always return 8 bytes")
		})
	}
}
