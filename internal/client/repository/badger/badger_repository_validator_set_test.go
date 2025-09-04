package badger

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/core/entity"
)

func TestRepository_ValidatorSet(t *testing.T) {
	repo := setupTestRepository(t)

	// Create two validator sets with different epochs
	vs1 := randomValidatorSet(t, 2)
	vs2 := randomValidatorSet(t, 1)

	// Test saving validator sets
	t.Run("save validator sets", func(t *testing.T) {
		// Save newer epoch first
		err := repo.SaveValidatorSet(t.Context(), vs1)
		require.NoError(t, err)

		// Save older epoch
		err = repo.SaveValidatorSet(t.Context(), vs2)
		require.NoError(t, err)

		// Try to save the same epoch again
		err = repo.SaveValidatorSet(t.Context(), vs1)
		assert.True(t, errors.Is(err, entity.ErrEntityAlreadyExist))
	})

	// Test getting validator set by epoch
	t.Run("get validator set by epoch", func(t *testing.T) {
		// Get newer epoch
		gotVS1, err := repo.GetValidatorSetByEpoch(t.Context(), vs1.Epoch)
		require.NoError(t, err)
		assert.Equal(t, vs1, gotVS1)

		// Get older epoch
		gotVS2, err := repo.GetValidatorSetByEpoch(t.Context(), vs2.Epoch)
		require.NoError(t, err)
		assert.Equal(t, vs2, gotVS2)

		// Get non-existent epoch
		_, err = repo.GetValidatorSetByEpoch(t.Context(), 999)
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
	})

	// Test getting latest validator set via epoch lookup
	t.Run("get latest validator set via epoch lookup", func(t *testing.T) {
		latestEpoch, err := repo.GetLatestValidatorSetEpoch(t.Context())
		require.NoError(t, err)

		latest, err := repo.GetValidatorSetByEpoch(t.Context(), latestEpoch)
		require.NoError(t, err)
		// Latest should be vs1 (epoch 2) even though we saved it first
		assert.Equal(t, vs1, latest)
	})

	// Test getting latest validator set epoch
	t.Run("get latest validator set epoch", func(t *testing.T) {
		latestEpoch, err := repo.GetLatestValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		// Latest epoch should be vs1's epoch (2) even though we saved it first
		assert.Equal(t, vs1.Epoch, latestEpoch)
	})

	// Test getting validator set header by epoch
	t.Run("get validator set header by epoch", func(t *testing.T) {
		// Get header for newer epoch
		gotHeader1, err := repo.GetValidatorSetHeaderByEpoch(t.Context(), vs1.Epoch)
		require.NoError(t, err)

		// Verify header matches expected values from validator set
		expectedHeader1, err := vs1.GetHeader()
		require.NoError(t, err)
		assert.Equal(t, expectedHeader1, gotHeader1)

		// Get header for older epoch
		gotHeader2, err := repo.GetValidatorSetHeaderByEpoch(t.Context(), vs2.Epoch)
		require.NoError(t, err)

		expectedHeader2, err := vs2.GetHeader()
		require.NoError(t, err)
		assert.Equal(t, expectedHeader2, gotHeader2)

		// Get non-existent epoch header
		_, err = repo.GetValidatorSetHeaderByEpoch(t.Context(), 999)
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
	})

	// Test getting latest validator set header
	t.Run("get latest validator set header", func(t *testing.T) {
		latestHeader, err := repo.GetLatestValidatorSetHeader(t.Context())
		require.NoError(t, err)

		// Latest header should be from vs1 (epoch 2)
		expectedHeader, err := vs1.GetHeader()
		require.NoError(t, err)
		assert.Equal(t, expectedHeader, latestHeader)
	})

	// Test getting individual validators by key
	t.Run("get validator by key", func(t *testing.T) {
		// Test with vs1 (epoch 2)
		if len(vs1.Validators) > 0 && len(vs1.Validators[0].Keys) > 0 {
			validator := vs1.Validators[0]
			key := validator.Keys[0]

			// Get validator by key should return the correct validator
			gotValidator, _, err := repo.GetValidatorByKey(t.Context(), vs1.Epoch, key.Tag, key.Payload)
			require.NoError(t, err)
			assert.Equal(t, validator, gotValidator)
		}

		// Test with vs2 (epoch 1)
		if len(vs2.Validators) > 0 && len(vs2.Validators[0].Keys) > 0 {
			validator := vs2.Validators[0]
			key := validator.Keys[0]

			// Get validator by key should return the correct validator
			gotValidator, _, err := repo.GetValidatorByKey(t.Context(), vs2.Epoch, key.Tag, key.Payload)
			require.NoError(t, err)
			assert.Equal(t, validator, gotValidator)
		}

		// Test non-existent validator
		fakeKey := []byte("fake-key-that-does-not-exist")
		_, _, err := repo.GetValidatorByKey(t.Context(), vs1.Epoch, entity.KeyTag(1), fakeKey)
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))

		// Test non-existent epoch
		if len(vs1.Validators) > 0 && len(vs1.Validators[0].Keys) > 0 {
			key := vs1.Validators[0].Keys[0]
			_, _, err := repo.GetValidatorByKey(t.Context(), 999, key.Tag, key.Payload)
			assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
		}
	})
}

