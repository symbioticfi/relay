package tests

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"
	"github.com/symbioticfi/relay/core/client/evm"
	"github.com/symbioticfi/relay/core/entity"
	valsetDeriver "github.com/symbioticfi/relay/core/usecase/valset-deriver"
	"github.com/testcontainers/testcontainers-go"
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
	deploymentData, err := loadDeploymentData(t.Context())
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

	nwConfig, err := evmClient.GetConfig(t.Context(), captureTimestamp)
	require.NoError(t, err, "Failed to get network config")

	valset, err := deriver.GetValidatorSet(t.Context(), currentEpoch, nwConfig)
	require.NoError(t, err, "Failed to get validator set")

	// next valset, we expect nothing to change apart from epoch details
	valset.Epoch++
	valset.CaptureTimestamp += deploymentData.Env.EpochTime

	aggIndices, commIndices, err := deriver.GetSchedulerInfo(t.Context(), valset, nwConfig)
	require.NoError(t, err, "Failed to get scheduler info")
	require.NotEmpty(t, aggIndices, "No aggregators found in scheduler info")
	require.NotEmpty(t, commIndices, "No committers found in scheduler info")
	valset.AggregatorIndices = aggIndices
	valset.CommitterIndices = commIndices

	for i := range globalTestEnv.SidecarConfigs {
		if valset.IsAggregator(globalTestEnv.SidecarConfigs[i].RequiredSymKey.PublicKey().OnChain()) {
			aggregatorIndexes = append(aggregatorIndexes, i)
		} else if !valset.IsCommitter(globalTestEnv.SidecarConfigs[i].RequiredSymKey.PublicKey().OnChain()) && valset.IsSigner(globalTestEnv.SidecarConfigs[i].RequiredSymKey.PublicKey().OnChain()) {
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
		container := globalTestEnv.Containers[aggIndex]
		err := container.Stop(ctx, nil)
		require.NoError(t, err, "Failed to stop aggregator container %d", aggIndex)
		stoppedAggregators = append(stoppedAggregators, aggIndex)
		t.Logf("Stopped aggregator container %d", aggIndex)
	}

	// Step 3: Wait for next epoch to trigger signature generation
	// During this time, signers will generate signatures but aggregators are offline
	t.Log("Step 3: Waiting for next epoch to trigger signature generation...")

	nextEpoch := currentEpoch + 1
	err = waitForEpoch(ctx, evmClient, nextEpoch, 2*time.Minute)
	require.NoError(t, err, "Failed to wait for next epoch")
	t.Logf("Reached epoch %d", nextEpoch)

	// Step 4: Verify signers have generated signatures
	t.Log("Step 4: Verifying signers have generated signatures...")
	expected := map[string]interface{}{
		"msg":   "Message signed",
		"epoch": float64(nextEpoch),
	}
	err = waitForLogLine(ctx, globalTestEnv.Containers[onlySignerIndex], expected, 1*time.Minute)
	require.NoError(t, err)

	// Step 5: Start aggregators back up
	t.Log("Step 5: Starting aggregator containers back up...")
	for _, aggIndex := range stoppedAggregators {
		container := globalTestEnv.Containers[aggIndex]
		err := container.Start(ctx)
		require.NoError(t, err, "Failed to restart aggregator container %d", aggIndex)
		t.Logf("Restarted aggregator container %d", aggIndex)

		mappedPort, err := container.MappedPort(ctx, "8080/tcp")
		require.NoError(t, err, "Failed to get mapped port for aggregator %d", aggIndex)
		globalTestEnv.ContainerPorts[aggIndex] = mappedPort.Port()
	}

	// Step 6: Verify aggregators have synced and generated proofs
	t.Log("Step 6: Verifying aggregators have synced signatures and generated proofs...")

	for _, aggIndex := range aggregatorIndexes {
		// Wait for aggregator to be healthy
		healthEndpoint := globalTestEnv.GetHealthEndpoint(aggIndex)
		err := waitForHealthy(ctx, healthEndpoint, 30*time.Second)
		require.NoError(t, err, "Aggregator %d failed to become healthy after restart", aggIndex)
		t.Logf("Aggregator %d is healthy", aggIndex)

		proofExpectedLog := map[string]interface{}{
			"msg":   "Proof created, trying to send aggregated signature message",
			"epoch": float64(nextEpoch),
		}
		err = waitForLogLine(ctx, globalTestEnv.Containers[aggIndex], proofExpectedLog, time.Minute)
		require.NoError(t, err)
		t.Logf("Aggregator %d has synced signatures and generated proof", aggIndex)
	}

	t.Log("✅ Signature sync test completed successfully")
}

// waitForHealthy polls a health endpoint until it returns 200 or timeout occurs
func waitForHealthy(ctx context.Context, healthURL string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("timeout waiting for health endpoint %s: %v", healthURL, ctx.Err())
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
			if err != nil {
				continue
			}

			resp, err := client.Do(req)
			if err != nil {
				continue // Continue trying on error
			}
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
	}
}

// waitForLogLine waits for a specific line to appear in container logs
func waitForLogLine(ctx context.Context, container testcontainers.Container, expectedLog map[string]interface{}, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("timeout waiting for log line '%v': %v", expectedLog, ctx.Err())
		case <-ticker.C:
			logs, err := container.Logs(ctx)
			if err != nil {
				continue // Continue trying on error
			}

			// Read logs content
			logBytes, err := io.ReadAll(logs)
			logs.Close()
			if err != nil {
				continue
			}

			// Check each line for JSON match
			scanner := bufio.NewScanner(strings.NewReader(string(logBytes)))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}

				// Try to parse line as JSON
				var logMap map[string]interface{}
				if err := json.Unmarshal([]byte(line), &logMap); err != nil {
					continue // Skip non-JSON lines
				}

				// Check if this log contains all expected key-value pairs
				if containsAllKeyValues(logMap, expectedLog) {
					return nil
				}
			}
		}
	}
}

// containsAllKeyValues checks if logMap contains all key-value pairs from expectedLog
func containsAllKeyValues(logMap, expectedLog map[string]interface{}) bool {
	for key, expectedValue := range expectedLog {
		actualValue, exists := logMap[key]
		if !exists {
			return false
		}

		// Compare values (handles different types appropriately)
		if actualValue != expectedValue {
			return false
		}
	}
	return true
}

// createEVMClient creates an EVM client for interacting with the blockchain
func createEVMClient(t *testing.T, deploymentData RelayContractsData) *evm.Client {
	t.Helper()
	config := evm.Config{
		ChainURLs: settlementChains,
		DriverAddress: entity.CrossChainAddress{
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
func waitForEpoch(ctx context.Context, client evm.IEvmClient, targetEpoch uint64, timeout time.Duration) error {
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
