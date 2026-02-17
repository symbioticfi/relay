package tests

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"

	evmtest "github.com/symbioticfi/relay/e2e/tests/evm"
	"github.com/symbioticfi/relay/symbiotic/client/evm"
	votingpowerclient "github.com/symbioticfi/relay/symbiotic/client/votingpower"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	valsetDeriver "github.com/symbioticfi/relay/symbiotic/usecase/valset-deriver"
)

const externalVotingPowerBonus = int64(100)
const externalVotingPowerChainID = votingpowerclient.ExternalVotingPowerChainIDMin

var externalProviderID = [10]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00}

func TestExternalVotingPowerProvider_AddsBonusForExistingOperators(t *testing.T) {
	t.Log("Starting external voting power provider test...")

	ctx := t.Context()

	deployData := loadDeploymentData(t)
	adminEVM := createTestEVM(t, settlementChains[0], getFunderPrivateKey(t))
	evmClient := createEVMClient(t, deployData)
	driverAddress := common.HexToAddress(deployData.GetDriverAddress())
	require.NoError(t, cleanupExternalVotingPowerProviders(ctx, evmClient, adminEVM, driverAddress))

	currentEpoch, err := evmClient.GetCurrentEpoch(ctx)
	require.NoError(t, err, "Failed to get current epoch")
	currentEpochStart, err := evmClient.GetEpochStart(ctx, currentEpoch)
	require.NoError(t, err, "Failed to get epoch start")
	currentCfg, err := evmClient.GetConfig(ctx, currentEpochStart, currentEpoch)
	require.NoError(t, err, "Failed to get current config")

	externalProvider := symbiotic.CrossChainAddress{
		ChainId: externalVotingPowerChainID,
		Address: providerAddressFromID(externalProviderID),
	}
	require.True(t, votingpowerclient.IsExternalVotingPowerChainID(externalProvider.ChainId))

	_, err = adminEVM.AddVotingPowerProvider(ctx, driverAddress, externalProvider.ChainId, externalProvider.Address)
	require.NoError(t, err, "Failed to add external voting power provider")
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		require.NoError(t, cleanupExternalVotingPowerProviders(cleanupCtx, evmClient, adminEVM, driverAddress))
	})

	targetEpoch := currentEpoch
	cfgWithoutExternal := currentCfg
	cfgWithoutExternal.VotingPowerProviders = cfgWithoutExternal.VotingPowerProviders[:0]
	for _, provider := range currentCfg.VotingPowerProviders {
		if provider.ChainId == 0 && provider.Address == (common.Address{}) {
			continue
		}
		if votingpowerclient.IsExternalVotingPowerChainID(provider.ChainId) {
			continue
		}
		cfgWithoutExternal.VotingPowerProviders = append(cfgWithoutExternal.VotingPowerProviders, provider)
	}

	cfgForExternalBonus := cfgWithoutExternal
	cfgForExternalBonus.VotingPowerProviders = append(cfgForExternalBonus.VotingPowerProviders, externalProvider)

	baselineDeriver, err := valsetDeriver.NewDeriver(evmClient, nil)
	require.NoError(t, err, "Failed to create baseline valset deriver")
	baselineValset, err := baselineDeriver.GetValidatorSet(ctx, targetEpoch, cfgWithoutExternal)
	require.NoError(t, err, "Failed to derive baseline validator set")
	require.NotEmpty(t, baselineValset.Validators)

	existingOperators := make([]common.Address, 0, len(baselineValset.Validators))
	for _, validator := range baselineValset.Validators {
		existingOperators = append(existingOperators, validator.Operator)
	}

	externalURL := startBonusVotingPowerServer(t, existingOperators, externalVotingPowerBonus)
	externalClient, err := votingpowerclient.NewClient(ctx, []votingpowerclient.ProviderConfig{{
		ID:  providerIDToHex(externalProviderID),
		URL: externalURL,
	}})
	require.NoError(t, err, "Failed to create external voting power client")
	t.Cleanup(func() {
		require.NoError(t, externalClient.Close())
	})

	deriverWithExternal, err := valsetDeriver.NewDeriver(evmClient, externalClient)
	require.NoError(t, err, "Failed to create valset deriver with external client")
	withExternalValset, err := deriverWithExternal.GetValidatorSet(ctx, targetEpoch, cfgForExternalBonus)
	require.NoError(t, err, "Failed to derive validator set with external provider")

	baselineByOperator := make(map[common.Address]*big.Int, len(baselineValset.Validators))
	withExternalByOperator := make(map[common.Address]*big.Int, len(withExternalValset.Validators))
	for _, validator := range baselineValset.Validators {
		baselineByOperator[validator.Operator] = new(big.Int).Set(validator.VotingPower.Int)
	}
	for _, validator := range withExternalValset.Validators {
		withExternalByOperator[validator.Operator] = new(big.Int).Set(validator.VotingPower.Int)
	}

	for operator, baselineVP := range baselineByOperator {
		withVP, ok := withExternalByOperator[operator]
		require.True(t, ok, "operator %s missing in valset with external provider", operator.Hex())
		expected := new(big.Int).Add(new(big.Int).Set(baselineVP), big.NewInt(externalVotingPowerBonus))
		require.Zero(t, expected.Cmp(withVP), "unexpected voting power for operator %s", operator.Hex())
	}

	_, err = adminEVM.RemoveVotingPowerProvider(ctx, driverAddress, externalProvider.ChainId, externalProvider.Address)
	require.NoError(t, err, "Failed to remove external voting power provider")
	require.NoError(t, ensureExternalProviderAbsent(ctx, evmClient, externalProvider, currentEpoch))

	t.Log("External voting power provider test completed successfully")
}

