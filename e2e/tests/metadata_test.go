package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	apiv1 "github.com/symbioticfi/relay/api/client/v1"
	"github.com/symbioticfi/relay/core/entity"
)

// TestGetValidatorSetMetadata tests the GetValidatorSetMetadata API endpoint
// and verifies that the signature target id can be used to retrieve the proof of a committed valset
func TestGetValidatorSetMetadata(t *testing.T) {
	t.Log("Starting validator set metadata API test...")

	_, err := loadDeploymentData(t.Context())
	require.NoError(t, err, "Failed to load deployment data")

	address := globalTestEnv.GetGRPCAddress(0)
	t.Logf("Testing validator set metadata API on %s", address)

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoErrorf(t, err, "Failed to connect to relay server at %s", address)
	defer conn.Close()

	client := apiv1.NewSymbioticClient(conn)

	// Get last committed epochs to find a committed epoch ≥1 for testing
	// We need committed epochs because that's when proofs and signatures are available
	var committedEpoch uint64
	const maxRetries = 10
	const retryDelay = 10 * time.Second

	t.Log("Waiting for committed epoch ≥1 on all chains...")
	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		lastAllCommittedResp, err := client.GetLastAllCommitted(ctx, &apiv1.GetLastAllCommittedRequest{})
		cancel()

		if err != nil {
			t.Logf("Attempt %d: Failed to get last committed epochs: %v", attempt, err)
			if attempt == maxRetries {
				t.Fatalf("Failed to get last committed epochs after %d attempts", maxRetries)
			}
			time.Sleep(retryDelay)
			continue
		}

		epochInfos := lastAllCommittedResp.GetEpochInfos()
		if len(epochInfos) == 0 {
			t.Logf("Attempt %d: No committed epochs found yet", attempt)
			if attempt == maxRetries {
				t.Fatalf("No committed epochs found after %d attempts", maxRetries)
			}
			time.Sleep(retryDelay)
			continue
		}

		// Find the minimum committed epoch across all chains that is ≥1
		minCommittedEpoch := uint64(0)
		allChainsReady := true

		for chainID, epochInfo := range epochInfos {
			lastCommitted := epochInfo.GetLastCommittedEpoch()
			if lastCommitted < 1 {
				t.Logf("Attempt %d: Chain %d has committed epoch %d (need ≥1)", attempt, chainID, lastCommitted)
				allChainsReady = false
				break
			}
			if minCommittedEpoch == 0 || lastCommitted < minCommittedEpoch {
				minCommittedEpoch = lastCommitted
			}
		}

		if allChainsReady && minCommittedEpoch >= 1 {
			committedEpoch = minCommittedEpoch
			t.Logf("Found committed epoch %d on all chains", committedEpoch)
			break
		}

		if attempt == maxRetries {
			t.Fatalf("Not all chains have committed epoch ≥1 after %d attempts", maxRetries)
		}

		t.Logf("Attempt %d: Waiting for all chains to have committed epoch ≥1...", attempt)
		time.Sleep(retryDelay)
	}

	// Test 1: Get metadata for committed epoch (should work and have proofs/signatures)
	t.Run("GetMetadataForCommittedEpoch", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()

		metadataResp, err := client.GetValidatorSetMetadata(ctx, &apiv1.GetValidatorSetMetadataRequest{
			Epoch: &committedEpoch,
		})
		require.NoError(t, err, "Failed to get validator set metadata for committed epoch")

		// Validate response structure
		require.NotEmpty(t, metadataResp.GetSignatureTargetId(), "Signature target id should not be empty")
		require.NotEmpty(t, metadataResp.GetCommitmentData(), "Commitment data should not be empty")
		// ExtraData can be empty, so we don't require it to be non-empty

		signatureTargetId := metadataResp.GetSignatureTargetId()
		t.Logf("Got metadata for committed epoch %d with signature target id: %s", committedEpoch, signatureTargetId)

		// Test 2: Use the signature target id to get signature request
		t.Run("GetSignatureRequestFromHash", func(t *testing.T) {
			sigReqResp, err := client.GetSignatureRequest(ctx, &apiv1.GetSignatureRequestRequest{
				SignatureTargetId: signatureTargetId,
			})
			require.NoError(t, err, "Failed to get signature request using signature target id from metadata")

			// Validate the signature request
			require.Equal(t, uint32(entity.ValsetHeaderKeyTag), sigReqResp.GetKeyTag(),
				"Key tag should be ValsetHeaderKeyTag")
			require.Equal(t, committedEpoch, sigReqResp.GetRequiredEpoch(),
				"Required epoch should match committed epoch")
			require.NotEmpty(t, sigReqResp.GetMessage(), "Message should not be empty")

			t.Logf("Successfully retrieved signature request for key tag %d, epoch %d",
				sigReqResp.GetKeyTag(), sigReqResp.GetRequiredEpoch())
		})

		// Test 3: Get aggregation proof (should exist for committed epochs)
		t.Run("GetAggregationProofFromHash", func(t *testing.T) {
			proofResp, err := client.GetAggregationProof(ctx, &apiv1.GetAggregationProofRequest{
				SignatureTargetId: signatureTargetId,
			})

			// For committed epochs, aggregation proof should be available
			require.NoError(t, err, "Failed to get aggregation proof for committed epoch signature target id %s", signatureTargetId)
			require.NotNil(t, proofResp.GetAggregationProof(), "Aggregation proof should not be nil for committed epoch")
			require.NotEmpty(t, proofResp.GetAggregationProof().GetProof(),
				"Aggregation proof data should not be empty for committed epoch")
			require.NotEmpty(t, proofResp.GetAggregationProof().GetMessageHash(),
				"Aggregation proof message hash should not be empty")
			t.Logf("Successfully retrieved aggregation proof for committed epoch signature target id %s", signatureTargetId)
		})

		// Test 4: Get signatures for the signature target id (should exist for committed epochs)
		t.Run("GetSignaturesFromHash", func(t *testing.T) {
			signaturesResp, err := client.GetSignatures(ctx, &apiv1.GetSignaturesRequest{
				SignatureTargetId: signatureTargetId,
			})

			// For committed epochs, signatures should be available
			require.NoError(t, err, "Failed to get signatures for committed epoch signature target id %s", signatureTargetId)
			require.NotEmpty(t, signaturesResp.GetSignatures(), "Should have signatures for committed epoch")

			t.Logf("Found %d signatures for committed epoch signature target id %s",
				len(signaturesResp.GetSignatures()), signatureTargetId)

			// Validate signatures structure
			for i, sig := range signaturesResp.GetSignatures() {
				require.NotEmpty(t, sig.GetSignature(), "Signature %d should not be empty", i)
				require.NotEmpty(t, sig.GetMessageHash(), "Message hash %d should not be empty", i)
				require.NotEmpty(t, sig.GetPublicKey(), "Public key %d should not be empty", i)
			}
		})
	})

	// Test 5: Get metadata without specifying epoch (should use current epoch)
	t.Run("GetMetadataWithoutEpoch", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()

		metadataResp, err := client.GetValidatorSetMetadata(ctx, &apiv1.GetValidatorSetMetadataRequest{})
		require.NoError(t, err, "Failed to get validator set metadata without specifying epoch")

		require.NotEmpty(t, metadataResp.GetSignatureTargetId(), "signature target id should not be empty")
		require.NotEmpty(t, metadataResp.GetCommitmentData(), "Commitment data should not be empty")

		t.Logf("Got metadata without epoch specification with signature target id: %s",
			metadataResp.GetSignatureTargetId())
	})

	// Test 6: Try to get metadata for a future epoch (should fail)
	t.Run("GetMetadataForFutureEpoch", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()

		// Get current epoch to determine what would be a future epoch
		currentEpochResp, err := client.GetCurrentEpoch(ctx, &apiv1.GetCurrentEpochRequest{})
		require.NoError(t, err, "Failed to get current epoch")
		futureEpoch := currentEpochResp.GetEpoch() + 100

		_, err = client.GetValidatorSetMetadata(ctx, &apiv1.GetValidatorSetMetadataRequest{
			Epoch: &futureEpoch,
		})
		require.Error(t, err, "Should fail to get metadata for future epoch %d", futureEpoch)
		t.Logf("Correctly failed to get metadata for future epoch %d: %v", futureEpoch, err)
	})

	t.Log("Validator set metadata API test completed successfully")
}

