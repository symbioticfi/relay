package badger

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/core/entity"
)

func TestBadgerRepository_SaveAggregationProofPending(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := entity.Epoch(100)
	requestID := common.HexToHash("0x123456789abcdef")

	t.Run("save new pending aggregation proof", func(t *testing.T) {
		err := repo.SaveAggregationProofPending(t.Context(), requestID, epoch)
		require.NoError(t, err)
	})

	t.Run("save duplicate pending aggregation proof should fail", func(t *testing.T) {
		err := repo.SaveAggregationProofPending(t.Context(), requestID, epoch)
		require.ErrorIs(t, err, entity.ErrEntityAlreadyExist)
	})

	t.Run("save different epoch should succeed", func(t *testing.T) {
		differentEpoch := entity.Epoch(101)
		err := repo.SaveAggregationProofPending(t.Context(), requestID, differentEpoch)
		require.NoError(t, err)
	})

	t.Run("save different hash should succeed", func(t *testing.T) {
		differentHash := common.HexToHash("0xabcdef123456789")
		err := repo.SaveAggregationProofPending(t.Context(), differentHash, epoch)
		require.NoError(t, err)
	})
}

func TestBadgerRepository_RemoveAggregationProofPending(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := entity.Epoch(200)
	requestID := common.HexToHash("0x987654321fedcba")

	t.Run("remove non-existent pending aggregation proof should fail", func(t *testing.T) {
		err := repo.RemoveAggregationProofPending(t.Context(), epoch, requestID)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
	})

	t.Run("save and remove pending aggregation proof", func(t *testing.T) {
		// First save
		err := repo.SaveAggregationProofPending(t.Context(), requestID, epoch)
		require.NoError(t, err)

		// Then remove
		err = repo.RemoveAggregationProofPending(t.Context(), epoch, requestID)
		require.NoError(t, err)

		// Try to remove again should fail
		err = repo.RemoveAggregationProofPending(t.Context(), epoch, requestID)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
	})

	t.Run("remove only affects specific epoch and hash", func(t *testing.T) {
		epoch1 := entity.Epoch(300)
		epoch2 := entity.Epoch(301)
		hash1 := common.HexToHash("0x111111")
		hash2 := common.HexToHash("0x222222")

		// Save multiple entries
		err := repo.SaveAggregationProofPending(t.Context(), hash1, epoch1)
		require.NoError(t, err)
		err = repo.SaveAggregationProofPending(t.Context(), hash2, epoch1)
		require.NoError(t, err)
		err = repo.SaveAggregationProofPending(t.Context(), hash1, epoch2)
		require.NoError(t, err)

		// Remove one specific entry
		err = repo.RemoveAggregationProofPending(t.Context(), epoch1, hash1)
		require.NoError(t, err)

		// Verify others still exist by trying to save them again (should fail)
		err = repo.SaveAggregationProofPending(t.Context(), hash2, epoch1)
		require.ErrorIs(t, err, entity.ErrEntityAlreadyExist)
		err = repo.SaveAggregationProofPending(t.Context(), hash1, epoch2)
		require.ErrorIs(t, err, entity.ErrEntityAlreadyExist)

		// But the removed one should be gone - we can save it again
		err = repo.SaveAggregationProofPending(t.Context(), hash1, epoch1)
		require.NoError(t, err)
	})
}

