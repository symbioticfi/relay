package entity

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestWantAggregationProofsRequest(t *testing.T) {
	t.Run("creates empty request", func(t *testing.T) {
		req := WantAggregationProofsRequest{}
		require.Empty(t, req.RequestHashes)
	})

	t.Run("creates request with hashes", func(t *testing.T) {
		hash1 := common.HexToHash("0x123456")
		hash2 := common.HexToHash("0x789abc")

		req := WantAggregationProofsRequest{
			RequestHashes: []common.Hash{hash1, hash2},
		}
		require.Len(t, req.RequestHashes, 2)
		require.Contains(t, req.RequestHashes, hash1)
		require.Contains(t, req.RequestHashes, hash2)
	})

	t.Run("handles duplicate hashes", func(t *testing.T) {
		hash := common.HexToHash("0x123456")

		req := WantAggregationProofsRequest{
			RequestHashes: []common.Hash{hash, hash, hash},
		}
		require.Len(t, req.RequestHashes, 3)
		for _, h := range req.RequestHashes {
			require.Equal(t, hash, h)
		}
	})
}

func TestWantAggregationProofsResponse(t *testing.T) {
	t.Run("creates empty response", func(t *testing.T) {
		resp := WantAggregationProofsResponse{}
		require.Nil(t, resp.Proofs)
	})

	t.Run("creates response with proofs", func(t *testing.T) {
		hash1 := common.HexToHash("0x123456")
		hash2 := common.HexToHash("0x789abc")

		proof1 := AggregationProof{
			VerificationType: VerificationTypeBlsBn254Simple,
			MessageHash:      []byte("message1"),
			Proof:            []byte("proof1"),
		}
		proof2 := AggregationProof{
			VerificationType: VerificationTypeBlsBn254ZK,
			MessageHash:      []byte("message2"),
			Proof:            []byte("proof2"),
		}

		resp := WantAggregationProofsResponse{
			Proofs: map[common.Hash]AggregationProof{
				hash1: proof1,
				hash2: proof2,
			},
		}

		require.Len(t, resp.Proofs, 2)
		require.Equal(t, proof1, resp.Proofs[hash1])
		require.Equal(t, proof2, resp.Proofs[hash2])
	})

	t.Run("handles missing hash lookup", func(t *testing.T) {
		hash := common.HexToHash("0x123456")
		nonExistentHash := common.HexToHash("0x999999")

		proof := AggregationProof{
			VerificationType: VerificationTypeBlsBn254Simple,
			MessageHash:      []byte("message"),
			Proof:            []byte("proof"),
		}

		resp := WantAggregationProofsResponse{
			Proofs: map[common.Hash]AggregationProof{
				hash: proof,
			},
		}

		// Existing hash should return proof
		foundProof, exists := resp.Proofs[hash]
		require.True(t, exists)
		require.Equal(t, proof, foundProof)

		// Non-existent hash should return zero value and false
		_, exists = resp.Proofs[nonExistentHash]
		require.False(t, exists)
	})
}

func TestAggregationProofProcessingStats(t *testing.T) {
	t.Run("creates zero stats", func(t *testing.T) {
		stats := AggregationProofProcessingStats{}

		require.Equal(t, 0, stats.ProcessedCount)
		require.Equal(t, 0, stats.UnrequestedProofCount)
		require.Equal(t, 0, stats.SignatureRequestErrorCount)
		require.Equal(t, 0, stats.VerificationErrorCount)
		require.Equal(t, 0, stats.ProcessingErrorCount)
		require.Equal(t, 0, stats.AlreadyExistCount)
		require.Equal(t, 0, stats.TotalErrors())
	})

	t.Run("calculates total errors correctly", func(t *testing.T) {
		stats := AggregationProofProcessingStats{
			ProcessedCount:             5, // Not included in errors
			UnrequestedProofCount:      2,
			SignatureRequestErrorCount: 3,
			VerificationErrorCount:     1,
			ProcessingErrorCount:       4,
			AlreadyExistCount:          2,
		}

		expectedErrors := 2 + 3 + 1 + 4 + 2 // 12
		require.Equal(t, expectedErrors, stats.TotalErrors())
	})

	t.Run("handles zero errors", func(t *testing.T) {
		stats := AggregationProofProcessingStats{
			ProcessedCount: 10,
			// All error counts remain 0
		}

		require.Equal(t, 0, stats.TotalErrors())
	})

	t.Run("handles only one type of error", func(t *testing.T) {
		stats := AggregationProofProcessingStats{
			ProcessedCount:         5,
			VerificationErrorCount: 3,
			// All other error counts remain 0
		}

		require.Equal(t, 3, stats.TotalErrors())
	})

	t.Run("handles large numbers", func(t *testing.T) {
		stats := AggregationProofProcessingStats{
			ProcessedCount:             1000000,
			UnrequestedProofCount:      50000,
			SignatureRequestErrorCount: 25000,
			VerificationErrorCount:     10000,
			ProcessingErrorCount:       15000,
			AlreadyExistCount:          5000,
		}

		expectedErrors := 50000 + 25000 + 10000 + 15000 + 5000 // 105000
		require.Equal(t, expectedErrors, stats.TotalErrors())
	})
}