func TestRepository_ValidatorSet_EmptyRepository(t *testing.T) {
	repo := setupTestRepository(t)

	t.Run("get latest validator set from empty repo", func(t *testing.T) {
		_, err := repo.GetLatestValidatorSetEpoch(t.Context())
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
	})

	t.Run("get latest validator set header from empty repo", func(t *testing.T) {
		_, err := repo.GetLatestValidatorSetHeader(t.Context())
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
	})

	t.Run("get latest validator set epoch from empty repo", func(t *testing.T) {
		_, err := repo.GetLatestValidatorSetEpoch(t.Context())
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
	})

	t.Run("get validator set by epoch from empty repo", func(t *testing.T) {
		_, err := repo.GetValidatorSetByEpoch(t.Context(), 1)
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
	})

	t.Run("get validator set header by epoch from empty repo", func(t *testing.T) {
		_, err := repo.GetValidatorSetHeaderByEpoch(t.Context(), 1)
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
	})

	t.Run("get validator by key from empty repo", func(t *testing.T) {
		fakeKey := []byte("fake-key")
		_, _, err := repo.GetValidatorByKey(t.Context(), 1, entity.KeyTag(1), fakeKey)
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
	})
}

func TestRepository_ValidatorSet_EpochOrdering(t *testing.T) {
	repo := setupTestRepository(t)

	// Create validator sets with different epochs in non-chronological order
	vs1 := randomValidatorSet(t, 5)
	vs2 := randomValidatorSet(t, 1)
	vs3 := randomValidatorSet(t, 10)
	vs4 := randomValidatorSet(t, 3)

	// Save them in random order
	require.NoError(t, repo.SaveValidatorSet(t.Context(), vs2)) // epoch 1
	require.NoError(t, repo.SaveValidatorSet(t.Context(), vs4)) // epoch 3
	require.NoError(t, repo.SaveValidatorSet(t.Context(), vs3)) // epoch 10
	require.NoError(t, repo.SaveValidatorSet(t.Context(), vs1)) // epoch 5

	t.Run("latest validator set should be highest epoch", func(t *testing.T) {
		latestEpoch, err := repo.GetLatestValidatorSetEpoch(t.Context())
		require.NoError(t, err)

		latest, err := repo.GetValidatorSetByEpoch(t.Context(), latestEpoch)
		require.NoError(t, err)
		assert.Equal(t, vs3, latest) // epoch 10 should be latest
	})

	t.Run("latest validator set header should be highest epoch", func(t *testing.T) {
		latestHeader, err := repo.GetLatestValidatorSetHeader(t.Context())
		require.NoError(t, err)

		expectedHeader, err := vs3.GetHeader()
		require.NoError(t, err)
		assert.Equal(t, expectedHeader, latestHeader)
	})

	t.Run("latest validator set epoch should be highest epoch", func(t *testing.T) {
		latestEpoch, err := repo.GetLatestValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, vs3.Epoch, latestEpoch) // epoch 10 should be latest
	})

	t.Run("can retrieve any validator set by epoch", func(t *testing.T) {
		tests := []entity.ValidatorSet{vs1, vs2, vs3, vs4}
		for _, expected := range tests {
			got, err := repo.GetValidatorSetByEpoch(t.Context(), expected.Epoch)
			require.NoError(t, err)
			assert.Equal(t, expected, got)
		}
	})
}

