package tests

import (
	"encoding/json"
	"os"

	"github.com/go-errors/errors"
	"github.com/kelseyhightower/envconfig"
	"github.com/symbioticfi/relay/core/entity"
	symbioticCrypto "github.com/symbioticfi/relay/core/usecase/crypto"
)

var (
	settlementChains = []string{
		"http://localhost:8545", // Main anvil chain
		"http://localhost:8546", // Settlement anvil chain
	}
)

// RelayServerEndpoint represents a relay server endpoint for testing
type RelayServerEndpoint struct {
	Address string
	Port    int
	Role    string // "committer", "aggregator", "signer"
}

// ContractAddress represents a contract address with chain ID
type ContractAddress struct {
	Addr    string `json:"addr"`
	ChainId uint64 `json:"chainId"`
}

// RelayContractsData represents the structure from relay_contracts.json
type RelayContractsData struct {
	Driver               ContractAddress   `json:"driver"`
	KeyRegistry          ContractAddress   `json:"keyRegisztry"` // Note: typo in original JSON
	Network              string            `json:"network"`
	Settlements          []ContractAddress `json:"settlements"`
	StakingTokens        []ContractAddress `json:"stakingTokens"`
	VotingPowerProviders []ContractAddress `json:"votingPowerProviders"`

	Env EnvInfo `json:"-"`
}

type EnvInfo struct {
	Operators        int64 `default:"4" split_words:"true"`
	Commiters        int64 `default:"1" split_words:"true"`
	Aggregators      int64 `default:"1" split_words:"true"`
	EpochTime        int64 `default:"30" split_words:"true"`
	VerificationType uint8 `default:"1" split_words:"true"`
}

// GetDriverAddress returns the driver address as a string for backward compatibility
func (d *RelayContractsData) GetDriverAddress() string {
	return d.Driver.Addr
}

// GetKeyRegistryAddress returns the key registry address as a string
func (d *RelayContractsData) GetKeyRegistryAddress() string {
	return d.KeyRegistry.Addr
}

// GetVotingPowerProviderAddress returns the first voting power provider address
func (d *RelayContractsData) GetVotingPowerProviderAddress() string {
	if len(d.VotingPowerProviders) > 0 {
		return d.VotingPowerProviders[0].Addr
	}
	return ""
}

// GetSettlementAddresses returns all settlement addresses
func (d *RelayContractsData) GetSettlementAddresses() []ContractAddress {
	return d.Settlements
}

func getRelayEndpoints(env EnvInfo) []RelayServerEndpoint {
	commiterCount := env.Commiters
	aggregatorCount := env.Aggregators
	endpoints := make([]RelayServerEndpoint, env.Operators)
	for i := 0; i < int(env.Operators); i++ {
		role := "signer"
		if commiterCount > 0 {
			role = "committer"
			commiterCount--
		} else if aggregatorCount > 0 {
			role = "aggregator"
			aggregatorCount--
		}
		endpoints[i] = RelayServerEndpoint{
			Address: "localhost",
			Port:    8081 + i,
			Role:    role,
		}
	}

	return endpoints
}

func loadDeploymentData() (*RelayContractsData, error) {
	path := "../temp-network/deploy-data/relay_contracts.json"
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Errorf("failed to read %s: %w", path, err)
	}
	var relayContracts RelayContractsData
	if err := json.Unmarshal(data, &relayContracts); err != nil {
		return nil, errors.Errorf("failed to unmarshal %s: %w", path, err)
	}
	relayContracts.Env = EnvInfo{}
	if err := envconfig.Process("", &relayContracts.Env); err != nil {
		return nil, errors.Errorf("failed to process environment variables: %w", err)
	}
	return &relayContracts, nil
}

// testMockKeyProvider is a mock key provider for testing
type testMockKeyProvider struct{}

func (m *testMockKeyProvider) GetPrivateKey(_ entity.KeyTag) (symbioticCrypto.PrivateKey, error) {
	return nil, errors.New("mock key provider - no keys available")
}

func (m *testMockKeyProvider) GetPrivateKeyByAlias(_ string) (symbioticCrypto.PrivateKey, error) {
	return nil, errors.New("mock key provider - no keys available")
}

func (m *testMockKeyProvider) GetPrivateKeyByNamespaceTypeId(_ string, _ entity.KeyType, _ int) (symbioticCrypto.PrivateKey, error) {
	return nil, errors.New("mock key provider - no keys available")
}

func (m *testMockKeyProvider) HasKey(_ entity.KeyTag) (bool, error) {
	return false, nil
}

func (m *testMockKeyProvider) HasKeyByAlias(_ string) (bool, error) {
	return false, nil
}

func (m *testMockKeyProvider) HasKeyByNamespaceTypeId(_ string, _ entity.KeyType, _ int) (bool, error) {
	return false, nil
}
