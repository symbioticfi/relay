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
	// common configs / flags
	Epoch     uint64                       `mapstructure:"epoch"`
	ChainsId  []uint64                     `mapstructure:"chains-id" validate:"required"`
	ChainsUrl []string                     `mapstructure:"chains-rpc-url" validate:"required"`
	Driver    entity2.CMDCrossChainAddress `mapstructure:"driver" validate:"required"`

	// info configs / flags
	Validators     bool `mapstructure:"validators"`
	ValidatorsFull bool `mapstructure:"validators-full"`
	Addresses      bool `mapstructure:"addresses"`

	// genesis configs / flags
	Commit     bool   `mapstructure:"commit"`
	SecretKey  string `mapstructure:"secret-key"`
	OutputFile string `mapstructure:"output-file"`
	Json       bool   `mapstructure:"json"`
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
	networkCmd.PersistentFlags().Uint64P("epoch", "e", 0, "Network epoch to fetch info")

	infoCmd.PersistentFlags().BoolP("validators", "v", false, "Print compact validators info")
	infoCmd.PersistentFlags().BoolP("validators-full", "V", false, "Print full validators info")
	infoCmd.PersistentFlags().BoolP("addresses", "a", false, "Print addresses")

	genesisCmd.PersistentFlags().Bool("commit", false, "Commit genesis flag")
	genesisCmd.PersistentFlags().String("secret-key", "", "Secret key for genesis commit")
	genesisCmd.PersistentFlags().BoolP("json", "j", false, "Print as json")
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

	// common flags
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
	if err := v.BindPFlag("epoch", cmd.InheritedFlags().Lookup("epoch")); err != nil {
		return errors.Errorf("Failed to bind flag: %w", err)
	}

	// info flags
	if flag := cmd.PersistentFlags().Lookup("validators"); flag != nil {
		if err := v.BindPFlag("validators", flag); err != nil {
			return errors.Errorf("Failed to bind flag: %w", err)
		}
	}
	if flag := cmd.PersistentFlags().Lookup("validators-full"); flag != nil {
		if err := v.BindPFlag("validators-full", flag); err != nil {
			return errors.Errorf("Failed to bind flag: %w", err)
		}
	}
	if flag := cmd.PersistentFlags().Lookup("addresses"); flag != nil {
		if err := v.BindPFlag("addresses", flag); err != nil {
			return errors.Errorf("Failed to bind flag: %w", err)
		}
	}

	// genesis flags
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
	if flag := cmd.PersistentFlags().Lookup("json"); flag != nil {
		if err := v.BindPFlag("json", flag); err != nil {
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
