package tests

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/symbiotic/client/evm"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// TestEpochProgression tests that epochs progress correctly over time
func TestEpochProgression(t *testing.T) {
	t.Log("Starting epoch progression test...")

	deployData := loadDeploymentData(t)

	config := evm.Config{
		ChainURLs: settlementChains,
		DriverAddress: symbiotic.CrossChainAddress{
			ChainId: deployData.Driver.ChainId,
			Address: common.HexToAddress(deployData.Driver.Addr),
		},
		RequestTimeout: 10 * time.Second,
		KeyProvider:    &testMockKeyProvider{},
	}

	evmClient, err := evm.NewEvmClient(t.Context(), config)
	require.NoError(t, err, "Failed to create EVM client")

	waitForNextCommitment(t, evmClient, deployData.Env.EpochTime)
}

func waitForNextCommitment(t *testing.T, evmClient *evm.Client, epochTime uint64) {
	t.Helper()

	deployData := loadDeploymentData(t)
	cfg, err := evmClient.GetConfig(t.Context(), symbiotic.Timestamp(time.Now().Unix()), 0)
	require.NoError(t, err, "Failed to get settlement config")

	initialEpoch, err := evmClient.GetLastCommittedHeaderEpoch(t.Context(), cfg.Settlements[0])
	require.NoError(t, err, "Failed to get initial epoch")
	t.Logf("Initial committed epoch: %d", initialEpoch)

	// ensure the current epoch gets committed, timeout after 2x epoch time and error if still not committed
	require.NoError(t, waitForErrorIsNil(t.Context(), time.Duration(deployData.Env.EpochTime*2)*time.Second, func() error {
		committed, err := evmClient.IsValsetHeaderCommittedAt(t.Context(), cfg.Settlements[0], initialEpoch)
		require.NoError(t, err, "Failed to check if epoch is committed")

		if committed {
			t.Logf("Initial Epoch %d has been committed", initialEpoch)
			return nil
		}

		return errors.Errorf("Epoch %d not yet committed", initialEpoch)
	}))

	waitForCommitmentInEpoch(t, evmClient, deployData.Env.EpochTime, initialEpoch+1)
}

func waitForCommitmentInEpoch(t *testing.T, evmClient *evm.Client, epochTime uint64, waitForEpoch symbiotic.Epoch) {
	t.Helper()

	cfg, err := evmClient.GetConfig(t.Context(), symbiotic.Timestamp(time.Now().Unix()), 0)
	require.NoError(t, err, "Failed to get settlement config")

	// start watching for any new epochs being committed, will keep timeout to 5x the epoch duration
	t.Log("Waiting for epoch progression...")
	require.NoError(t, waitForErrorIsNil(t.Context(), time.Duration(epochTime*5)*time.Second, func() error {
		currentEpoch, err := evmClient.GetLastCommittedHeaderEpoch(t.Context(), cfg.Settlements[0])
		require.NoError(t, err, "Failed to get current epoch")
		t.Logf("Last committed epoch: %d", currentEpoch)

		if currentEpoch >= waitForEpoch {
			t.Logf("New epoch detected: %d", currentEpoch)
			return nil
		}

		return errors.Errorf("No new epoch committed yet, current: %d", currentEpoch)
	}))
}
