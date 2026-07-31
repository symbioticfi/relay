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

func TestValidatorSetHeaderRoundTrip_PreservesCommitterOrder(t *testing.T) {
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
		CommitterIndices: []uint32{2, 0, 1},
	}

	require.True(
		t,
		valset.IsActiveCommitter(context.Background(), slotDuration, 1050, 0, publicKey3),
		"sanity check: validator 2 should own the first slot before storage round-trip",
	)

	headerBytes, err := ValidatorSetHeaderToBytes(valset)
	require.NoError(t, err)

	_, committerIndices, err := ExtractAdditionalInfoFromHeaderData(headerBytes)
	require.NoError(t, err)

	assert.Equal(
		t,
		valset.CommitterIndices,
		committerIndices,
		"header round-trip should preserve committer order because slot rotation depends on it",
	)

	roundTripped := valset
	roundTripped.CommitterIndices = committerIndices

	assert.True(
		t,
		roundTripped.IsActiveCommitter(context.Background(), slotDuration, 1050, 0, publicKey3),
		"the same validator should still own the first slot after header round-trip",
	)
}
