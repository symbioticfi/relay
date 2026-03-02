package network

import (
	"log/slog"
	"time"

	cmdhelpers "github.com/symbioticfi/relay/cmd/utils/cmd-helpers"
	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	"github.com/symbioticfi/relay/internal/usecase/metrics"
	"github.com/symbioticfi/relay/symbiotic/client/evm"
	"github.com/symbioticfi/relay/symbiotic/client/votingpower"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	valsetDeriver "github.com/symbioticfi/relay/symbiotic/usecase/valset-deriver"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/pterm/pterm"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
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
			DriverAddress: symbiotic.CrossChainAddress{
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

		deriver, err := valsetDeriver.NewDeriver(evmClient, externalVPClient)
		if err != nil {
			return errors.Errorf("failed to create deriver: %w", err)
		}

		epoch := symbiotic.Epoch(globalFlags.Epoch)
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

		networkConfig, err := evmClient.GetConfig(ctx, captureTimestamp, epoch)
		if err != nil {
			return errors.Errorf("Failed to get config: %w", err)
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
					printNetworkInfo(epoch, captureTimestamp, &networkConfig, &valset),
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
					printAddresses(symbiotic.CrossChainAddress{
						ChainId: globalFlags.DriverChainId,
						Address: common.HexToAddress(globalFlags.DriverAddress),
					}, &networkConfig),
				)},
			})
		}

		// row with settlements info
		if infoFlags.Settlement {
			settlementData := make([]settlementReplicaData, len(networkConfig.Settlements))

			eg, egCtx := errgroup.WithContext(ctx)
			eg.SetLimit(5)
			for i, settlement := range networkConfig.Settlements {
				eg.Go(func() error {
					isCommitted, err := evmClient.IsValsetHeaderCommittedAt(egCtx, settlement, epoch)
					if err != nil {
						return errors.Errorf("Failed to get latest epoch: %w", err)
					}
					settlementData[i].IsCommitted = isCommitted

					if isCommitted {
						headerHash, err := evmClient.GetHeaderHashAt(egCtx, settlement, epoch)
						if err != nil {
							return errors.Errorf("Failed to get header hash: %w", err)
						}
						settlementData[i].HeaderHash = headerHash
					}

					lastCommittedHeaderEpoch, err := evmClient.GetLastCommittedHeaderEpoch(ctx, settlement)
					if err != nil {
						return errors.Errorf("Failed to get last committed header epoch: %w", err)
					}
					settlementData[i].LastCommittedHeaderEpoch = uint64(lastCommittedHeaderEpoch)

					allEpochsFromZero := lo.RepeatBy(int(epoch+1), func(i int) symbiotic.Epoch {
						return symbiotic.Epoch(i)
					})

					commitmentResults, err := evmClient.IsValsetHeaderCommittedAtEpochs(egCtx, settlement, allEpochsFromZero)
					if err != nil {
						return errors.Errorf("Failed to check epoch commitments: %w", err)
					}

					settlementData[i].MissedEpochs = uint64(lo.CountBy(commitmentResults, func(committed bool) bool { return !committed }))

					return nil
				})
			}

			if err := eg.Wait(); err != nil {
				return err
			}
			header, err := valset.GetHeader()
			if err != nil {
				return errors.Errorf("Failed to get header: %w", err)
			}
			panels = append(panels, []pterm.Panel{
				{Data: pterm.DefaultBox.WithTitle("Settlement").Sprint(
					printSettlementData(header, networkConfig, settlementData),
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
