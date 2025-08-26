package badger

import (
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/core/entity"
)

func TestBadgerRepository_SignatureRequest(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	req := randomSignatureRequest(t)

	err := repo.SaveSignatureRequest(t.Context(), req)
	require.NoError(t, err)

	loadedConfig, err := repo.GetSignatureRequest(t.Context(), req.Hash())
	require.NoError(t, err)
	require.Equal(t, req, loadedConfig)
}

func TestBadgerRepository_GetSignatureRequestsByEpoch(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := entity.Epoch(100)

	// Create multiple signature requests for the same epoch
	requests := make([]entity.SignatureRequest, 5)
	for i := 0; i < 5; i++ {
		req := randomSignatureRequestForEpoch(t, epoch)
		requests[i] = req
		err := repo.SaveSignatureRequest(t.Context(), req)
		require.NoError(t, err)
	}

	// Sort requests by hash (lexicographic order) to match expected retrieval order
	sort.Slice(requests, func(i, j int) bool {
		return requests[i].Hash().Hex() < requests[j].Hash().Hex()
	})

	t.Run("get all requests for epoch", func(t *testing.T) {
		results, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 0, common.Hash{})
		require.NoError(t, err)
		require.Len(t, results, 5)

		// Verify they are in correct order
		for i, result := range results {
			require.Equal(t, requests[i], result)
		}
	})

	t.Run("get requests with limit", func(t *testing.T) {
		results, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 3, common.Hash{})
		require.NoError(t, err)
		require.Len(t, results, 3)

		// Verify first 3 requests
		for i, result := range results {
			require.Equal(t, requests[i], result)
		}
	})

	t.Run("cursor-based pagination", func(t *testing.T) {
		// Get first page (2 items)
		firstPage, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 2, common.Hash{})
		require.NoError(t, err)
		require.Len(t, firstPage, 2)
		require.Equal(t, requests[0], firstPage[0])
		require.Equal(t, requests[1], firstPage[1])

		// Get second page using cursor
		lastHash := firstPage[len(firstPage)-1].Hash()
		secondPage, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 2, lastHash)
		require.NoError(t, err)
		require.Len(t, secondPage, 2)
		require.Equal(t, requests[2], secondPage[0])
		require.Equal(t, requests[3], secondPage[1])

		// Get third page using cursor
		lastHash = secondPage[len(secondPage)-1].Hash()
		thirdPage, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 2, lastHash)
		require.NoError(t, err)
		require.Len(t, thirdPage, 1)
		require.Equal(t, requests[4], thirdPage[0])

		// Get fourth page (should be empty)
		lastHash = thirdPage[len(thirdPage)-1].Hash()
		fourthPage, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 2, lastHash)
		require.NoError(t, err)
		require.Empty(t, fourthPage)
	})

	t.Run("empty epoch", func(t *testing.T) {
		emptyEpoch := entity.Epoch(999)
		results, err := repo.GetSignatureRequestsByEpoch(t.Context(), emptyEpoch, 0, common.Hash{})
		require.NoError(t, err)
		require.Empty(t, results)
	})

	t.Run("invalid cursor hash", func(t *testing.T) {
		// Use a hash that doesn't exist in this epoch
		nonExistentHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		results, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 0, nonExistentHash)
		require.NoError(t, err)
		// Should return results after the cursor position in lexicographic order
		require.LessOrEqual(t, len(results), 5)
	})
}

func TestBadgerRepository_GetSignatureRequestsByEpoch_MultipleEpochs(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch1 := entity.Epoch(100)
	epoch2 := entity.Epoch(200)

	// Create requests for epoch1
	for i := 0; i < 3; i++ {
		req := randomSignatureRequestForEpoch(t, epoch1)
		err := repo.SaveSignatureRequest(t.Context(), req)
		require.NoError(t, err)
	}

	// Create requests for epoch2
	for i := 0; i < 2; i++ {
		req := randomSignatureRequestForEpoch(t, epoch2)
		err := repo.SaveSignatureRequest(t.Context(), req)
		require.NoError(t, err)
	}

	// Query epoch1 should return only epoch1 requests
	epoch1Results, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch1, 0, common.Hash{})
	require.NoError(t, err)
	require.Len(t, epoch1Results, 3)
	for _, result := range epoch1Results {
		require.Equal(t, epoch1, result.RequiredEpoch)
	}

	// Query epoch2 should return only epoch2 requests
	epoch2Results, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch2, 0, common.Hash{})
	require.NoError(t, err)
	require.Len(t, epoch2Results, 2)
	for _, result := range epoch2Results {
		require.Equal(t, epoch2, result.RequiredEpoch)
	}
}

func randomSignatureRequest(t *testing.T) entity.SignatureRequest {
	t.Helper()
	return entity.SignatureRequest{
		KeyTag:        entity.KeyTag(15),
		RequiredEpoch: entity.Epoch(randomBigInt(t).Uint64()),
		Message:       randomBytes(t, 32),
	}
}

func randomSignatureRequestForEpoch(t *testing.T, epoch entity.Epoch) entity.SignatureRequest {
	t.Helper()
	return entity.SignatureRequest{
		KeyTag:        entity.KeyTag(15),
		RequiredEpoch: epoch,
		Message:       randomBytes(t, 32),
	}
}
