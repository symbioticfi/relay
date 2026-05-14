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

	hash := ap.RequestID()

	err := repo.saveAggregationProof(t.Context(), hash, ap)
	require.NoError(t, err)
	err = repo.saveAggregationProof(t.Context(), hash, ap)
	require.ErrorIs(t, err, entity.ErrEntityAlreadyExist)

	loadedConfig, err := repo.GetAggregationProof(t.Context(), hash)
	require.NoError(t, err)
	require.Equal(t, ap, loadedConfig)
}

func TestKeyAggregationProofPendingBinaryFormat(t *testing.T) {
	epoch := symbiotic.Epoch(11)
	requestID := common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111")

	key := keyAggregationProofPending(epoch, requestID)

	extracted, err := extractRequestIDFromEpochDelimitedKey(key, aggregationProofPendingPrefix)
	require.NoError(t, err)
	require.Equal(t, requestID, extracted)
}

func TestBadgerRepository_GetAggregationProofsByEpoch(t *testing.T) {
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

	hash1 := ap1.RequestID()
	hash2 := ap2.RequestID()
	hash3 := ap3.RequestID()

	err := repo.saveAggregationProof(t.Context(), hash1, ap1)
	require.NoError(t, err)

	err = repo.saveAggregationProof(t.Context(), hash2, ap2)
	require.NoError(t, err)

	err = repo.saveAggregationProof(t.Context(), hash3, ap3)
	require.NoError(t, err)

	t.Run("get aggregation proofs for epoch 2", func(t *testing.T) {
		proofs, _, err := repo.GetAggregationProofsByEpoch(t.Context(), 2, 0, nil)
		require.NoError(t, err)

		require.Len(t, proofs, 1)
		require.Equal(t, symbiotic.Epoch(2), proofs[0].Epoch)
		require.Equal(t, ap2, proofs[0])
	})

	t.Run("get aggregation proofs for epoch 1", func(t *testing.T) {
		proofs, _, err := repo.GetAggregationProofsByEpoch(t.Context(), 1, 0, nil)
		require.NoError(t, err)

		require.Len(t, proofs, 1)
		require.Equal(t, symbiotic.Epoch(1), proofs[0].Epoch)
		require.Equal(t, ap1, proofs[0])
	})

	t.Run("get aggregation proofs for epoch 3", func(t *testing.T) {
		proofs, _, err := repo.GetAggregationProofsByEpoch(t.Context(), 3, 0, nil)
		require.NoError(t, err)

		require.Len(t, proofs, 1)
		require.Equal(t, symbiotic.Epoch(3), proofs[0].Epoch)
		require.Equal(t, ap3, proofs[0])
	})

	t.Run("get aggregation proofs for non-existent epoch", func(t *testing.T) {
		proofs, _, err := repo.GetAggregationProofsByEpoch(t.Context(), 10, 0, nil)
		require.NoError(t, err)
		require.Empty(t, proofs)
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
