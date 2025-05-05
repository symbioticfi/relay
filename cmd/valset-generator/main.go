package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"middleware-offchain/internal/client/eth"
	"middleware-offchain/pkg/log"
	"middleware-offchain/valset"
)

func main() {
	cobra.OnInitialize(initConfig)

	log.Init()

	// Add commands
	rootCmd.AddCommand(startCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("error executing command", "error", err)
		os.Exit(1)
	}
	slog.Info("Valset generator completed successfully")
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "valset-generator",
	Short: "Generate valset headers",
	Long:  "Generate valset headers",
}

type config struct {
	EthEndpoint   string `mapstructure:"eth_endpoint"`
	ContractAddr  string `mapstructure:"contract_addr"`
	EthPrivateKey string `mapstructure:"eth_private_key"`
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the generate valset header app",
	Long:  "Start the generate valset header app",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := signalContext(context.Background())

		var cfg config
		err := viper.Unmarshal(&cfg)
		if err != nil {
			return errors.Errorf("failed to unmarshal config: %w", err)
		}

		privateKeyInt, ok := new(big.Int).SetString(cfg.EthPrivateKey, 10)
		if !ok {
			return errors.Errorf("failed to parse private key: %s", cfg.EthPrivateKey)
		}

		ethClient, err := eth.NewEthClient(eth.Config{
			MasterRPCURL:  cfg.EthEndpoint,
			MasterAddress: cfg.ContractAddr,
			PrivateKey:    privateKeyInt.Bytes(),
		})
		if err != nil {
			return errors.Errorf("failed to create eth client: %w", err)
		}

		deriver, err := valset.NewValsetDeriver(ethClient)
		if err != nil {
			return errors.Errorf("failed to create valset deriver: %w", err)
		}

		generator, err := valset.NewValsetGenerator(deriver, ethClient)
		if err != nil {
			return errors.Errorf("failed to create valset generator: %w", err)
		}

		header, err := generator.GenerateValidatorSetHeader(ctx)
		if err != nil {
			return errors.Errorf("failed to generate valset header: %w", err)
		}

		encodedJSON, err := header.EncodeJSON()
		if err != nil {
			return errors.Errorf("failed to encode valset header: %w", err)
		}

		fmt.Println(string(encodedJSON))

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

// initConfig reads in config file and ENV variables if set
func initConfig() {
	viper.SetConfigType("yaml")

	viper.AddConfigPath(".")
	viper.SetConfigName("middleware-offchain.config.yaml")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in
	err := viper.ReadInConfig()
	if err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
