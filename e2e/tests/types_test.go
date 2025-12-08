package tests

import (
	"context"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strconv"
	"testing"

	"github.com/go-errors/errors"
	"github.com/kelseyhightower/envconfig"
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

// RelayContractsData represents the structure from relay_contracts.json
type RelayContractsData struct {
	Driver ContractAddress `json:"driver"`

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

func loadDeploymentData() (RelayContractsData, error) {
	relayContracts := RelayContractsData{
		Driver: ContractAddress{
			Addr:    "0x43C27243F96591892976FFf886511807B65a33d5",
			ChainId: 31337,
		},
	}

	relayContracts.Env = EnvInfo{}
	if err := envconfig.Process("", &relayContracts.Env); err != nil {
		return RelayContractsData{}, errors.Errorf("failed to process environment variables: %w", err)
	}

	return relayContracts, nil
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
