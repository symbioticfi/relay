package bbolt

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

	vs1 := randomValidatorSet(t, 2)
	vs2 := randomValidatorSet(t, 1)

	t.Run("save validator sets", func(t *testing.T) {
		err := repo.saveValidatorSet(t.Context(), vs1)
		require.NoError(t, err)

		err = repo.saveValidatorSet(t.Context(), vs2)
		require.NoError(t, err)

		err = repo.saveValidatorSet(t.Context(), vs1)
		assert.True(t, errors.Is(err, entity.ErrEntityAlreadyExist))
	})

	t.Run("get validator set by epoch", func(t *testing.T) {
		gotVS1, err := repo.GetValidatorSetByEpoch(t.Context(), vs1.Epoch)
		require.NoError(t, err)
		assert.Equal(t, vs1, gotVS1)

		gotVS2, err := repo.GetValidatorSetByEpoch(t.Context(), vs2.Epoch)
		require.NoError(t, err)
		assert.Equal(t, vs2, gotVS2)

		_, err = repo.GetValidatorSetByEpoch(t.Context(), 999)
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
	})

	t.Run("get latest validator set epoch", func(t *testing.T) {
		latestEpoch, err := repo.GetLatestValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, vs1.Epoch, latestEpoch)
	})

	t.Run("get latest validator set header", func(t *testing.T) {
		latestHeader, err := repo.GetLatestValidatorSetHeader(t.Context())
		require.NoError(t, err)

		expectedHeader, err := vs1.GetHeader()
		require.NoError(t, err)
		assert.Equal(t, expectedHeader, latestHeader)
	})

	t.Run("get validator set header by epoch", func(t *testing.T) {
		gotHeader1, err := repo.GetValidatorSetHeaderByEpoch(t.Context(), vs1.Epoch)
		require.NoError(t, err)

		expectedHeader1, err := vs1.GetHeader()
		require.NoError(t, err)
		assert.Equal(t, expectedHeader1, gotHeader1)

		_, err = repo.GetValidatorSetHeaderByEpoch(t.Context(), 999)
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
	})

	t.Run("get validator by key", func(t *testing.T) {
		if len(vs1.Validators) > 0 && len(vs1.Validators[0].Keys) > 0 {
			validator := vs1.Validators[0]
			key := validator.Keys[0]

			gotValidator, _, err := repo.GetValidatorByKey(t.Context(), vs1.Epoch, key.Tag, key.Payload)
			require.NoError(t, err)
			assert.Equal(t, validator, gotValidator)
		}

		fakeKey := []byte("fake-key-that-does-not-exist")
		_, _, err := repo.GetValidatorByKey(t.Context(), vs1.Epoch, symbiotic.KeyTag(1), fakeKey)
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
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

	t.Run("get validator set by epoch from empty repo", func(t *testing.T) {
		_, err := repo.GetValidatorSetByEpoch(t.Context(), 1)
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

	vs1 := randomValidatorSet(t, 5)
	vs2 := randomValidatorSet(t, 1)
	vs3 := randomValidatorSet(t, 10)
	vs4 := randomValidatorSet(t, 3)

	require.NoError(t, repo.saveValidatorSet(t.Context(), vs2))
	require.NoError(t, repo.saveValidatorSet(t.Context(), vs4))
	require.NoError(t, repo.saveValidatorSet(t.Context(), vs3))
	require.NoError(t, repo.saveValidatorSet(t.Context(), vs1))

	t.Run("latest validator set should be highest epoch", func(t *testing.T) {
		latestEpoch, err := repo.GetLatestValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, vs3.Epoch, latestEpoch)
	})

	t.Run("can retrieve any validator set by epoch", func(t *testing.T) {
		for _, expected := range []symbiotic.ValidatorSet{vs1, vs2, vs3, vs4} {
			got, err := repo.GetValidatorSetByEpoch(t.Context(), expected.Epoch)
			require.NoError(t, err)
			assert.Equal(t, expected, got)
		}
	})
}

func TestRepository_ValidatorSet_ActiveIndex(t *testing.T) {
	repo := setupTestRepository(t)

	vs := randomValidatorSet(t, 1)
	vs.Validators = []symbiotic.Validator{
		{
			Operator:    common.HexToAddress("0x0000000000000000000000000000000000000000"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(300)),
			IsActive:    true,
			Keys:        []symbiotic.ValidatorKey{{Tag: symbiotic.KeyTag(1), Payload: []byte("active_key_0")}},
		},
		{
			Operator:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(200)),
			IsActive:    false,
			Keys:        []symbiotic.ValidatorKey{{Tag: symbiotic.KeyTag(1), Payload: []byte("inactive_key")}},
		},
		{
			Operator:    common.HexToAddress("0x3333333333333333333333333333333333333333"),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
			IsActive:    true,
			Keys:        []symbiotic.ValidatorKey{{Tag: symbiotic.KeyTag(1), Payload: []byte("active_key_2")}},
		},
	}

	err := repo.saveValidatorSet(t.Context(), vs)
	require.NoError(t, err)

	t.Run("active validator gets correct index", func(t *testing.T) {
		_, activeIndex, err := repo.GetValidatorByKey(t.Context(), vs.Epoch, symbiotic.KeyTag(1), []byte("active_key_0"))
		require.NoError(t, err)
		assert.Equal(t, uint32(0), activeIndex)
	})

	t.Run("second active validator gets correct index", func(t *testing.T) {
		_, activeIndex, err := repo.GetValidatorByKey(t.Context(), vs.Epoch, symbiotic.KeyTag(1), []byte("active_key_2"))
		require.NoError(t, err)
		assert.Equal(t, uint32(1), activeIndex)
	})

	t.Run("inactive validator gets index 0", func(t *testing.T) {
		_, activeIndex, err := repo.GetValidatorByKey(t.Context(), vs.Epoch, symbiotic.KeyTag(1), []byte("inactive_key"))
		require.NoError(t, err)
		assert.Equal(t, uint32(0), activeIndex)
	})
}

func TestRepository_FirstUncommittedValidatorSetEpoch(t *testing.T) {
	repo := setupTestRepository(t)

	t.Run("save and get first uncommitted epoch", func(t *testing.T) {
		err := repo.SaveFirstUncommittedValidatorSetEpoch(t.Context(), 42)
		require.NoError(t, err)

		epoch, err := repo.GetFirstUncommittedValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, symbiotic.Epoch(42), epoch)

		err = repo.SaveFirstUncommittedValidatorSetEpoch(t.Context(), 100)
		require.NoError(t, err)

		epoch, err = repo.GetFirstUncommittedValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, symbiotic.Epoch(100), epoch)
	})
}

