package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-errors/errors"
	"github.com/spf13/cobra"

	"middleware-offchain/internal/client/symbiotic"
	"middleware-offchain/internal/entity"
	valsetDeriver "middleware-offchain/internal/usecase/valset-deriver"
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
	rootCmd.PersistentFlags().StringVar(&cfg.masterAddress, "master-address", "", "Master contract address")
	rootCmd.PersistentFlags().StringVarP(&cfg.outputFile, "output", "o", "", "Output file path (default: stdout)")
	rootCmd.PersistentFlags().StringVar(&cfg.logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&cfg.logMode, "log-mode", "text", "Log mode (text, pretty)")

	if err := rootCmd.MarkPersistentFlagRequired("rpc-url"); err != nil {
		return errors.Errorf("failed to mark rpc-url as required: %w", err)
	}
	if err := rootCmd.MarkPersistentFlagRequired("master-address"); err != nil {
		return errors.Errorf("failed to mark master-address as required: %w", err)
	}

	return rootCmd.Execute()
}

type config struct {
	rpcURL        string
	masterAddress string
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

		client, err := symbiotic.NewEVMClient(symbiotic.Config{
			MasterRPCURL:   cfg.rpcURL,
			MasterAddress:  cfg.masterAddress,
			RequestTimeout: time.Second * 5,
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

		captureTimestamp, err := client.GetCaptureTimestamp(ctx)
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

		// extra data generation is also clear but still in deriver
		extraData, err := deriver.GenerateExtraData(newValset, networkConfig)
		if err != nil {
			return errors.Errorf("failed to generate extra data: %w", err)
		}

		jsonData, err := entity.ValidatorSetHeaderWithExtraData{
			ValidatorSetHeader: header,
			ExtraDataList:      extraData,
		}.EncodeJSON()
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
