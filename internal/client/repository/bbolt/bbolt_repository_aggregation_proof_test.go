package bbolt

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestRepository_AggregationProof(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	ap := randomAggregationProof(t)
	hash := ap.RequestID()

	err := repo.saveAggregationProof(t.Context(), hash, ap)
	require.NoError(t, err)

	err = repo.saveAggregationProof(t.Context(), hash, ap)
	require.ErrorIs(t, err, entity.ErrEntityAlreadyExist)

	loadedProof, err := repo.GetAggregationProof(t.Context(), hash)
	require.NoError(t, err)
	require.Equal(t, ap, loadedProof)
}

func TestRepository_GetAggregationProofsStartingFromEpoch(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)

	ap1 := symbiotic.AggregationProof{
		MessageHash: randomBytes(t, 32),
		KeyTag:      symbiotic.KeyTag(15),
		Epoch:       1,
		Proof:       randomBytes(t, 32),
	}

	ap2 := symbiotic.AggregationProof{
		MessageHash: randomBytes(t, 32),
		KeyTag:      symbiotic.KeyTag(15),
		Epoch:       2,
		Proof:       randomBytes(t, 32),
	}

	ap3 := symbiotic.AggregationProof{
		MessageHash: randomBytes(t, 32),
		KeyTag:      symbiotic.KeyTag(15),
		Epoch:       3,
		Proof:       randomBytes(t, 32),
	}

	require.NoError(t, repo.saveAggregationProof(t.Context(), ap1.RequestID(), ap1))
	require.NoError(t, repo.saveAggregationProof(t.Context(), ap2.RequestID(), ap2))
	require.NoError(t, repo.saveAggregationProof(t.Context(), ap3.RequestID(), ap3))

	t.Run("get aggregation proofs starting from epoch 2", func(t *testing.T) {
		proofs, err := repo.GetAggregationProofsStartingFromEpoch(t.Context(), 2)
		require.NoError(t, err)
		require.Len(t, proofs, 2)
		require.Equal(t, symbiotic.Epoch(2), proofs[0].Epoch)
		require.Equal(t, symbiotic.Epoch(3), proofs[1].Epoch)
	})

	t.Run("get aggregation proofs starting from epoch 1", func(t *testing.T) {
		proofs, err := repo.GetAggregationProofsStartingFromEpoch(t.Context(), 1)
		require.NoError(t, err)
		require.Len(t, proofs, 3)
	})

	t.Run("get aggregation proofs starting from non-existent epoch", func(t *testing.T) {
		proofs, err := repo.GetAggregationProofsStartingFromEpoch(t.Context(), 10)
		require.NoError(t, err)
		require.Empty(t, proofs)
	})
}

func TestRepository_GetAggregationProofsByEpoch(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)

	ap1 := symbiotic.AggregationProof{
		MessageHash: randomBytes(t, 32),
		KeyTag:      symbiotic.KeyTag(15),
		Epoch:       1,
		Proof:       randomBytes(t, 32),
	}

	ap2 := symbiotic.AggregationProof{
		MessageHash: randomBytes(t, 32),
		KeyTag:      symbiotic.KeyTag(15),
		Epoch:       2,
		Proof:       randomBytes(t, 32),
	}

	require.NoError(t, repo.saveAggregationProof(t.Context(), ap1.RequestID(), ap1))
	require.NoError(t, repo.saveAggregationProof(t.Context(), ap2.RequestID(), ap2))

	t.Run("get aggregation proofs for epoch 1", func(t *testing.T) {
		proofs, err := repo.GetAggregationProofsByEpoch(t.Context(), 1)
		require.NoError(t, err)
		require.Len(t, proofs, 1)
		require.Equal(t, ap1, proofs[0])
	})

	t.Run("get aggregation proofs for non-existent epoch", func(t *testing.T) {
		proofs, err := repo.GetAggregationProofsByEpoch(t.Context(), 10)
		require.NoError(t, err)
		require.Empty(t, proofs)
	})
}
