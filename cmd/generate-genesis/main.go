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

	"middleware-offchain/internal/client/eth"
	"middleware-offchain/pkg/log"
	"middleware-offchain/valset"
)

// generate_genesis --master-address 0x5081a39b8A5f0E35a8D959395a630b68B74Dd30f --rpc-url http://127.0.0.1:8545
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
}

var cfg config

var rootCmd = &cobra.Command{
	Use:   "generate_genesis",
	Short: "Generate genesis validator set header",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Init(cfg.logLevel)

		ctx := signalContext(context.Background())

		client, err := eth.NewEthClient(eth.Config{
			MasterRPCURL:   cfg.rpcURL,
			MasterAddress:  cfg.masterAddress,
			PrivateKey:     nil,
			RequestTimeout: time.Second * 5,
		})
		if err != nil {
			return errors.Errorf("failed to create eth client: %w", err)
		}

		deriver, err := valset.NewValsetDeriver(client)
		if err != nil {
			return errors.Errorf("failed to create valset deriver: %w", err)
		}

		generator, err := valset.NewValsetGenerator(deriver, client)
		if err != nil {
			return errors.Errorf("failed to create valset generator: %w", err)
		}

		header, err := generator.GenerateValidatorSetHeader(ctx)
		if err != nil {
			return errors.Errorf("failed to generate validator set header: %w", err)
		}

		slog.Info("Valset header generated!")

		jsonData, err := header.EncodeJSON()
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
