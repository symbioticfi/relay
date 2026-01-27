package tests

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/go-errors/errors"
	"github.com/kelseyhightower/envconfig"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	apiv1 "github.com/symbioticfi/relay/api/client/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

var (
	settlementChains = []string{
		"http://localhost:8545", // Main anvil chain
		"http://localhost:8546", // Settlement anvil chain
	}
)

// ContractAddress represents a contract address with chain ID
type ContractAddress struct {
	Addr    string `json:"addr"`
	ChainId uint64 `json:"chainId"`
}

// ChainAddresses holds all contract addresses for a chain
type ChainAddresses struct {
	KeyRegistry                 string `json:"key_registry,omitempty"`
	VaultFactory                string `json:"vault_factory"`
	DelegatorFactory            string `json:"delegator_factory"`
	SlasherFactory              string `json:"slasher_factory"`
	NetworkRegistry             string `json:"network_registry"`
	OperatorRegistry            string `json:"operator_registry"`
	OperatorMetadataService     string `json:"operator_metadata_service"`
	NetworkMetadataService      string `json:"network_metadata_service"`
	NetworkMiddlewareService    string `json:"network_middleware_service"`
	OperatorVaultOptInService   string `json:"operator_vault_opt_in_service"`
	OperatorNetworkOptInService string `json:"operator_network_opt_in_service"`
	VaultConfigurator           string `json:"vault_configurator"`
	Network                     string `json:"network"`
	StakingToken                string `json:"staking_token,omitempty"`
	VotingPowerProvider         string `json:"voting_power_provider,omitempty"`
	Settlement                  string `json:"settlement"`
	SumTask                     string `json:"sum_task"`
	ValSetDriver                string `json:"val_set_driver,omitempty"`
}

// ChainConfig holds configuration for a specific chain
type ChainConfig struct {
	ChainID     uint64         `json:"chain_id"`
	EndpointURL string         `json:"endpoint_url"`
	Addresses   ChainAddresses `json:"addresses"`
}

// NetworkConfig holds the network-level configuration
type NetworkConfig struct {
	NetworkID           uint64   `json:"network_id"`
	EndpointURL         string   `json:"endpoint_url"`
	KeyRegistry         uint64   `json:"key_registry"`
	VotingPowerProvider []uint64 `json:"voting_power_provider"`
	Settlement          []uint64 `json:"settlement"`
	ValSetDriver        uint64   `json:"val_set_driver"`
}

// RelayContractsData represents the structure from relay_contracts.json
type RelayContractsData struct {
	Driver ContractAddress `json:"driver"`

	// Parsed chains from TOML
	MainChain       ChainConfig   `json:"main_chain"`
	SettlementChain ChainConfig   `json:"settlement_chain"`
	Network         NetworkConfig `json:"network"`

	Env EnvInfo `json:"-"`
}

type EnvInfo struct {
	Operators        int64  `default:"4" split_words:"true"`
	Commiters        int64  `default:"1" split_words:"true"`
	Aggregators      int64  `default:"1" split_words:"true"`
	EpochTime        uint64 `default:"30" split_words:"true"`
	VerificationType uint32 `default:"1" split_words:"true"`
}

type RelaySidecarConfig struct {
	ContainerName  string
	RequiredSymKey crypto.PrivateKey
}

func (i EnvInfo) GetSidecarConfigs() []RelaySidecarConfig {
	const basePrivateKey = 1000000000000000000

	configs := make([]RelaySidecarConfig, 0, i.Operators)
	for op := range i.Operators {
		keyIndex := op
		symbPrivateKeyDecimal := basePrivateKey + keyIndex
		symbPrivateKeyHex := fmt.Sprintf("%064x", symbPrivateKeyDecimal)

		privBytes, err := hex.DecodeString(symbPrivateKeyHex)
		if err != nil {
			panic(fmt.Sprintf("failed to decode symb private key hex: %v", err))
		}
		symbKey, err := crypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, privBytes)
		if err != nil {
			panic(fmt.Sprintf("failed to create symb private key: %v", err))
		}

		configs = append(configs, RelaySidecarConfig{
			RequiredSymKey: symbKey,
			ContainerName:  fmt.Sprintf("relay-sidecar-%d", op+1),
		})
	}
	return configs
}

