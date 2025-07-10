package operator

import (
	"context"
	"io/fs"
	"middleware-offchain/core/entity"
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
	Driver     entity2.CMDCrossChainAddress `mapstructure:"driver" validate:"required"`
	ChainsId   []uint64                     `mapstructure:"chains-id" validate:"required"`
	ChainsUrl  []string                     `mapstructure:"chains-rpc-url" validate:"required"`
	Path       string                       `mapstructure:"path"`
	Password   string                       `mapstructure:"password"`
	KeyTag     uint8                        `mapstructure:"key-tag"`
	PrivateKey string                       `mapstructure:"private-key"`
	Compact    bool                         `mapstructure:"compact"`
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
	operatorCmd.PersistentFlags().StringVar(&configFile, "config", "config.utils.operator.yaml", "Path to config file")
	operatorCmd.PersistentFlags().StringSlice("chains-id", nil, "Chains id")
	operatorCmd.PersistentFlags().StringSlice("chains-rpc-url", nil, "Chains rpc urls")
	operatorCmd.PersistentFlags().Uint64("driver.chain-id", 0, "Driver contract chain ID")
	operatorCmd.PersistentFlags().String("driver.address", "", "Driver contract address")

	infoCmd.PersistentFlags().Uint64("epoch", 0, "Network epoch")
	infoCmd.PersistentFlags().String("path", "", "Keystore path")
	infoCmd.PersistentFlags().String("password", "", "Keystore password")
	infoCmd.PersistentFlags().Uint8("key-tag", uint8(entity.KeyTypeInvalid), "Key tag of operator key")
	infoCmd.PersistentFlags().Bool("compact", false, "Compact operator info print")

	registerCmd.PersistentFlags().String("private-key", "", "Private key of operator")
	registerCmd.PersistentFlags().Uint64("chain-id", 0, "Chain id where to register")

	registerKeyCmd.PersistentFlags().String("private-key", "", "Private key of operator")
	registerKeyCmd.PersistentFlags().String("path", "", "Keystore path")
	registerKeyCmd.PersistentFlags().String("password", "", "Keystore password")
	registerKeyCmd.PersistentFlags().Uint8("key-tag", uint8(entity.KeyTypeInvalid), "Key tag of operator key")
}

func initConfig(cmd *cobra.Command, _ []string) error {
	var cfg config

	v := viper.New()

	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	v.SetEnvPrefix("SYMB")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.AutomaticEnv()

	// Bind flags to viper
	if err := v.BindPFlag("chains-id", cmd.InheritedFlags().Lookup("chains-id")); err != nil {
		return errors.Errorf("failed to bind chains-id: %w", err)
	}

	if err := v.BindPFlag("chains-rpc-url", cmd.InheritedFlags().Lookup("chains-rpc-url")); err != nil {
		return errors.Errorf("failed to bind chains-rpc-url: %w", err)
	}

	if err := v.BindPFlag("driver.chain-id", cmd.InheritedFlags().Lookup("driver.chain-id")); err != nil {
		return errors.Errorf("failed to bind driver.chain-id: %w", err)
	}

	if err := v.BindPFlag("driver.address", cmd.InheritedFlags().Lookup("driver.address")); err != nil {
		return errors.Errorf("failed to bind driver.address: %w", err)
	}

	if flag := cmd.PersistentFlags().Lookup("epoch"); flag != nil {
		if err := v.BindPFlag("epoch", flag); err != nil {
			return errors.Errorf("failed to bind epoch: %w", err)
		}
	}

	if flag := cmd.PersistentFlags().Lookup("path"); flag != nil {
		if err := v.BindPFlag("path", flag); err != nil {
			return errors.Errorf("failed to bind path: %w", err)
		}
	}

	if flag := cmd.PersistentFlags().Lookup("password"); flag != nil {
		if err := v.BindPFlag("password", flag); err != nil {
			return errors.Errorf("failed to bind password: %w", err)
		}
	}

	if flag := cmd.PersistentFlags().Lookup("key-tag"); flag != nil {
		if err := v.BindPFlag("key-tag", flag); err != nil {
			return errors.Errorf("failed to bind key-tag: %w", err)
		}
	}

	if flag := cmd.PersistentFlags().Lookup("private-key"); flag != nil {
		if err := v.BindPFlag("private-key", flag); err != nil {
			return errors.Errorf("failed to bind private-key: %w", err)
		}
	}

	if flag := cmd.PersistentFlags().Lookup("compact"); flag != nil {
		if err := v.BindPFlag("compact", flag); err != nil {
			return errors.Errorf("failed to bind compact: %w", err)
		}
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
