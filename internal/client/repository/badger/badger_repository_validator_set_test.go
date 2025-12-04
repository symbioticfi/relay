package badger

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestRepository_ValidatorSet(t *testing.T) {
	repo := setupTestRepository(t)

	// Create two validator sets with different epochs
	vs1 := randomValidatorSet(t, 2)
	vs2 := randomValidatorSet(t, 1)

	// Test saving validator sets
	t.Run("save validator sets", func(t *testing.T) {
		// Save newer epoch first
		err := repo.saveValidatorSet(t.Context(), vs1)
		require.NoError(t, err)

		// Save older epoch
		err = repo.saveValidatorSet(t.Context(), vs2)
		require.NoError(t, err)

		// Try to save the same epoch again
		err = repo.saveValidatorSet(t.Context(), vs1)
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
		_, _, err := repo.GetValidatorByKey(t.Context(), vs1.Epoch, symbiotic.KeyTag(1), fakeKey)
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
		_, _, err := repo.GetValidatorByKey(t.Context(), 1, symbiotic.KeyTag(1), fakeKey)
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
	})
}

func TestRepository_GetOldestValidatorSetEpoch(t *testing.T) {
	t.Parallel()
	t.Run("returns error when repository is empty", func(t *testing.T) {
		repo := setupTestRepository(t)
		_, err := repo.GetOldestValidatorSetEpoch(t.Context())
		require.Error(t, err)
		assert.ErrorIs(t, err, entity.ErrEntityNotFound)
	})

	t.Run("returns earliest epoch", func(t *testing.T) {
		repo := setupTestRepository(t)

		valsets := []symbiotic.ValidatorSet{
			randomValidatorSet(t, 10),
			randomValidatorSet(t, 5),
			randomValidatorSet(t, 7),
		}

		for _, vs := range valsets {
			require.NoError(t, repo.saveValidatorSet(t.Context(), vs))
		}

		epoch, err := repo.GetOldestValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, symbiotic.Epoch(5), epoch)
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
	require.NoError(t, repo.saveValidatorSet(t.Context(), vs2)) // epoch 1
	require.NoError(t, repo.saveValidatorSet(t.Context(), vs4)) // epoch 3
	require.NoError(t, repo.saveValidatorSet(t.Context(), vs3)) // epoch 10
	require.NoError(t, repo.saveValidatorSet(t.Context(), vs1)) // epoch 5

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
		tests := []symbiotic.ValidatorSet{vs1, vs2, vs3, vs4}
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
	require.NoError(t, repo.saveValidatorSet(t.Context(), vs))

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
		vs.Validators[0].Keys = []symbiotic.ValidatorKey{
			{Tag: symbiotic.KeyTag(1), Payload: randomBytes(t, 32)},
			{Tag: symbiotic.KeyTag(2), Payload: randomBytes(t, 32)},
			{Tag: symbiotic.KeyTag(3), Payload: randomBytes(t, 32)},
		}

		require.NoError(t, repo.saveValidatorSet(t.Context(), vs))

		// Retrieve the validator set and check if deduplication works correctly
		retrievedVS, err := repo.GetValidatorSetByEpoch(t.Context(), vs.Epoch)
		require.NoError(t, err)

		// The validator should appear only once despite having multiple keys
		assert.Len(t, retrievedVS.Validators, len(vs.Validators), "Validators should not be duplicated")

		// Find our multi-key validator in the retrieved set
		var foundValidator *symbiotic.Validator
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

func TestRepository_ValidatorSet_ActiveIndex(t *testing.T) {
	repo := setupTestRepository(t)

	// Create validator set with mixed active/inactive validators
	vs := randomValidatorSet(t, 1)

	// Modify validators to have specific active/inactive states and addresses
	// Note: validators must be sorted by operator address ascending
	vs.Validators = []symbiotic.Validator{
		{
			Operator:    common.HexToAddress("0x0000000000000000000000000000000000000000"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(300)),
			IsActive:    true, // Should get active index 0 (first when sorted by address)
			Keys: []symbiotic.ValidatorKey{
				{Tag: symbiotic.KeyTag(1), Payload: []byte("active_key_0")},
			},
		},
		{
			Operator:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(200)),
			IsActive:    false, // Should get active index 0 (inactive) - positioned between two active validators
			Keys: []symbiotic.ValidatorKey{
				{Tag: symbiotic.KeyTag(1), Payload: []byte("inactive_key")},
			},
		},
		{
			Operator:    common.HexToAddress("0x3333333333333333333333333333333333333333"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
			IsActive:    true, // Should get active index 1 (second active validator, despite inactive validator in between)
			Keys: []symbiotic.ValidatorKey{
				{Tag: symbiotic.KeyTag(1), Payload: []byte("active_key_2")},
			},
		},
	}

	// Save validator set
	err := repo.saveValidatorSet(t.Context(), vs)
	require.NoError(t, err)

	t.Run("active validator gets correct index", func(t *testing.T) {
		// Test first active validator (0x0000... should be index 0)
		validator, activeIndex, err := repo.GetValidatorByKey(t.Context(), vs.Epoch, symbiotic.KeyTag(1), []byte("active_key_0"))
		require.NoError(t, err)

		assert.Equal(t, common.HexToAddress("0x0000000000000000000000000000000000000000"), validator.Operator)
		assert.True(t, validator.IsActive)
		assert.Equal(t, uint32(0), activeIndex, "First active validator (by address sort) should have index 0")
	})

	t.Run("second active validator gets correct index despite inactive validator in between", func(t *testing.T) {
		// Test second active validator (0x3333... should be index 1, even though inactive 0x2222... is between them)
		validator, activeIndex, err := repo.GetValidatorByKey(t.Context(), vs.Epoch, symbiotic.KeyTag(1), []byte("active_key_2"))
		require.NoError(t, err)

		assert.Equal(t, common.HexToAddress("0x3333333333333333333333333333333333333333"), validator.Operator)
		assert.True(t, validator.IsActive)
		assert.Equal(t, uint32(1), activeIndex, "Second active validator should have index 1, inactive validators should not affect active indexing")
	})

	t.Run("inactive validator gets index 0", func(t *testing.T) {
		// Test inactive validator
		validator, activeIndex, err := repo.GetValidatorByKey(t.Context(), vs.Epoch, symbiotic.KeyTag(1), []byte("inactive_key"))
		require.NoError(t, err)

		assert.Equal(t, common.HexToAddress("0x2222222222222222222222222222222222222222"), validator.Operator)
		assert.False(t, validator.IsActive)
		assert.Equal(t, uint32(0), activeIndex, "Inactive validator should have index 0")
	})
}

func TestRepository_FirstUncommittedValidatorSetEpoch(t *testing.T) {
	repo := setupTestRepository(t)

	t.Run("save and get first uncommitted epoch", func(t *testing.T) {
		// Save first uncommitted epoch
		err := repo.SaveFirstUncommittedValidatorSetEpoch(t.Context(), 42)
		require.NoError(t, err)

		// Get first uncommitted epoch
		epoch, err := repo.GetFirstUncommittedValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, symbiotic.Epoch(42), epoch)

		// Update to a different epoch
		err = repo.SaveFirstUncommittedValidatorSetEpoch(t.Context(), 100)
		require.NoError(t, err)

		// Verify it was updated
		epoch, err = repo.GetFirstUncommittedValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, symbiotic.Epoch(100), epoch)
	})
}

