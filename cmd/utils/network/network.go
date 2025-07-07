package network

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"middleware-offchain/core/client/evm"
	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/aggregator"
	valsetDeriver "middleware-offchain/core/usecase/valset-deriver"
	utils_app "middleware-offchain/internal/usecase/utils-app"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
)

type config struct {
	epoch         uint64
	rpcURL        string
	driverAddress string
	compact       bool
	commit        bool
	secretKey     string
	outputFile    string

	driverCrossChainAddress entity.CrossChainAddress
	client                  *evm.Client
	deriver                 *valsetDeriver.Deriver
}

var cfg config

func NewNetworkCmd() (*cobra.Command, error) {
	networkCmd.PersistentFlags().StringVar(&cfg.rpcURL, "rpc-url", "", "RPC URL")
	networkCmd.PersistentFlags().StringVar(&cfg.driverAddress, "driver-address", "", "Driver contract address")
	if err := networkCmd.MarkPersistentFlagRequired("rpc-url"); err != nil {
		return nil, errors.Errorf("failed to mark rpc-url as required: %w", err)
	}
	if err := networkCmd.MarkPersistentFlagRequired("driver-address"); err != nil {
		return nil, errors.Errorf("failed to mark driver-address as required: %w", err)
	}

	infoCmd.PersistentFlags().Uint64Var(&cfg.epoch, "epoch", 0, "Network epoch")

	valsetCmd.PersistentFlags().BoolVar(&cfg.compact, "compact", false, "Compact valset print")

	genesisCmd.PersistentFlags().BoolVar(&cfg.commit, "commit", false, "Commit genesis flag (default: false)")
	genesisCmd.PersistentFlags().StringVar(&cfg.secretKey, "secret-key", "", "Secret key for genesis commit")
	genesisCmd.PersistentFlags().StringVarP(&cfg.outputFile, "output", "o", "", "Output file path (default: stdout)")

	networkCmd.AddCommand(infoCmd)
	networkCmd.AddCommand(valsetCmd)
	networkCmd.AddCommand(genesisCmd)

	return networkCmd, nil
}

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Network tool",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx := signalContext(context.Background())

		var err error
		var privateKey []byte

		if cfg.secretKey != "" {
			b, ok := new(big.Int).SetString(cfg.secretKey, 10)
			if !ok {
				return errors.Errorf("failed to parse secret key as big.Int")
			}

			privateKey = b.FillBytes(make([]byte, 32))
		}

		cfg.driverCrossChainAddress = entity.CrossChainAddress{ChainId: 111, Address: common.HexToAddress(cfg.driverAddress)}
		cfg.client, err = evm.NewEVMClient(ctx, evm.Config{
			Chains: []entity.ChainURL{{
				ChainID: 111,
				RPCURL:  cfg.rpcURL,
			}},
			DriverAddress:  cfg.driverCrossChainAddress,
			RequestTimeout: time.Second * 5,
			PrivateKey:     privateKey,
		})
		if err != nil {
			return errors.Errorf("failed to create symbiotic client: %w", err)
		}

		cfg.deriver, err = valsetDeriver.NewDeriver(cfg.client)
		if err != nil {
			return errors.Errorf("failed to create valset deriver: %w", err)
		}
		return nil
	},
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print network information",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		ctx := signalContext(context.Background())

		if cfg.epoch == 0 {
			cfg.epoch, err = cfg.client.GetCurrentEpoch(ctx)
			if err != nil {
				return errors.Errorf("failed to get current epoch: %w", err)
			}
		}

		captureTimestamp, err := cfg.client.GetEpochStart(ctx, cfg.epoch)
		if err != nil {
			return errors.Errorf("failed to get capture timestamp: %w", err)
		}

		networkConfig, err := cfg.client.GetConfig(ctx, captureTimestamp)
		if err != nil {
			return errors.Errorf("failed to get config: %w", err)
		}

		epoch, err := cfg.client.GetLastCommittedHeaderEpoch(ctx, networkConfig.Replicas[0])
		if err != nil {
			return errors.Errorf("failed to get valset header: %w", err)
		}

		valset, err := cfg.deriver.GetValidatorSet(ctx, epoch, networkConfig)
		if err != nil {
			return errors.Errorf("failed to get validator set: %w", err)
		}

		slog.InfoContext(ctx, "Network Info")
		slog.InfoContext(ctx, "-", "epoch", cfg.epoch)
		slog.InfoContext(ctx, "-", "active_operators", valset.GetTotalActiveValidators())
		slog.InfoContext(ctx, "-", "total_voting_power", valset.GetTotalActiveVotingPower())
		slog.InfoContext(ctx, "- config")

		slog.InfoContext(ctx, "    - voting power providers")
		for _, addr := range networkConfig.VotingPowerProviders {
			slog.InfoContext(ctx, "        -", "addr", addr.Address, "chain_id", addr.ChainId)
		}

		slog.InfoContext(ctx, "    -", "keys_provider", networkConfig.KeysProvider)
		slog.InfoContext(ctx, "    - replicas")
		for _, addr := range networkConfig.Replicas {
			slog.InfoContext(ctx, "        -", "addr", addr.Address, "chain_id", addr.ChainId)
		}

		slog.InfoContext(ctx, "    -", "verification_type", networkConfig.VerificationType)
		slog.InfoContext(ctx, "    -", "max_voting_power", networkConfig.MaxVotingPower)
		slog.InfoContext(ctx, "    -", "min_inclusion_voting_power", networkConfig.MinInclusionVotingPower)
		slog.InfoContext(ctx, "    -", "max_validators_count", networkConfig.MaxValidatorsCount)

		slog.InfoContext(ctx, "    - key tags")
		for i, tag := range networkConfig.RequiredKeyTags {
			bytes, err := tag.MarshalText()
			if err != nil {
				return errors.Errorf("failed to format network config: %w", err)
			}
			slog.InfoContext(ctx, "        -", strconv.Itoa(i), string(bytes))
		}

		bytes, err := networkConfig.RequiredHeaderKeyTag.MarshalText()
		if err != nil {
			return errors.Errorf("failed to format network config: %w", err)
		}
		slog.InfoContext(ctx, "    -", "required_header_key_tag", string(bytes))

		slog.InfoContext(ctx, "    - quorum thresholds")
		for _, t := range networkConfig.QuorumThresholds {
			bytes, err = t.KeyTag.MarshalText()
			if err != nil {
				return errors.Errorf("failed to format network config: %w", err)
			}

			slog.InfoContext(ctx, "        -", "key_tag", string(bytes), "value", t.QuorumThreshold)
		}

		return nil
	},
}

