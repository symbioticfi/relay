package badger

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestBadgerRepository_GetSignatureRequestsWithoutAggregationProof(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := symbiotic.Epoch(400)

	t.Run("empty epoch returns empty list", func(t *testing.T) {
		requests, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch, 10, common.Hash{})
		require.NoError(t, err)
		require.Empty(t, requests)
	})

	t.Run("returns requests without proofs", func(t *testing.T) {
		testEpoch := symbiotic.Epoch(410)
		sigReq := randomSignatureRequestForEpoch(t, testEpoch)
		sig := randomSignatureExtendedForEpoch(t, testEpoch)

		// SaveSignatureRequest creates both signature request and pending marker
		err := repo.SaveSignatureRequest(t.Context(), sig.RequestID(), sigReq)
		require.NoError(t, err)

		requests, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), testEpoch, 10, common.Hash{})
		require.NoError(t, err)
		require.Len(t, requests, 1)
		require.Equal(t, sig.RequestID(), requests[0].RequestID)
	})

	t.Run("skips requests with existing proofs", func(t *testing.T) {
		testEpoch := symbiotic.Epoch(420)

		sig1 := randomSignatureExtendedForEpoch(t, testEpoch)
		sig2 := randomSignatureExtendedForEpoch(t, testEpoch)

		err := repo.SaveSignatureRequest(t.Context(), sig1.RequestID(), randomSignatureRequestForEpoch(t, testEpoch))
		require.NoError(t, err)
		err = repo.SaveSignatureRequest(t.Context(), sig2.RequestID(), randomSignatureRequestForEpoch(t, testEpoch))
		require.NoError(t, err)

		// SaveProof removes pending marker
		proof := symbiotic.AggregationProof{
			MessageHash: sig1.MessageHash,
			KeyTag:      sig1.KeyTag,
			Epoch:       testEpoch,
			Proof:       randomBytes(t, 32),
		}
		err = repo.SaveProof(t.Context(), proof)
		require.NoError(t, err)

		// Only sig2 should appear (sig1's pending was removed by SaveProof)
		requests, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), testEpoch, 10, common.Hash{})
		require.NoError(t, err)
		require.Len(t, requests, 1)
		require.Equal(t, sig2.RequestID(), requests[0].RequestID)
	})

	t.Run("pagination works correctly", func(t *testing.T) {
		testEpoch := symbiotic.Epoch(500)

		var sigs []symbiotic.Signature
		for range 5 {
			sigReq := randomSignatureRequestForEpoch(t, testEpoch)
			sig := randomSignatureExtendedForEpoch(t, testEpoch)
			sigs = append(sigs, sig)

			err := repo.SaveSignatureRequest(t.Context(), sig.RequestID(), sigReq)
			require.NoError(t, err)
		}

		firstPage, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), testEpoch, 3, common.Hash{})
		require.NoError(t, err)
		require.Len(t, firstPage, 3)

		lastHash := firstPage[len(firstPage)-1].RequestID
		secondPage, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), testEpoch, 3, lastHash)
		require.NoError(t, err)
		require.Len(t, secondPage, 2)

		// No overlap
		firstPageHashes := make(map[common.Hash]bool)
		for _, req := range firstPage {
			firstPageHashes[req.RequestID] = true
		}
		for _, req := range secondPage {
			require.False(t, firstPageHashes[req.RequestID])
		}

		// All found
		allFound := make(map[common.Hash]bool)
		for _, req := range append(firstPage, secondPage...) {
			allFound[req.RequestID] = true
		}
		for _, sig := range sigs {
			require.True(t, allFound[sig.RequestID()])
		}
	})

	t.Run("handles multiple epochs correctly", func(t *testing.T) {
		epoch1 := symbiotic.Epoch(700)
		epoch2 := symbiotic.Epoch(701)

		sig1 := randomSignatureExtendedForEpoch(t, epoch1)
		err := repo.SaveSignatureRequest(t.Context(), sig1.RequestID(), randomSignatureRequestForEpoch(t, epoch1))
		require.NoError(t, err)

		sig2 := randomSignatureExtendedForEpoch(t, epoch2)
		err = repo.SaveSignatureRequest(t.Context(), sig2.RequestID(), randomSignatureRequestForEpoch(t, epoch2))
		require.NoError(t, err)

		requests1, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch1, 10, common.Hash{})
		require.NoError(t, err)
		require.Len(t, requests1, 1)
		require.Equal(t, sig1.RequestID(), requests1[0].RequestID)

		requests2, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), epoch2, 10, common.Hash{})
		require.NoError(t, err)
		require.Len(t, requests2, 1)
		require.Equal(t, sig2.RequestID(), requests2[0].RequestID)
	})

	t.Run("RemoveAggregationProofPending removes from results", func(t *testing.T) {
		testEpoch := symbiotic.Epoch(800)
		sig := randomSignatureExtendedForEpoch(t, testEpoch)

		err := repo.SaveSignatureRequest(t.Context(), sig.RequestID(), randomSignatureRequestForEpoch(t, testEpoch))
		require.NoError(t, err)

		// Should appear
		requests, err := repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), testEpoch, 10, common.Hash{})
		require.NoError(t, err)
		require.Len(t, requests, 1)

		// Remove pending directly
		err = repo.RemoveAggregationProofPending(t.Context(), testEpoch, sig.RequestID())
		require.NoError(t, err)

		// Should no longer appear
		requests, err = repo.GetSignatureRequestsWithoutAggregationProof(t.Context(), testEpoch, 10, common.Hash{})
		require.NoError(t, err)
		require.Empty(t, requests)
	})
}
