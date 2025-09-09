package tests

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/core/client/evm"
	"github.com/symbioticfi/relay/core/entity"
)

// TestGenesisDone tests that the genesis validator set header has been committed
func TestGenesisDone(t *testing.T) {
	t.Log("Starting genesis generation test...")

	deployData, err := loadDeploymentData(t.Context())
	require.NoError(t, err, "Failed to load deployment data")

	config := evm.Config{
		ChainURLs: settlementChains,
		DriverAddress: entity.CrossChainAddress{
			ChainId: deployData.Driver.ChainId,
			Address: common.HexToAddress(deployData.Driver.Addr),
		},
		RequestTimeout: 10 * time.Second,
		KeyProvider:    &testMockKeyProvider{},
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	evmClient, err := evm.NewEvmClient(ctx, config)
	require.NoError(t, err, "Failed to create EVM client")

	for _, settlement := range deployData.Settlements {
		t.Logf("Checking settlement %s on chain %d", settlement.Addr, settlement.ChainId)
		_, err := evmClient.GetValSetHeader(ctx, entity.CrossChainAddress{
			ChainId: settlement.ChainId,
			Address: common.HexToAddress(settlement.Addr),
		})
		require.NoError(t, err, "Failed to get validator set header for settlement %s", settlement.Addr)
	}

	t.Log("Genesis generation test completed successfully")
}

// TestContractData tests that the data in the contract matches expected values
func TestContractData(t *testing.T) {
	deployData, err := loadDeploymentData(t.Context())
	require.NoError(t, err, "Failed to load deployment data")

	expectedContractData := getExpectedDataFromContracts(t, deployData)

	require.Equal(t, len(expectedContractData.ValidatorSet.Validators), int(deployData.Env.Operators), "Validator set length does not match expected number of operators")
	require.Equal(t, uint32(expectedContractData.NetworkConfig.VerificationType), deployData.Env.VerificationType, "Verification type does not match expected value")
	require.Equal(t, expectedContractData.CurrentEpochDuration, deployData.Env.EpochTime, "Epoch time does not match expected value")
}
