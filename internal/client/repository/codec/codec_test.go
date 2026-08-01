package codec

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// Committer indices serialize as a bitmap, so the round-trip returns the set in
// ascending order rather than any insertion order (see the deriver, which sorts
// them for exactly this reason). Slot rotation indexes into that slice, so a
// dropped or spurious member would hand a slot to the wrong validator.
func TestValidatorSetHeaderRoundTrip_PreservesCommitterSet(t *testing.T) {
	t.Parallel()

	const slotDuration = uint64(100)
	keyTag := symbiotic.KeyTag(15)

	publicKey1 := []byte("committer1_key")
	publicKey2 := []byte("committer2_key")
	publicKey3 := []byte("committer3_key")

	valset := symbiotic.ValidatorSet{
		RequiredKeyTag:   keyTag,
		CaptureTimestamp: 1000,
		Validators: symbiotic.Validators{
			{
				Operator:    common.HexToAddress("0x1111111111111111111111111111111111111111"),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
				IsActive:    true,
				Keys: []symbiotic.ValidatorKey{
					{Tag: keyTag, Payload: publicKey1},
				},
			},
			{
				Operator:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(200)),
				IsActive:    true,
				Keys: []symbiotic.ValidatorKey{
					{Tag: keyTag, Payload: publicKey2},
				},
			},
			{
				Operator:    common.HexToAddress("0x3333333333333333333333333333333333333333"),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(150)),
				IsActive:    true,
				Keys: []symbiotic.ValidatorKey{
					{Tag: keyTag, Payload: publicKey3},
				},
			},
		},
		// A strict subset, ascending, as the deriver emits it.
		CommitterIndices: []uint32{0, 2},
	}

	require.True(
		t,
		valset.IsActiveCommitter(context.Background(), slotDuration, 1050, 0, publicKey1),
		"sanity check: validator 0 should own the first slot before storage round-trip",
	)

	headerBytes, err := ValidatorSetHeaderToBytes(valset)
	require.NoError(t, err)

	_, committerIndices, err := ExtractAdditionalInfoFromHeaderData(headerBytes)
	require.NoError(t, err)

	assert.Equal(
		t,
		valset.CommitterIndices,
		committerIndices,
		"header round-trip should preserve the committer set in ascending order",
	)

	roundTripped := valset
	roundTripped.CommitterIndices = committerIndices

	assert.True(
		t,
		roundTripped.IsActiveCommitter(context.Background(), slotDuration, 1050, 0, publicKey1),
		"the same validator should still own the first slot after header round-trip",
	)
	assert.False(
		t,
		roundTripped.IsActiveCommitter(context.Background(), slotDuration, 1050, 0, publicKey2),
		"a non-committer should not gain a slot through the round-trip",
	)
}
