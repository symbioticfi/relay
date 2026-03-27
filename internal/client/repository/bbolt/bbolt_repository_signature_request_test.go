package bbolt

import (
	"sort"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func signatureRequestID(t *testing.T, req symbiotic.SignatureRequest) common.Hash {
	t.Helper()
	priv, err := crypto.GeneratePrivateKey(req.KeyTag.Type())
	require.NoError(t, err)
	_, messageHash, err := priv.Sign(req.Message)
	require.NoError(t, err)

	sig := symbiotic.Signature{
		KeyTag:      req.KeyTag,
		Epoch:       req.RequiredEpoch,
		MessageHash: messageHash,
		PublicKey:   priv.PublicKey(),
	}
	return sig.RequestID()
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

type reqWithTargetID struct {
	req  symbiotic.SignatureRequest
	hash common.Hash
}

func TestRepository_SignatureRequest(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	req := randomSignatureRequest(t)
	requestId := signatureRequestID(t, req)

	err := repo.SaveSignatureRequest(t.Context(), requestId, req)
	require.NoError(t, err)

	loadedReq, err := repo.GetSignatureRequest(t.Context(), requestId)
	require.NoError(t, err)
	require.Equal(t, req, loadedReq)
}

func TestRepository_GetSignatureRequestsByEpoch(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := symbiotic.Epoch(100)

	requests := make([]reqWithTargetID, 5)
	for i := range 5 {
		req := randomSignatureRequestForEpoch(t, epoch)
		req.Message = append([]byte(strconv.Itoa(i)+"-"), req.Message...)
		requests[i].req = req
		requests[i].hash = signatureRequestID(t, req)

		err := repo.SaveSignatureRequest(t.Context(), requests[i].hash, req)
		require.NoError(t, err)
	}

	sort.Slice(requests, func(i, j int) bool {
		return requests[i].hash.Hex() < requests[j].hash.Hex()
	})

	t.Run("get all requests for epoch", func(t *testing.T) {
		results, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 0, common.Hash{})
		require.NoError(t, err)
		require.Len(t, results, 5)

		for i, result := range results {
			require.Equal(t, requests[i].req, result)
		}
	})

	t.Run("get requests with limit", func(t *testing.T) {
		results, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 3, common.Hash{})
		require.NoError(t, err)
		require.Len(t, results, 3)
	})

	t.Run("cursor-based pagination", func(t *testing.T) {
		firstPage, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 2, common.Hash{})
		require.NoError(t, err)
		require.Len(t, firstPage, 2)

		lastHash := requests[1].hash
		secondPage, err := repo.GetSignatureRequestsByEpoch(t.Context(), epoch, 2, lastHash)
		require.NoError(t, err)
		require.Len(t, secondPage, 2)
		require.Equal(t, requests[2].req, secondPage[0])
		require.Equal(t, requests[3].req, secondPage[1])
	})

	t.Run("empty epoch", func(t *testing.T) {
		results, err := repo.GetSignatureRequestsByEpoch(t.Context(), symbiotic.Epoch(999), 0, common.Hash{})
		require.NoError(t, err)
		require.Empty(t, results)
	})
}

func TestRepository_GetSignatureRequestIDsByEpoch(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := symbiotic.Epoch(100)

	expectedIDs := make([]common.Hash, 5)
	for i := range 5 {
		req := randomSignatureRequestForEpoch(t, epoch)
		req.Message = append([]byte(strconv.Itoa(i)+"-"), req.Message...)
		requestID := signatureRequestID(t, req)
		expectedIDs[i] = requestID

		err := repo.SaveSignatureRequest(t.Context(), requestID, req)
		require.NoError(t, err)
	}

	sort.Slice(expectedIDs, func(i, j int) bool {
		return expectedIDs[i].Hex() < expectedIDs[j].Hex()
	})

	t.Run("get all request IDs for epoch", func(t *testing.T) {
		results, err := repo.GetSignatureRequestIDsByEpoch(t.Context(), epoch)
		require.NoError(t, err)
		require.Len(t, results, 5)

		for i, result := range results {
			require.Equal(t, expectedIDs[i], result)
		}
	})

	t.Run("empty epoch returns no IDs", func(t *testing.T) {
		results, err := repo.GetSignatureRequestIDsByEpoch(t.Context(), symbiotic.Epoch(999))
		require.NoError(t, err)
		require.Empty(t, results)
	})
}

func TestRepository_GetSignatureRequestsWithIDByEpoch(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	epoch := symbiotic.Epoch(100)

	requests := make([]reqWithTargetID, 5)
	for i := range 5 {
		req := randomSignatureRequestForEpoch(t, epoch)
		req.Message = append([]byte(strconv.Itoa(i)+"-"), req.Message...)
		requests[i].req = req
		requests[i].hash = signatureRequestID(t, req)

		err := repo.SaveSignatureRequest(t.Context(), requests[i].hash, req)
		require.NoError(t, err)
	}

	sort.Slice(requests, func(i, j int) bool {
		return requests[i].hash.Hex() < requests[j].hash.Hex()
	})

	t.Run("get all requests with IDs for epoch", func(t *testing.T) {
		results, err := repo.GetSignatureRequestsWithIDByEpoch(t.Context(), epoch)
		require.NoError(t, err)
		require.Len(t, results, 5)

		for i, result := range results {
			require.Equal(t, requests[i].hash, result.RequestID)
			require.Equal(t, requests[i].req, result.SignatureRequest)
		}
	})

	t.Run("empty epoch returns empty list", func(t *testing.T) {
		results, err := repo.GetSignatureRequestsWithIDByEpoch(t.Context(), symbiotic.Epoch(999))
		require.NoError(t, err)
		require.Empty(t, results)
	})
}
