package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"offchain-middleware/bls"
	"offchain-middleware/eth"
	"offchain-middleware/network"
	"offchain-middleware/p2p"
	"offchain-middleware/storage"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// Config holds all application configuration
type Config struct {
	ListenAddr    string
	EthEndpoint   string   `mapstructure:"eth"`
	ContractAddr  string   `mapstructure:"contract"`
	EthPrivateKey []byte   `mapstructure:"eth-private-key"`
	BlsPrivateKey []byte   `mapstructure:"bls-private-key"`
	Peers         []string `mapstructure:"peers"`
}

// App represents the application and its components
type App struct {
	config         Config
	storage        *storage.Storage
	ethClient      eth.IEthClient
	p2pService     *p2p.P2PService
	networkService *network.NetworkService
}

// NewApp creates a new application instance with the provided configuration
func NewApp(config Config) *App {
	return &App{
		config: config,
	}
}

// Initialize sets up all components of the application
func (a *App) Initialize(ctx context.Context) error {
	// Parse the listen address
	addr, err := multiaddr.NewMultiaddr(a.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("invalid listen address: %w", err)
	}

	// Create storage
	a.storage = storage.NewStorage()

	// Create Ethereum client
	if !viper.GetBool("test") {
		a.ethClient, err = eth.NewEthClient(a.config.EthEndpoint, a.config.ContractAddr, a.config.EthPrivateKey)
		if err != nil {
			return fmt.Errorf("failed to create ETH service: %w", err)
		}
	} else {
		a.ethClient = eth.NewMockEthClient()
	}

	key, err := crypto.UnmarshalSecp256k1PrivateKey(a.config.EthPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to unmarshal ETH private key: %w", err)
	}

	// Create the P2P service
	a.p2pService, err = p2p.NewP2PService(ctx, key, []multiaddr.Multiaddr{addr}, a.config.Peers, a.storage)
	if err != nil {
		return fmt.Errorf("failed to create P2P service: %w", err)
	}

	// Create network service
	a.networkService, err = network.NewNetworkService(a.p2pService, a.ethClient, a.storage, bls.ComputeKeyPair(a.config.BlsPrivateKey))
	if err != nil {
		return fmt.Errorf("failed to create network service: %w", err)
	}

	return nil
}

// Start begins all services
func (a *App) Start() error {
	// Start the P2P service
	if err := a.p2pService.Start(); err != nil {
		return fmt.Errorf("failed to start P2P service: %w", err)
	}

	// Start the network service
	if err := a.networkService.Start(time.Minute); err != nil {
		return fmt.Errorf("failed to start network service: %w", err)
	}

	return nil
}

// Stop gracefully shuts down all services
func (a *App) Stop() {
	if a.p2pService != nil {
		a.p2pService.Stop()
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "offchain-middleware",
	Short: "Offchain middleware for signature aggregation",
	Long:  `A P2P service for collecting and aggregating signatures for Ethereum contracts.`,
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the offchain middleware service",
	Long:  `Start the offchain middleware service with the specified configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create config from viper
		// Create an instance of AppConfig
		var config Config
		// Unmarshal the config file into the AppConfig struct
		err := viper.Unmarshal(&config)
		if err != nil {
			log.Fatalf("Unable to decode into struct, %v", err)
		}

		if len(config.BlsPrivateKey) == 0 {
			log.Fatalf("Config is missing BLS private key")
		}

		config.ListenAddr = viper.GetString("listen")

		// Create application
		app := NewApp(config)

		// Create context with cancellation
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Initialize application
		if err := app.Initialize(ctx); err != nil {
			log.Fatalf("Failed to initialize application: %s", err)
		}

		// Start application
		if err := app.Start(); err != nil {
			log.Fatalf("Failed to start application: %s", err)
		}
		defer app.Stop()

		// Set up signal handling for graceful shutdown
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		// Wait for termination signal
		<-sigCh
		fmt.Println("Shutting down...")
	},
}

// generateConfigCmd represents the generate config command
var generateConfigCmd = &cobra.Command{
	Use:   "generate-config",
	Short: "Generate a default configuration file",
	Long:  `Generate a default configuration file with all available options.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Set default values
		viper.Set("eth", "http://localhost:8545")
		viper.Set("contract", "")
		ethPrivateKey, err := eth.GeneratePrivateKey()
		if err != nil {
			log.Fatalf("Failed to generate ETH private key: %s", err)
		}
		viper.Set("eth-private-key", ethPrivateKey)
		blsPrivateKey, err := bls.GenerateKey()
		if err != nil {
			log.Fatalf("Failed to generate BLS private key: %s", err)
		}
		viper.Set("bls-private-key", blsPrivateKey)
		viper.Set("peers", []string{})

		// Create config directory if it doesn't exist
		configDir := path.Dir(viper.ConfigFileUsed())
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			if err := os.MkdirAll(configDir, 0755); err != nil {
				log.Fatalf("Failed to create config directory: %s", err)
			}
			fmt.Printf("Created config directory: %s\n", configDir)
		}

		if err := viper.WriteConfig(); err != nil {
			log.Fatalf("Failed to write config: %s", err)
		}

		fmt.Printf("Configuration file generated at: %s\n", viper.ConfigFileUsed())
	},
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	viper.SetConfigType("yaml")

	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to find home directory: %s", err)
		}

		// Search config in home directory with name ".offchain-middleware" (without extension)
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func main() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.offchain-middleware.yaml)")

	// Start command flags
	startCmd.Flags().String("listen", "/ip4/127.0.0.1/tcp/8000", "Address to listen on")
	startCmd.Flags().Bool("test", false, "Test mode, use mock eth client")

	// Bind flags to viper
	viper.BindPFlag("listen", startCmd.Flags().Lookup("listen"))
	viper.BindPFlag("test", startCmd.Flags().Lookup("test"))

	// Add commands
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(generateConfigCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
