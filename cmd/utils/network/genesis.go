package network

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	cmdhelpers "github.com/symbioticfi/relay/cmd/utils/cmd-helpers"
	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	"github.com/symbioticfi/relay/symbiotic/client/evm"
	"github.com/symbioticfi/relay/symbiotic/client/votingpower"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/aggregator"
	symbioticCrypto "github.com/symbioticfi/relay/symbiotic/usecase/crypto"
	valsetDeriver "github.com/symbioticfi/relay/symbiotic/usecase/valset-deriver"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var genesisCmd = &cobra.Command{
	Use:   "generate-genesis",
	Short: "Generate genesis validator set header",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := signalContext(cmd.Context())

		kp, err := keyprovider.NewSimpleKeystoreProvider()
		if err != nil {
			return err
		}

		evmClient, err := evm.NewEvmClient(ctx, evm.Config{
			ChainURLs: globalFlags.Chains,
			DriverAddress: symbiotic.CrossChainAddress{
				ChainId: globalFlags.DriverChainId,
				Address: common.HexToAddress(globalFlags.DriverAddress),
			},
			RequestTimeout: 5 * time.Second,
			KeyProvider:    kp,
		})
		if err != nil {
			return err
		}

		if genesisFlags.Commit {
			privateKeyInput := pterm.DefaultInteractiveTextInput.WithMask("*")
			for _, chainId := range evmClient.GetChains() {
				secret, ok := genesisFlags.Secrets.Secrets[chainId]
				if !ok {
					secret, _ = privateKeyInput.Show("Enter private key for chain with ID: " + strconv.Itoa(int(chainId)))
				}
				pk, err := symbioticCrypto.NewPrivateKey(symbiotic.KeyTypeEcdsaSecp256k1, common.FromHex(secret))
				if err != nil {
					return err
				}
				err = kp.AddKeyByNamespaceTypeId(
					keyprovider.EVM_KEY_NAMESPACE,
					symbiotic.KeyTypeEcdsaSecp256k1,
					int(chainId),
					pk,
				)
				if err != nil {
					return err
				}
			}
		}

		providerConfigs, err := cmdhelpers.ExternalVotingPowerProviderConfigs(globalFlags.ExternalVotingPowerProviders)
		if err != nil {
			return err
		}

		var externalVPClient *votingpower.Client
		if len(providerConfigs) > 0 {
			externalVPClient, err = votingpower.NewClient(ctx, providerConfigs)
			if err != nil {
				return errors.Errorf("failed to create external voting power client: %w", err)
			}
			defer func() {
				if err := externalVPClient.Close(); err != nil {
					slog.WarnContext(ctx, "Failed to close external voting power client", "error", err)
				}
			}()
		}

		spinner := getSpinner("Fetching on-chain network config...")

		deriver, err := valsetDeriver.NewDeriver(evmClient, externalVPClient)
		if err != nil {
			return errors.Errorf("failed to create deriver: %w", err)
		}

		currentOnchainEpoch := symbiotic.Epoch(0)
		if genesisFlags.Epoch >= 0 {
			currentOnchainEpoch = symbiotic.Epoch(genesisFlags.Epoch)
		} else {
			currentOnchainEpoch, err = evmClient.GetCurrentEpoch(ctx)
			if err != nil {
				return errors.Errorf("failed to get current epoch: %w", err)
			}
		}

		captureTimestamp, err := evmClient.GetEpochStart(ctx, currentOnchainEpoch)
		if err != nil {
			return errors.Errorf("failed to get capture timestamp: %w", err)
		}

		networkConfig, err := evmClient.GetConfig(ctx, captureTimestamp, currentOnchainEpoch)
		if err != nil {
			return errors.Errorf("failed to get config: %w", err)
		}
		spinner.Success()

		spinner = getSpinner("Fetching on-chain validators data...")

		newValset, err := deriver.GetValidatorSet(ctx, currentOnchainEpoch, networkConfig)
		if err != nil {
			return errors.Errorf("failed to get validator set extra for epoch %d: %w", currentOnchainEpoch, err)
		}

		spinner.Success()

		spinner = getSpinner("Building header and extra data...")

		// header generation is clear now
		header, err := newValset.GetHeader()
		if err != nil {
			return errors.Errorf("failed to generate validator set header: %w", err)
		}

		aggregator, err := aggregator.NewAggregator(networkConfig.VerificationType, nil)
		if err != nil {
			return errors.Errorf("failed to create aggregator: %w", err)
		}

		// extra data generation is also clear but still in deriver
		extraData, err := aggregator.GenerateExtraData(ctx, newValset, networkConfig.RequiredKeyTags)
		if err != nil {
			return errors.Errorf("failed to generate extra data: %w", err)
		}

		spinner.Success()

		jsonData := printHeaderWithExtraDataToJSON(header, extraData)

		if !genesisFlags.Json {
			panels := pterm.Panels{
				{
					{Data: pterm.DefaultBox.WithTitle("Validator Set Header").Sprint(
						printHeaderTable(header),
					)},
				},
				{
					{Data: pterm.DefaultBox.WithTitle("Extra Data").Sprint(
						printExtraDataTable(extraData),
					)},
				},
			}

			pterm.DefaultPanel.WithPanels(panels).Render()
		} else {
			pterm.Println(jsonData)
		}

		if genesisFlags.Output != "" {
			err = os.WriteFile(genesisFlags.Output, []byte(jsonData), 0600)
			if err != nil {
				return errors.Errorf("failed to write output file: %w", err)
			}
			pterm.Success.Println("Genesis data written to " + genesisFlags.Output)
		}

		if genesisFlags.Commit {
			for _, settlement := range networkConfig.Settlements {
				spinner = getSpinner("Setting genesis on " + settlement.Address.String())
				txResult, err := evmClient.SetGenesis(
					ctx,
					settlement,
					header,
					extraData)
				if err != nil {
					spinner.Fail("Transaction failed: ", err)
					return errors.Errorf("failed to set genesis for network %d: %w", settlement.ChainId, err)
				}
				spinner.Success("Transaction hash: ", txResult.TxHash.String())
			}
		}

		return nil
	},
}

func getSpinner(text string) *pterm.SpinnerPrinter {
	spinner, _ := pterm.DefaultSpinner.
		WithTimerRoundingFactor(time.Millisecond).
		WithWriter(os.Stderr).
		WithDelay(time.Millisecond * 100).
		Start(text)
	return spinner
}