func TestAggregationProofProcessingStats_Comparison_With_SignatureProcessingStats(t *testing.T) {
	t.Run("both stats follow similar patterns", func(t *testing.T) {
		// Create similar error scenarios for both types
		sigStats := SignatureProcessingStats{
			ProcessedCount:             5,
			UnrequestedSignatureCount:  2,
			SignatureRequestErrorCount: 3,
			ProcessingErrorCount:       4,
			AlreadyExistCount:          2,
			// Signature-specific errors
			UnrequestedHashCount:        1,
			PublicKeyErrorCount:         1,
			ValidatorInfoErrorCount:     1,
			ValidatorIndexMismatchCount: 1,
		}

		aggStats := AggregationProofProcessingStats{
			ProcessedCount:             5,
			UnrequestedProofCount:      2,
			SignatureRequestErrorCount: 3,
			ProcessingErrorCount:       4,
			AlreadyExistCount:          2,
			// Aggregation-specific errors
			VerificationErrorCount: 1, // Fewer errors than signature-specific errors
		}

		// Both should have meaningful error counts
		require.Positive(t, sigStats.TotalErrors())
		require.Positive(t, aggStats.TotalErrors())

		// Aggregation stats should be simpler (fewer error types)
		require.Less(t, aggStats.TotalErrors(), sigStats.TotalErrors())
	})
}

func TestAggregationProofSync_Integration(t *testing.T) {
	t.Run("complete aggregation proof sync flow", func(t *testing.T) {
		// Step 1: Create request for missing aggregation proofs
		hash1 := common.HexToHash("0x123456")
		hash2 := common.HexToHash("0x789abc")
		hash3 := common.HexToHash("0xdef012")

		request := WantAggregationProofsRequest{
			RequestHashes: []common.Hash{hash1, hash2, hash3},
		}

		require.Len(t, request.RequestHashes, 3)

		// Step 2: Create response with available proofs (only 2 out of 3)
		proof1 := AggregationProof{
			VerificationType: VerificationTypeBlsBn254Simple,
			MessageHash:      []byte("message1"),
			Proof:            []byte("proof1"),
		}
		proof2 := AggregationProof{
			VerificationType: VerificationTypeBlsBn254ZK,
			MessageHash:      []byte("message2"),
			Proof:            []byte("proof2"),
		}

		response := WantAggregationProofsResponse{
			Proofs: map[common.Hash]AggregationProof{
				hash1: proof1,
				hash2: proof2,
				// hash3 is missing - peer doesn't have it
			},
		}

		require.Len(t, response.Proofs, 2)

		// Step 3: Process response and track stats
		stats := AggregationProofProcessingStats{}

		// Simulate processing each requested hash
		for _, reqHash := range request.RequestHashes {
			if proof, exists := response.Proofs[reqHash]; exists {
				// Proof found and processed successfully
				_ = proof // Use the proof
				stats.ProcessedCount++
			}
			// Proof not found in response - no error, just missing
			// This is normal behavior when peers don't have all proofs
		}

		require.Equal(t, 2, stats.ProcessedCount)
		require.Equal(t, 0, stats.TotalErrors())

		// Step 4: Simulate some processing errors for remaining proof
		stats.VerificationErrorCount = 1 // hash3 had verification error
		require.Equal(t, 1, stats.TotalErrors())
	})
}
