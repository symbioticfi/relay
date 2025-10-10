package badger

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestBadgerRepository_SaveProofCommitPending(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := symbiotic.Epoch(100)
	requestID := common.HexToHash("0x123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef01")

	t.Run("save new pending proof commit", func(t *testing.T) {
		err := repo.SaveProofCommitPending(t.Context(), epoch, requestID)
		require.NoError(t, err)
	})

	t.Run("save duplicate pending proof commit should fail", func(t *testing.T) {
		err := repo.SaveProofCommitPending(t.Context(), epoch, requestID)
		require.ErrorIs(t, err, entity.ErrEntityAlreadyExist)
	})

	t.Run("save different epoch should succeed", func(t *testing.T) {
		differentEpoch := symbiotic.Epoch(101)
		err := repo.SaveProofCommitPending(t.Context(), differentEpoch, requestID)
		require.NoError(t, err)
	})

	t.Run("save different hash should succeed", func(t *testing.T) {
		differentHash := common.HexToHash("0xabcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789")
		err := repo.SaveProofCommitPending(t.Context(), epoch, differentHash)
		require.NoError(t, err)
	})

	t.Run("save with zero epoch should succeed", func(t *testing.T) {
		zeroEpoch := symbiotic.Epoch(0)
		zeroHash := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000")
		err := repo.SaveProofCommitPending(t.Context(), zeroEpoch, zeroHash)
		require.NoError(t, err)
	})
}

func TestBadgerRepository_RemoveProofCommitPending(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := symbiotic.Epoch(200)
	requestID := common.HexToHash("0x987654321fedcba0987654321fedcba0987654321fedcba0987654321fedcba09")

	t.Run("remove non-existent pending proof commit should fail", func(t *testing.T) {
		err := repo.RemoveProofCommitPending(t.Context(), epoch, requestID)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
	})

	t.Run("save and remove pending proof commit", func(t *testing.T) {
		// First save
		err := repo.SaveProofCommitPending(t.Context(), epoch, requestID)
		require.NoError(t, err)

		// Then remove
		err = repo.RemoveProofCommitPending(t.Context(), epoch, requestID)
		require.NoError(t, err)

		// Try to remove again should fail
		err = repo.RemoveProofCommitPending(t.Context(), epoch, requestID)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
	})

	t.Run("remove only affects specific epoch and hash", func(t *testing.T) {
		epoch1 := symbiotic.Epoch(300)
		epoch2 := symbiotic.Epoch(301)
		hash1 := common.HexToHash("0x111111111111111111111111111111111111111111111111111111111111111")
		hash2 := common.HexToHash("0x222222222222222222222222222222222222222222222222222222222222222")

		// Save multiple entries
		err := repo.SaveProofCommitPending(t.Context(), epoch1, hash1)
		require.NoError(t, err)
		err = repo.SaveProofCommitPending(t.Context(), epoch1, hash2)
		require.NoError(t, err)
		err = repo.SaveProofCommitPending(t.Context(), epoch2, hash1)
		require.NoError(t, err)

		// Remove one specific entry
		err = repo.RemoveProofCommitPending(t.Context(), epoch1, hash1)
		require.NoError(t, err)

		// Verify others still exist by trying to save them again (should fail)
		err = repo.SaveProofCommitPending(t.Context(), epoch1, hash2)
		require.ErrorIs(t, err, entity.ErrEntityAlreadyExist)
		err = repo.SaveProofCommitPending(t.Context(), epoch2, hash1)
		require.ErrorIs(t, err, entity.ErrEntityAlreadyExist)

		// But the removed one should be gone - we can save it again
		err = repo.SaveProofCommitPending(t.Context(), epoch1, hash1)
		require.NoError(t, err)
	})
}