// GetDriverAddress returns the driver address as a string for backward compatibility
func (d RelayContractsData) GetDriverAddress() string {
	return d.Driver.Addr
}

// tomlDeploymentData represents the structure of my-relay-deploy.toml
type tomlDeploymentData struct {
	Chain31337 tomlChainData   `toml:"31337"`
	Chain31338 tomlChainData   `toml:"31338"`
	Network    tomlNetworkData `toml:"1234567890"`
}

type tomlChainData struct {
	EndpointURL string        `toml:"endpoint_url"`
	Address     tomlAddresses `toml:"address"`
}

type tomlAddresses struct {
	KeyRegistry                 string `toml:"key_registry"`
	VaultFactory                string `toml:"vault_factory"`
	DelegatorFactory            string `toml:"delegator_factory"`
	SlasherFactory              string `toml:"slasher_factory"`
	NetworkRegistry             string `toml:"network_registry"`
	OperatorRegistry            string `toml:"operator_registry"`
	OperatorMetadataService     string `toml:"operator_metadata_service"`
	NetworkMetadataService      string `toml:"network_metadata_service"`
	NetworkMiddlewareService    string `toml:"network_middleware_service"`
	OperatorVaultOptInService   string `toml:"operator_vault_opt_in_service"`
	OperatorNetworkOptInService string `toml:"operator_network_opt_in_service"`
	VaultConfigurator           string `toml:"vault_configurator"`
	Network                     string `toml:"network"`
	StakingToken                string `toml:"staking_token"`
	VotingPowerProvider         string `toml:"voting_power_provider"`
	Settlement                  string `toml:"settlement"`
	SumTask                     string `toml:"sum_task"`
	ValSetDriver                string `toml:"val_set_driver"`
}

type tomlNetworkData struct {
	EndpointURL         string   `toml:"endpoint_url"`
	KeyRegistry         uint64   `toml:"keyRegistry"`
	VotingPowerProvider []uint64 `toml:"votingPowerProvider"`
	Settlement          []uint64 `toml:"settlement"`
	ValSetDriver        uint64   `toml:"valSetDriver"`
}

func loadDeploymentData(t *testing.T) RelayContractsData {
	t.Helper()

	// Read TOML file from temp-network directory
	tomlPath := filepath.Join("..", "temp-network", "my-relay-deploy.toml")
	data, err := os.ReadFile(tomlPath)
	require.NoError(t, err, "Failed to read deployment TOML file")

	// Parse TOML
	var deployData tomlDeploymentData
	err = toml.Unmarshal(data, &deployData)
	require.NoError(t, err, "Failed to parse deployment TOML")

	// Extract driver address from chain 31337
	require.NotEmpty(t, deployData.Chain31337.Address.ValSetDriver, "val_set_driver address not found in deployment TOML for chain 31337")

	relayContracts := RelayContractsData{
		Driver: ContractAddress{
			Addr:    deployData.Chain31337.Address.ValSetDriver,
			ChainId: 31337,
		},
		MainChain: ChainConfig{
			ChainID:     31337,
			EndpointURL: deployData.Chain31337.EndpointURL,
			Addresses: ChainAddresses{
				KeyRegistry:                 deployData.Chain31337.Address.KeyRegistry,
				VaultFactory:                deployData.Chain31337.Address.VaultFactory,
				DelegatorFactory:            deployData.Chain31337.Address.DelegatorFactory,
				SlasherFactory:              deployData.Chain31337.Address.SlasherFactory,
				NetworkRegistry:             deployData.Chain31337.Address.NetworkRegistry,
				OperatorRegistry:            deployData.Chain31337.Address.OperatorRegistry,
				OperatorMetadataService:     deployData.Chain31337.Address.OperatorMetadataService,
				NetworkMetadataService:      deployData.Chain31337.Address.NetworkMetadataService,
				NetworkMiddlewareService:    deployData.Chain31337.Address.NetworkMiddlewareService,
				OperatorVaultOptInService:   deployData.Chain31337.Address.OperatorVaultOptInService,
				OperatorNetworkOptInService: deployData.Chain31337.Address.OperatorNetworkOptInService,
				VaultConfigurator:           deployData.Chain31337.Address.VaultConfigurator,
				Network:                     deployData.Chain31337.Address.Network,
				StakingToken:                deployData.Chain31337.Address.StakingToken,
				VotingPowerProvider:         deployData.Chain31337.Address.VotingPowerProvider,
				Settlement:                  deployData.Chain31337.Address.Settlement,
				SumTask:                     deployData.Chain31337.Address.SumTask,
				ValSetDriver:                deployData.Chain31337.Address.ValSetDriver,
			},
		},
		SettlementChain: ChainConfig{
			ChainID:     31338,
			EndpointURL: deployData.Chain31338.EndpointURL,
			Addresses: ChainAddresses{
				VaultFactory:                deployData.Chain31338.Address.VaultFactory,
				DelegatorFactory:            deployData.Chain31338.Address.DelegatorFactory,
				SlasherFactory:              deployData.Chain31338.Address.SlasherFactory,
				NetworkRegistry:             deployData.Chain31338.Address.NetworkRegistry,
				OperatorRegistry:            deployData.Chain31338.Address.OperatorRegistry,
				OperatorMetadataService:     deployData.Chain31338.Address.OperatorMetadataService,
				NetworkMetadataService:      deployData.Chain31338.Address.NetworkMetadataService,
				NetworkMiddlewareService:    deployData.Chain31338.Address.NetworkMiddlewareService,
				OperatorVaultOptInService:   deployData.Chain31338.Address.OperatorVaultOptInService,
				OperatorNetworkOptInService: deployData.Chain31338.Address.OperatorNetworkOptInService,
				VaultConfigurator:           deployData.Chain31338.Address.VaultConfigurator,
				Network:                     deployData.Chain31338.Address.Network,
				Settlement:                  deployData.Chain31338.Address.Settlement,
				SumTask:                     deployData.Chain31338.Address.SumTask,
			},
		},
		Network: NetworkConfig{
			NetworkID:           1234567890,
			EndpointURL:         deployData.Network.EndpointURL,
			KeyRegistry:         deployData.Network.KeyRegistry,
			VotingPowerProvider: deployData.Network.VotingPowerProvider,
			Settlement:          deployData.Network.Settlement,
			ValSetDriver:        deployData.Network.ValSetDriver,
		},
	}

	relayContracts.Env = EnvInfo{}
	err = envconfig.Process("", &relayContracts.Env)
	require.NoError(t, err, "Failed to process environment variables")

	return relayContracts
}

