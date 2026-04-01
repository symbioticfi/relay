package operator

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	key_registerer "github.com/symbioticfi/relay/internal/usecase/key-registerer"
	"github.com/symbioticfi/relay/internal/usecase/metrics"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	symbioticCrypto "github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

func TestRegisterKeyArtifactECDSA(t *testing.T) {
	keyTag, err := symbiotic.KeyTagFromTypeAndId(symbiotic.KeyTypeEcdsaSecp256k1, 1)
	require.NoError(t, err)

	relayKeyBytes := make([]byte, 32)
	relayKeyBytes[len(relayKeyBytes)-1] = 0x7b

	relayKey, err := symbioticCrypto.NewPrivateKey(symbiotic.KeyTypeEcdsaSecp256k1, relayKeyBytes)
	require.NoError(t, err)

	keystorePath := createOperatorTestKeystore(t, keyTag, relayKey)
	operatorAddress := common.HexToAddress("0x0000000000000000000000000000000000000044")

	expectedArtifact := key_registerer.RegistrationArtifact{
		KeyTag:    keyTag,
		Key:       []byte{0x01, 0x02},
		Signature: []byte{0x03, 0x04},
	}

	restoreDeps := stubRegisterKeyDependencies(t, expectedArtifact)
	defer restoreDeps()

	output, err := runRegisterKeyArtifactCommand(t, keyTag, keystorePath, operatorAddress)
	require.NoError(t, err, output)

	var actual registrationArtifactJSON
	require.NoError(t, json.Unmarshal([]byte(output), &actual))
	require.Equal(t, registrationArtifactJSON{
		KeyTag:       uint8(expectedArtifact.KeyTag),
		KeyHex:       "0x0102",
		SignatureHex: "0x0304",
		ExtraDataHex: "0x",
	}, actual)
}

func TestRegisterKeyArtifactUsesOperatorAddress(t *testing.T) {
	keyTag, err := symbiotic.KeyTagFromTypeAndId(symbiotic.KeyTypeEcdsaSecp256k1, 1)
	require.NoError(t, err)

	relayKeyBytes := make([]byte, 32)
	relayKeyBytes[len(relayKeyBytes)-1] = 0x55

	relayKey, err := symbioticCrypto.NewPrivateKey(symbiotic.KeyTypeEcdsaSecp256k1, relayKeyBytes)
	require.NoError(t, err)

	keystorePath := createOperatorTestKeystore(t, keyTag, relayKey)
	operatorAddress := common.HexToAddress("0x00000000000000000000000000000000000000ab")

	expectedArtifact := key_registerer.RegistrationArtifact{
		KeyTag:    keyTag,
		Key:       []byte{0x11, 0x12},
		Signature: []byte{0x13, 0x14},
	}

	builder := &staticRegisterKeyArtifactBuilder{artifact: expectedArtifact}
	restoreDeps := stubRegisterKeyDependenciesWithBuilder(t, builder)
	defer restoreDeps()

	output, err := runRegisterKeyArtifactCommand(t, keyTag, keystorePath, operatorAddress)
	require.NoError(t, err, output)

	var actual registrationArtifactJSON
	require.NoError(t, json.Unmarshal([]byte(output), &actual))
	require.Equal(t, registrationArtifactJSON{
		KeyTag:       uint8(expectedArtifact.KeyTag),
		KeyHex:       "0x1112",
		SignatureHex: "0x1314",
		ExtraDataHex: "0x",
	}, actual)
	require.Equal(t, operatorAddress, builder.operatorAddress)
	require.True(t, builder.buildArtifactCalled)
	require.False(t, builder.registerCalled)
}

func TestRegisterKeyArtifactRequiresOperatorAddress(t *testing.T) {
	keyTag, err := symbiotic.KeyTagFromTypeAndId(symbiotic.KeyTypeEcdsaSecp256k1, 1)
	require.NoError(t, err)

	relayKeyBytes := make([]byte, 32)
	relayKeyBytes[len(relayKeyBytes)-1] = 0x56

	relayKey, err := symbioticCrypto.NewPrivateKey(symbiotic.KeyTypeEcdsaSecp256k1, relayKeyBytes)
	require.NoError(t, err)

	keystorePath := createOperatorTestKeystore(t, keyTag, relayKey)
	builder := &staticRegisterKeyArtifactBuilder{}
	restoreDeps := stubRegisterKeyDependenciesWithBuilder(t, builder)
	defer restoreDeps()

	output, err := runOperatorCommand(t,
		"register-key-artifact",
		"--key-tag", keyTagArg(keyTag),
		"--path", keystorePath,
		"--password", testKeystorePassword,
	)
	require.EqualError(t, err, "--operator-address is required")
	require.Contains(t, output, "--operator-address is required")
	require.False(t, builder.buildArtifactCalled)
	require.False(t, builder.registerCalled)
}

