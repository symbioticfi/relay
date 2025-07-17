package operator

import (
	"middleware-offchain/core/client/evm"
	"middleware-offchain/core/entity"
	keyprovider "middleware-offchain/core/usecase/key-provider"
	valsetDeriver "middleware-offchain/core/usecase/valset-deriver"
	cmdhelpers "middleware-offchain/internal/usecase/cmd-helpers"
	"middleware-offchain/internal/usecase/metrics"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print operator information",
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

		if infoFlags.Password == "" {
			infoFlags.Password, err = cmdhelpers.GetPassword()
			if err != nil {
				return errors.Errorf("failed to get password: %w", err)
			}
		}

		if infoFlags.Epoch == 0 {
			infoFlags.Epoch, err = client.GetCurrentEpoch(ctx)
			if err != nil {
				return errors.Errorf("failed to get current epoch: %w", err)
			}
		}

		captureTimestamp, err := client.GetEpochStart(ctx, infoFlags.Epoch)
		if err != nil {
			return errors.Errorf("failed to get capture timestamp: %w", err)
		}

		networkConfig, err := client.GetConfig(ctx, captureTimestamp)
		if err != nil {
			return errors.Errorf("failed to get config: %w", err)
		}

		epoch, err := client.GetLastCommittedHeaderEpoch(ctx, networkConfig.Replicas[0])
		if err != nil {
			return errors.Errorf("failed to get valset header: %w", err)
		}

		deriver, err := valsetDeriver.NewDeriver(client)
		if err != nil {
			return errors.Errorf("failed to create valset deriver: %w", err)
		}

		valset, err := deriver.GetValidatorSet(ctx, epoch, networkConfig)
		if err != nil {
			return errors.Errorf("failed to get validator set: %w", err)
		}

		keyStore, err := keyprovider.NewKeystoreProvider(infoFlags.Path, infoFlags.Password)
		if err != nil {
			return err
		}

		kt := entity.KeyTag(infoFlags.KeyTag)
		pk, err := keyStore.GetPrivateKey(kt)
		if err != nil {
			return err
		}

		validator, found := valset.FindValidatorByKey(kt, pk.PublicKey().Raw())
		if !found {
			return errors.Errorf("validator not found for key: %d %s", kt, common.Bytes2Hex(pk.PublicKey().Raw()))
		}

		leveledList := pterm.LeveledList{}
		leveledList = cmdhelpers.PrintTreeValidator(leveledList, validator, valset.GetTotalActiveVotingPower().Int)
		text, _ := pterm.DefaultTree.WithRoot(putils.TreeFromLeveledList(leveledList)).Srender()
		panels := pterm.Panels{{{Data: pterm.DefaultBox.WithTitle("Operator info").Sprint(text)}}}
		pterm.DefaultPanel.WithPanels(panels).Render()

		return nil
	},
}
