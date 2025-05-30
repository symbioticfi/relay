package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-errors/errors"
	"github.com/spf13/cobra"

	"middleware-offchain/internal/client/eth"
	"middleware-offchain/internal/entity"
	valsetDeriver "middleware-offchain/internal/uc/valset-deriver"
	"middleware-offchain/pkg/log"
)

// generate_genesis --master-address 0x04C89607413713Ec9775E14b954286519d836FEf --rpc-url http://127.0.0.1:8545
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

		client, err := eth.NewEthClient(eth.Config{
			MasterRPCURL:   cfg.rpcURL,
			MasterAddress:  cfg.masterAddress,
			RequestTimeout: time.Second * 5,
		})
		if err != nil {
			return errors.Errorf("failed to create eth client: %w", err)
		}

		deriver, err := valsetDeriver.NewDeriver(client)
		if err != nil {
			return errors.Errorf("failed to create valset deriver: %w", err)
		}

		currentOnchainEpoch, err := client.GetCurrentEpoch(ctx)
		if err != nil {
			return errors.Errorf("failed to get current epoch: %w", err)
		}

		newValset, err := deriver.GetValidatorSetExtraForEpoch(ctx, currentOnchainEpoch)
		if err != nil {
			return errors.Errorf("failed to get validator set extra for epoch %s: %w", currentOnchainEpoch, err)
		}

		header, err := deriver.MakeValsetHeader(ctx, newValset)
		if err != nil {
			return errors.Errorf("failed to generate validator set header: %w", err)
		}

		slog.Info("Valset header generated!")

		jsonData, err := valsetHeaderMarshalJSON(header)
		if err != nil {
			return errors.Errorf("failed to marshal header to JSON: %w", err)
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

func valsetHeaderMarshalJSON(v entity.ValidatorSetHeader) ([]byte, error) {
	// Convert byte arrays to hex strings before JSON marshaling
	type key struct {
		Tag     uint8  `json:"tag"`
		Payload string `json:"payload"` // hex string
	}
	type jsonHeader struct {
		Version                uint8    `json:"version"`
		ActiveAggregatedKeys   []key    `json:"activeAggregatedKeys"`
		ValidatorsSszMRoot     string   `json:"validatorsSszMRoot"` // hex string
		ExtraData              string   `json:"extraData"`
		TotalActiveVotingPower *big.Int `json:"totalActiveVotingPower"`
	}

	jsonHeaderData := jsonHeader{
		Version:                v.Version,
		ActiveAggregatedKeys:   make([]key, len(v.ActiveAggregatedKeys)),
		ValidatorsSszMRoot:     fmt.Sprintf("0x%064x", v.ValidatorsSszMRoot),
		ExtraData:              fmt.Sprintf("0x%064x", v.ExtraData),
		TotalActiveVotingPower: v.TotalActiveVotingPower,
	}

	for i, key := range v.ActiveAggregatedKeys {
		jsonHeaderData.ActiveAggregatedKeys[i].Tag = key.Tag
		jsonHeaderData.ActiveAggregatedKeys[i].Payload = fmt.Sprintf("0x%0128x", key.Payload)
	}

	jsonData, err := json.MarshalIndent(jsonHeaderData, "", "  ")
	if err != nil {
		return nil, errors.Errorf("failed to marshal header to JSON: %w", err)
	}

	return jsonData, nil
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
