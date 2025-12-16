package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/symbioticfi/relay/api/client/v1"
	"github.com/symbioticfi/relay/symbiotic/client/evm"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	valsetDeriver "github.com/symbioticfi/relay/symbiotic/usecase/valset-deriver"
)

// TestAggregatorSignatureSync tests that aggregators can sync missed signatures
// and generate proofs even when they were offline during signature collection.
//
// Test scenario:
// 1. Get current epoch from EVM client
// 2. Stop all aggregators before signatures are generated
// 3. Wait for next epoch to trigger signature generation by signers
// 4. Verify signers have generated signatures
// 5. Start aggregators back up
// 6. Verify aggregators sync missed signatures and generate proofs
func TestAggregatorSignatureSync(t *testing.T) {
	ctx := t.Context()

	// Load deployment data to get contract addresses and environment info
	deploymentData, err := loadDeploymentData()
	require.NoError(t, err, "Failed to load deployment data")

	// Identify aggregators
	onlySignerIndex := -1
	var aggregatorIndexes []int

	// Step 1: Get current epoch from EVM client
	evmClient := createEVMClient(t, deploymentData)
	currentEpoch, err := evmClient.GetCurrentEpoch(ctx)
	require.NoError(t, err, "Failed to get current epoch")
	t.Logf("Step 1: Current epoch: %d", currentEpoch)

	t.Logf("Identifying aggregators and signer-only nodes for next epoch %d...", currentEpoch+1)
	deriver, err := valsetDeriver.NewDeriver(evmClient)
	require.NoError(t, err, "Failed to create valset deriver")

	captureTimestamp, err := evmClient.GetEpochStart(ctx, currentEpoch)
	require.NoError(t, err, "Failed to get epoch start timestamp")

	nwConfig, err := evmClient.GetConfig(ctx, captureTimestamp, currentEpoch)
	require.NoError(t, err, "Failed to get network config")

	nextEpoch := currentEpoch + 1
	nextValset, err := deriver.GetValidatorSet(ctx, nextEpoch, nwConfig)
	require.NoError(t, err, "Failed to get validator set")

	for i, sidecarConfig := range deploymentData.Env.GetSidecarConfigs() {
		onChainKey := sidecarConfig.RequiredSymKey.PublicKey().OnChain()
		if nextValset.IsAggregator(onChainKey) {
			aggregatorIndexes = append(aggregatorIndexes, i)
		} else if !nextValset.IsCommitter(onChainKey) && nextValset.IsSigner(onChainKey) {
			onlySignerIndex = i
		}
	}

	require.Greater(t, onlySignerIndex, -1, "No signer-only node found in test environment")
	require.NotEmpty(t, aggregatorIndexes, "No aggregators found in test environment")

	t.Logf("Found %d aggregators", len(aggregatorIndexes))
	t.Logf("Signer-only node index: %d", onlySignerIndex)

	// Step 2: Stop all aggregator containers
	t.Log("Step 2: Stopping all aggregator containers...")
	stoppedAggregators := make([]int, 0, len(aggregatorIndexes))
	for _, aggIndex := range aggregatorIndexes {
		container := deploymentData.Env.GetSidecarConfigs()[aggIndex].ContainerName
		err := stopContainer(ctx, container)
		require.NoError(t, err, "Failed to stop aggregator container %d", aggIndex)
		stoppedAggregators = append(stoppedAggregators, aggIndex)
		t.Logf("Stopped aggregator container %d", aggIndex)
	}

	// Step 3: Wait for next epoch to trigger signature generation
	// During this time, signers will generate signatures but aggregators are offline
	t.Log("Step 3: Waiting for next epoch to trigger signature generation...")

	err = waitForEpoch(ctx, evmClient, nextEpoch, 2*time.Minute)
	require.NoError(t, err, "Failed to wait for next epoch")
	t.Logf("Reached epoch %d", nextEpoch)

	// Step 4: Verify signers have generated signatures
	t.Log("Step 4: Verifying signers have generated signatures...")
	client := getGRPCClient(t, onlySignerIndex)
	err = waitForErrorIsNil(ctx, time.Second*30, func() error {
		_, err := client.GetValidatorSetMetadata(ctx, &apiv1.GetValidatorSetMetadataRequest{
			Epoch: (*uint64)(&nextEpoch),
		})
		return err
	})
	require.NoError(t, err)

	metadataResp, err := client.GetValidatorSetMetadata(ctx, &apiv1.GetValidatorSetMetadataRequest{
		Epoch: (*uint64)(&nextEpoch),
	})
	require.NoError(t, err)

	// Step 5: Start aggregators back up
	t.Log("Step 5: Starting aggregator containers back up...")
	for _, aggIndex := range stoppedAggregators {
		container := deploymentData.Env.GetSidecarConfigs()[aggIndex].ContainerName
		err := startContainer(ctx, container)
		require.NoError(t, err, "Failed to restart aggregator container %d", aggIndex)
		t.Logf("Restarted aggregator container %d", aggIndex)
	}

	// Step 6: Verify aggregators have synced and generated proofs
	t.Log("Step 6: Verifying aggregators have synced signatures and generated proofs...")

	for _, aggIndex := range aggregatorIndexes {
		// Wait for aggregator to be healthy
		healthEndpoint := getHealthEndpoint(aggIndex)
		err := waitForHealthy(ctx, healthEndpoint, 60*time.Second)
		require.NoError(t, err, "Aggregator %d failed to become healthy after restart", aggIndex)
		t.Logf("Aggregator %d is healthy", aggIndex)

		err = waitForErrorIsNil(ctx, time.Second*30, func() error {
			_, err = client.GetAggregationProof(ctx, &apiv1.GetAggregationProofRequest{
				RequestId: metadataResp.GetRequestId(),
			})
			return err
		})
		require.NoError(t, err, "Failed to get aggregation proof from aggregator %d", aggIndex)

		t.Logf("Aggregator %d has synced signatures and generated proof", aggIndex)
	}

	t.Log("✅ Signature sync test completed successfully")
}