func TestRepository_FirstUncommittedValidatorSetEpoch_EmptyRepository(t *testing.T) {
	repo := setupTestRepository(t)

	t.Run("get first uncommitted epoch from empty repo", func(t *testing.T) {
		epoch, err := repo.GetFirstUncommittedValidatorSetEpoch(t.Context())
		require.NoError(t, err)
		assert.Equal(t, symbiotic.Epoch(0), epoch)
	})
}

func TestRepository_GetValidatorSetsByEpoch(t *testing.T) {
	repo := setupTestRepository(t)

	vs1 := randomValidatorSet(t, 1)
	vs2 := randomValidatorSet(t, 2)
	vs3 := randomValidatorSet(t, 3)

	require.NoError(t, repo.saveValidatorSet(t.Context(), vs1))
	require.NoError(t, repo.saveValidatorSet(t.Context(), vs2))
	require.NoError(t, repo.saveValidatorSet(t.Context(), vs3))

	t.Run("get validator sets starting from epoch 2", func(t *testing.T) {
		validatorSets, err := repo.GetValidatorSetsStartingFromEpoch(t.Context(), 2)
		require.NoError(t, err)
		require.Len(t, validatorSets, 2)
		assert.Equal(t, symbiotic.Epoch(2), validatorSets[0].Epoch)
		assert.Equal(t, symbiotic.Epoch(3), validatorSets[1].Epoch)
	})

	t.Run("get validator sets starting from epoch 1", func(t *testing.T) {
		validatorSets, err := repo.GetValidatorSetsStartingFromEpoch(t.Context(), 1)
		require.NoError(t, err)
		require.Len(t, validatorSets, 3)
	})

	t.Run("get validator sets starting from non-existent epoch", func(t *testing.T) {
		validatorSets, err := repo.GetValidatorSetsStartingFromEpoch(t.Context(), 10)
		require.NoError(t, err)
		assert.Empty(t, validatorSets)
	})
}
