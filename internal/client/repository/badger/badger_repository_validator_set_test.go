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

	// Test getting latest validator set
	t.Run("get latest validator set", func(t *testing.T) {
		latest, err := repo.GetLatestValidatorSet(t.Context())
		require.NoError(t, err)
		// Latest should be vs1 (epoch 2) even though we saved it first
		assert.Equal(t, vs1, latest)
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
