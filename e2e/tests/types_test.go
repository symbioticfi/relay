package tests

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/go-errors/errors"
	"github.com/kelseyhightower/envconfig"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	symbioticCrypto "github.com/symbioticfi/relay/symbiotic/usecase/crypto"
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
	Port    string
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
	KeyRegistry          ContractAddress   `json:"keyRegistry"`
	Network              string            `json:"network"`
	Settlements          []ContractAddress `json:"settlements"`
	StakingTokens        []ContractAddress `json:"stakingTokens"`
	VotingPowerProviders []ContractAddress `json:"votingPowerProviders"`

	Env EnvInfo `json:"-"`
}

type EnvInfo struct {
	Operators        int64  `default:"4" split_words:"true"`
	Commiters        int64  `default:"1" split_words:"true"`
	Aggregators      int64  `default:"1" split_words:"true"`
	EpochTime        uint64 `default:"30" split_words:"true"`
	VerificationType uint32 `default:"1" split_words:"true"`
}

// GetDriverAddress returns the driver address as a string for backward compatibility
func (d RelayContractsData) GetDriverAddress() string {
	return d.Driver.Addr
}

// GetKeyRegistryAddress returns the key registry address as a string
func (d RelayContractsData) GetKeyRegistryAddress() string {
	return d.KeyRegistry.Addr
}

// GetVotingPowerProviderAddress returns the first voting power provider address
func (d RelayContractsData) GetVotingPowerProviderAddress() string {
	if len(d.VotingPowerProviders) > 0 {
		return d.VotingPowerProviders[0].Addr
	}
	return ""
}

// GetSettlementAddresses returns all settlement addresses
func (d RelayContractsData) GetSettlementAddresses() []ContractAddress {
	return d.Settlements
}

func loadDeploymentData(ctx context.Context) (RelayContractsData, error) {
	path := "../temp-network/deploy-data/relay_contracts.json"

	// Wait for relay_contracts.json to be created by shell script
	const maxWaitTime = 60 * time.Second
	const checkInterval = 500 * time.Millisecond
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return RelayContractsData{}, errors.Errorf("context cancelled while waiting for %s: %w", path, ctx.Err())
		default:
		}

		if _, err := os.Stat(path); err == nil {
			break // File exists, break the loop
		}

		if time.Since(startTime) > maxWaitTime {
			return RelayContractsData{}, errors.Errorf("timeout waiting for %s to be created after %v", path, maxWaitTime)
		}

		select {
		case <-ctx.Done():
			return RelayContractsData{}, errors.Errorf("context cancelled while waiting for %s: %w", path, ctx.Err())
		case <-time.After(checkInterval):
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return RelayContractsData{}, errors.Errorf("failed to read %s: %w", path, err)
	}

	var relayContracts RelayContractsData
	if err := json.Unmarshal(data, &relayContracts); err != nil {
		return RelayContractsData{}, errors.Errorf("failed to unmarshal %s: %w", path, err)
	}

	relayContracts.Env = EnvInfo{}
	if err := envconfig.Process("", &relayContracts.Env); err != nil {
		return RelayContractsData{}, errors.Errorf("failed to process environment variables: %w", err)
	}

	return relayContracts, nil
}

// testMockKeyProvider is a mock key provider for testing
type testMockKeyProvider struct{}

func (m *testMockKeyProvider) GetPrivateKey(_ symbiotic.KeyTag) (symbioticCrypto.PrivateKey, error) {
	return nil, errors.New("mock key provider - no keys available")
}

func (m *testMockKeyProvider) GetPrivateKeyByAlias(_ string) (symbioticCrypto.PrivateKey, error) {
	return nil, errors.New("mock key provider - no keys available")
}

func (m *testMockKeyProvider) GetPrivateKeyByNamespaceTypeId(_ string, _ symbiotic.KeyType, _ int) (symbioticCrypto.PrivateKey, error) {
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
