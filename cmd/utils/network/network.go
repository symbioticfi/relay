package network

import (
	"context"
	"fmt"
	"log/slog"
	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/aggregator"
	valsetDeriver "middleware-offchain/core/usecase/valset-deriver"
	utils_app "middleware-offchain/internal/usecase/utils-app"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
)

func NewNetworkCmd() *cobra.Command {
	networkCmd.AddCommand(infoCmd)
	networkCmd.AddCommand(valsetCmd)
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

		_, epoch, err := deriver.GetLastCommittedHeaderEpoch(ctx, networkConfig)
		if err != nil {
			return errors.Errorf("Failed to get valset header: %w", err)
		}

		valset, err := deriver.GetValidatorSet(ctx, epoch, networkConfig)
		if err != nil {
			return errors.Errorf("Failed to get validator set: %w", err)
		}

		fmt.Printf("\nNetwork Info:\n")
		fmt.Printf("   Epoch: %v\n", cfg.Epoch)
		fmt.Printf("   Operators: %v\n", valset.GetTotalActiveValidators())
		fmt.Printf("   Voting Power: %v\n", valset.GetTotalActiveVotingPower())

		fmt.Printf("\nVoting Power Providers (%d):\n", len(networkConfig.VotingPowerProviders))
		fmt.Printf("   # | Address | Chain ID\n")
		for i, addr := range networkConfig.VotingPowerProviders {
			fmt.Printf("   %d | %s | %d\n", i+1, addr.Address, addr.ChainId)
		}

		fmt.Printf("\nKeys Provider: %v\n", networkConfig.KeysProvider)

		fmt.Printf("\nReplicas (%d):\n", len(networkConfig.Replicas))
		fmt.Printf("   # | Address | Chain ID\n")
		for i, addr := range networkConfig.Replicas {
			fmt.Printf("   %d | %s | %d\n", i+1, addr.Address, addr.ChainId)
		}

		verificationType, err := networkConfig.VerificationType.MarshalText()
		if err != nil {
			return errors.Errorf("Failed to marshal verification type: %w", err)
		}

		fmt.Printf("\nConfig:\n")
		fmt.Printf("   Verification Type: %s\n", verificationType)
		fmt.Printf("   Max Voting Power: %v\n", networkConfig.MaxVotingPower)
		fmt.Printf("   Min Inclusion Voting Power: %v\n", networkConfig.MinInclusionVotingPower)
		fmt.Printf("   Max Validators Count: %v\n", networkConfig.MaxValidatorsCount)

		fmt.Printf("\nKey Tags (%d):\n", len(networkConfig.RequiredKeyTags))
		fmt.Printf("   # | Tag\n")
		for i, tag := range networkConfig.RequiredKeyTags {
			bytes, err := tag.MarshalText()
			if err != nil {
				return errors.Errorf("Failed to format network config: %w", err)
			}

			fmt.Printf("   %d | %s\n", i+1, string(bytes))
		}

		bytes, err := networkConfig.RequiredHeaderKeyTag.MarshalText()
		if err != nil {
			return errors.Errorf("Failed to format network config: %w", err)
		}

		fmt.Printf("\nHeader Key Tag: %s\n", string(bytes))

		fmt.Printf("\nQuorum Thresholds (%d):\n", len(networkConfig.QuorumThresholds))
		fmt.Printf("   # | Key Tag | Threshold\n")
		for i, t := range networkConfig.QuorumThresholds {
			bytes, err = t.KeyTag.MarshalText()
			if err != nil {
				return errors.Errorf("Failed to format network config: %w", err)
			}

			fmt.Printf("   %d | %s | %v\n", i+1, string(bytes), t.QuorumThreshold)
		}

		return nil
	},
}

var valsetCmd = &cobra.Command{
	Use:   "valset",
	Short: "Print validator set information",
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

		_, epoch, err := deriver.GetLastCommittedHeaderEpoch(ctx, networkConfig)
		if err != nil {
			return errors.Errorf("Failed to get valset header: %w", err)
		}

		valset, err := deriver.GetValidatorSet(ctx, epoch, networkConfig)
		if err != nil {
			return errors.Errorf("Failed to get validator set: %w", err)
		}

		fmt.Printf("\nValidators Info:\n")
		fmt.Printf("   Current Epoch: %v\n", cfg.Epoch)
		if cfg.Epoch != epoch {
			fmt.Printf("   Valset Committed Epoch: %v\n", epoch)
		}
		fmt.Printf("   Operators: %d\n", len(valset.Validators))
		fmt.Printf("   Total Voting Power: %v\n", valset.GetTotalActiveVotingPower())

		for _, validator := range valset.Validators {
			str, err := utils_app.MarshalTextValidator(validator, cfg.Compact)
			if err != nil {
				return errors.Errorf("Failed to log validator: %w", err)
			}
			fmt.Print(str)
		}

		return nil
	},
}

var genesisCmd = &cobra.Command{
	Use:   "generate-genesis",
	Short: "Generate genesis validator set header",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		newValset, err := deriver.GetValidatorSet(ctx, currentOnchainEpoch, networkConfig)
		if err != nil {
			return errors.Errorf("Failed to get validator set extra for epoch %d: %w", currentOnchainEpoch, err)
		}

		// header generation is clear now
		header, err := newValset.GetHeader()
		if err != nil {
			return errors.Errorf("Failed to generate validator set header: %w", err)
		}

		slog.Info("Valset header generated!")

		aggregator := aggregator.NewAggregator(nil)

		// extra data generation is also clear but still in deriver
		extraData, err := aggregator.GenerateExtraData(newValset, networkConfig)
		if err != nil {
			return errors.Errorf("Failed to generate extra data: %w", err)
		}

		jsonData, err := EncodeValidatorSetHeaderWithExtraDataToJSON(header, extraData)
		if err != nil {
			return errors.Errorf("Failed to encode validator set header with extra data to JSON: %w", err)
		}

		if cfg.OutputFile != "" {
			err = os.WriteFile(cfg.OutputFile, jsonData, 0600)
			if err != nil {
				return errors.Errorf("Failed to write output file: %w", err)
			}
		} else {
			fmt.Println(string(jsonData)) //nolint:forbidigo // ok to print result to stdout
		}

		if !cfg.Commit {
			return nil
		}

		errs := make([]error, len(networkConfig.Replicas))
		for i, replica := range networkConfig.Replicas {
			var txResult entity.TxResult
			txResult, errs[i] = client.SetGenesis(
				ctx,
				entity.CrossChainAddress{
					ChainId: cfg.Driver.ChainID,
					Address: common.HexToAddress(cfg.Driver.Address),
				},
				header,
				extraData)
			if errs[i] != nil {
				slog.ErrorContext(ctx, "Failed to set genesis on replica", "replica", replica, "error", errs[i])
			} else {
				slog.InfoContext(ctx, "Genesis valset set on replica", "replica", replica, "txHash", txResult.TxHash.String())
			}
		}
		if err := errors.Join(errs...); err != nil {
			return errors.Errorf("Failed to commit valset header: %w", err)
		}

		return nil
	},
}

// signalContext returns a context that is canceled if either SIGTERM or SIGINT signal is received.
func signalContext(ctx context.Context) context.Context {
	cnCtx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-c
		slog.Info("Received signal", "signal", sig)
		cancel()
	}()

	return cnCtx
}
