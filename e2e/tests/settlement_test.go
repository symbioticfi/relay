package tests

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/symbioticfi/relay/api/client/v1"
	"github.com/symbioticfi/relay/symbiotic/entity"
	valsetDeriver "github.com/symbioticfi/relay/symbiotic/usecase/valset-deriver"
)

// TestRemoveSettlement verifies the settlement lifecycle management:
// 1. Removes a settlement from the network configuration
// 2. Waits for the next epoch and confirms removal took effect (2 → 1 settlement)
// 3. Re-adds the settlement back to the configuration
// 4. Sets genesis for the re-added settlement chain
// 5. Waits for another epoch and confirms both settlements are active and committed (1 → 2 settlements)
//
// This test validates that the relay network can dynamically handle settlement chain
// configuration changes across epoch boundaries and properly commits validator sets
// to all active settlement chains.
func TestRemoveAndAddSettlement(t *testing.T) {
	t.Log("Starting TestRemoveSettlement - testing settlement lifecycle management")

	deploymentData := loadDeploymentData(t)
	evmClient := createEVMClient(t, deploymentData)
	deriver, err := valsetDeriver.NewDeriver(evmClient, nil)
	require.NoError(t, err)

	var currentEpoch entity.Epoch
	var currentConfig entity.NetworkConfig
	t.Run("remove settlement", func(t *testing.T) {
		t.Log("Step 1: Removing settlement from network configuration")

		currentEpoch, err = evmClient.GetCurrentEpoch(t.Context())
		require.NoError(t, err, "Failed to get current epoch")
		t.Logf("Current epoch: %d", currentEpoch)

		captureTimestamp, err := evmClient.GetEpochStart(t.Context(), currentEpoch)
		require.NoError(t, err, "Failed to get epoch start timestamp")

		currentConfig, err = evmClient.GetConfig(t.Context(), captureTimestamp, currentEpoch)
		require.NoError(t, err, "Failed to get network config")

		require.Lenf(t, currentConfig.Settlements, 2, "Expected exactly two settlement")
		t.Logf("Initial settlements count: %d", len(currentConfig.Settlements))
		t.Logf("Settlement to remove - ChainID: %d, Address: %s",
			currentConfig.Settlements[0].ChainId,
			currentConfig.Settlements[0].Address.Hex())

		txResult, err := evmClient.RemoveSettlement(t.Context(), currentConfig.Settlements[0])
		require.NoError(t, err, "Failed to remove settlement")

		t.Logf("Successfully removed settlement. Tx hash: %s", txResult.TxHash.Hex())
	})

	oneSettlementEpoch := currentEpoch + 2
	t.Logf("Will wait for epoch %d to verify removal", oneSettlementEpoch)

	var oneSettlementConfig entity.NetworkConfig
	t.Run("add settlement back", func(t *testing.T) {
		t.Log("Step 2: Verifying settlement removal and re-adding settlement")

		err = waitForEpoch(t.Context(), evmClient, oneSettlementEpoch, time.Minute*4)
		require.NoError(t, err, "Failed to wait for next epoch")
		t.Logf("Reached epoch %d", oneSettlementEpoch)

		nextCaptureTimestamp, err := evmClient.GetEpochStart(t.Context(), oneSettlementEpoch)
		require.NoError(t, err, "Failed to get epoch start timestamp")
		t.Logf("Epoch %d start timestamp: %d", oneSettlementEpoch, nextCaptureTimestamp)

		oneSettlementConfig, err = evmClient.GetConfig(t.Context(), nextCaptureTimestamp, oneSettlementEpoch)
		require.NoError(t, err, "Failed to get network config")

		require.Len(t, oneSettlementConfig.Settlements, 1, "Expected exactly one settlement after removal")
		t.Logf("Settlement removal confirmed in epoch %d - settlements count: %d", oneSettlementEpoch, len(oneSettlementConfig.Settlements))
		t.Logf("Remaining settlement - ChainID: %d, Address: %s",
			oneSettlementConfig.Settlements[0].ChainId,
			oneSettlementConfig.Settlements[0].Address.Hex())

		t.Logf("Re-adding settlement - ChainID: %d, Address: %s",
			currentConfig.Settlements[0].ChainId,
			currentConfig.Settlements[0].Address.Hex())
		txResult, err := evmClient.AddSettlement(t.Context(), currentConfig.Settlements[0])
		require.NoError(t, err, "Failed to add settlement back")
		t.Logf("Successfully added settlement back. Tx hash: %s", txResult.TxHash.Hex())
	})

	t.Run("set genesis with re-added settlement", func(t *testing.T) {
		t.Log("Step 3: Setting genesis for re-added settlement chain")

		// Query the last committed epoch on the unmodified settlement to handle edge cases
		// where the addition happens slightly delayed and oneSettlementEpoch+1 might've been committed
		unmodifiedSettlement := oneSettlementConfig.Settlements[0]
		lastCommittedEpoch, err := evmClient.GetLastCommittedHeaderEpoch(t.Context(), unmodifiedSettlement)
		require.NoError(t, err, "Failed to get last committed epoch on unmodified settlement")
		t.Logf("Last committed epoch on unmodified settlement (ChainID: %d): %d", unmodifiedSettlement.ChainId, lastCommittedEpoch)

		genesisEpoch := lastCommittedEpoch + 1
		t.Logf("Using epoch %d for genesis (last committed + 1)", genesisEpoch)

		client := getGRPCClient(t, 0)
		t.Log("Waiting for validator set metadata to be available")
		var nextMetadata *apiv1.GetValidatorSetMetadataResponse
		err = waitForErrorIsNil(t.Context(), time.Second*30, func() error {
			nextMetadata, err = client.GetValidatorSetMetadata(t.Context(), &apiv1.GetValidatorSetMetadataRequest{
				Epoch: (*uint64)(&genesisEpoch),
			})
			return err
		})
		require.NoError(t, err)
		t.Logf("Retrieved validator set metadata for epoch %d with %d extra data entries", genesisEpoch, len(nextMetadata.ExtraData))

		// Get config for the genesis epoch
		genesisCaptureTimestamp, err := evmClient.GetEpochStart(t.Context(), genesisEpoch)
		require.NoError(t, err, "Failed to get epoch start timestamp for genesis epoch")

		genesisConfig, err := evmClient.GetConfig(t.Context(), genesisCaptureTimestamp, genesisEpoch)
		require.NoError(t, err, "Failed to get network config for genesis epoch")

		extraData := lo.Map(nextMetadata.ExtraData, func(item *apiv1.ExtraData, index int) entity.ExtraData {
			return entity.ExtraData{
				Key:   common.BytesToHash(item.Key),
				Value: common.BytesToHash(item.Value),
			}
		})

		t.Logf("Deriving validator set for epoch %d", genesisEpoch)
		newValset, err := deriver.GetValidatorSet(t.Context(), genesisEpoch, genesisConfig)
		require.NoError(t, err)
		header, err := newValset.GetHeader()
		require.NoError(t, err)

		t.Logf("Setting genesis for re-added settlement - ChainID: %d", currentConfig.Settlements[0].ChainId)
		txResult, err := evmClient.SetGenesis(t.Context(), currentConfig.Settlements[0], header, extraData)
		require.NoError(t, err)

		t.Logf("Successfully set genesis with re-added settlement. Tx hash: %s", txResult.TxHash.Hex())
	})

	backTwoSettlementsEpoch := oneSettlementEpoch + 1
	t.Logf("Will wait for epoch %d to verify settlement re-addition and commitment", backTwoSettlementsEpoch)

	t.Run("final check", func(t *testing.T) {
		t.Log("Step 4: Verifying settlement re-addition and commitment to both chains")

		err = waitForEpoch(t.Context(), evmClient, backTwoSettlementsEpoch, waitEpochTimeout)
		require.NoError(t, err, "Failed to wait for epoch after next")
		t.Logf("Reached epoch %d", backTwoSettlementsEpoch)

		finalCaptureTimestamp, err := evmClient.GetEpochStart(t.Context(), backTwoSettlementsEpoch)
		require.NoError(t, err, "Failed to get epoch start timestamp")

		finalConfig, err := evmClient.GetConfig(t.Context(), finalCaptureTimestamp, backTwoSettlementsEpoch)
		require.NoError(t, err, "Failed to get network config")

		require.Len(t, finalConfig.Settlements, 2, "Expected exactly two settlements after re-adding")
		t.Logf("Settlement re-addition confirmed in epoch %d - settlements count: %d", backTwoSettlementsEpoch, len(finalConfig.Settlements))
		for i, settlement := range finalConfig.Settlements {
			t.Logf("Settlement %d - ChainID: %d, Address: %s", i, settlement.ChainId, settlement.Address.Hex())
		}

		t.Logf("Waiting for validator set commitments on both settlements")
		require.NoError(t, waitForErrorIsNil(t.Context(), time.Minute*10, func() error {
			allCommitted := true
			for _, settlement := range finalConfig.Settlements {
				lastCommitted, err := evmClient.GetLastCommittedHeaderEpoch(t.Context(), settlement)
				if err != nil {
					t.Logf("Error getting last committed epoch on settlement ChainID %d: %v", settlement.ChainId, err)
					return err
				}
				t.Logf("Last committed epoch on settlement ChainID %d: %d", settlement.ChainId, lastCommitted)

				committed, err := evmClient.IsValsetHeaderCommittedAt(t.Context(), settlement, backTwoSettlementsEpoch)
				if err != nil {
					t.Logf("Error checking valset header commitment on settlement ChainID %d: %v", settlement.ChainId, err)
					return err
				}
				if !committed {
					t.Logf("Valset header not yet committed on settlement ChainID %d for epoch %d", settlement.ChainId, backTwoSettlementsEpoch)
					allCommitted = false
				}
			}

			if !allCommitted {
				t.Logf("Not all settlements have committed the validator set for epoch %d yet", backTwoSettlementsEpoch)
				return errors.New("not all settlements committed")
			}
			return nil
		}))
		t.Logf("Successfully verified validator set committed on all %d settlements for epoch %d", len(finalConfig.Settlements), backTwoSettlementsEpoch)
	})

	t.Log("TestRemoveSettlement completed successfully")
}
