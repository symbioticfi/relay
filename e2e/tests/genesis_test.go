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

// TestGenesisDone tests that the genesis validator set header has been committed
func TestGenesisDone(t *testing.T) {
	t.Log("Starting genesis generation test...")

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

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	evmClient, err := evm.NewEvmClient(ctx, config)
	require.NoError(t, err, "Failed to create EVM client")

	cfg, err := evmClient.GetConfig(ctx, symbiotic.Timestamp(time.Now().Unix()), 0)
	require.NoError(t, err, "Failed to get settlement config")

	for _, settlement := range cfg.Settlements {
		t.Logf("Checking settlement %s on chain %d", settlement.Address, settlement.ChainId)
		_, err := evmClient.GetValSetHeader(ctx, settlement)
		require.NoErrorf(t, err, "Failed to get validator set header for settlement %s", settlement.Address)
	}

	t.Log("Genesis generation test completed successfully")
}

// TestContractData tests that the data in the contract matches expected values
func TestContractData(t *testing.T) {
	deployData := loadDeploymentData(t)

	expectedContractData := getExpectedDataFromContracts(t, deployData)

	require.Equal(t, len(expectedContractData.ValidatorSet.Validators), int(deployData.Env.Operators), "Validator set length does not match expected number of operators")
	require.Equal(t, uint32(expectedContractData.NetworkConfig.VerificationType), deployData.Env.VerificationType, "Verification type does not match expected value")
	require.Equal(t, expectedContractData.CurrentEpochDuration, deployData.Env.EpochTime, "Epoch time does not match expected value")
}