func TestRepository_ValidatorSet_ValidatorIndexing(t *testing.T) {
	repo := setupTestRepository(t)

	// Create a validator set with multiple validators having multiple keys
	vs := randomValidatorSet(t, 1)
	require.NoError(t, repo.SaveValidatorSet(t.Context(), vs))

	t.Run("find validator by different key tags", func(t *testing.T) {
		for _, validator := range vs.Validators {
			for _, key := range validator.Keys {
				// Should be able to find validator by any of their keys
				foundValidator, _, err := repo.GetValidatorByKey(t.Context(), vs.Epoch, key.Tag, key.Payload)
				require.NoError(t, err)
				assert.Equal(t, validator, foundValidator)
			}
		}
	})

	t.Run("validator lookup with wrong key tag", func(t *testing.T) {
		if len(vs.Validators) > 0 && len(vs.Validators[0].Keys) > 0 {
			validator := vs.Validators[0]
			key := validator.Keys[0]
			wrongKeyTag := key.Tag + 100 // Use a different key tag

			// Should not find validator with wrong key tag but same payload
			_, _, err := repo.GetValidatorByKey(t.Context(), vs.Epoch, wrongKeyTag, key.Payload)
			assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
		}
	})
}

func TestRepository_ValidatorSet_MultiKeyStorageProblem(t *testing.T) {
	repo := setupTestRepository(t)

	t.Run("validator storage duplication with multiple keys", func(t *testing.T) {
		// Create a validator with multiple keys to test storage duplication
		vs := randomValidatorSet(t, 1)

		// Modify the first validator to have 3 different keys
		vs.Validators[0].Keys = []entity.ValidatorKey{
			{Tag: entity.KeyTag(1), Payload: randomBytes(t, 32)},
			{Tag: entity.KeyTag(2), Payload: randomBytes(t, 32)},
			{Tag: entity.KeyTag(3), Payload: randomBytes(t, 32)},
		}

		require.NoError(t, repo.SaveValidatorSet(t.Context(), vs))

		// Retrieve the validator set and check if deduplication works correctly
		retrievedVS, err := repo.GetValidatorSetByEpoch(t.Context(), vs.Epoch)
		require.NoError(t, err)

		// The validator should appear only once despite having multiple keys
		assert.Len(t, retrievedVS.Validators, len(vs.Validators), "Validators should not be duplicated")

		// Find our multi-key validator in the retrieved set
		var foundValidator *entity.Validator
		for _, v := range retrievedVS.Validators {
			if v.Operator == vs.Validators[0].Operator {
				foundValidator = &v
				break
			}
		}
		require.NotNil(t, foundValidator, "Multi-key validator should be found")

		// Verify the validator has all its keys
		assert.Equal(t, vs.Validators[0], *foundValidator, "Retrieved validator should match original")
		assert.Len(t, foundValidator.Keys, 3, "Validator should have all 3 keys")
	})
}

