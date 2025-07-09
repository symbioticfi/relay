package keys

import (
	"context"
	"io/fs"
	"middleware-offchain/core/entity"
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
	Path       string `mapstructure:"path" validate:"required"`
	Password   string `mapstructure:"password"`
	KeyTag     uint8  `mapstructure:"key-tag"`
	PrivateKey string `mapstructure:"private-key"`
	Generate   bool   `mapstructure:"generate"`
	Force      bool   `mapstructure:"force"`
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
	keysCmd.PersistentFlags().StringVar(&configFile, "config", "config.utils.keys.yaml", "Path to config file")
	keysCmd.PersistentFlags().StringP("path", "p", "./keystore.jks", "Path to keystore")
	keysCmd.PersistentFlags().String("password", "", "Keystore password")

	addKeyCmd.PersistentFlags().Uint8("key-tag", uint8(entity.KeyTypeInvalid), "key tag")
	addKeyCmd.PersistentFlags().String("private-key", "", "private key")
	addKeyCmd.PersistentFlags().Bool("generate", false, "generate key")
	addKeyCmd.PersistentFlags().Bool("force", false, "force overwrite key")

	removeKeyCmd.PersistentFlags().Uint8("key-tag", uint8(entity.KeyTypeInvalid), "key tag")

	updateKeyCmd.PersistentFlags().Uint8("key-tag", uint8(entity.KeyTypeInvalid), "key tag")
	updateKeyCmd.PersistentFlags().String("private-key", "", "private key")
}

func initConfig(cmd *cobra.Command, _ []string) error {
	var cfg config

	v := viper.New()

	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	v.SetEnvPrefix("SYMB")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.AutomaticEnv()

	if err := v.BindPFlag("path", cmd.InheritedFlags().Lookup("path")); err != nil {
		return errors.Errorf("Failed to bind path: %w", err)
	}

	if err := v.BindPFlag("password", cmd.InheritedFlags().Lookup("password")); err != nil {
		return errors.Errorf("Failed to bind password: %w", err)
	}

	if flag := cmd.PersistentFlags().Lookup("key-tag"); flag != nil {
		if err := v.BindPFlag("key-tag", flag); err != nil {
			return errors.Errorf("Failed to bind flag: %w", err)
		}
	}

	if flag := cmd.PersistentFlags().Lookup("private-key"); flag != nil {
		if err := v.BindPFlag("private-key", flag); err != nil {
			return errors.Errorf("Failed to bind flag: %w", err)
		}
	}

	if flag := cmd.PersistentFlags().Lookup("generate"); flag != nil {
		if err := v.BindPFlag("generate", cmd.PersistentFlags().Lookup("generate")); err != nil {
			return errors.Errorf("Failed to bind flag: %w", err)
		}
	}

	if flag := cmd.PersistentFlags().Lookup("force"); flag != nil {
		if err := v.BindPFlag("force", flag); err != nil {
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
