package badger

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestBadgerRepository_AggregationProof(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	ap := randomAggregationProof(t)

	hash := common.BytesToHash(randomBytes(t, 32))

	err := repo.saveAggregationProof(t.Context(), hash, ap)
	require.NoError(t, err)
	err = repo.saveAggregationProof(t.Context(), hash, ap)
	require.ErrorIs(t, err, entity.ErrEntityAlreadyExist)

	loadedConfig, err := repo.GetAggregationProof(t.Context(), 10, hash)
	require.NoError(t, err)
	require.Equal(t, ap, loadedConfig)
}

func TestBadgerRepository_GetAggregationProofsByEpoch(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)

	// Create three aggregation proofs with epochs 1, 2, 3
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

	hash1 := common.BytesToHash(randomBytes(t, 32))
	hash2 := common.BytesToHash(randomBytes(t, 32))
	hash3 := common.BytesToHash(randomBytes(t, 32))

	// Save all three aggregation proofs
	err := repo.saveAggregationProof(t.Context(), hash1, ap1)
	require.NoError(t, err)

	err = repo.saveAggregationProof(t.Context(), hash2, ap2)
	require.NoError(t, err)

	err = repo.saveAggregationProof(t.Context(), hash3, ap3)
	require.NoError(t, err)

	t.Run("get aggregation proofs starting from epoch 2", func(t *testing.T) {
		// Query starting from epoch 2
		proofs, err := repo.GetAggregationProofsByEpoch(t.Context(), 2)
		require.NoError(t, err)

		// Should return exactly 2 proofs (epochs 2 and 3)
		require.Len(t, proofs, 2, "Should return 2 aggregation proofs (epochs 2 and 3)")

		// Verify epochs are 2 and 3
		require.Equal(t, symbiotic.Epoch(2), proofs[0].Epoch)
		require.Equal(t, ap2, proofs[0])

		require.Equal(t, symbiotic.Epoch(3), proofs[1].Epoch)
		require.Equal(t, ap3, proofs[1])
	})

	t.Run("get aggregation proofs starting from epoch 1", func(t *testing.T) {
		// Query starting from epoch 1 - should return all 3
		proofs, err := repo.GetAggregationProofsByEpoch(t.Context(), 1)
		require.NoError(t, err)

		require.Len(t, proofs, 3, "Should return all 3 aggregation proofs")
		require.Equal(t, symbiotic.Epoch(1), proofs[0].Epoch)
		require.Equal(t, symbiotic.Epoch(2), proofs[1].Epoch)
		require.Equal(t, symbiotic.Epoch(3), proofs[2].Epoch)
	})

	t.Run("get aggregation proofs starting from epoch 3", func(t *testing.T) {
		// Query starting from epoch 3 - should return only epoch 3
		proofs, err := repo.GetAggregationProofsByEpoch(t.Context(), 3)
		require.NoError(t, err)

		require.Len(t, proofs, 1, "Should return only 1 aggregation proof (epoch 3)")
		require.Equal(t, symbiotic.Epoch(3), proofs[0].Epoch)
		require.Equal(t, ap3, proofs[0])
	})

	t.Run("get aggregation proofs starting from non-existent epoch", func(t *testing.T) {
		// Query starting from epoch 10 (doesn't exist) - should return empty
		proofs, err := repo.GetAggregationProofsByEpoch(t.Context(), 10)
		require.NoError(t, err)
		require.Empty(t, proofs, "Should return empty slice for non-existent epoch")
	})
}

func randomAggregationProof(t *testing.T) symbiotic.AggregationProof {
	t.Helper()

	return symbiotic.AggregationProof{
		MessageHash: randomBytes(t, 32),
		KeyTag:      symbiotic.KeyTag(15),
		Epoch:       10,
		Proof:       randomBytes(t, 32),
	}
}