func TestRepository_FirstUncommittedValidatorSetEpoch_EmptyRepository(t *testing.T) {
	repo := setupTestRepository(t)

	t.Run("get first uncommitted epoch from empty repo", func(t *testing.T) {
		// Should return 0 and no error when not set (based on implementation)
		epoch, err := repo.GetFirstUncommittedValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, symbiotic.Epoch(0), epoch)
	})
}

func TestRepository_GetValidatorSetsByEpoch(t *testing.T) {
	repo := setupTestRepository(t)

	// Create three validator sets with epochs 1, 2, 3
	vs1 := randomValidatorSet(t, 1)
	vs2 := randomValidatorSet(t, 2)
	vs3 := randomValidatorSet(t, 3)

	// Save all three validator sets
	require.NoError(t, repo.saveValidatorSet(t.Context(), vs1))
	require.NoError(t, repo.saveValidatorSet(t.Context(), vs2))
	require.NoError(t, repo.saveValidatorSet(t.Context(), vs3))

	t.Run("get validator sets starting from epoch 2", func(t *testing.T) {
		// Query starting from epoch 2
		validatorSets, err := repo.GetValidatorSetsStartingFromEpoch(t.Context(), 2)
		require.NoError(t, err)

		// Should return exactly 2 validator sets (epochs 2 and 3)
		require.Len(t, validatorSets, 2, "Should return 2 validator sets (epochs 2 and 3)")

		// First result should be epoch 2
		assert.Equal(t, symbiotic.Epoch(2), validatorSets[0].Epoch)
		assert.Equal(t, vs2, validatorSets[0])

		// Second result should be epoch 3
		assert.Equal(t, symbiotic.Epoch(3), validatorSets[1].Epoch)
		assert.Equal(t, vs3, validatorSets[1])
	})

	t.Run("get validator sets starting from epoch 1", func(t *testing.T) {
		// Query starting from epoch 1 - should return all 3
		validatorSets, err := repo.GetValidatorSetsStartingFromEpoch(t.Context(), 1)
		require.NoError(t, err)

		require.Len(t, validatorSets, 3, "Should return all 3 validator sets")
		assert.Equal(t, symbiotic.Epoch(1), validatorSets[0].Epoch)
		assert.Equal(t, symbiotic.Epoch(2), validatorSets[1].Epoch)
		assert.Equal(t, symbiotic.Epoch(3), validatorSets[2].Epoch)
	})

	t.Run("get validator sets starting from epoch 3", func(t *testing.T) {
		// Query starting from epoch 3 - should return only epoch 3
		validatorSets, err := repo.GetValidatorSetsStartingFromEpoch(t.Context(), 3)
		require.NoError(t, err)

		require.Len(t, validatorSets, 1, "Should return only 1 validator set (epoch 3)")
		assert.Equal(t, symbiotic.Epoch(3), validatorSets[0].Epoch)
		assert.Equal(t, vs3, validatorSets[0])
	})

	t.Run("get validator sets starting from non-existent epoch", func(t *testing.T) {
		// Query starting from epoch 10 (doesn't exist) - should return empty
		validatorSets, err := repo.GetValidatorSetsStartingFromEpoch(t.Context(), 10)
		require.NoError(t, err)
		assert.Empty(t, validatorSets, "Should return empty slice for non-existent epoch")
	})
}

