package main

import (
	"context"
	"fmt"
	"io/fs"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"github.com/symbioticfi/relay/pkg/signals"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/symbioticfi/relay/internal/entity"
)

type CMDSecretKey struct {
	Namespace string `validate:"required"`
	KeyType   uint8  `validate:"required"`
	KeyId     int    `validate:"required"`
	Secret    string `validate:"required"`
}

func (c *CMDSecretKey) String() string {
	return fmt.Sprintf("%s/%d/%d/%s", c.Namespace, c.KeyType, c.KeyId, c.Secret)
}

func (c *CMDSecretKey) Set(str string) error {
	strs := strings.Split(str, "/")
	if len(strs) != 4 {
		return errors.Errorf("invalid secret key format: %s, expected {namespace}/{type}/{id}", str)
	}
	c.Namespace = strs[0]
	c.Secret = strs[3]

	v, err := strconv.Atoi(strs[1])
	if err != nil {
		return err
	}
	c.KeyType = uint8(v)

	v, err = strconv.Atoi(strs[2])
	if err != nil {
		return err
	}
	c.KeyId = v
	return nil
}

func (c *CMDSecretKey) FromStr(str string) (*CMDSecretKey, error) {
	err := c.Set(str)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *CMDSecretKey) Type() string {
	return "secret-key"
}

type CMDSecretKeySlice []CMDSecretKey

func (s *CMDSecretKeySlice) String() string {
	strs := make([]string, len(*s))
	for i, ss := range *s {
		strs[i] = ss.String()
	}
	return strings.Join(strs, ",")
}

func (s *CMDSecretKeySlice) Set(str string) error {
	strs := strings.Split(str, ",")
	for _, elem := range strs {
		key := CMDSecretKey{}
		err := key.Set(elem)
		if err != nil {
			return err
		}
		*s = append(*s, key)
	}
	return nil
}

func (s *CMDSecretKeySlice) Type() string {
	return "secret-key-slice"
}

// The config can be populated from command-line flags, environment variables, and a config.yaml file.
// Priority order (highest to lowest):
// 1. Command-line flags
// 2. Environment variables (prefixed with SYMB_ and dashes replaced by underscores)
// 3. config.yaml file (specified by --config or default "config.yaml")
type config struct {
	Driver            entity.CMDCrossChainAddress `mapstructure:"driver" validate:"required"`
	LogLevel          string                      `mapstructure:"log-level" validate:"oneof=debug info warn error"`
	LogMode           string                      `mapstructure:"log-mode" validate:"oneof=json text pretty"`
	P2PListenAddress  string                      `mapstructure:"p2p-listen" validate:"required"`
	HTTPListenAddr    string                      `mapstructure:"http-listen" validate:"required"`
	MetricsListenAddr string                      `mapstructure:"metrics-listen"`
	SecretKeys        CMDSecretKeySlice           `mapstructure:"secret-keys"`
	IsAggregator      bool                        `mapstructure:"aggregator"`
	IsSigner          bool                        `mapstructure:"signer"`
	IsCommitter       bool                        `mapstructure:"committer"`
	StorageDir        string                      `mapstructure:"storage-dir"`
	Chains            []string                    `mapstructure:"chains" validate:"required"`
	CircuitsDir       string                      `mapstructure:"circuits-dir"`
	KeyStore          entity.KeyStore             `mapstructure:"keystore"`
	Bootnodes         []string                    `mapstructure:"bootnodes"`
	DHTMode           string                      `mapstructure:"dht-mode" validate:"oneof=auto server client disabled"`
	MDnsEnabled       bool                        `mapstructure:"enable-mdns"`
	MaxCalls          int                         `mapstructure:"evm-max-calls"`
	SignalCfg         signals.SignalCfg           `mapstructure:"signal"`
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
	rootCmd.PersistentFlags().String("log-mode", "json", "Log mode (text, pretty, json)")
	rootCmd.PersistentFlags().String("p2p-listen", "", "P2P listen address")
	rootCmd.PersistentFlags().String("http-listen", "", "Http listener address")
	rootCmd.PersistentFlags().String("metrics-listen", "", "Http listener address for metrics endpoint")
	rootCmd.PersistentFlags().Bool("aggregator", false, "Is Aggregator Node")
	rootCmd.PersistentFlags().Bool("signer", true, "Is Signer Node")
	rootCmd.PersistentFlags().Bool("committer", false, "Is Committer Node")
	rootCmd.PersistentFlags().String("storage-dir", ".data", "Dir to store data")
	rootCmd.PersistentFlags().StringSlice("chains", nil, "Chains, comma separated rpc-url,..")
	rootCmd.PersistentFlags().Var(&CMDSecretKeySlice{}, "secret-keys", "Secret keys, comma separated {namespace}/{type}/{id}/{key},..")
	rootCmd.PersistentFlags().String("circuits-dir", "", "Directory path to load zk circuits from, if empty then zp prover is disabled")
	rootCmd.PersistentFlags().String("keystore.path", "", "Path to optional keystore file, if provided will be used instead of secret-keys flag")
	rootCmd.PersistentFlags().String("keystore.password", "", "Password for the keystore file, if provided will be used to decrypt the keystore file")
	rootCmd.PersistentFlags().StringSlice("bootnodes", nil, "List of bootnodes in multiaddr format")
	rootCmd.PersistentFlags().String("dht-mode", "server", "DHT mode: auto, server, client, disabled")
	rootCmd.PersistentFlags().Bool("enable-mdns", false, "Enable mDNS discovery for P2P")
	rootCmd.PersistentFlags().Int64("signal.worker-count", 10, "Signal worker count")
	rootCmd.PersistentFlags().Int64("signal.buffer-size", 20, "Signal buffer size")
	rootCmd.PersistentFlags().Int("evm-max-calls", 0, "Max calls in multicall")
}

