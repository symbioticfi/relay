package main

import (
	"context"
	"io/fs"
	"strings"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"middleware-offchain/internal/entity"
)

// The config can be populated from command-line flags, environment variables, and a config.yaml file.
// Priority order (highest to lowest):
// 1. Command-line flags
// 2. Environment variables (prefixed with SYMB_ and dashes replaced by underscores)
// 3. config.yaml file (specified by --config or default "config.yaml")
type config struct {
	Driver           entity.CMDCrossChainAddress `mapstructure:"driver" validate:"required"`
	LogLevel         string                      `mapstructure:"log-level" validate:"oneof=debug info warn error"`
	LogMode          string                      `mapstructure:"log-mode" validate:"oneof=text pretty"`
	P2PListenAddress string                      `mapstructure:"p2p-listen"`
	HTTPListenAddr   string                      `mapstructure:"http-listen" validate:"required"`
	SecretKey        string                      `mapstructure:"secret-key" validate:"required"`
	IsAggregator     bool                        `mapstructure:"aggregator"`
	IsSigner         bool                        `mapstructure:"signer"`
	IsCommitter      bool                        `mapstructure:"committer"`
	StorageDir       string                      `mapstructure:"storage-dir"`
	ChainsId         []uint64                    `mapstructure:"chains-id" validate:"required"`
	ChainsUrl        []string                    `mapstructure:"chains-url" validate:"required"`
}

func (c config) Validate() error {
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return errors.Errorf("invalid config: %w", err)
	}

	return nil
}

var (
	configFile string
)

func addRootFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&configFile, "config", "config.yaml", "Path to config file")

	rootCmd.PersistentFlags().Uint64("driver.chain-id", 0, "Driver contract chain id")
	rootCmd.PersistentFlags().String("driver.address", "", "Driver contract address")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-mode", "text", "Log mode (text, pretty)")
	rootCmd.PersistentFlags().String("p2p-listen", "", "P2P listen address")
	rootCmd.PersistentFlags().String("http-listen", "", "Http listener address")
	rootCmd.PersistentFlags().String("secret-key", "", "Secret key for BLS signature generation")
	rootCmd.PersistentFlags().Bool("aggregator", false, "Is Aggregator Node")
	rootCmd.PersistentFlags().Bool("signer", true, "Is Signer Node")
	rootCmd.PersistentFlags().Bool("committer", false, "Is Committer Node")
	rootCmd.PersistentFlags().String("storage-dir", ".data", "Dir to store data")
	rootCmd.PersistentFlags().UintSlice("chains-id", nil, "Chains ids")
	rootCmd.PersistentFlags().StringSlice("chains-rpc-url", nil, "Chains RPC URLS")
}

func initConfig(cmd *cobra.Command, _ []string) error {
	var cfg config

	v := viper.New()

	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	v.SetEnvPrefix("SYMB")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.AutomaticEnv()

	if err := v.BindPFlag("driver.chain-id", cmd.PersistentFlags().Lookup("driver.chain-id")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("driver.address", cmd.PersistentFlags().Lookup("driver.address")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("log-level", cmd.PersistentFlags().Lookup("log-level")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("log-mode", cmd.PersistentFlags().Lookup("log-mode")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("p2p-listen", cmd.PersistentFlags().Lookup("p2p-listen")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("http-listen", cmd.PersistentFlags().Lookup("http-listen")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("secret-key", cmd.PersistentFlags().Lookup("secret-key")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("aggregator", cmd.PersistentFlags().Lookup("aggregator")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("signer", cmd.PersistentFlags().Lookup("signer")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("committer", cmd.PersistentFlags().Lookup("committer")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("storage-dir", cmd.PersistentFlags().Lookup("storage-dir")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("chains-id", cmd.PersistentFlags().Lookup("chains-id")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("chains-rpc-url", cmd.PersistentFlags().Lookup("chains-rpc-url")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}

	err := v.ReadInConfig()
	if err != nil && !errors.Is(err, viper.ConfigFileNotFoundError{}) && !errors.As(err, lo.ToPtr(&fs.PathError{})) {
		return errors.Errorf("failed to read config file: %w", err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return errors.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return errors.Errorf("invalid config: %w", err)
	}

	cmd.SetContext(ctxWithCfg(cmd.Context(), cfg))

	return nil
}

type contextKeyStruct struct{}

func ctxWithCfg(ctx context.Context, cfg config) context.Context {
	return context.WithValue(ctx, contextKeyStruct{}, cfg)
}
func cfgFromCtx(ctx context.Context) config {
	return ctx.Value(contextKeyStruct{}).(config)
}
