package root

import (
	"context"
	"fmt"
	"io/fs"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/symbioticfi/relay/pkg/signals"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// The config can be populated from command-line flags, environment variables, and a config.yaml file.
// Priority order (highest to lowest):
// 1. Command-line flags
// 2. Environment variables (prefixed with SYMB_ and dashes replaced by underscores)
// 3. config.yaml file (specified by --config or default "config.yaml")
type config struct {
	StorageDir   string `mapstructure:"storage-dir"`
	CircuitsDir  string `mapstructure:"circuits-dir"`
	MaxUnsigners uint64 `mapstructure:"aggregation-policy-max-unsigners"`

	Log        LogConfig            `mapstructure:"log" validate:"required"`
	API        APIConfig            `mapstructure:"api" validate:"required"`
	Metrics    MetricsConfig        `mapstructure:"metrics"`
	Driver     CMDCrossChainAddress `mapstructure:"driver" validate:"required"`
	SecretKeys CMDSecretKeySlice    `mapstructure:"secret-keys"`
	KeyStore   KeyStore             `mapstructure:"keystore"`
	SignalCfg  signals.Config       `mapstructure:"signal"`
	Cache      CacheConfig          `mapstructure:"cache"`
	Sync       SyncConfig           `mapstructure:"sync"`
	KeyCache   KeyCache             `mapstructure:"key-cache"`
	P2P        P2PConfig            `mapstructure:"p2p" validate:"required"`
	Evm        EvmConfig            `mapstructure:"evm" validate:"required"`
	ForceRole  ForceRole            `mapstructure:"force-role"`
}

type LogConfig struct {
	Level string `mapstructure:"level" validate:"oneof=debug info warn error"`
	Mode  string `mapstructure:"mode" validate:"oneof=json text pretty"`
}

type APIConfig struct {
	ListenAddress     string `mapstructure:"listen" validate:"required"`
	MaxAllowedStreams uint64 `mapstructure:"max-allowed-streams" validate:"required"`
	VerboseLogging    bool   `mapstructure:"verbose-logging"`
	HTTPGateway       bool   `mapstructure:"http-gateway"`
}

type MetricsConfig struct {
	ListenAddress string `mapstructure:"listen"`
	PprofEnabled  bool   `mapstructure:"pprof"`
}

type CMDCrossChainAddress struct {
	ChainID uint64 `mapstructure:"chain-id" validate:"required"`
	Address string `mapstructure:"address" validate:"required"`
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

type KeyStore struct {
	Path     string `json:"path"`
	Password string `json:"password"`
}
type CacheConfig struct {
	NetworkConfigCacheSize int `mapstructure:"network-config-size"`
	ValidatorSetCacheSize  int `mapstructure:"validator-set-size"`
}

type SyncConfig struct {
	Enabled      bool          `mapstructure:"enabled"`
	Period       time.Duration `mapstructure:"period"`
	Timeout      time.Duration `mapstructure:"timeout"`
	EpochsToSync uint64        `mapstructure:"epochs"`
}

type KeyCache struct {
	// max size of the cache
	Size    int  `mapstructure:"size"`
	Enabled bool `mapstructure:"enabled"`
}

type P2PConfig struct {
	ListenAddress string   `mapstructure:"listen" validate:"required"`
	Bootnodes     []string `mapstructure:"bootnodes"`
	DHTMode       string   `mapstructure:"dht-mode" validate:"oneof=auto server client disabled"`
	MDnsEnabled   bool     `mapstructure:"mdns"`
}

type EvmConfig struct {
	Chains   []string `mapstructure:"chains" validate:"required"`
	MaxCalls int      `mapstructure:"max-calls"`
}

type ForceRole struct {
	Aggregator bool `mapstructure:"aggregator"`
	Committer  bool `mapstructure:"committer"`
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

	rootCmd.PersistentFlags().String("log.level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log.mode", "json", "Log mode (text, pretty, json)")
	rootCmd.PersistentFlags().String("storage-dir", ".data", "Dir to store data")
	rootCmd.PersistentFlags().String("circuits-dir", "", "Directory path to load zk circuits from, if empty then zp prover is disabled")
	rootCmd.PersistentFlags().Uint64("aggregation-policy-max-unsigners", 50, "Max unsigners for low cost agg policy")
	rootCmd.PersistentFlags().String("api.listen", "", "API Server listener address")
	rootCmd.PersistentFlags().Uint64("api.max-allowed-streams", 100, "Max allowed streams count API Server")
	rootCmd.PersistentFlags().Bool("api.verbose-logging", false, "Enable verbose logging for the API Server on /api/v1/* path")
	rootCmd.PersistentFlags().Bool("api.http-gateway", false, "Enable HTTP/JSON REST API gateway")
	rootCmd.PersistentFlags().String("metrics.listen", "", "Http listener address for metrics endpoint")
	rootCmd.PersistentFlags().Bool("metrics.pprof", false, "Enable pprof debug endpoints")
	rootCmd.PersistentFlags().Uint64("driver.chain-id", 0, "Driver contract chain id")
	rootCmd.PersistentFlags().String("driver.address", "", "Driver contract address")
	rootCmd.PersistentFlags().Var(&CMDSecretKeySlice{}, "secret-keys", "Secret keys, comma separated {namespace}/{type}/{id}/{key},..")
	rootCmd.PersistentFlags().String("keystore.path", "", "Path to optional keystore file, if provided will be used instead of secret-keys flag")
	rootCmd.PersistentFlags().String("keystore.password", "", "Password for the keystore file, if provided will be used to decrypt the keystore file")
	rootCmd.PersistentFlags().Int64("signal.worker-count", 10, "Signal worker count")
	rootCmd.PersistentFlags().Int64("signal.buffer-size", 20, "Signal buffer size")
	rootCmd.PersistentFlags().Int("cache.network-config-size", 10, "Network config cache size")
	rootCmd.PersistentFlags().Int("cache.validator-set-size", 10, "Validator set cache size")
	rootCmd.PersistentFlags().Bool("sync.enabled", true, "Enable signature syncer")
	rootCmd.PersistentFlags().Duration("sync.period", time.Second*5, "Signature sync period")
	rootCmd.PersistentFlags().Duration("sync.timeout", time.Minute, "Signature sync timeout")
	rootCmd.PersistentFlags().Uint64("sync.epochs", 5, "Epochs to sync")
	rootCmd.PersistentFlags().Int("key-cache.size", 100, "Key cache size")
	rootCmd.PersistentFlags().Bool("key-cache.enabled", true, "Enable key cache")
	rootCmd.PersistentFlags().String("p2p.listen", "", "P2P listen address")
	rootCmd.PersistentFlags().StringSlice("p2p.bootnodes", nil, "List of bootnodes in multiaddr format")
	rootCmd.PersistentFlags().String("p2p.dht-mode", "server", "DHT mode: auto, server, client, disabled")
	rootCmd.PersistentFlags().Bool("p2p.mdns", false, "Enable mDNS discovery for P2P")
	rootCmd.PersistentFlags().StringSlice("evm.chains", nil, "Chains, comma separated rpc-url,..")
	rootCmd.PersistentFlags().Int("evm.max-calls", 0, "Max calls in multicall")
	rootCmd.PersistentFlags().Bool("force-role.aggregator", false, "Force node to act as aggregator regardless of deterministic scheduling")
	rootCmd.PersistentFlags().Bool("force-role.committer", false, "Force node to act as committer regardless of deterministic scheduling")
}

func DecodeFlagToStruct(fromType reflect.Type, toType reflect.Type, from interface{}) (interface{}, error) {
	if fromType.Kind() != reflect.String {
		// if not string return as is
		return from, nil
	}

	// Handle time.Duration specifically
	if toType == reflect.TypeOf(time.Duration(0)) {
		duration, err := time.ParseDuration(from.(string))
		if err != nil {
			return nil, errors.Errorf("failed to parse duration: %w", err)
		}
		return duration, nil
	}

	flagType := reflect.TypeOf((*pflag.Value)(nil))

	// if fromType implements pflag.Value then we can parse it from string
	if reflect.PointerTo(toType).Implements(flagType.Elem()) {
		res := reflect.New(toType).Interface().(pflag.Value)
		if err := res.Set(from.(string)); err != nil {
			return nil, errors.Errorf("failed to set flag value: %w", err)
		}
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

	if err := v.BindPFlag("log.level", cmd.PersistentFlags().Lookup("log.level")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("log.mode", cmd.PersistentFlags().Lookup("log.mode")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("storage-dir", cmd.PersistentFlags().Lookup("storage-dir")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("circuits-dir", cmd.PersistentFlags().Lookup("circuits-dir")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("aggregation-policy-max-unsigners", cmd.PersistentFlags().Lookup("aggregation-policy-max-unsigners")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("api.listen", cmd.PersistentFlags().Lookup("api.listen")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("api.verbose-logging", cmd.PersistentFlags().Lookup("api.verbose-logging")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("api.http-gateway", cmd.PersistentFlags().Lookup("api.http-gateway")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("api.max-allowed-streams", cmd.PersistentFlags().Lookup("api.max-allowed-streams")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("metrics.listen", cmd.PersistentFlags().Lookup("metrics.listen")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("metrics.pprof", cmd.PersistentFlags().Lookup("metrics.pprof")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("driver.chain-id", cmd.PersistentFlags().Lookup("driver.chain-id")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("driver.address", cmd.PersistentFlags().Lookup("driver.address")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("secret-keys", cmd.PersistentFlags().Lookup("secret-keys")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("keystore.path", cmd.PersistentFlags().Lookup("keystore.path")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("keystore.password", cmd.PersistentFlags().Lookup("keystore.password")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("signal.buffer-size", cmd.PersistentFlags().Lookup("signal.buffer-size")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("signal.worker-count", cmd.PersistentFlags().Lookup("signal.worker-count")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("cache.network-config-size", cmd.PersistentFlags().Lookup("cache.network-config-size")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("cache.validator-set-size", cmd.PersistentFlags().Lookup("cache.validator-set-size")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("sync.enabled", cmd.PersistentFlags().Lookup("sync.enabled")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("sync.timeout", cmd.PersistentFlags().Lookup("sync.timeout")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("sync.period", cmd.PersistentFlags().Lookup("sync.period")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("sync.epochs", cmd.PersistentFlags().Lookup("sync.epochs")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("key-cache.size", cmd.PersistentFlags().Lookup("key-cache.size")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("key-cache.enabled", cmd.PersistentFlags().Lookup("key-cache.enabled")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("p2p.listen", cmd.PersistentFlags().Lookup("p2p.listen")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("p2p.bootnodes", cmd.PersistentFlags().Lookup("p2p.bootnodes")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("p2p.dht-mode", cmd.PersistentFlags().Lookup("p2p.dht-mode")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("p2p.mdns", cmd.PersistentFlags().Lookup("p2p.mdns")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("evm.chains", cmd.PersistentFlags().Lookup("evm.chains")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("evm.max-calls", cmd.PersistentFlags().Lookup("evm.max-calls")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("force-role.aggregator", cmd.PersistentFlags().Lookup("force-role.aggregator")); err != nil {
		return errors.Errorf("failed to bind flag: %w", err)
	}
	if err := v.BindPFlag("force-role.committer", cmd.PersistentFlags().Lookup("force-role.committer")); err != nil {
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