func DecodeFlagToStruct(fromType reflect.Type, toType reflect.Type, from interface{}) (interface{}, error) {
	if fromType.Kind() != reflect.String {
		// if not string return as is
		return from, nil
	}

	flagType := reflect.TypeOf((*pflag.Value)(nil))

	// if fromType implements pflag.Value then we can parse it from string
	if reflect.PointerTo(toType).Implements(flagType.Elem()) {
		res := reflect.New(toType).Interface().(pflag.Value)
		res.Set(from.(string))
		return res, nil
	}
	return from, nil
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
	if err := v.BindPFlag("metrics-listen", cmd.PersistentFlags().Lookup("metrics-listen")); err != nil {
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
	if err := v.BindPFlag("secret-keys", cmd.PersistentFlags().Lookup("secret-keys")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("chains", cmd.PersistentFlags().Lookup("chains")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("circuits-dir", cmd.PersistentFlags().Lookup("circuits-dir")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("keystore.path", cmd.PersistentFlags().Lookup("keystore.path")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("keystore.password", cmd.PersistentFlags().Lookup("keystore.password")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("bootnodes", cmd.PersistentFlags().Lookup("bootnodes")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("dht-mode", cmd.PersistentFlags().Lookup("dht-mode")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("enable-mdns", cmd.PersistentFlags().Lookup("enable-mdns")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("signal.buffer-size", cmd.PersistentFlags().Lookup("signal.buffer-size")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("signal.worker-count", cmd.PersistentFlags().Lookup("signal.worker-count")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("evm-max-calls", cmd.PersistentFlags().Lookup("evm-max-calls")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}

	err := v.ReadInConfig()
	if err != nil && !errors.Is(err, viper.ConfigFileNotFoundError{}) && !errors.As(err, lo.ToPtr(&fs.PathError{})) {
		return errors.Errorf("failed to read config file: %w", err)
	}

	// pflags allows to implement custom types by implementing pflag.Value (we can define how to parse struct from string)
	// but[1] viper converts back struct defined flags to string automatically using String() method :(
	// but[2] fortunately viper allows to pass decoder that we can use to convert string back to struct :D
	if err := v.Unmarshal(&cfg, viper.DecodeHook(DecodeFlagToStruct)); err != nil {
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
