package network

import (
	"context"
	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/aggregator"
	valsetDeriver "middleware-offchain/core/usecase/valset-deriver"
	utils_app "middleware-offchain/internal/usecase/utils-app"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-errors/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func NewNetworkCmd() *cobra.Command {
	networkCmd.AddCommand(infoCmd)
	networkCmd.AddCommand(genesisCmd)

	addFlags()

	return networkCmd
}

var networkCmd = &cobra.Command{
	Use:               "network",
	Short:             "Network tool",
	PersistentPreRunE: initConfig,
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print network information",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		ctx := signalContext(cmd.Context())
		cfg := cfgFromCtx(ctx)
		client, err := utils_app.GetEvmClient(ctx, cfg.SecretKey, cfg.Driver, cfg.ChainsId, cfg.ChainsUrl)
		if err != nil {
			return errors.Errorf("Failed to get evm client: %v", err)
		}
		deriver, err := valsetDeriver.NewDeriver(client)
		if err != nil {
			return errors.Errorf("Failed to create deriver: %v", err)
		}

		if cfg.Epoch == 0 {
			cfg.Epoch, err = client.GetCurrentEpoch(ctx)
			if err != nil {
				return errors.Errorf("Failed to get current epoch: %w", err)
			}
		}

		captureTimestamp, err := client.GetEpochStart(ctx, cfg.Epoch)
		if err != nil {
			return errors.Errorf("Failed to get capture timestamp: %w", err)
		}

		networkConfig, err := client.GetConfig(ctx, captureTimestamp)
		if err != nil {
			return errors.Errorf("Failed to get config: %w", err)
		}

		_, committedEpoch, err := deriver.GetLastCommittedHeaderEpoch(ctx, networkConfig)
		if err != nil {
			return errors.Errorf("Failed to get valset header: %w", err)
		}

		epochDuration, err := client.GetEpochDuration(ctx, cfg.Epoch)
		if err != nil {
			return errors.Errorf("Failed to get epoch duration: %w", err)
		}

		valset, err := deriver.GetValidatorSet(ctx, cfg.Epoch, networkConfig)
		if err != nil {
			return errors.Errorf("Failed to get validator set: %w", err)
		}

		// first row with info and config
		panels := pterm.Panels{
			{
				{Data: pterm.DefaultBox.WithTitle("Network info").Sprint(
					printNetworkInfo(cfg.Epoch, committedEpoch, captureTimestamp, &networkConfig, &valset),
				)},
				{Data: pterm.DefaultBox.WithTitle("Network config").Sprint(
					printNetworkConfig(epochDuration, &networkConfig),
				)},
			},
		}

		// second row with addresses [optional]
		if cfg.Addresses {
			panels = append(panels, []pterm.Panel{
				{Data: pterm.DefaultBox.WithTitle("Addresses").Sprint(
					printAddresses(cfg.Driver, &networkConfig),
				)},
			})
		}

		// third row with validators [optional]
		if cfg.ValidatorsFull {
			panels = append(panels, []pterm.Panel{
				{Data: pterm.DefaultBox.WithTitle("Validators").Sprint(
					printValidatorsTree(&valset),
				)},
			})
		} else if cfg.Validators {
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

var genesisCmd = &cobra.Command{
	Use:   "generate-genesis",
	Short: "Generate genesis validator set header",
	RunE: func(cmd *cobra.Command, args []string) error {
		spinner := getSpinner("Fetching on-chain network config...")

		cfg := cfgFromCtx(signalContext(cmd.Context()))

		if cfg.Commit && cfg.SecretKey == "" {
			return errors.New("If commit true secret-key must be set")
		}

		ctx := signalContext(context.Background())
		client, err := utils_app.GetEvmClient(ctx, cfg.SecretKey, cfg.Driver, cfg.ChainsId, cfg.ChainsUrl)
		if err != nil {
			return errors.Errorf("Failed to get evm client: %v", err)
		}
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

		aggregator, err := aggregator.NewAggregator(entity.VerificationTypeSimple, nil)
		if err != nil {
			return errors.Errorf("Failed to create aggregator: %w", err)
		}

		// extra data generation is also clear but still in deriver
		extraData, err := aggregator.GenerateExtraData(newValset, networkConfig)
		if err != nil {
			return errors.Errorf("Failed to generate extra data: %w", err)
		}

		spinner.Success()

		jsonData := printHeaderWithExtraDataToJSON(header, extraData)

		if !cfg.Json {
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

		if cfg.OutputFile != "" {
			err = os.WriteFile(cfg.OutputFile, []byte(jsonData), 0600)
			if err != nil {
				return errors.Errorf("Failed to write output file: %w", err)
			}
			pterm.Success.Println("Genesis data written to " + cfg.OutputFile)
		}

		if cfg.Commit {
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

// signalContext returns a context that is canceled if either SIGTERM or SIGINT signal is received.
func signalContext(ctx context.Context) context.Context {
	cnCtx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-c
		pterm.Warning.Println("Received termination signal, shutting down...")
		cancel()
	}()

	return cnCtx
}
