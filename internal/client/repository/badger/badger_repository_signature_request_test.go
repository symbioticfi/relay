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

	t.Run("cursor hash between stored keys - no off-by-one", func(t *testing.T) {
		// This test validates that when a cursor hash falls between stored keys,
		// we don't skip the first valid item after the seek (off-by-one bug)
		// Get the first two requests to determine a cursor that falls between them
		firstTwo, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 2, common.Hash{})
		require.NoError(t, err)
		require.Len(t, firstTwo, 2)

		// Create a cursor hash that falls lexicographically between the first two stored hashes
		firstHash := firstTwo[0].Hash()
		secondHash := firstTwo[1].Hash()

		// Create a hash that's lexicographically between first and second
		var betweenHash common.Hash
		copy(betweenHash[:], firstHash[:])
		// Increment the last byte to create a hash between first and second
		if betweenHash[31] < 255 {
			betweenHash[31]++
		}

		// Verify the hash is actually between the two stored hashes
		require.Negative(t, firstHash.Cmp(betweenHash), "betweenHash should be greater than firstHash")
		require.Negative(t, betweenHash.Cmp(secondHash), "betweenHash should be less than secondHash")

		// Query with this between-hash cursor - should start from secondHash (not skip it)
		results, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 0, betweenHash)
		require.NoError(t, err)

		// Should return all items starting from the second item (no off-by-one skip)
		require.Len(t, results, 4) // Should have 4 remaining items (total 5 - first 1)
		require.Equal(t, secondHash, results[0].Hash(), "should start from second item, not skip it")

		// Verify the sequence is correct
		for i := 0; i < len(results); i++ {
			expectedIndex := i + 1 // Skip first item, start from second
			require.Equal(t, requests[expectedIndex], results[i])
		}
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

func TestBadgerRepository_GetSignatureRequestsByEpochPending(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := entity.Epoch(100)

	// Create multiple pending signature requests for the same epoch
	requests := make([]entity.SignatureRequest, 5)
	for i := 0; i < 5; i++ {
		req := randomSignatureRequestForEpoch(t, epoch)
		requests[i] = req
		err := repo.SaveSignatureRequest(t.Context(), req)
		require.NoError(t, err)
		err = repo.SaveSignatureRequestPending(t.Context(), req)
		require.NoError(t, err)
	}

	// Sort requests by hash (lexicographic order) to match expected retrieval order
	sort.Slice(requests, func(i, j int) bool {
		return requests[i].Hash().Hex() < requests[j].Hash().Hex()
	})

	t.Run("get all pending requests for epoch", func(t *testing.T) {
		results, err := repo.GetSignatureRequestsByEpochPending(t.Context(), epoch, 0, common.Hash{})
		require.NoError(t, err)
		require.Len(t, results, 5)

		// Verify they are in correct order
		for i, result := range results {
			require.Equal(t, requests[i], result)
		}
	})

	t.Run("get pending requests with limit", func(t *testing.T) {
		results, err := repo.GetSignatureRequestsByEpochPending(t.Context(), epoch, 3, common.Hash{})
		require.NoError(t, err)
		require.Len(t, results, 3)

		// Verify first 3 requests
		for i, result := range results {
			require.Equal(t, requests[i], result)
		}
	})

	t.Run("cursor-based pagination for pending requests", func(t *testing.T) {
		// Get first page (2 items)
		firstPage, err := repo.GetSignatureRequestsByEpochPending(t.Context(), epoch, 2, common.Hash{})
		require.NoError(t, err)
		require.Len(t, firstPage, 2)
		require.Equal(t, requests[0], firstPage[0])
		require.Equal(t, requests[1], firstPage[1])

		// Get second page using cursor
		lastHash := firstPage[len(firstPage)-1].Hash()
		secondPage, err := repo.GetSignatureRequestsByEpochPending(t.Context(), epoch, 2, lastHash)
		require.NoError(t, err)
		require.Len(t, secondPage, 2)
		require.Equal(t, requests[2], secondPage[0])
		require.Equal(t, requests[3], secondPage[1])

		// Get third page using cursor
		lastHash = secondPage[len(secondPage)-1].Hash()
		thirdPage, err := repo.GetSignatureRequestsByEpochPending(t.Context(), epoch, 2, lastHash)
		require.NoError(t, err)
		require.Len(t, thirdPage, 1)
		require.Equal(t, requests[4], thirdPage[0])

		// Get fourth page (should be empty)
		lastHash = thirdPage[len(thirdPage)-1].Hash()
		fourthPage, err := repo.GetSignatureRequestsByEpochPending(t.Context(), epoch, 2, lastHash)
		require.NoError(t, err)
		require.Empty(t, fourthPage)
	})

	t.Run("empty epoch for pending requests", func(t *testing.T) {
		emptyEpoch := entity.Epoch(999)
		results, err := repo.GetSignatureRequestsByEpochPending(t.Context(), emptyEpoch, 0, common.Hash{})
		require.NoError(t, err)
		require.Empty(t, results)
	})
}

