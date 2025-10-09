package badger

import (
	"sort"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestBadgerRepository_SignatureRequest(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	req := randomSignatureRequest(t)
	requestId := common.BytesToHash(randomBytes(t, 32))

	err := repo.SaveSignatureRequest(t.Context(), requestId, req)
	require.NoError(t, err)

	loadedConfig, err := repo.GetSignatureRequest(t.Context(), requestId)
	require.NoError(t, err)
	require.Equal(t, req, loadedConfig)
}

type reqWithTargetID struct {
	req  symbiotic.SignatureRequest
	hash common.Hash
}

func TestBadgerRepository_GetSignatureRequestsByEpoch(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := symbiotic.Epoch(100)

	requests := make([]reqWithTargetID, 5)
	for i := 0; i < 5; i++ {
		req := randomSignatureRequestForEpoch(t, epoch)
		req.Message = append([]byte(strconv.Itoa(i)+"-"), req.Message...)
		requests[i].req = req
		requests[i].hash = common.BytesToHash(randomBytes(t, 32))

		err := repo.SaveSignatureRequest(t.Context(), requests[i].hash, req)
		require.NoError(t, err)
	}

	// Sort requests by hash (lexicographic order) to match expected retrieval order
	sort.Slice(requests, func(i, j int) bool {
		return requests[i].hash.Hex() < requests[j].hash.Hex()
	})

	t.Run("get all requests for epoch", func(t *testing.T) {
		results, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 0, common.Hash{})
		require.NoError(t, err)
		require.Len(t, results, 5)

		// Verify they are in correct order
		for i, result := range results {
			require.Equal(t, requests[i].req, result)
		}
	})

	t.Run("get requests with limit", func(t *testing.T) {
		results, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 3, common.Hash{})
		require.NoError(t, err)
		require.Len(t, results, 3)

		// Verify first 3 requests
		for i, result := range results {
			require.Equal(t, requests[i].req, result)
		}
	})

	t.Run("cursor-based pagination", func(t *testing.T) {
		// Get first page (2 items)
		firstPage, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 2, common.Hash{})
		require.NoError(t, err)
		require.Len(t, firstPage, 2)
		require.Equal(t, requests[0].req, firstPage[0])
		require.Equal(t, requests[1].req, firstPage[1])

		// Get second page using cursor
		lastHash := requests[len(firstPage)-1].hash
		secondPage, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 2, lastHash)
		require.NoError(t, err)
		require.Len(t, secondPage, 2)
		require.Equal(t, requests[2].req, secondPage[0])
		require.Equal(t, requests[3].req, secondPage[1])

		// Get third page using cursor
		lastHash = requests[len(firstPage)+len(secondPage)-1].hash
		thirdPage, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 2, lastHash)
		require.NoError(t, err)
		require.Len(t, thirdPage, 1)
		require.Equal(t, requests[4].req, thirdPage[0])

		// Get fourth page (should be empty)
		lastHash = requests[len(firstPage)+len(secondPage)+len(thirdPage)-1].hash
		fourthPage, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 2, lastHash)
		require.NoError(t, err)
		require.Empty(t, fourthPage)
	})

	t.Run("empty epoch", func(t *testing.T) {
		emptyEpoch := symbiotic.Epoch(999)
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
		firstHash := requests[0].hash
		secondHash := requests[1].hash

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

		// Verify the sequence is correct
		for i := 0; i < len(results); i++ {
			expectedIndex := i + 1 // Skip first item, start from second
			require.Equal(t, requests[expectedIndex].req, results[i])
		}
	})
}

func TestBadgerRepository_GetSignatureRequestsByEpoch_MultipleEpochs(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch1 := symbiotic.Epoch(100)
	epoch2 := symbiotic.Epoch(200)

	// Create requests for epoch1
	for i := 0; i < 3; i++ {
		req := randomSignatureRequestForEpoch(t, epoch1)
		err := repo.SaveSignatureRequest(t.Context(), common.BytesToHash(randomBytes(t, 32)), req)
		require.NoError(t, err)
	}

	// Create requests for epoch2
	for i := 0; i < 2; i++ {
		req := randomSignatureRequestForEpoch(t, epoch2)
		err := repo.SaveSignatureRequest(t.Context(), common.BytesToHash(randomBytes(t, 32)), req)
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

func randomSignatureRequest(t *testing.T) symbiotic.SignatureRequest {
	t.Helper()
	return symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(15),
		RequiredEpoch: symbiotic.Epoch(randomBigInt(t).Uint64()),
		Message:       randomBytes(t, 32),
	}
}

func randomSignatureRequestForEpoch(t *testing.T, epoch symbiotic.Epoch) symbiotic.SignatureRequest {
	t.Helper()
	return symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(15),
		RequiredEpoch: epoch,
		Message:       randomBytes(t, 32),
	}
}

func randomSignatureExtendedForEpoch(t *testing.T, epoch symbiotic.Epoch) symbiotic.SignatureExtended {
	t.Helper()
	return symbiotic.SignatureExtended{
		MessageHash: randomBytes(t, 32),
		KeyTag:      15,
		Epoch:       epoch,
		Signature:   []byte("signature1"),
		PublicKey:   []byte("publickey1"),
	}
}