func providerIDToHex(id [10]byte) string {
	return "0x" + hex.EncodeToString(id[:])
}

func providerAddressFromID(id [10]byte) common.Address {
	var addr common.Address
	copy(addr[:10], id[:])
	return addr
}

func cleanupExternalVotingPowerProviders(
	ctx context.Context,
	evmClient *evm.Client,
	adminEVM *evmtest.Client,
	driverAddress common.Address,
) error {
	for iteration := 0; iteration < 3; iteration++ {
		currentEpoch, err := evmClient.GetCurrentEpoch(ctx)
		if err != nil {
			return err
		}

		providersToRemove := map[string]symbiotic.CrossChainAddress{}
		targetEpochs := []symbiotic.Epoch{currentEpoch, currentEpoch + 1}
		for _, epoch := range targetEpochs {
			epochStart, epochErr := evmClient.GetEpochStart(ctx, epoch)
			if epochErr != nil {
				continue
			}

			cfg, cfgErr := evmClient.GetConfig(ctx, epochStart, epoch)
			if cfgErr != nil {
				continue
			}

			for _, provider := range cfg.VotingPowerProviders {
				if !votingpowerclient.IsExternalVotingPowerChainID(provider.ChainId) {
					continue
				}
				key := fmt.Sprintf("%d:%s", provider.ChainId, provider.Address.Hex())
				providersToRemove[key] = provider
			}
		}

		if len(providersToRemove) == 0 {
			break
		}

		for _, provider := range providersToRemove {
			_, removeErr := adminEVM.RemoveVotingPowerProvider(ctx, driverAddress, provider.ChainId, provider.Address)
			if removeErr != nil && !strings.Contains(removeErr.Error(), "ValSetDriver_NotAdded") {
				return removeErr
			}
		}

		nextEpoch := currentEpoch + 1
		if err := waitForEpoch(ctx, evmClient, nextEpoch, waitEpochTimeout); err != nil {
			return err
		}
	}

	return nil
}

func ensureExternalProviderAbsent(
	ctx context.Context,
	evmClient *evm.Client,
	externalProvider symbiotic.CrossChainAddress,
	currentEpoch symbiotic.Epoch,
) error {
	for epoch := currentEpoch + 1; epoch <= currentEpoch+8; epoch++ {
		if err := waitForEpoch(ctx, evmClient, epoch, waitEpochTimeout); err != nil {
			return err
		}

		epochStart, err := evmClient.GetEpochStart(ctx, epoch)
		if err != nil {
			return err
		}

		cfg, err := evmClient.GetConfig(ctx, epochStart, epoch)
		if err != nil {
			return err
		}

		for _, provider := range cfg.VotingPowerProviders {
			if provider.ChainId == externalProvider.ChainId && provider.Address == externalProvider.Address {
				return errors.Errorf("external provider still present in epoch %d config", epoch)
			}
		}
	}

	return nil
}