// TestGetLastAllCommitted tests the GetLastAllCommitted API endpoint
// and validates whether it returns proper epoch info for the contracts
func TestGetLastAllCommitted(t *testing.T) {
	t.Log("Starting GetLastAllCommitted API test...")

	deploymentData, err := loadDeploymentData(t.Context())
	require.NoError(t, err, "Failed to load deployment data")

	address := globalTestEnv.GetGRPCAddress(0)
	t.Logf("Testing GetLastAllCommitted API on %s", address)

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoErrorf(t, err, "Failed to connect to relay server at %s", address)
	defer conn.Close()

	client := apiv1.NewSymbioticClient(conn)

	// Get expected data from contracts to validate against
	expected := getExpectedDataFromContracts(t, deploymentData)
	t.Logf("Expected network config has %d settlement chains", len(expected.NetworkConfig.Settlements))

	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()

	// Test the GetLastAllCommitted API
	lastAllCommittedResp, err := client.GetLastAllCommitted(ctx, &apiv1.GetLastAllCommittedRequest{})
	require.NoError(t, err, "Failed to get last all committed epochs")

	epochInfos := lastAllCommittedResp.GetEpochInfos()
	require.NotEmpty(t, epochInfos, "Epoch infos should not be empty")

	t.Logf("GetLastAllCommitted returned epoch info for %d chains", len(epochInfos))

	// Get current epoch for validation
	currentEpochResp, err := client.GetCurrentEpoch(ctx, &apiv1.GetCurrentEpochRequest{})
	require.NoError(t, err, "Failed to get current epoch for validation")
	currentEpoch := currentEpochResp.GetEpoch()

	// Validate that we have epoch info for all settlement chains in the network config
	for _, settlement := range expected.NetworkConfig.Settlements {
		chainID := settlement.ChainId
		epochInfo, exists := epochInfos[chainID]
		require.True(t, exists, "Should have epoch info for settlement chain %d", chainID)

		require.NotNil(t, epochInfo, "Epoch info should not be nil for chain %d", chainID)
		require.NotNil(t, epochInfo.GetStartTime(), "Start time should not be nil for chain %d", chainID)

		lastCommittedEpoch := epochInfo.GetLastCommittedEpoch()
		startTime := epochInfo.GetStartTime()

		// Validate that the epoch is reasonable (not 0 and not way in the future)
		require.Positive(t, lastCommittedEpoch, "Last committed epoch should be positive for chain %d", chainID)
		require.LessOrEqual(t, lastCommittedEpoch, currentEpoch,
			"Last committed epoch %d should not exceed current epoch %d for chain %d",
			lastCommittedEpoch, currentEpoch, chainID)

		// Validate start time is reasonable (not zero and not in the future)
		require.Positive(t, startTime.GetSeconds(), "Start time should be positive for chain %d", chainID)
		require.True(t, startTime.AsTime().Before(time.Now().Add(time.Minute)),
			"Start time should not be in the future for chain %d", chainID)

		t.Logf("Chain %d: Last committed epoch %d, start time %s",
			chainID, lastCommittedEpoch, startTime.AsTime().Format(time.RFC3339))
	}

	// Test individual GetLastCommitted for each chain to ensure consistency
	t.Run("ValidateIndividualChainQueries", func(t *testing.T) {
		for chainID, expectedEpochInfo := range epochInfos {
			t.Run(fmt.Sprintf("Chain_%d", chainID), func(t *testing.T) {
				individualResp, err := client.GetLastCommitted(ctx, &apiv1.GetLastCommittedRequest{
					SettlementChainId: chainID,
				})
				require.NoError(t, err, "Failed to get last committed for chain %d", chainID)

				require.Equal(t, chainID, individualResp.GetSettlementChainId(),
					"Settlement chain ID should match for chain %d", chainID)

				individualEpochInfo := individualResp.GetEpochInfo()
				require.NotNil(t, individualEpochInfo, "Epoch info should not be nil for chain %d", chainID)

				// Compare with the result from GetLastAllCommitted
				require.Equal(t, expectedEpochInfo.GetLastCommittedEpoch(),
					individualEpochInfo.GetLastCommittedEpoch(),
					"Last committed epoch should match between GetLastAllCommitted and GetLastCommitted for chain %d", chainID)

				require.Equal(t, expectedEpochInfo.GetStartTime().GetSeconds(),
					individualEpochInfo.GetStartTime().GetSeconds(),
					"Start time should match between GetLastAllCommitted and GetLastCommitted for chain %d", chainID)

				t.Logf("Chain %d individual query matches GetLastAllCommitted result", chainID)
			})
		}
	})

	// Test edge case: query for non-existent chain
	t.Run("QueryNonExistentChain", func(t *testing.T) {
		nonExistentChainID := uint64(999999)
		_, err := client.GetLastCommitted(ctx, &apiv1.GetLastCommittedRequest{
			SettlementChainId: nonExistentChainID,
		})
		require.Error(t, err, "Should fail to get last committed for non-existent chain %d", nonExistentChainID)
		t.Logf("Correctly failed to get data for non-existent chain %d: %v", nonExistentChainID, err)
	})

	t.Log("GetLastAllCommitted API test completed successfully")
}