func TestRepository_LatestSignedValidatorSetEpoch(t *testing.T) {
	repo := setupTestRepository(t)

	// Create validator sets with different epochs
	vs1 := randomValidatorSet(t, 1)
	vs2 := randomValidatorSet(t, 5)
	vs3 := randomValidatorSet(t, 3)

	// Save the validator sets first (required for referential integrity)
	require.NoError(t, repo.SaveValidatorSet(t.Context(), vs1))
	require.NoError(t, repo.SaveValidatorSet(t.Context(), vs2))
	require.NoError(t, repo.SaveValidatorSet(t.Context(), vs3))

	t.Run("save and get latest signed validator set epoch", func(t *testing.T) {
		// Save latest signed epoch for vs1 (epoch 1)
		err := repo.SaveLatestSignedValidatorSetEpoch(t.Context(), vs1)
		require.NoError(t, err)

		// Get latest signed epoch - should be vs1.Epoch (1)
		latestSignedEpoch, err := repo.GetLatestSignedValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, vs1.Epoch, latestSignedEpoch)

		// Save latest signed epoch for vs3 (epoch 3) - should overwrite
		err = repo.SaveLatestSignedValidatorSetEpoch(t.Context(), vs3)
		require.NoError(t, err)

		// Get latest signed epoch - should now be vs3.Epoch (3)
		latestSignedEpoch, err = repo.GetLatestSignedValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, vs3.Epoch, latestSignedEpoch)

		// Save latest signed epoch for vs2 (epoch 5) - should overwrite again
		err = repo.SaveLatestSignedValidatorSetEpoch(t.Context(), vs2)
		require.NoError(t, err)

		// Get latest signed epoch - should now be vs2.Epoch (5)
		latestSignedEpoch, err = repo.GetLatestSignedValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, vs2.Epoch, latestSignedEpoch)
	})

	t.Run("latest signed epoch vs latest validator set epoch", func(t *testing.T) {
		// At this point:
		// - Latest validator set epoch should be 5 (highest saved epoch)
		// - Latest signed validator set epoch should be 5 (from previous test)

		latestEpoch, err := repo.GetLatestValidatorSetEpoch(t.Context())
		require.NoError(t, err)

		latestSignedEpoch, err := repo.GetLatestSignedValidatorSetEpoch(t.Context())
		require.NoError(t, err)

		assert.Equal(t, uint64(5), latestEpoch)
		assert.Equal(t, uint64(5), latestSignedEpoch)

		// Now save a signed epoch that's lower than the latest validator set
		err = repo.SaveLatestSignedValidatorSetEpoch(t.Context(), vs1) // epoch 1
		require.NoError(t, err)

		// Latest validator set epoch should remain 5
		latestEpoch, err = repo.GetLatestValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, uint64(5), latestEpoch)

		// But latest signed epoch should now be 1
		latestSignedEpoch, err = repo.GetLatestSignedValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, uint64(1), latestSignedEpoch)
	})
}

func TestRepository_LatestSignedValidatorSetEpoch_EmptyRepository(t *testing.T) {
	repo := setupTestRepository(t)

	t.Run("get latest signed validator set epoch from empty repo", func(t *testing.T) {
		_, err := repo.GetLatestSignedValidatorSetEpoch(t.Context())
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
	})

	t.Run("save latest signed epoch without validator set", func(t *testing.T) {
		// This should work - the implementation doesn't enforce referential integrity
		vs := randomValidatorSet(t, 42)
		err := repo.SaveLatestSignedValidatorSetEpoch(t.Context(), vs)
		require.NoError(t, err)

		// Should be able to retrieve it
		latestSignedEpoch, err := repo.GetLatestSignedValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, vs.Epoch, latestSignedEpoch)
	})
}

func TestRepository_LatestSignedValidatorSetEpoch_Ordering(t *testing.T) {
	repo := setupTestRepository(t)

	// Create validator sets with different epochs
	epochs := []uint64{10, 1, 15, 5, 8, 3}
	validatorSets := make([]entity.ValidatorSet, len(epochs))

	for i, epoch := range epochs {
		validatorSets[i] = randomValidatorSet(t, epoch)
		require.NoError(t, repo.SaveValidatorSet(t.Context(), validatorSets[i]))
	}

	t.Run("save latest signed epochs in random order", func(t *testing.T) {
		// Save signed epochs in a specific order and verify each one
		testOrder := []int{2, 0, 4, 1, 5, 3} // Indices into validatorSets array

		for _, idx := range testOrder {
			vs := validatorSets[idx]
			err := repo.SaveLatestSignedValidatorSetEpoch(t.Context(), vs)
			require.NoError(t, err)

			// Verify it was saved correctly
			latestSignedEpoch, err := repo.GetLatestSignedValidatorSetEpoch(t.Context())
			require.NoError(t, err)
			assert.Equal(t, vs.Epoch, latestSignedEpoch,
				"Latest signed epoch should be %d after saving validator set with epoch %d",
				vs.Epoch, vs.Epoch)
		}

		// Final latest signed epoch should be from the last saved (index 3 = epoch 5)
		latestSignedEpoch, err := repo.GetLatestSignedValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, uint64(5), latestSignedEpoch)

		// But latest validator set epoch should be the highest saved epoch (15)
		latestEpoch, err := repo.GetLatestValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, uint64(15), latestEpoch)
	})
}