func TestBadgerRepository_GetSignatureRequestsWithoutAggregationProof(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := entity.Epoch(400)

	t.Run("empty epoch returns empty list", func(t *testing.T) {
		requests, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch, 10, common.Hash{})
		require.NoError(t, err)
		require.Empty(t, requests)
	})

	t.Run("returns signature requests for pending aggregation proofs", func(t *testing.T) {
		// Create signature requests
		req1 := randomSignatureRequestForEpoch(t, epoch)
		sigID1 := randomRequestID(t)
		req2 := randomSignatureRequestForEpoch(t, epoch)
		sigID2 := randomRequestID(t)
		req3 := randomSignatureRequestForEpoch(t, epoch)
		sigID3 := randomRequestID(t)

		// Save signature requests
		err := repo.SaveSignatureRequest(t.Context(), sigID1, req1)
		require.NoError(t, err)
		err = repo.SaveSignatureRequest(t.Context(), sigID2, req2)
		require.NoError(t, err)
		err = repo.SaveSignatureRequest(t.Context(), sigID3, req3)
		require.NoError(t, err)

		// Mark only req1 and req3 as pending aggregation proof
		err = repo.SaveAggregationProofPending(t.Context(), sigID1, epoch)
		require.NoError(t, err)
		err = repo.SaveAggregationProofPending(t.Context(), sigID3, epoch)
		require.NoError(t, err)

		// Get requests without aggregation proof
		requests, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch, 10, common.Hash{})
		require.NoError(t, err)
		require.Len(t, requests, 2)

		// Verify we got req1 and req3 (order might vary due to map iteration)
		hashes := []common.Hash{requests[0].RequestID, requests[1].RequestID}
		require.Contains(t, hashes, sigID1)
		require.Contains(t, hashes, sigID3)
		require.NotContains(t, hashes, sigID2)
	})

	t.Run("pagination works correctly", func(t *testing.T) {
		testEpoch := entity.Epoch(500)

		// Create multiple signature requests
		var requests []entity.SignatureRequestWithID
		for i := 0; i < 5; i++ {
			req := randomSignatureRequestForEpoch(t, testEpoch)
			sigID := randomRequestID(t)
			requests = append(requests, entity.SignatureRequestWithID{
				SignatureRequest: req,
				RequestID:        sigID,
			})

			// Save signature request
			err := repo.SaveSignatureRequest(t.Context(), sigID, req)
			require.NoError(t, err)

			// Mark as pending aggregation proof
			err = repo.SaveAggregationProofPending(t.Context(), sigID, testEpoch)
			require.NoError(t, err)
		}

		// Get first page (limit 3)
		firstPage, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), testEpoch, 3, common.Hash{})
		require.NoError(t, err)
		require.Len(t, firstPage, 3)

		// Get second page using last hash from first page
		lastHash := firstPage[len(firstPage)-1].RequestID
		secondPage, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), testEpoch, 3, lastHash)
		require.NoError(t, err)
		require.Len(t, secondPage, 2) // Remaining 2 requests

		// Verify no overlap between pages
		firstPageHashes := make(map[common.Hash]bool)
		for _, req := range firstPage {
			firstPageHashes[req.RequestID] = true
		}

		for _, req := range secondPage {
			require.False(t, firstPageHashes[req.RequestID], "Found duplicate request id between pages")
		}

		// Verify all original requests are found across both pages
		allFoundHashes := make(map[common.Hash]bool)
		for _, req := range append(firstPage, secondPage...) {
			allFoundHashes[req.RequestID] = true
		}

		for _, originalReq := range requests {
			require.True(t, allFoundHashes[originalReq.RequestID], "Original request not found in paginated results")
		}
	})

	t.Run("skips entries with missing signature requests", func(t *testing.T) {
		testEpoch := entity.Epoch(600)

		// Create one valid signature request
		req := randomSignatureRequestForEpoch(t, testEpoch)
		sigID := randomRequestID(t)
		err := repo.SaveSignatureRequest(t.Context(), sigID, req)
		require.NoError(t, err)
		err = repo.SaveAggregationProofPending(t.Context(), sigID, testEpoch)
		require.NoError(t, err)

		// Create a pending aggregation proof marker without corresponding signature request
		orphanHash := common.HexToHash("0xorphan")
		err = repo.SaveAggregationProofPending(t.Context(), orphanHash, testEpoch)
		require.NoError(t, err)

		// Should only return the valid request, skipping the orphan
		requests, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), testEpoch, 10, common.Hash{})
		require.NoError(t, err)
		require.Len(t, requests, 1)
		require.Equal(t, sigID, requests[0].RequestID)
	})

	t.Run("handles multiple epochs correctly", func(t *testing.T) {
		epoch1 := entity.Epoch(700)
		epoch2 := entity.Epoch(701)

		// Create requests in different epochs
		req1 := randomSignatureRequestForEpoch(t, epoch1)
		sigID1 := randomRequestID(t)
		req2 := randomSignatureRequestForEpoch(t, epoch2)
		sigID2 := randomRequestID(t)

		err := repo.SaveSignatureRequest(t.Context(), sigID1, req1)
		require.NoError(t, err)
		err = repo.SaveSignatureRequest(t.Context(), sigID2, req2)
		require.NoError(t, err)

		err = repo.SaveAggregationProofPending(t.Context(), sigID1, epoch1)
		require.NoError(t, err)
		err = repo.SaveAggregationProofPending(t.Context(), sigID2, epoch2)
		require.NoError(t, err)

		// Query epoch1 should only return req1
		requests1, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch1, 10, common.Hash{})
		require.NoError(t, err)
		require.Len(t, requests1, 1)
		require.Equal(t, sigID1, requests1[0].RequestID)

		// Query epoch2 should only return req2
		requests2, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch2, 10, common.Hash{})
		require.NoError(t, err)
		require.Len(t, requests2, 1)
		require.Equal(t, sigID2, requests2[0].RequestID)
	})
}

func TestBadgerRepository_AggregationProofPendingIntegration(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := entity.Epoch(800)
	req := randomSignatureRequestForEpoch(t, epoch)
	sigID := randomRequestID(t)

	// Save signature request
	err := repo.SaveSignatureRequest(t.Context(), sigID, req)
	require.NoError(t, err)

	// Initially, no pending aggregation proofs
	requests, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch, 10, common.Hash{})
	require.NoError(t, err)
	require.Empty(t, requests)

	// Mark as pending aggregation proof
	err = repo.SaveAggregationProofPending(t.Context(), sigID, epoch)
	require.NoError(t, err)

	// Now it should appear in pending list
	requests, err = repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch, 10, common.Hash{})
	require.NoError(t, err)
	require.Len(t, requests, 1)
	require.Equal(t, sigID, requests[0].RequestID)

	// Remove from pending
	err = repo.RemoveAggregationProofPending(t.Context(), epoch, sigID)
	require.NoError(t, err)

	// Should no longer appear in pending list
	requests, err = repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch, 10, common.Hash{})
	require.NoError(t, err)
	require.Empty(t, requests)
}
