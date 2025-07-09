package network

import (
	"context"
	"io/fs"
	entity2 "middleware-offchain/internal/entity"
	"strings"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// The config can be populated from command-line flags, environment variables, and a config.yaml file.
// Priority order (highest to lowest):
// 1. Command-line flags
// 2. Environment variables (dashes replaced by underscores)
// 3. config.yaml file (specified by --config or default "config.yaml")
type config struct {
	Epoch      uint64                       `mapstructure:"epoch"`
	ChainsId   []uint64                     `mapstructure:"chains-id" validate:"required"`
	ChainsUrl  []string                     `mapstructure:"chains-rpc-url" validate:"required"`
	Driver     entity2.CMDCrossChainAddress `mapstructure:"driver" validate:"required"`
	Compact    bool                         `mapstructure:"compact"`
	Commit     bool                         `mapstructure:"commit"`
	SecretKey  string                       `mapstructure:"secret-key"`
	OutputFile string                       `mapstructure:"output-file"`
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

func addFlags() {
	networkCmd.PersistentFlags().StringVar(&configFile, "config", "config.utils.network.yaml", "Path to config file")
	networkCmd.PersistentFlags().StringSlice("chains-rpc-url", nil, "Chains rpc urls")
	networkCmd.PersistentFlags().IntSlice("chains-id", nil, "Chains ids")
	networkCmd.PersistentFlags().String("driver.address", "", "Driver contract address")
	networkCmd.PersistentFlags().Uint64("driver.chain-id", 0, "Driver contract chain id")
	networkCmd.PersistentFlags().Uint64("epoch", 0, "Network epoch")
	networkCmd.PersistentFlags().Bool("compact", false, "Compact valset print")

	genesisCmd.PersistentFlags().Bool("commit", false, "Commit genesis flag")
	genesisCmd.PersistentFlags().String("secret-key", "", "Secret key for genesis commit")
	genesisCmd.PersistentFlags().StringP("output", "o", "", "Output file path")
}

func initConfig(cmd *cobra.Command, _ []string) error {
	var cfg config

	v := viper.New()

	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	v.SetEnvPrefix("SYMB")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.AutomaticEnv()

	if err := v.BindPFlag("chains-rpc-url", cmd.InheritedFlags().Lookup("chains-rpc-url")); err != nil {
		return errors.Errorf("Failed to bind rpc-url: %w", err)
	}
	if err := v.BindPFlag("chains-id", cmd.InheritedFlags().Lookup("chains-id")); err != nil {
		return errors.Errorf("Failed to bind rpc-url: %w", err)
	}

	if err := v.BindPFlag("driver.address", cmd.InheritedFlags().Lookup("driver.address")); err != nil {
		return errors.Errorf("Failed to bind driver-address: %w", err)
	}
	if err := v.BindPFlag("driver.chain-id", cmd.InheritedFlags().Lookup("driver.chain-id")); err != nil {
		return errors.Errorf("Failed to bind driver-address: %w", err)
	}

	if flag := cmd.PersistentFlags().Lookup("epoch"); flag != nil {
		if err := v.BindPFlag("epoch", flag); err != nil {
			return errors.Errorf("Failed to bind flag: %w", err)
		}
	}

	if flag := cmd.PersistentFlags().Lookup("compact"); flag != nil {
		if err := v.BindPFlag("compact", flag); err != nil {
			return errors.Errorf("Failed to bind flag: %w", err)
		}
	}

	if flag := cmd.PersistentFlags().Lookup("commit"); flag != nil {
		if err := v.BindPFlag("commit", flag); err != nil {
			return errors.Errorf("Failed to bind flag: %w", err)
		}
	}

	if flag := cmd.PersistentFlags().Lookup("secret-key"); flag != nil {
		if err := v.BindPFlag("secret-key", flag); err != nil {
			return errors.Errorf("Failed to bind flag: %w", err)
		}
	}

	if flag := cmd.PersistentFlags().Lookup("output"); flag != nil {
		if err := v.BindPFlag("output-file", flag); err != nil {
			return errors.Errorf("Failed to bind flag: %w", err)
		}
	}

	err := v.ReadInConfig()

	if err != nil && !errors.Is(err, viper.ConfigFileNotFoundError{}) && !errors.As(err, lo.ToPtr(&fs.PathError{})) {
		return errors.Errorf("Failed to read config file: %w", err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return errors.Errorf("Failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return errors.Errorf("Invalid config: %w", err)
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
