package badger

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"

	"github.com/symbiotic/relay/core/entity"
)

func TestRepository_PendingValidatorSet(t *testing.T) {
	repo := setupTestRepository(t)

	// Create two validator sets for different request hashes
	reqHash1 := common.BytesToHash(randomBytes(t, 32))

	// Create a validator set with some test data
	vs1 := randomValidatorSet(t, 100, entity.HeaderPending)

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

func randomValidatorSet(t *testing.T, epoch uint64, status entity.ValidatorSetStatus) entity.ValidatorSet {
	t.Helper()
	return entity.ValidatorSet{
		Version:            1,
		RequiredKeyTag:     entity.KeyTag(15),
		Epoch:              epoch,
		CaptureTimestamp:   1234567890,
		QuorumThreshold:    entity.ToVotingPower(big.NewInt(1000)),
		PreviousHeaderHash: common.BytesToHash(randomBytes(t, 32)),
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
		Status: status,
	}
}
