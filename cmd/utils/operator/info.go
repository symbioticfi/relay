package operator

import (
	"time"

	cmdhelpers "github.com/symbioticfi/relay/internal/usecase/cmd-helpers"
	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	"github.com/symbioticfi/relay/internal/usecase/metrics"
	"github.com/symbioticfi/relay/symbiotic/client/evm"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	valsetDeriver "github.com/symbioticfi/relay/symbiotic/usecase/valset-deriver"

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

		if infoFlags.Password == "" {
			infoFlags.Password, err = cmdhelpers.GetPassword()
			if err != nil {
				return errors.Errorf("failed to get password: %w", err)
			}
		}

		if infoFlags.Epoch == 0 {
			epoch, err := evmClient.GetCurrentEpoch(ctx)
			if err != nil {
				return errors.Errorf("failed to get current epoch: %w", err)
			}
			infoFlags.Epoch = uint64(epoch)
		}

		captureTimestamp, err := evmClient.GetEpochStart(ctx, symbiotic.Epoch(infoFlags.Epoch))
		if err != nil {
			return errors.Errorf("failed to get capture timestamp: %w", err)
		}

		networkConfig, err := evmClient.GetConfig(ctx, captureTimestamp, symbiotic.Epoch(infoFlags.Epoch))
		if err != nil {
			return errors.Errorf("failed to get config: %w", err)
		}

		epoch, err := evmClient.GetLastCommittedHeaderEpoch(ctx, networkConfig.Settlements[0])
		if err != nil {
			return errors.Errorf("failed to get valset header: %w", err)
		}

		deriver, err := valsetDeriver.NewDeriver(evmClient)
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

		kt := symbiotic.KeyTag(infoFlags.KeyTag)
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