func TestRepository_ValidatorSet_ActiveIndex(t *testing.T) {
	repo := setupTestRepository(t)

	// Create validator set with mixed active/inactive validators
	vs := randomValidatorSet(t, 1)

	// Modify validators to have specific active/inactive states and addresses
	// Note: validators must be sorted by operator address ascending
	vs.Validators = []entity.Validator{
		{
			Operator:    common.HexToAddress("0x0000000000000000000000000000000000000000"),
			VotingPower: entity.ToVotingPower(big.NewInt(300)),
			IsActive:    true, // Should get active index 0 (first when sorted by address)
			Keys: []entity.ValidatorKey{
				{Tag: entity.KeyTag(1), Payload: []byte("active_key_0")},
			},
		},
		{
			Operator:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
			VotingPower: entity.ToVotingPower(big.NewInt(200)),
			IsActive:    false, // Should get active index 0 (inactive) - positioned between two active validators
			Keys: []entity.ValidatorKey{
				{Tag: entity.KeyTag(1), Payload: []byte("inactive_key")},
			},
		},
		{
			Operator:    common.HexToAddress("0x3333333333333333333333333333333333333333"),
			VotingPower: entity.ToVotingPower(big.NewInt(100)),
			IsActive:    true, // Should get active index 1 (second active validator, despite inactive validator in between)
			Keys: []entity.ValidatorKey{
				{Tag: entity.KeyTag(1), Payload: []byte("active_key_2")},
			},
		},
	}

	// Save validator set
	err := repo.SaveValidatorSet(t.Context(), vs)
	require.NoError(t, err)

	t.Run("active validator gets correct index", func(t *testing.T) {
		// Test first active validator (0x0000... should be index 0)
		validator, activeIndex, err := repo.GetValidatorByKey(t.Context(), vs.Epoch, entity.KeyTag(1), []byte("active_key_0"))
		require.NoError(t, err)

		assert.Equal(t, common.HexToAddress("0x0000000000000000000000000000000000000000"), validator.Operator)
		assert.True(t, validator.IsActive)
		assert.Equal(t, uint32(0), activeIndex, "First active validator (by address sort) should have index 0")
	})

	t.Run("second active validator gets correct index despite inactive validator in between", func(t *testing.T) {
		// Test second active validator (0x3333... should be index 1, even though inactive 0x2222... is between them)
		validator, activeIndex, err := repo.GetValidatorByKey(t.Context(), vs.Epoch, entity.KeyTag(1), []byte("active_key_2"))
		require.NoError(t, err)

		assert.Equal(t, common.HexToAddress("0x3333333333333333333333333333333333333333"), validator.Operator)
		assert.True(t, validator.IsActive)
		assert.Equal(t, uint32(1), activeIndex, "Second active validator should have index 1, inactive validators should not affect active indexing")
	})

	t.Run("inactive validator gets index 0", func(t *testing.T) {
		// Test inactive validator
		validator, activeIndex, err := repo.GetValidatorByKey(t.Context(), vs.Epoch, entity.KeyTag(1), []byte("inactive_key"))
		require.NoError(t, err)

		assert.Equal(t, common.HexToAddress("0x2222222222222222222222222222222222222222"), validator.Operator)
		assert.False(t, validator.IsActive)
		assert.Equal(t, uint32(0), activeIndex, "Inactive validator should have index 0")
	})
}

func setupTestRepository(t *testing.T) *Repository {
	t.Helper()
	repo, err := New(Config{Dir: t.TempDir()})
	require.NoError(t, err)
	t.Cleanup(func() {
		err := repo.Close()
		require.NoError(t, err)
	})
	return repo
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