func TestRegisterKeyArtifactBLSIncludesExtraData(t *testing.T) {
	keyTag, err := symbiotic.KeyTagFromTypeAndId(symbiotic.KeyTypeBlsBn254, 1)
	require.NoError(t, err)

	relayKey, err := symbioticCrypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, []byte("bls-private-key-material"))
	require.NoError(t, err)

	keystorePath := createOperatorTestKeystore(t, keyTag, relayKey)
	operatorAddress := common.HexToAddress("0x0000000000000000000000000000000000000045")

	expectedArtifact := key_registerer.RegistrationArtifact{
		KeyTag:    keyTag,
		Key:       []byte{0x05, 0x06},
		Signature: []byte{0x07, 0x08},
		ExtraData: []byte{0x09, 0x0a},
	}

	restoreDeps := stubRegisterKeyDependencies(t, expectedArtifact)
	defer restoreDeps()

	output, err := runRegisterKeyArtifactCommand(t, keyTag, keystorePath, operatorAddress)
	require.NoError(t, err, output)

	var actual registrationArtifactJSON
	require.NoError(t, json.Unmarshal([]byte(output), &actual))
	require.Equal(t, registrationArtifactJSON{
		KeyTag:       uint8(expectedArtifact.KeyTag),
		KeyHex:       "0x0506",
		SignatureHex: "0x0708",
		ExtraDataHex: "0x090a",
	}, actual)
}

func TestRegisterKeyArtifactBLS12381IncludesExtraData(t *testing.T) {
	keyTag, err := symbiotic.KeyTagFromTypeAndId(symbiotic.KeyTypeBls12381, 1)
	require.NoError(t, err)

	relayKey, err := symbioticCrypto.NewPrivateKey(symbiotic.KeyTypeBls12381, []byte("bls12381-private-key-material"))
	require.NoError(t, err)

	keystorePath := createOperatorTestKeystore(t, keyTag, relayKey)
	operatorAddress := common.HexToAddress("0x0000000000000000000000000000000000000046")

	expectedArtifact := key_registerer.RegistrationArtifact{
		KeyTag:    keyTag,
		Key:       []byte{0x0b, 0x0c},
		Signature: []byte{0x0d, 0x0e},
		ExtraData: []byte{0x0f, 0x10},
	}

	restoreDeps := stubRegisterKeyDependencies(t, expectedArtifact)
	defer restoreDeps()

	output, err := runRegisterKeyArtifactCommand(t, keyTag, keystorePath, operatorAddress)
	require.NoError(t, err, output)

	var actual registrationArtifactJSON
	require.NoError(t, json.Unmarshal([]byte(output), &actual))
	require.Equal(t, registrationArtifactJSON{
		KeyTag:       uint8(expectedArtifact.KeyTag),
		KeyHex:       "0x0b0c",
		SignatureHex: "0x0d0e",
		ExtraDataHex: "0x0f10",
	}, actual)
}

func stubRegisterKeyDependencies(t *testing.T, artifact key_registerer.RegistrationArtifact) func() {
	t.Helper()

	builder := &staticRegisterKeyArtifactBuilder{artifact: artifact}
	return stubRegisterKeyDependenciesWithBuilder(t, builder)
}

func stubRegisterKeyDependenciesWithBuilder(t *testing.T, builder *staticRegisterKeyArtifactBuilder) func() {
	t.Helper()

	originalFactory := newRegisterKeyArtifactBuilder
	originalMetricsFactory := newRegisterKeyArtifactMetrics

	newRegisterKeyArtifactMetrics = func() *metrics.Metrics {
		return metrics.New(metrics.Config{Registerer: prometheus.NewRegistry()})
	}
	newRegisterKeyArtifactBuilder = func(registerKeyBuildConfig) (registerKeyArtifactBuilder, error) {
		return builder, nil
	}

	return func() {
		newRegisterKeyArtifactMetrics = originalMetricsFactory
		newRegisterKeyArtifactBuilder = originalFactory
	}
}