func TestAggregatorProofSync(t *testing.T) {
	ctx := t.Context()

	// Load deployment data to get contract addresses and environment info
	deploymentData, err := loadDeploymentData()
	require.NoError(t, err, "Failed to load deployment data")

	// Step 1: Get current epoch from EVM client
	evmClient := createEVMClient(t, deploymentData)
	currentEpoch, err := evmClient.GetCurrentEpoch(ctx)
	require.NoError(t, err, "Failed to get current epoch")
	t.Logf("Step 1: Current epoch: %d", currentEpoch)

	t.Logf("Identifying signer-only nodes for next epoch %d...", currentEpoch+1)
	deriver, err := valsetDeriver.NewDeriver(evmClient)
	require.NoError(t, err, "Failed to create valset deriver")

	captureTimestamp, err := evmClient.GetEpochStart(ctx, currentEpoch)
	require.NoError(t, err, "Failed to get epoch start timestamp")

	nwConfig, err := evmClient.GetConfig(ctx, captureTimestamp, currentEpoch)
	require.NoError(t, err, "Failed to get network config")

	nextEpoch := currentEpoch + 1
	nextValset, err := deriver.GetValidatorSet(ctx, nextEpoch, nwConfig)
	require.NoError(t, err, "Failed to get validator set")

	// Identify only signers
	onlySignerIndex := -1
	for i, sidecarConfig := range deploymentData.Env.GetSidecarConfigs() {
		onChainKey := sidecarConfig.RequiredSymKey.PublicKey().OnChain()
		if nextValset.IsSigner(onChainKey) && !nextValset.IsAggregator(onChainKey) {
			onlySignerIndex = i
		}
	}
	require.Greater(t, onlySignerIndex, -1, "No signer-only node found in test environment")

	t.Logf("Signer-only node index: %d", onlySignerIndex)

	// Step 2: Stop signer container
	t.Log("Step 2: Stopping only signer container...")
	container := deploymentData.Env.GetSidecarConfigs()[onlySignerIndex].ContainerName
	err = stopContainer(ctx, container)
	require.NoError(t, err, "Failed to stop only signer container %d", onlySignerIndex)
	t.Logf("Stopped only signer %d", onlySignerIndex)

	// Step 3: Wait for next epoch to trigger proof generation
	t.Log("Step 3: Waiting for next epoch to trigger proof generation...")

	err = waitForEpoch(ctx, evmClient, nextEpoch, 2*time.Minute)
	require.NoError(t, err, "Failed to wait for next epoch")
	t.Logf("Reached epoch %d", nextEpoch)

	t.Log("Step 4: Verifying other operator has proof...")
	anotherSignerIndex := lo.If(onlySignerIndex == 0, 1).Else(0)
	client := getGRPCClient(t, anotherSignerIndex)
	var metadataResp *apiv1.GetValidatorSetMetadataResponse
	err = waitForErrorIsNil(ctx, time.Second*30, func() error {
		metadataResp, err = client.GetValidatorSetMetadata(ctx, &apiv1.GetValidatorSetMetadataRequest{
			Epoch: (*uint64)(&nextEpoch),
		})
		return err
	})
	require.NoError(t, err)

	// Step 5: Start signer back up
	t.Log("Step 5: Starting aggregator containers back up...")
	err = startContainer(ctx, container)
	require.NoError(t, err, "Failed to restart only signer container %d", onlySignerIndex)
	t.Logf("Restarted only signer container %d", onlySignerIndex)

	t.Log("Step 6: Verifying only signer have synced proof...")

	healthEndpoint := getHealthEndpoint(onlySignerIndex)
	err = waitForHealthy(ctx, healthEndpoint, 60*time.Second)
	require.NoError(t, err, "Only signer %d failed to become healthy after restart", onlySignerIndex)
	t.Logf("Only signer %d is healthy", onlySignerIndex)

	err = waitForErrorIsNil(ctx, time.Second*30, func() error {
		_, err = client.GetAggregationProof(ctx, &apiv1.GetAggregationProofRequest{
			RequestId: metadataResp.GetRequestId(),
		})
		return err
	})
	require.NoError(t, err, "Failed to get aggregation proof from only signer %d", onlySignerIndex)

	t.Logf("Only signer %d has synced signatures and generated proof", onlySignerIndex)

	t.Log("✅ Proof sync test completed successfully")
}

