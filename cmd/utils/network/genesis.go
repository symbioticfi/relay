package network

import (
	"middleware-offchain/core/client/evm"
	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/aggregator"
	symbioticCrypto "middleware-offchain/core/usecase/crypto"
	keyprovider "middleware-offchain/core/usecase/key-provider"
	valsetDeriver "middleware-offchain/core/usecase/valset-deriver"
	"middleware-offchain/internal/usecase/metrics"
	"os"
	"strconv"
	"time"

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

		client, err := evm.NewEVMClient(ctx, evm.Config{
			ChainURLs: globalFlags.Chains,
			DriverAddress: entity.CrossChainAddress{
				ChainId: globalFlags.DriverChainId,
				Address: common.HexToAddress(globalFlags.DriverAddress),
			},
			RequestTimeout: 5 * time.Second,
			KeyProvider:    kp,
			Metrics:        metrics.New(metrics.Config{}),
		})
		if err != nil {
			return err
		}

		if genesisFlags.Commit {
			privateKeyInput := pterm.DefaultInteractiveTextInput.WithMask("*")
			for _, chainId := range client.GetChains() {
				secret, ok := genesisFlags.Secrets.Secrets[chainId]
				if !ok {
					secret, _ = privateKeyInput.Show("Enter private key for chain with ID: " + strconv.Itoa(int(chainId)))
				}
				pk, err := symbioticCrypto.NewPrivateKey(entity.KeyTypeEcdsaSecp256k1, common.Hex2Bytes(secret))
				if err != nil {
					return err
				}
				err = kp.AddKeyByNamespaceTypeId(
					keyprovider.EVM_KEY_NAMESPACE,
					entity.KeyTypeEcdsaSecp256k1,
					int(chainId),
					pk,
				)
				if err != nil {
					return err
				}
			}
		}

		spinner := getSpinner("Fetching on-chain network config...")

		deriver, err := valsetDeriver.NewDeriver(client)
		if err != nil {
			return errors.Errorf("Failed to create deriver: %v", err)
		}

		currentOnchainEpoch, err := client.GetCurrentEpoch(ctx)
		if err != nil {
			return errors.Errorf("Failed to get current epoch: %w", err)
		}

		captureTimestamp, err := client.GetEpochStart(ctx, currentOnchainEpoch)
		if err != nil {
			return errors.Errorf("Failed to get capture timestamp: %w", err)
		}

		networkConfig, err := client.GetConfig(ctx, captureTimestamp)
		if err != nil {
			return errors.Errorf("Failed to get config: %w", err)
		}
		spinner.Success()

		spinner = getSpinner("Fetching on-chain validators data...")

		newValset, err := deriver.GetValidatorSet(ctx, currentOnchainEpoch, networkConfig)
		if err != nil {
			return errors.Errorf("Failed to get validator set extra for epoch %d: %w", currentOnchainEpoch, err)
		}

		spinner.Success()

		spinner = getSpinner("Building header and extra data...")

		// header generation is clear now
		header, err := newValset.GetHeader()
		if err != nil {
			return errors.Errorf("Failed to generate validator set header: %w", err)
		}

		aggregator := aggregator.NewAggregator(nil)

		// extra data generation is also clear but still in deriver
		extraData, err := aggregator.GenerateExtraData(newValset, networkConfig)
		if err != nil {
			return errors.Errorf("Failed to generate extra data: %w", err)
		}

		spinner.Success()

		jsonData := printHeaderWithExtraDataToJSON(header, extraData)

		if !genesisFlags.Json {
			panels := pterm.Panels{
				{
					{Data: pterm.DefaultBox.WithTitle("Validator Set Header").Sprint(
						printHeaderTable(&header),
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
				return errors.Errorf("Failed to write output file: %w", err)
			}
			pterm.Success.Println("Genesis data written to " + genesisFlags.Output)
		}

		if genesisFlags.Commit {
			for _, replica := range networkConfig.Replicas {
				spinner = getSpinner("Setting genesis on " + replica.Address.String())
				txResult, err := client.SetGenesis(
					ctx,
					replica,
					header,
					extraData)
				if err != nil {
					spinner.Fail("Transaction failed: ", err)
				} else {
					spinner.Success("Transaction hash: ", txResult.TxHash.String())
				}
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