type staticRegisterKeyArtifactBuilder struct {
	artifact            key_registerer.RegistrationArtifact
	buildArtifactCalled bool
	registerCalled      bool
	operatorAddress     common.Address
}

type registrationArtifactJSON struct {
	KeyTag       uint8  `json:"keyTag"`
	KeyHex       string `json:"keyHex"`
	SignatureHex string `json:"signatureHex"`
	ExtraDataHex string `json:"extraDataHex"`
}

func (s *staticRegisterKeyArtifactBuilder) BuildArtifact(
	_ context.Context,
	_ symbioticCrypto.PrivateKey,
	_ symbiotic.KeyTag,
	operatorAddress common.Address,
) (key_registerer.RegistrationArtifact, error) {
	s.buildArtifactCalled = true
	s.operatorAddress = operatorAddress
	return s.artifact, nil
}

func (s *staticRegisterKeyArtifactBuilder) Register(
	_ context.Context,
	_ symbioticCrypto.PrivateKey,
	_ symbiotic.KeyTag,
	operatorAddress common.Address,
) (symbiotic.TxResult, error) {
	s.registerCalled = true
	s.operatorAddress = operatorAddress
	return symbiotic.TxResult{}, nil
}

func createOperatorTestKeystore(t *testing.T, keyTag symbiotic.KeyTag, privateKey symbioticCrypto.PrivateKey) string {
	t.Helper()

	path := t.TempDir() + "/keystore.jks"
	keyStore, err := keyprovider.NewKeystoreProvider(path, testKeystorePassword)
	require.NoError(t, err)
	require.NoError(t, keyStore.AddKey(keyprovider.SYMBIOTIC_KEY_NAMESPACE, keyTag, privateKey, testKeystorePassword, false))
	return path
}

func runOperatorCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()

	registerKeyFlags = RegisterKeyFlags{}
	registerKeyArtifactFlags = RegisterKeyArtifactFlags{}
	rpcServer := newTestRegisterKeyRPCServer(t)

	testOperatorCmdOnce.Do(func() {
		operatorCmd = &cobra.Command{
			Use:   "operator",
			Short: "Operator tool",
		}
		testOperatorCmd = NewOperatorCmd()
	})
	cmd := testOperatorCmd
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(out)
	baseArgs := []string{
		"--chains", rpcServer.URL,
		"--driver.address", "0x0000000000000000000000000000000000000001",
		"--driver.chainid", "1",
		"--voting-provider-chain-id", "1",
	}
	cmd.SetArgs(append(baseArgs, args...))
	cmd.SetIn(bytes.NewReader(nil))

	err := cmd.Execute()
	return out.String(), err
}

func runRegisterKeyArtifactCommand(t *testing.T, keyTag symbiotic.KeyTag, keystorePath string, operatorAddress common.Address) (string, error) {
	t.Helper()

	return runOperatorCommand(t,
		"register-key-artifact",
		"--key-tag", keyTagArg(keyTag),
		"--path", keystorePath,
		"--password", testKeystorePassword,
		"--operator-address", operatorAddress.Hex(),
	)
}

func newTestRegisterKeyRPCServer(t *testing.T) *httptest.Server {
	t.Helper()

	registerKeyRPCServerOnce.Do(func() {
		registerKeyRPCServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var req struct {
				ID     json.RawMessage `json:"id"`
				Method string          `json:"method"`
			}
			if err := json.Unmarshal(body, &req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			var result any
			switch req.Method {
			case "eth_chainId":
				result = "0x1"
			case "eth_maxPriorityFeePerGas":
				result = "0x3b9aca00"
			default:
				result = "0x1"
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result":  result,
			}); err != nil {
				t.Errorf("encode rpc response: %v", err)
			}
		}))
	})
	return registerKeyRPCServer
}

var (
	registerKeyRPCServerOnce sync.Once
	registerKeyRPCServer     *httptest.Server
	testOperatorCmdOnce      sync.Once
	testOperatorCmd          *cobra.Command
)

func keyTagArg(keyTag symbiotic.KeyTag) string {
	return strconv.FormatUint(uint64(keyTag), 10)
}

const testKeystorePassword = "password"
