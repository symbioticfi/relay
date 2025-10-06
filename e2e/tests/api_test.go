package tests

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/symbioticfi/relay/api/client/v1"
	"github.com/symbioticfi/relay/core/client/evm"
	"github.com/symbioticfi/relay/core/entity"
	valsetDeriver "github.com/symbioticfi/relay/core/usecase/valset-deriver"
)

// ContractExpectedData holds expected values derived from smart contracts
type ContractExpectedData struct {
	CurrentEpoch         entity.Epoch
	EpochStartTime       entity.Timestamp
	CurrentEpochDuration uint64
	ValidatorSet         entity.ValidatorSet
	NetworkConfig        entity.NetworkConfig
	IsEpochCommitted     bool
}

// getExpectedDataFromContracts retrieves expected values directly from smart contracts
func getExpectedDataFromContracts(t *testing.T, relayContracts RelayContractsData) *ContractExpectedData {
	t.Helper()

	config := evm.Config{
		ChainURLs: settlementChains,
		DriverAddress: entity.CrossChainAddress{
			ChainId: 31337,
			Address: common.HexToAddress(relayContracts.GetDriverAddress()),
		},
		RequestTimeout: 10 * time.Second,
		KeyProvider:    &testMockKeyProvider{},
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	evmClient, err := evm.NewEvmClient(ctx, config)
	require.NoError(t, err, "Failed to create EVM client")

	currentEpoch, err := evmClient.GetCurrentEpoch(ctx)
	require.NoError(t, err, "Failed to get current epoch from contract")

	currentEpochDuration, err := evmClient.GetCurrentEpochDuration(ctx)
	require.NoError(t, err, "Failed to get current epoch duration from contract")

	epochStart, err := evmClient.GetEpochStart(ctx, currentEpoch)
	require.NoError(t, err, "Failed to get epoch start time from contract")

	networkConfig, err := evmClient.GetConfig(ctx, epochStart)
	require.NoError(t, err, "Failed to get network config from contract")

	deriver, err := valsetDeriver.NewDeriver(evmClient)
	require.NoError(t, err, "Failed to create validator set deriver")

	expectedValset, err := deriver.GetValidatorSet(ctx, currentEpoch, networkConfig)
	require.NoError(t, err, "Failed to derive expected validator set")

	// Check if current epoch is committed
	isCurrentEpochCommitted := true
	for _, settlement := range networkConfig.Settlements {
		committed, err := evmClient.IsValsetHeaderCommittedAt(ctx, settlement, currentEpoch)
		if err != nil || !committed {
			isCurrentEpochCommitted = false
			break
		}
	}

	return &ContractExpectedData{
		CurrentEpoch:         currentEpoch,
		EpochStartTime:       epochStart,
		ValidatorSet:         expectedValset,
		NetworkConfig:        networkConfig,
		IsEpochCommitted:     isCurrentEpochCommitted,
		CurrentEpochDuration: currentEpochDuration,
	}
}

// validateValidatorSetAgainstExpected compares API response with expected contract data
func validateValidatorSetAgainstExpected(t *testing.T, apiResponse *apiv1.GetValidatorSetResponse, expected *ContractExpectedData) {
	t.Helper()

	require.Equal(t, expected.ValidatorSet.Epoch, entity.Epoch(apiResponse.Epoch), "API epoch should match contract epoch")
	require.Equal(t, expected.ValidatorSet.CaptureTimestamp, entity.Timestamp(apiResponse.CaptureTimestamp.GetSeconds()),
		"API capture timestamp should match contract timestamp")
	require.Equal(t, uint32(expected.ValidatorSet.Version), apiResponse.Version,
		"API version should match contract version")

	expectedQuorum := expected.ValidatorSet.QuorumThreshold.String()
	require.Equal(t, expectedQuorum, apiResponse.QuorumThreshold,
		"API quorum threshold should match contract quorum threshold")
	require.Len(t, apiResponse.Validators, len(expected.ValidatorSet.Validators),
		"API should return same number of validators as contract")

	for i, expectedValidator := range expected.ValidatorSet.Validators {
		require.Less(t, i, len(apiResponse.Validators), "API response should have validator %d", i)
		apiValidator := apiResponse.Validators[i]

		require.Equal(t, expectedValidator.Operator.Hex(), apiValidator.GetOperator(),
			"Validator %d operator should match contract", i)
		require.Equal(t, expectedValidator.VotingPower.String(), apiValidator.GetVotingPower(),
			"Validator %d voting power should match contract", i)
		require.Equal(t, expectedValidator.IsActive, apiValidator.GetIsActive(),
			"Validator %d active status should match contract", i)

		require.Len(t, apiValidator.GetKeys(), len(expectedValidator.Keys),
			"Validator %d should have same number of keys as contract", i)

		for j, expectedKey := range expectedValidator.Keys {
			require.Less(t, j, len(apiValidator.GetKeys()), "API should have key %d for validator %d", j, i)
			apiKey := apiValidator.GetKeys()[j]

			require.Equal(t, uint32(expectedKey.Tag), apiKey.GetTag(),
				"Validator %d key %d tag should match contract", i, j)

			expectedPayload := []byte(expectedKey.Payload)
			require.Equal(t, expectedPayload, apiKey.GetPayload(),
				"Validator %d key %d payload should match contract", i, j)
		}

		require.Len(t, apiValidator.GetVaults(), len(expectedValidator.Vaults),
			"Validator %d should have same number of vaults as contract", i)

		for k, expectedVault := range expectedValidator.Vaults {
			require.Less(t, k, len(apiValidator.GetVaults()), "API should have vault %d for validator %d", k, i)
			apiVault := apiValidator.GetVaults()[k]

			require.Equal(t, expectedVault.ChainID, apiVault.GetChainId(),
				"Validator %d vault %d chain ID should match contract", i, k)
			require.Equal(t, expectedVault.Vault.Hex(), apiVault.GetVault(),
				"Validator %d vault %d address should match contract", i, k)
			require.Equal(t, expectedVault.VotingPower.String(), apiVault.GetVotingPower(),
				"Validator %d vault %d voting power should match contract", i, k)
		}
	}

	t.Logf("Validator set validation passed: %d validators, epoch %d, quorum %s",
		len(apiResponse.Validators), apiResponse.Epoch, apiResponse.QuorumThreshold)
}

// TestRelayAPIConnectivity tests that all relay servers are accessible via gRPC
func TestRelayAPIConnectivity(t *testing.T) {
	t.Log("Starting relay API connectivity test...")

	for i := range globalTestEnv.SidecarConfigs {
		t.Run(fmt.Sprintf("Connect_%s", globalTestEnv.GetContainerPort(i)), func(t *testing.T) {
			t.Logf("Testing connection to %d", i)

			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()

			client := globalTestEnv.GetGRPCClient(t, i)
			epochResp, err := client.GetCurrentEpoch(ctx, &apiv1.GetCurrentEpochRequest{})
			require.NoErrorf(t, err, "Failed to get current epoch from %d", i)
			require.NotNil(t, epochResp.GetStartTime(), "Epoch start time should be set")

			t.Logf("Successfully connected to %d - Current epoch: %d", i, epochResp.GetEpoch())
		})
	}

	t.Log("Relay API connectivity test completed successfully")
}

// TestValidatorSetAPI tests the GetValidatorSet API endpoint
func TestValidatorSetAPI(t *testing.T) {
	t.Log("Starting validator set API test...")

	deploymentData, err := loadDeploymentData(t.Context())
	require.NoError(t, err, "Failed to load deployment data")

	client := globalTestEnv.GetGRPCClient(t, 0)

	const retryAttempts = 4
	for i := 0; i < retryAttempts; i++ {
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		valsetResp, err := client.GetValidatorSet(ctx, &apiv1.GetValidatorSetRequest{})
		cancel()

		if err != nil {
			t.Logf("Attempt %d: Failed to get validator set from %d: %v", i+1, i, err)
			if i == retryAttempts-1 {
				t.Fatalf("Failed to get validator set after %d attempts: %v", retryAttempts, err)
			}
			time.Sleep(5 * time.Second)
			continue
		}

		t.Logf("Performing contract validation...")
		expected := getExpectedDataFromContracts(t, deploymentData)

		if expected.ValidatorSet.Epoch != entity.Epoch(valsetResp.GetEpoch()) {
			continue
		}

		validateValidatorSetAgainstExpected(t, valsetResp, expected)
		t.Logf("✓ Contract validation passed")

		require.Positive(t, valsetResp.GetVersion(), "Validator set version should be greater than 0")
		require.NotNil(t, valsetResp.GetCaptureTimestamp(), "Capture timestamp should be set")
		require.NotEmpty(t, valsetResp.GetQuorumThreshold(), "Quorum threshold should not be empty")
		require.NotEmpty(t, valsetResp.GetValidators(), "Validator set should contain validators")

		t.Logf("Validator set from %d: Epoch %d, %d validators, Version %d",
			i, valsetResp.GetEpoch(), len(valsetResp.GetValidators()), valsetResp.GetVersion())

		for i, validator := range valsetResp.GetValidators() {
			require.NotEmpty(t, validator.GetOperator(), "Validator %d operator should not be empty", i)
			require.NotEmpty(t, validator.GetVotingPower(), "Validator %d voting power should not be empty", i)
			require.NotEmpty(t, validator.GetKeys(), "Validator %d should have keys", i)

			votingPower, err := new(big.Int).SetString(validator.GetVotingPower(), 10)
			require.True(t, err, "Validator %d voting power should be a valid big integer", i)
			require.Positive(t, votingPower.Sign(), "Validator %d voting power should be positive", i)

			for j, key := range validator.GetKeys() {
				require.NotEmpty(t, key.GetPayload(), "Validator %d key %d payload should not be empty", i, j)
				require.Positive(t, key.GetTag(), "Validator %d key %d tag should be greater than 0", i, j)
			}

			for k, vault := range validator.GetVaults() {
				require.Positive(t, vault.GetChainId(), "Validator %d vault %d chain ID should be greater than 0", i, k)
				require.True(t, common.IsHexAddress(vault.GetVault()), "Validator %d vault %d should be valid address", i, k)
				require.NotEmpty(t, vault.GetVotingPower(), "Validator %d vault %d voting power should not be empty", i, k)
			}

			t.Logf("  Validator %d: Operator %s, Voting Power %s, Active: %v, Keys: %d, Vaults: %d",
				i, validator.GetOperator(), validator.GetVotingPower(), validator.GetIsActive(), len(validator.GetKeys()), len(validator.GetVaults()))
		}

		if valsetResp.GetEpoch() > 0 {
			ctx, cancel = context.WithTimeout(t.Context(), 5*time.Second)
			specificEpochResp, err := client.GetValidatorSet(ctx, &apiv1.GetValidatorSetRequest{
				Epoch: &valsetResp.Epoch,
			})
			cancel()

			if err == nil {
				require.Equal(t, valsetResp.GetEpoch(), specificEpochResp.GetEpoch(), "Epoch should match")
				require.Len(t, specificEpochResp.GetValidators(), len(valsetResp.GetValidators()), "Number of validators should match")
				require.Equal(t, valsetResp.GetQuorumThreshold(), specificEpochResp.GetQuorumThreshold(), "Quorum threshold should match")
			}
		}
		return
	}

	t.Fatalf("Validator set API test failed to verify after %d attempts", retryAttempts)
}