func waitForErrorIsNil(ctx context.Context, timeout time.Duration, f func() error) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("timeout waiting for function to succeed: %v", ctx.Err())
		case <-ticker.C:
			err := f()
			if err == nil {
				return nil
			}
		}
	}
}

// waitForHealthy polls a health endpoint until it returns 200 or timeout occurs
func waitForHealthy(ctx context.Context, healthURL string, timeout time.Duration) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	return waitForErrorIsNil(ctx, timeout, func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		}

		return errors.Errorf("health check returned status %d", resp.StatusCode)
	})
}

// createEVMClient creates an EVM client for interacting with the blockchain
func createEVMClient(t *testing.T, deploymentData RelayContractsData) *evm.Client {
	t.Helper()
	config := evm.Config{
		ChainURLs: settlementChains,
		DriverAddress: symbiotic.CrossChainAddress{
			ChainId: 31337,
			Address: common.HexToAddress(deploymentData.GetDriverAddress()),
		},
		RequestTimeout: 10 * time.Second,
		KeyProvider:    &testMockKeyProvider{},
	}

	evmClient, err := evm.NewEvmClient(t.Context(), config)
	require.NoError(t, err)

	return evmClient
}

// waitForEpoch waits until the specified epoch is reached
func waitForEpoch(ctx context.Context, client evm.IEvmClient, targetEpoch symbiotic.Epoch, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("timeout waiting for epoch %d: %v", targetEpoch, ctx.Err())
		case <-ticker.C:
			currentEpoch, err := client.GetCurrentEpoch(ctx)
			if err != nil {
				// Continue trying on error, but log it
				continue
			}

			if currentEpoch >= targetEpoch {
				return nil
			}
		}
	}
}
