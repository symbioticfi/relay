package network

import (
	"time"

	"github.com/symbioticfi/relay/core/client/evm"
	"github.com/symbioticfi/relay/core/entity"
	keyprovider "github.com/symbioticfi/relay/core/usecase/key-provider"
	valsetDeriver "github.com/symbioticfi/relay/core/usecase/valset-deriver"
	"github.com/symbioticfi/relay/internal/usecase/metrics"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print network information",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		ctx := signalContext(cmd.Context())

		kp, err := keyprovider.NewSimpleKeystoreProvider()
		if err != nil {
			return err
		}

		evmClient, err := evm.NewEvmClient(ctx, evm.Config{
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

		if err != nil {
			return errors.Errorf("Failed to get evm client: %v", err)
		}
		deriver, err := valsetDeriver.NewDeriver(evmClient)
		if err != nil {
			return errors.Errorf("Failed to create deriver: %v", err)
		}

		epoch := globalFlags.Epoch
		if globalFlags.Epoch == 0 {
			epoch, err = evmClient.GetCurrentEpoch(ctx)
			if err != nil {
				return errors.Errorf("Failed to get current epoch: %w", err)
			}
		}

		captureTimestamp, err := evmClient.GetEpochStart(ctx, epoch)
		if err != nil {
			return errors.Errorf("Failed to get capture timestamp: %w", err)
		}

		networkConfig, err := evmClient.GetConfig(ctx, captureTimestamp)
		if err != nil {
			return errors.Errorf("Failed to get config: %w", err)
		}

		_, committedEpoch, err := deriver.GetLastCommittedHeaderEpoch(ctx, networkConfig)
		if err != nil {
			return errors.Errorf("Failed to get valset header: %w", err)
		}

		epochDuration, err := evmClient.GetEpochDuration(ctx, epoch)
		if err != nil {
			return errors.Errorf("Failed to get epoch duration: %w", err)
		}

		valset, err := deriver.GetValidatorSet(ctx, epoch, networkConfig)
		if err != nil {
			return errors.Errorf("Failed to get validator set: %w", err)
		}

		// row with info and config
		panels := pterm.Panels{
			{
				{Data: pterm.DefaultBox.WithTitle("Network info").Sprint(
					printNetworkInfo(epoch, committedEpoch, captureTimestamp, &networkConfig, &valset),
				)},
				{Data: pterm.DefaultBox.WithTitle("Network config").Sprint(
					printNetworkConfig(epochDuration, &networkConfig),
				)},
			},
		}

		// row with addresses [optional]
		if infoFlags.Addresses {
			panels = append(panels, []pterm.Panel{
				{Data: pterm.DefaultBox.WithTitle("Addresses").Sprint(
					printAddresses(entity.CrossChainAddress{
						ChainId: globalFlags.DriverChainId,
						Address: common.HexToAddress(globalFlags.DriverAddress),
					}, &networkConfig),
				)},
			})
		}

		// row with settlements info
		if infoFlags.Settlement {
			settlementData := make([]SettlementReplicaData, len(networkConfig.Replicas))
			for i, replica := range networkConfig.Replicas {
				settlementData[i].IsCommitted, err = evmClient.IsValsetHeaderCommittedAt(ctx, replica, epoch)
				if err != nil {
					return errors.Errorf("Failed to get latest epoch: %w", err)
				}
				if settlementData[i].IsCommitted {
					settlementData[i].HeaderHash, err = evmClient.GetHeaderHashAt(ctx, replica, epoch)
					if err != nil {
						return errors.Errorf("Failed to get header hash: %w", err)
					}
				}
			}
			header, err := valset.GetHeader()
			if err != nil {
				return errors.Errorf("Failed to get header: %w", err)
			}
			panels = append(panels, []pterm.Panel{
				{Data: pterm.DefaultBox.WithTitle("Settlement").Sprint(
					printSettlementData(header, networkConfig, settlementData, committedEpoch),
				)},
			})
		}

		// row with validators [optional]
		if infoFlags.ValidatorsFull {
			panels = append(panels, []pterm.Panel{
				{Data: pterm.DefaultBox.WithTitle("Validators").Sprint(
					printValidatorsTree(&valset),
				)},
			})
		} else if infoFlags.Validators {
			panels = append(panels, []pterm.Panel{
				{Data: pterm.DefaultBox.WithTitle("Validators").Sprint(
					printValidatorsTable(&valset),
				)},
			})
		}

		pterm.DefaultPanel.WithPanels(panels).Render()

		return nil
	},
}
