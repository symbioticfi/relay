package entity

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestAggregationProofSync_Integration(t *testing.T) {
	t.Run("complete aggregation proof sync flow", func(t *testing.T) {
		// Step 1: Create request for missing aggregation proofs
		hash1 := common.HexToHash("0x123456")
		hash2 := common.HexToHash("0x789abc")
		hash3 := common.HexToHash("0xdef012")

		request := WantAggregationProofsRequest{
			RequestIDs: []common.Hash{hash1, hash2, hash3},
		}

		require.Len(t, request.RequestIDs, 3)

		// Step 2: Create response with available proofs (only 2 out of 3)
		proof1 := AggregationProof{
			MessageHash: []byte("message1"),
			KeyTag:      KeyTag(15),
			Epoch:       10,
			Proof:       []byte("proof1"),
		}
		proof2 := AggregationProof{
			MessageHash: []byte("message2"),
			KeyTag:      KeyTag(15),
			Epoch:       10,
			Proof:       []byte("proof2"),
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
		for _, requestID := range request.RequestIDs {
			if proof, exists := response.Proofs[requestID]; exists {
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
		stats.VerificationFailCount = 1 // hash3 had verification error
		require.Equal(t, 1, stats.TotalErrors())
	})
}
