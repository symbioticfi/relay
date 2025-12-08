package tests

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/symbiotic/client/evm"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// TestEpochProgression tests that epochs progress correctly over time
func TestEpochProgression(t *testing.T) {
	t.Log("Starting epoch progression test...")

	deployData, err := loadDeploymentData()
	require.NoError(t, err, "Failed to load deployment data")

	config := evm.Config{
		ChainURLs: settlementChains,
		DriverAddress: symbiotic.CrossChainAddress{
			ChainId: deployData.Driver.ChainId,
			Address: common.HexToAddress(deployData.Driver.Addr),
		},
		RequestTimeout: 10 * time.Second,
		KeyProvider:    &testMockKeyProvider{},
	}

	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)

	evmClient, err := evm.NewEvmClient(ctx, config)
	require.NoError(t, err, "Failed to create EVM client")

	cfg, err := evmClient.GetConfig(ctx, symbiotic.Timestamp(time.Now().Unix()), 0)
	require.NoError(t, err, "Failed to get settlement config")

	initialEpoch, err := evmClient.GetLastCommittedHeaderEpoch(ctx, cfg.Settlements[0])
	require.NoError(t, err, "Failed to get initial epoch")
	t.Logf("Initial committed epoch: %d", initialEpoch)

	cancel()

	// ensure the current epoch gets committed, timeout after 2x epoch time and error if still not committed
	ctx, cancel = context.WithTimeout(t.Context(), time.Duration(deployData.Env.EpochTime*2)*time.Second)
initialEpochCheck:
	for {
		select {
		case <-ctx.Done():
			t.Fatal("Timed out waiting for epoch to commit")
		default:
			// Check if the epoch has been committed
			committed, err := evmClient.IsValsetHeaderCommittedAt(ctx, cfg.Settlements[0], initialEpoch)
			require.NoError(t, err, "Failed to check if epoch is committed")
			if committed {
				t.Logf("Initial Epoch %d has been committed", initialEpoch)
				break initialEpochCheck
			}
			t.Logf("Waiting for epoch to commit, current: %d", initialEpoch)
			time.Sleep(5 * time.Second)
		}
	}
	cancel()

	// start watching for any new epochs being committed, will keep timeout to 5x the epoch duration
	t.Log("Waiting for epoch progression...")
	ctx, cancel = context.WithTimeout(t.Context(), time.Duration(deployData.Env.EpochTime*5)*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("Timed out waiting for epoch progression")
		default:
			currentEpoch, err := evmClient.GetLastCommittedHeaderEpoch(ctx, cfg.Settlements[0])
			require.NoError(t, err, "Failed to get current epoch")
			t.Logf("Last committed epoch: %d", currentEpoch)

			if currentEpoch > initialEpoch {
				t.Logf("New epoch detected: %d", currentEpoch)
				return
			}

			t.Logf("Waiting for new epoch commit, current: %d", currentEpoch)
			time.Sleep(5 * time.Second)
		}
	}
}