func TestBadgerRepository_RemoveSignatureRequestPending(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := entity.Epoch(100)
	req := randomSignatureRequestForEpoch(t, epoch)

	t.Run("successfully removes existing pending request", func(t *testing.T) {
		// Save signature request (creates both main and pending entries)
		err := repo.SaveSignatureRequest(t.Context(), req)
		require.NoError(t, err)
		err = repo.SaveSignatureRequestPending(t.Context(), req)
		require.NoError(t, err)

		// Verify pending request exists
		pendingReqs, err := repo.GetSignatureRequestsByEpochPending(t.Context(), epoch, 0, common.Hash{})
		require.NoError(t, err)
		require.Len(t, pendingReqs, 1)
		require.Equal(t, req, pendingReqs[0])

		// Remove pending request
		err = repo.RemoveSignatureRequestPending(t.Context(), epoch, req.Hash())
		require.NoError(t, err)

		// Verify pending request is removed
		pendingReqs, err = repo.GetSignatureRequestsByEpochPending(t.Context(), epoch, 0, common.Hash{})
		require.NoError(t, err)
		require.Empty(t, pendingReqs)

		// Verify main signature request still exists
		retrievedReq, err := repo.GetSignatureRequest(t.Context(), req.Hash())
		require.NoError(t, err)
		require.Equal(t, req, retrievedReq)
	})

	t.Run("returns error when removing non-existent pending request", func(t *testing.T) {
		nonExistentHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

		err := repo.RemoveSignatureRequestPending(t.Context(), epoch, nonExistentHash)
		require.Error(t, err)
		require.Contains(t, err.Error(), "pending signature request not found")
		require.Contains(t, err.Error(), entity.ErrEntityNotFound)
	})

	t.Run("returns error when removing already removed pending request", func(t *testing.T) {
		// Save and then remove a pending request
		req2 := randomSignatureRequestForEpoch(t, epoch)
		err := repo.SaveSignatureRequest(t.Context(), req2)
		require.NoError(t, err)
		err = repo.SaveSignatureRequestPending(t.Context(), req2)
		require.NoError(t, err)

		err = repo.RemoveSignatureRequestPending(t.Context(), epoch, req2.Hash())
		require.NoError(t, err)

		// Try to remove again
		err = repo.RemoveSignatureRequestPending(t.Context(), epoch, req2.Hash())
		require.Error(t, err)
		require.Contains(t, err.Error(), "pending signature request not found")
		require.Contains(t, err.Error(), entity.ErrEntityNotFound)
	})
}

