package bbolt

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestRepository_SaveProofCommitPending(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := symbiotic.Epoch(100)
	requestID := common.HexToHash("0x123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef01")

	t.Run("save new pending proof commit", func(t *testing.T) {
		err := repo.saveProofCommitPending(t.Context(), epoch, requestID)
		require.NoError(t, err)
	})

	t.Run("save duplicate pending proof commit should fail", func(t *testing.T) {
		err := repo.saveProofCommitPending(t.Context(), epoch, requestID)
		require.ErrorIs(t, err, entity.ErrEntityAlreadyExist)
	})

	t.Run("save different epoch should succeed", func(t *testing.T) {
		err := repo.saveProofCommitPending(t.Context(), symbiotic.Epoch(101), requestID)
		require.NoError(t, err)
	})
}

func TestRepository_RemoveProofCommitPending(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := symbiotic.Epoch(200)
	requestID := common.HexToHash("0x987654321fedcba0987654321fedcba0987654321fedcba0987654321fedcba09")

	t.Run("remove non-existent pending proof commit should fail", func(t *testing.T) {
		err := repo.removeProofCommitPending(t.Context(), epoch)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
	})

	t.Run("save and remove pending proof commit", func(t *testing.T) {
		err := repo.saveProofCommitPending(t.Context(), epoch, requestID)
		require.NoError(t, err)

		err = repo.removeProofCommitPending(t.Context(), epoch)
		require.NoError(t, err)

		err = repo.removeProofCommitPending(t.Context(), epoch)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
	})
}

func TestRepository_GetPendingProofCommitsSinceEpoch(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	t.Run("empty repository returns empty list", func(t *testing.T) {
		commits, err := repo.GetPendingProofCommitsSinceEpoch(t.Context(), symbiotic.Epoch(100), 10)
		require.NoError(t, err)
		require.Empty(t, commits)
	})

	t.Run("returns commits for epochs >= since epoch", func(t *testing.T) {
		epoch50 := symbiotic.Epoch(50)
		epoch100 := symbiotic.Epoch(100)
		epoch150 := symbiotic.Epoch(150)

		hash1 := common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111")
		hash2 := common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222")
		hash3 := common.HexToHash("0x3333333333333333333333333333333333333333333333333333333333333333")

		require.NoError(t, repo.saveProofCommitPending(t.Context(), epoch50, hash1))
		require.NoError(t, repo.saveProofCommitPending(t.Context(), epoch100, hash2))
		require.NoError(t, repo.saveProofCommitPending(t.Context(), epoch150, hash3))

		commits, err := repo.GetPendingProofCommitsSinceEpoch(t.Context(), epoch100, 10)
		require.NoError(t, err)
		require.Len(t, commits, 2)

		epochsFound := make(map[symbiotic.Epoch]bool)
		for _, commit := range commits {
			epochsFound[commit.Epoch] = true
		}
		require.True(t, epochsFound[epoch100])
		require.True(t, epochsFound[epoch150])
		require.False(t, epochsFound[epoch50])
	})

	t.Run("limit parameter works correctly", func(t *testing.T) {
		testEpoch := symbiotic.Epoch(600)

		for i := range 5 {
			hash := common.HexToHash(string(rune('a'+i)) + "000000000000000000000000000000000000000000000000000000000000000")
			require.NoError(t, repo.saveProofCommitPending(t.Context(), symbiotic.Epoch(uint64(testEpoch)+uint64(i)), hash))
		}

		commits, err := repo.GetPendingProofCommitsSinceEpoch(t.Context(), testEpoch, 3)
		require.NoError(t, err)
		require.Len(t, commits, 3)

		commits, err = repo.GetPendingProofCommitsSinceEpoch(t.Context(), testEpoch, 0)
		require.NoError(t, err)
		require.Len(t, commits, 5)
	})
}