func setupTestRepository(t *testing.T) *Repository {
	t.Helper()
	repo, err := New(Config{Dir: t.TempDir(), Metrics: DoNothingMetrics{}})
	require.NoError(t, err)
	t.Cleanup(func() {
		err := repo.Close()
		require.NoError(t, err)
	})
	return repo
}

func randomValidatorSet(t *testing.T, epoch symbiotic.Epoch) symbiotic.ValidatorSet {
	t.Helper()
	return symbiotic.ValidatorSet{
		Version:          1,
		RequiredKeyTag:   symbiotic.KeyTag(15),
		Epoch:            epoch,
		CaptureTimestamp: 1234567890,
		QuorumThreshold:  symbiotic.ToVotingPower(big.NewInt(1000)),
		Validators: []symbiotic.Validator{
			{
				Operator:    common.BytesToAddress(randomBytes(t, 20)),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(500)),
				IsActive:    true,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(15),
						Payload: randomBytes(t, 32),
					},
				},
				Vaults: []symbiotic.ValidatorVault{
					{
						ChainID:     1,
						Vault:       common.BytesToAddress(randomBytes(t, 20)),
						VotingPower: symbiotic.ToVotingPower(big.NewInt(500)),
					},
				},
			},
		},
		Status:            symbiotic.ValidatorSetStatus(symbiotic.HeaderCommitted),
		AggregatorIndices: []uint32{},
		CommitterIndices:  []uint32{},
	}
}
