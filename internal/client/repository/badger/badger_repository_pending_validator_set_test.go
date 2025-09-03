package badger

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/core/entity"
)

func TestRepository_PendingValidatorSet(t *testing.T) {
	repo := setupTestRepository(t)

	// Create two validator sets for different request hashes
	reqHash1 := common.BytesToHash(randomBytes(t, 32))

	// Create a validator set with some test data
	vs1 := randomValidatorSet(t, 100)

	// Save the first validator set
	err := repo.SavePendingValidatorSet(t.Context(), reqHash1, vs1)
	require.NoError(t, err)

	// Try to save the same validator set again (should fail)
	err = repo.SavePendingValidatorSet(t.Context(), reqHash1, vs1)
	require.Error(t, err)
	require.True(t, errors.Is(err, entity.ErrEntityAlreadyExist))

	// Get the validator set and verify its contents
	retrievedVS, err := repo.GetPendingValidatorSet(t.Context(), reqHash1)
	require.NoError(t, err)
	require.Equal(t, vs1, retrievedVS)

	// Try to get a non-existent validator set
	reqHash2 := common.BytesToHash(randomBytes(t, 32))
	_, err = repo.GetPendingValidatorSet(t.Context(), reqHash2)
	require.Error(t, err)
	require.True(t, errors.Is(err, entity.ErrEntityNotFound))
}

func TestRepository_GetLatestPendingValidatorSet(t *testing.T) {
	repo := setupTestRepository(t)

	// Test getting latest when no validator sets exist
	_, err := repo.GetLatestPendingValidatorSet(t.Context())
	require.Error(t, err)
	require.True(t, errors.Is(err, entity.ErrEntityNotFound))

	// Create validator sets with different epochs
	reqHash1 := common.BytesToHash(randomBytes(t, 32))
	reqHash2 := common.BytesToHash(randomBytes(t, 32))
	reqHash3 := common.BytesToHash(randomBytes(t, 32))

	vs1 := randomValidatorSet(t, 100) // Lower epoch
	vs2 := randomValidatorSet(t, 200) // Higher epoch
	vs3 := randomValidatorSet(t, 150) // Middle epoch

	// Save validator sets in non-chronological order
	err = repo.SavePendingValidatorSet(t.Context(), reqHash1, vs1)
	require.NoError(t, err)

	// After saving first validator set, it should be the latest
	latestVS, err := repo.GetLatestPendingValidatorSet(t.Context())
	require.NoError(t, err)
	require.Equal(t, vs1, latestVS)

	// Save a validator set with higher epoch
	err = repo.SavePendingValidatorSet(t.Context(), reqHash2, vs2)
	require.NoError(t, err)

	// Now vs2 should be the latest (higher epoch)
	latestVS, err = repo.GetLatestPendingValidatorSet(t.Context())
	require.NoError(t, err)
	require.Equal(t, vs2, latestVS)

	// Save a validator set with middle epoch
	err = repo.SavePendingValidatorSet(t.Context(), reqHash3, vs3)
	require.NoError(t, err)

	// Verify we can still get individual validator sets by hash
	retrievedVS1, err := repo.GetPendingValidatorSet(t.Context(), reqHash1)
	require.NoError(t, err)
	require.Equal(t, vs1, retrievedVS1)

	retrievedVS3, err := repo.GetPendingValidatorSet(t.Context(), reqHash3)
	require.NoError(t, err)
	require.Equal(t, vs3, retrievedVS3)
}

func randomValidatorSet(t *testing.T, epoch uint64) entity.ValidatorSet {
	t.Helper()
	return entity.ValidatorSet{
		Version:          1,
		RequiredKeyTag:   entity.KeyTag(15),
		Epoch:            epoch,
		CaptureTimestamp: 1234567890,
		QuorumThreshold:  entity.ToVotingPower(big.NewInt(1000)),
		Validators: []entity.Validator{
			{
				Operator:    common.BytesToAddress(randomBytes(t, 20)),
				VotingPower: entity.ToVotingPower(big.NewInt(500)),
				IsActive:    true,
				Keys: []entity.ValidatorKey{
					{
						Tag:     entity.KeyTag(15),
						Payload: randomBytes(t, 32),
					},
				},
				Vaults: []entity.ValidatorVault{
					{
						ChainID:     1,
						Vault:       common.BytesToAddress(randomBytes(t, 20)),
						VotingPower: entity.ToVotingPower(big.NewInt(500)),
					},
				},
			},
		},
		Status: entity.HeaderDerived,
	}
}