// testMockKeyProvider is a mock key provider for testing
type testMockKeyProvider struct{}

func (m *testMockKeyProvider) GetPrivateKey(_ symbiotic.KeyTag) (crypto.PrivateKey, error) {
	return nil, errors.New("mock key provider - no keys available")
}

func (m *testMockKeyProvider) GetPrivateKeyByAlias(_ string) (crypto.PrivateKey, error) {
	return nil, errors.New("mock key provider - no keys available")
}

func (m *testMockKeyProvider) GetPrivateKeyByNamespaceTypeId(_ string, _ symbiotic.KeyType, _ int) (crypto.PrivateKey, error) {
	return nil, errors.New("mock key provider - no keys available")
}

func (m *testMockKeyProvider) HasKey(_ symbiotic.KeyTag) (bool, error) {
	return false, nil
}

func (m *testMockKeyProvider) HasKeyByAlias(_ string) (bool, error) {
	return false, nil
}

func (m *testMockKeyProvider) HasKeyByNamespaceTypeId(_ string, _ symbiotic.KeyType, _ int) (bool, error) {
	return false, nil
}

func getContainerPort(i int) int {
	return 8081 + i
}

func getGRPCClient(t *testing.T, index int) *apiv1.SymbioticClient {
	t.Helper()
	address := "localhost:" + strconv.Itoa(getContainerPort(index))
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoErrorf(t, err, "Failed to connect to relay server at %s", address)
	t.Cleanup(func() {
		conn.Close()
	})

	return apiv1.NewSymbioticClient(conn)
}

func getHealthEndpoint(i int) string {
	return fmt.Sprintf("http://localhost:%d/healthz", getContainerPort(i))
}

func startContainer(ctx context.Context, container string) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", "../temp-network/docker-compose.yml", "restart", container)
	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.Errorf("failed to start container: %v, %s", err, output)
	}
	return nil
}

func stopContainer(ctx context.Context, container string) error {
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", "../temp-network/docker-compose.yml", "stop", container)
	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.Errorf("failed to stop container: %v, %s", err, output)
	}
	return nil
}