var valsetCmd = &cobra.Command{
	Use:   "valset",
	Short: "Print validator set information",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		ctx := signalContext(context.Background())

		if cfg.epoch == 0 {
			cfg.epoch, err = cfg.client.GetCurrentEpoch(ctx)
			if err != nil {
				return errors.Errorf("failed to get current epoch: %w", err)
			}
		}

		captureTimestamp, err := cfg.client.GetEpochStart(ctx, cfg.epoch)
		if err != nil {
			return errors.Errorf("failed to get capture timestamp: %w", err)
		}

		networkConfig, err := cfg.client.GetConfig(ctx, captureTimestamp)
		if err != nil {
			return errors.Errorf("failed to get config: %w", err)
		}

		epoch, err := cfg.client.GetLastCommittedHeaderEpoch(ctx, networkConfig.Replicas[0])
		if err != nil {
			return errors.Errorf("failed to get valset header: %w", err)
		}

		valset, err := cfg.deriver.GetValidatorSet(ctx, epoch, networkConfig)
		if err != nil {
			return errors.Errorf("failed to get validator set: %w", err)
		}

		slog.InfoContext(ctx, "Validators Info")
		slog.InfoContext(ctx, "-", "current_epoch", cfg.epoch)
		if cfg.epoch != epoch {
			slog.InfoContext(ctx, "-", "valset_committed_epoch", epoch)
		}
		slog.InfoContext(ctx, "-", "operators", len(valset.Validators))
		slog.InfoContext(ctx, "-", "total_voting_power", valset.GetTotalActiveVotingPower())

		for _, validator := range valset.Validators {
			if err := utils_app.LogValidator(ctx, validator, cfg.compact); err != nil {
				return errors.Errorf("failed to log validator: %w", err)
			}
		}

		return nil
	},
}

var genesisCmd = &cobra.Command{
	Use:   "generate_genesis",
	Short: "Generate genesis validator set header",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.commit && cfg.secretKey == "" {
			return errors.New("if commit true secret-key must be set")
		}

		ctx := signalContext(context.Background())

		currentOnchainEpoch, err := cfg.client.GetCurrentEpoch(ctx)
		if err != nil {
			return errors.Errorf("failed to get current epoch: %w", err)
		}

		captureTimestamp, err := cfg.client.GetEpochStart(ctx, currentOnchainEpoch)
		if err != nil {
			return errors.Errorf("failed to get capture timestamp: %w", err)
		}

		networkConfig, err := cfg.client.GetConfig(ctx, captureTimestamp)
		if err != nil {
			return errors.Errorf("failed to get config: %w", err)
		}

		newValset, err := cfg.deriver.GetValidatorSet(ctx, currentOnchainEpoch, networkConfig)
		if err != nil {
			return errors.Errorf("failed to get validator set extra for epoch %d: %w", currentOnchainEpoch, err)
		}

		// header generation is clear now
		header, err := newValset.GetHeader()
		if err != nil {
			return errors.Errorf("failed to generate validator set header: %w", err)
		}

		slog.Info("Valset header generated!")

		aggregator := aggregator.NewAggregator(nil)

		// extra data generation is also clear but still in deriver
		extraData, err := aggregator.GenerateExtraData(newValset, networkConfig)
		if err != nil {
			return errors.Errorf("failed to generate extra data: %w", err)
		}

		jsonData, err := EncodeValidatorSetHeaderWithExtraDataToJSON(header, extraData)
		if err != nil {
			return errors.Errorf("failed to encode validator set header with extra data to JSON: %w", err)
		}

		if cfg.outputFile != "" {
			err = os.WriteFile(cfg.outputFile, jsonData, 0600)
			if err != nil {
				return errors.Errorf("failed to write output file: %w", err)
			}
		} else {
			fmt.Println(string(jsonData)) //nolint:forbidigo // ok to print result to stdout
		}

		if !cfg.commit {
			return nil
		}

		errs := make([]error, len(networkConfig.Replicas))
		for i, replica := range networkConfig.Replicas {
			var txResult entity.TxResult
			txResult, errs[i] = cfg.client.SetGenesis(ctx, cfg.driverCrossChainAddress, header, extraData)
			if errs[i] != nil {
				slog.ErrorContext(ctx, "failed to set genesis on replica", "replica", replica, "error", errs[i])
			} else {
				slog.InfoContext(ctx, "genesis valset set on replica", "replica", replica, "txHash", txResult.TxHash.String())
			}
		}
		if err := errors.Join(errs...); err != nil {
			return errors.Errorf("failed to commit valset header: %w", err)
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
		slog.Info("received signal", "signal", sig)
		cancel()
	}()

	return cnCtx
}
