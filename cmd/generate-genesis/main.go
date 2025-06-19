package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"

	"middleware-offchain/core/client/evm"
	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/aggregator"
	valsetDeriver "middleware-offchain/core/usecase/valset-deriver"
	"middleware-offchain/pkg/log"
)

// generate_genesis --master-address 0x1f5fE7682E49c20289C20a4cFc8b45d5EB410690 --rpc-url http://127.0.0.1:8545
func main() {
	slog.Info("Running generate_genesis command", "args", os.Args)

	if err := run(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("error executing command", "error", err)
		os.Exit(1)
	}
	slog.Info("Generate genesis completed successfully")
}

func run() error {
	rootCmd.PersistentFlags().StringVar(&cfg.rpcURL, "rpc-url", "", "RPC URL")
	rootCmd.PersistentFlags().StringVar(&cfg.driverAddress, "driver-address", "", "Driver contract address")
	rootCmd.PersistentFlags().BoolVar(&cfg.commit, "commit", false, "Commit genesis flag (default: false)")
	rootCmd.PersistentFlags().StringVar(&cfg.secretKey, "secret-key", "", "Secret key for genesis commit")
	rootCmd.PersistentFlags().StringVarP(&cfg.outputFile, "output", "o", "", "Output file path (default: stdout)")
	rootCmd.PersistentFlags().StringVar(&cfg.logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&cfg.logMode, "log-mode", "text", "Log mode (text, pretty)")

	if err := rootCmd.MarkPersistentFlagRequired("rpc-url"); err != nil {
		return errors.Errorf("failed to mark rpc-url as required: %w", err)
	}
	if err := rootCmd.MarkPersistentFlagRequired("driver-address"); err != nil {
		return errors.Errorf("failed to mark driver-address as required: %w", err)
	}

	return rootCmd.Execute()
}

type config struct {
	rpcURL        string
	driverAddress string
	commit        bool
	secretKey     string
	outputFile    string
	logLevel      string
	logMode       string
}

var cfg config

var rootCmd = &cobra.Command{
	Use:   "generate_genesis",
	Short: "Generate genesis validator set header",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Init(cfg.logLevel, cfg.logMode)

		ctx := signalContext(context.Background())

		if cfg.commit && cfg.secretKey == "" {
			return errors.New("if commit true secret-key must be set")
		}

		var privateKey []byte

		if cfg.secretKey != "" {
			b, ok := new(big.Int).SetString(cfg.secretKey, 10)
			if !ok {
				return errors.Errorf("failed to parse secret key as big.Int")
			}

			privateKey = b.FillBytes(make([]byte, 32))
		}

		driverAddress := entity.CrossChainAddress{ChainId: 111, Address: common.HexToAddress(cfg.driverAddress)}
		client, err := evm.NewEVMClient(ctx, evm.Config{
			Chains: []entity.ChainURL{{
				ChainID: 111,
				RPCURL:  cfg.rpcURL,
			}},
			DriverAddress:  driverAddress,
			RequestTimeout: time.Second * 5,
			PrivateKey:     privateKey,
		})
		if err != nil {
			return errors.Errorf("failed to create symbiotic client: %w", err)
		}

		deriver, err := valsetDeriver.NewDeriver(client)
		if err != nil {
			return errors.Errorf("failed to create valset deriver: %w", err)
		}

		currentOnchainEpoch, err := client.GetCurrentEpoch(ctx)
		if err != nil {
			return errors.Errorf("failed to get current epoch: %w", err)
		}

		captureTimestamp, err := client.GetEpochStart(ctx, currentOnchainEpoch)
		if err != nil {
			return errors.Errorf("failed to get capture timestamp: %w", err)
		}

		networkConfig, err := client.GetConfig(ctx, captureTimestamp)
		if err != nil {
			return errors.Errorf("failed to get config: %w", err)
		}

		newValset, err := deriver.GetValidatorSet(ctx, currentOnchainEpoch, networkConfig)
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

		result, err := client.SetGenesis(ctx, driverAddress, header, extraData)
		if err != nil {
			return errors.Errorf("failed to commit valset header: %w", err)
		}

		slog.InfoContext(ctx, "genesis valset committed", "txHash", result.TxHash.String(), "epoch", newValset.Epoch)

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