func TestBadgerRepository_GetPendingProofCommitsSinceEpoch(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	t.Run("empty repository returns empty list", func(t *testing.T) {
		commits, err := repo.GetPendingProofCommitsSinceEpoch(t.Context(), symbiotic.Epoch(100), 10)
		require.NoError(t, err)
		require.Empty(t, commits)
	})

	t.Run("returns commits for epochs >= since epoch", func(t *testing.T) {
		// Save commits in different epochs
		epoch50 := symbiotic.Epoch(50)
		epoch100 := symbiotic.Epoch(100)
		epoch150 := symbiotic.Epoch(150)

		hash1 := common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111")
		hash2 := common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222")
		hash3 := common.HexToHash("0x3333333333333333333333333333333333333333333333333333333333333333")

		err := repo.SaveProofCommitPending(t.Context(), epoch50, hash1)
		require.NoError(t, err)
		err = repo.SaveProofCommitPending(t.Context(), epoch100, hash2)
		require.NoError(t, err)
		err = repo.SaveProofCommitPending(t.Context(), epoch150, hash3)
		require.NoError(t, err)

		// Get commits since epoch 100
		commits, err := repo.GetPendingProofCommitsSinceEpoch(t.Context(), epoch100, 10)
		require.NoError(t, err)
		require.Len(t, commits, 2)

		// Should return commits for epochs 100 and 150, not 50
		epochsFound := make(map[symbiotic.Epoch]bool)
		hashesFound := make(map[common.Hash]bool)
		for _, commit := range commits {
			epochsFound[commit.Epoch] = true
			hashesFound[commit.RequestID] = true
		}

		require.True(t, epochsFound[epoch100])
		require.True(t, epochsFound[epoch150])
		require.False(t, epochsFound[epoch50])

		require.True(t, hashesFound[hash2])
		require.True(t, hashesFound[hash3])
		require.False(t, hashesFound[hash1])
	})

	t.Run("sorting works correctly - epoch ascending, then hash ascending", func(t *testing.T) {
		testEpoch := symbiotic.Epoch(500)

		// Create commits with same epoch but different hashes to test hash sorting
		hash1 := common.HexToHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		hash2 := common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
		hash3 := common.HexToHash("0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")

		// Save in reverse order to test sorting
		err := repo.SaveProofCommitPending(t.Context(), testEpoch+2, hash3)
		require.NoError(t, err)
		err = repo.SaveProofCommitPending(t.Context(), testEpoch, hash2)
		require.NoError(t, err)
		err = repo.SaveProofCommitPending(t.Context(), testEpoch+1, hash1)
		require.NoError(t, err)
		err = repo.SaveProofCommitPending(t.Context(), testEpoch, hash1)
		require.NoError(t, err)

		commits, err := repo.GetPendingProofCommitsSinceEpoch(t.Context(), testEpoch, 10)
		require.NoError(t, err)
		require.Len(t, commits, 4)

		// Verify epoch ordering (ascending)
		for i := 1; i < len(commits); i++ {
			require.LessOrEqual(t, commits[i-1].Epoch, commits[i].Epoch,
				"Epochs not in ascending order: %d > %d", commits[i-1].Epoch, commits[i].Epoch)
		}

		// Verify hash ordering within same epoch
		sameEpochCommits := make(map[symbiotic.Epoch][]symbiotic.ProofCommitKey)
		for _, commit := range commits {
			sameEpochCommits[commit.Epoch] = append(sameEpochCommits[commit.Epoch], commit)
		}

		for epoch, epochCommits := range sameEpochCommits {
			if len(epochCommits) > 1 {
				for i := 1; i < len(epochCommits); i++ {
					require.LessOrEqual(t,
						epochCommits[i-1].RequestID.Big().Cmp(epochCommits[i].RequestID.Big()), 0,
						"Hashes not in ascending order within epoch %d", epoch)
				}
			}
		}
	})

	t.Run("limit parameter works correctly", func(t *testing.T) {
		testEpoch := symbiotic.Epoch(600)

		// Create multiple commits
		for i := 0; i < 5; i++ {
			hash := common.HexToHash(string(rune('a'+i)) + "000000000000000000000000000000000000000000000000000000000000000")
			err := repo.SaveProofCommitPending(t.Context(), testEpoch, hash)
			require.NoError(t, err)
		}

		// Test limit = 3
		commits, err := repo.GetPendingProofCommitsSinceEpoch(t.Context(), testEpoch, 3)
		require.NoError(t, err)
		require.Len(t, commits, 3)

		// Test limit = 0 (should return all)
		commits, err = repo.GetPendingProofCommitsSinceEpoch(t.Context(), testEpoch, 0)
		require.NoError(t, err)
		require.Len(t, commits, 5)

		// Test limit > available (should return all available)
		commits, err = repo.GetPendingProofCommitsSinceEpoch(t.Context(), testEpoch, 10)
		require.NoError(t, err)
		require.Len(t, commits, 5)
	})

	t.Run("handles multiple epochs correctly", func(t *testing.T) {
		epoch700 := symbiotic.Epoch(700)
		epoch701 := symbiotic.Epoch(701)
		epoch702 := symbiotic.Epoch(702)

		hash1 := common.HexToHash("0x7000000000000000000000000000000000000000000000000000000000000001")
		hash2 := common.HexToHash("0x7000000000000000000000000000000000000000000000000000000000000002")
		hash3 := common.HexToHash("0x7000000000000000000000000000000000000000000000000000000000000003")

		// Save commits in different epochs
		err := repo.SaveProofCommitPending(t.Context(), epoch700, hash1)
		require.NoError(t, err)
		err = repo.SaveProofCommitPending(t.Context(), epoch701, hash2)
		require.NoError(t, err)
		err = repo.SaveProofCommitPending(t.Context(), epoch702, hash3)
		require.NoError(t, err)

		// Query from epoch 701 should return epochs 701 and 702
		commits, err := repo.GetPendingProofCommitsSinceEpoch(t.Context(), epoch701, 10)
		require.NoError(t, err)
		require.Len(t, commits, 2)

		require.Equal(t, epoch701, commits[0].Epoch)
		require.Equal(t, hash2, commits[0].RequestID)
		require.Equal(t, epoch702, commits[1].Epoch)
		require.Equal(t, hash3, commits[1].RequestID)
	})

	t.Run("handles invalid keys gracefully", func(t *testing.T) {
		// This test verifies that the function handles malformed keys in the database
		// We can't directly insert malformed keys through the public API, but the function
		// should be robust against them. This test mainly verifies no panic occurs.

		testEpoch := symbiotic.Epoch(800)
		hash := common.HexToHash("0x8000000000000000000000000000000000000000000000000000000000000001")

		err := repo.SaveProofCommitPending(t.Context(), testEpoch, hash)
		require.NoError(t, err)

		commits, err := repo.GetPendingProofCommitsSinceEpoch(t.Context(), testEpoch, 10)
		require.NoError(t, err)
		require.Len(t, commits, 1)
		require.Equal(t, testEpoch, commits[0].Epoch)
		require.Equal(t, hash, commits[0].RequestID)
	})
}
