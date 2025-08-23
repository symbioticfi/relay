package badger

import (
	"testing"

	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/core/entity"
)

func TestRepository_ValidatorSet(t *testing.T) {
	repo := setupTestRepository(t)

	// Create two validator sets with different epochs
	vs1 := randomValidatorSet(t, 2, entity.HeaderCommitted)
	vs2 := randomValidatorSet(t, 1, entity.HeaderCommitted)

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
			gotValidator, err := repo.GetValidatorByKey(t.Context(), vs1.Epoch, key.Tag, key.Payload)
			require.NoError(t, err)
			assert.Equal(t, validator, gotValidator)
		}

		// Test with vs2 (epoch 1)
		if len(vs2.Validators) > 0 && len(vs2.Validators[0].Keys) > 0 {
			validator := vs2.Validators[0]
			key := validator.Keys[0]

			// Get validator by key should return the correct validator
			gotValidator, err := repo.GetValidatorByKey(t.Context(), vs2.Epoch, key.Tag, key.Payload)
			require.NoError(t, err)
			assert.Equal(t, validator, gotValidator)
		}

		// Test non-existent validator
		fakeKey := []byte("fake-key-that-does-not-exist")
		_, err := repo.GetValidatorByKey(t.Context(), vs1.Epoch, entity.KeyTag(1), fakeKey)
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))

		// Test non-existent epoch
		if len(vs1.Validators) > 0 && len(vs1.Validators[0].Keys) > 0 {
			key := vs1.Validators[0].Keys[0]
			_, err := repo.GetValidatorByKey(t.Context(), 999, key.Tag, key.Payload)
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
		_, err := repo.GetValidatorByKey(t.Context(), 1, entity.KeyTag(1), fakeKey)
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
	})
}

func TestRepository_ValidatorSet_EpochOrdering(t *testing.T) {
	repo := setupTestRepository(t)

	// Create validator sets with different epochs in non-chronological order
	vs1 := randomValidatorSet(t, 5, entity.HeaderCommitted)
	vs2 := randomValidatorSet(t, 1, entity.HeaderCommitted)
	vs3 := randomValidatorSet(t, 10, entity.HeaderCommitted)
	vs4 := randomValidatorSet(t, 3, entity.HeaderCommitted)

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
	vs := randomValidatorSet(t, 1, entity.HeaderCommitted)
	require.NoError(t, repo.SaveValidatorSet(t.Context(), vs))

	t.Run("find validator by different key tags", func(t *testing.T) {
		for _, validator := range vs.Validators {
			for _, key := range validator.Keys {
				// Should be able to find validator by any of their keys
				foundValidator, err := repo.GetValidatorByKey(t.Context(), vs.Epoch, key.Tag, key.Payload)
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
			_, err := repo.GetValidatorByKey(t.Context(), vs.Epoch, wrongKeyTag, key.Payload)
			assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
		}
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