func TestBadgerRepository_PendingSignatureRequests_Integration(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := entity.Epoch(100)

	t.Run("pending and regular requests are independent", func(t *testing.T) {
		// Create multiple requests
		req1 := randomSignatureRequestForEpoch(t, epoch)
		req2 := randomSignatureRequestForEpoch(t, epoch)
		req3 := randomSignatureRequestForEpoch(t, epoch)

		// Save all requests
		err := repo.SaveSignatureRequest(t.Context(), req1)
		require.NoError(t, err)
		err = repo.SaveSignatureRequestPending(t.Context(), req1)
		require.NoError(t, err)
		err = repo.SaveSignatureRequest(t.Context(), req2)
		require.NoError(t, err)
		err = repo.SaveSignatureRequestPending(t.Context(), req2)
		require.NoError(t, err)
		err = repo.SaveSignatureRequest(t.Context(), req3)
		require.NoError(t, err)
		err = repo.SaveSignatureRequestPending(t.Context(), req3)
		require.NoError(t, err)

		// Verify all are pending
		pendingReqs, err := repo.GetSignatureRequestsByEpochPending(t.Context(), epoch, 0, common.Hash{})
		require.NoError(t, err)
		require.Len(t, pendingReqs, 3)

		// Verify all are in regular collection
		allReqs, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 0, common.Hash{})
		require.NoError(t, err)
		require.Len(t, allReqs, 3)

		// Remove one from pending
		err = repo.RemoveSignatureRequestPending(t.Context(), epoch, req2.Hash())
		require.NoError(t, err)

		// Verify pending count decreased
		pendingReqs, err = repo.GetSignatureRequestsByEpochPending(t.Context(), epoch, 0, common.Hash{})
		require.NoError(t, err)
		require.Len(t, pendingReqs, 2)

		// Verify regular collection unchanged
		allReqs, err = repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 0, common.Hash{})
		require.NoError(t, err)
		require.Len(t, allReqs, 3)

		// Verify the removed request is not in pending
		for _, pendingReq := range pendingReqs {
			require.NotEqual(t, req2.Hash(), pendingReq.Hash())
		}

		// Verify the removed request still accessible via regular method
		retrievedReq, err := repo.GetSignatureRequest(t.Context(), req2.Hash())
		require.NoError(t, err)
		require.Equal(t, req2, retrievedReq)
	})

	t.Run("pending requests across different epochs", func(t *testing.T) {
		epoch1 := entity.Epoch(200)
		epoch2 := entity.Epoch(300)

		req1 := randomSignatureRequestForEpoch(t, epoch1)
		req2 := randomSignatureRequestForEpoch(t, epoch2)

		// Save requests for different epochs
		err := repo.SaveSignatureRequest(t.Context(), req1)
		require.NoError(t, err)
		err = repo.SaveSignatureRequestPending(t.Context(), req1)
		require.NoError(t, err)
		err = repo.SaveSignatureRequest(t.Context(), req2)
		require.NoError(t, err)
		err = repo.SaveSignatureRequestPending(t.Context(), req2)
		require.NoError(t, err)

		// Verify each epoch has its pending request
		pendingReqs1, err := repo.GetSignatureRequestsByEpochPending(t.Context(), epoch1, 0, common.Hash{})
		require.NoError(t, err)
		require.Len(t, pendingReqs1, 1)
		require.Equal(t, req1, pendingReqs1[0])

		pendingReqs2, err := repo.GetSignatureRequestsByEpochPending(t.Context(), epoch2, 0, common.Hash{})
		require.NoError(t, err)
		require.Len(t, pendingReqs2, 1)
		require.Equal(t, req2, pendingReqs2[0])

		// Remove pending from epoch1
		err = repo.RemoveSignatureRequestPending(t.Context(), epoch1, req1.Hash())
		require.NoError(t, err)

		// Verify epoch1 has no pending, epoch2 still has pending
		pendingReqs1, err = repo.GetSignatureRequestsByEpochPending(t.Context(), epoch1, 0, common.Hash{})
		require.NoError(t, err)
		require.Empty(t, pendingReqs1)

		pendingReqs2, err = repo.GetSignatureRequestsByEpochPending(t.Context(), epoch2, 0, common.Hash{})
		require.NoError(t, err)
		require.Len(t, pendingReqs2, 1)
		require.Equal(t, req2, pendingReqs2[0])
	})
}

func TestBadgerRepository_SaveSignatureRequestPending_DuplicateHandling(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := entity.Epoch(100)
	req := randomSignatureRequestForEpoch(t, epoch)

	// Save signature request to main collection first
	err := repo.SaveSignatureRequest(t.Context(), req)
	require.NoError(t, err)

	// Save to pending collection - should succeed
	err = repo.SaveSignatureRequestPending(t.Context(), req)
	require.NoError(t, err)

	// Try to save the same request to pending again - should fail
	err = repo.SaveSignatureRequestPending(t.Context(), req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "pending signature request already exists")
	require.ErrorIs(t, err, entity.ErrEntityAlreadyExist)

	// Verify the pending request still exists and is functional
	pendingReqs, err := repo.GetSignatureRequestsByEpochPending(t.Context(), epoch, 0, common.Hash{})
	require.NoError(t, err)
	require.Len(t, pendingReqs, 1)
	require.Equal(t, req, pendingReqs[0])
}
