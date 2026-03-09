package keys

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"

	keyprovider "github.com/symbioticfi/relay/internal/usecase/key-provider"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	relaycrypto "github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

func TestKeysSignECDSARelayKeyOutputsNormalizedSignature(t *testing.T) {
	t.Parallel()

	keyTag, err := symbiotic.KeyTagFromTypeAndId(symbiotic.KeyTypeEcdsaSecp256k1, 1)
	require.NoError(t, err)

	keyBytes := make([]byte, 32)
	keyBytes[len(keyBytes)-1] = 0x7b
	privateKey, err := relaycrypto.NewPrivateKey(symbiotic.KeyTypeEcdsaSecp256k1, keyBytes)
	require.NoError(t, err)

	keystorePath := createRelayKeyStore(t, keyTag, privateKey)
	messageHex := "0x01020304aabbcc"
	messageBytes, err := hexutil.Decode(messageHex)
	require.NoError(t, err)

	expectedSig, _, err := privateKey.Sign(messageBytes)
	require.NoError(t, err)
	expectedSig = append([]byte(nil), expectedSig...)
	expectedSig[64] += 27

	output, err := runUtilsCLI(t,
		"keys", "sign",
		"--key-tag", keyTagArg(keyTag),
		"--message-hex", messageHex,
		"--path", keystorePath,
		"--password", testKeystorePassword,
	)
	require.NoError(t, err, output)

	actualSig, err := hexutil.Decode(strings.TrimSpace(output))
	require.NoError(t, err)
	require.Len(t, actualSig, 65)
	require.Contains(t, []byte{27, 28}, actualSig[64])
	require.Equal(t, []byte(expectedSig[:64]), actualSig[:64])
	require.Equal(t, expectedSig[64], actualSig[64])
}

func TestKeysSignBLSRelayKeyOutputsVerifiableSignature(t *testing.T) {
	t.Parallel()

	keyTag, err := symbiotic.KeyTagFromTypeAndId(symbiotic.KeyTypeBlsBn254, 1)
	require.NoError(t, err)

	privateKey, err := relaycrypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, []byte("bls-private-key-material"))
	require.NoError(t, err)

	keystorePath := createRelayKeyStore(t, keyTag, privateKey)
	messageHex := "0xdeadbeef"
	messageBytes, err := hexutil.Decode(messageHex)
	require.NoError(t, err)

	output, err := runUtilsCLI(t,
		"keys", "sign",
		"--key-tag", keyTagArg(keyTag),
		"--message-hex", messageHex,
		"--path", keystorePath,
		"--password", testKeystorePassword,
	)
	require.NoError(t, err, output)

	signature, err := hexutil.Decode(strings.TrimSpace(output))
	require.NoError(t, err)
	require.NotEmpty(t, signature)
	require.NoError(t, privateKey.PublicKey().Verify(messageBytes, signature))
}

func TestKeysSignBLS12381RelayKeyOutputsVerifiableSignature(t *testing.T) {
	t.Parallel()

	keyTag, err := symbiotic.KeyTagFromTypeAndId(symbiotic.KeyTypeBls12381, 1)
	require.NoError(t, err)

	privateKey, err := relaycrypto.NewPrivateKey(symbiotic.KeyTypeBls12381, []byte("bls12381-private-key-material"))
	require.NoError(t, err)

	keystorePath := createRelayKeyStore(t, keyTag, privateKey)
	messageHex := "0xcafebabe"
	messageBytes, err := hexutil.Decode(messageHex)
	require.NoError(t, err)

	output, err := runUtilsCLI(t,
		"keys", "sign",
		"--key-tag", keyTagArg(keyTag),
		"--message-hex", messageHex,
		"--path", keystorePath,
		"--password", testKeystorePassword,
	)
	require.NoError(t, err, output)

	signature, err := hexutil.Decode(strings.TrimSpace(output))
	require.NoError(t, err)
	require.NotEmpty(t, signature)
	require.NoError(t, privateKey.PublicKey().Verify(messageBytes, signature))
}

func TestKeysSignValidationErrors(t *testing.T) {
	t.Parallel()

	keyTag, err := symbiotic.KeyTagFromTypeAndId(symbiotic.KeyTypeBlsBn254, 1)
	require.NoError(t, err)

	t.Run("legacy relay flag is rejected", func(t *testing.T) {
		t.Parallel()

		output, err := runUtilsCLI(t,
			"keys", "sign",
			"--relay",
			"--key-tag", keyTagArg(keyTag),
			"--message-hex", "0x01",
			"--path", filepath.Join(t.TempDir(), "keystore.jks"),
			"--password", testKeystorePassword,
		)
		require.Error(t, err)
		require.Contains(t, output, "unknown flag: --relay")
	})

	t.Run("missing key tag", func(t *testing.T) {
		t.Parallel()

		output, err := runUtilsCLI(t,
			"keys", "sign",
			"--message-hex", "0x01",
			"--path", filepath.Join(t.TempDir(), "keystore.jks"),
			"--password", testKeystorePassword,
		)
		require.Error(t, err)
		require.Contains(t, output, "key tag is required")
	})

	t.Run("invalid key tag", func(t *testing.T) {
		t.Parallel()

		output, err := runUtilsCLI(t,
			"keys", "sign",
			"--key-tag", strconv.FormatUint(uint64(symbiotic.KeyTypeInvalid), 10),
			"--message-hex", "0x01",
			"--path", filepath.Join(t.TempDir(), "keystore.jks"),
			"--password", testKeystorePassword,
		)
		require.Error(t, err)
		require.Contains(t, output, "invalid key tag, type not supported")
	})

	t.Run("malformed message hex", func(t *testing.T) {
		t.Parallel()

		output, err := runUtilsCLI(t,
			"keys", "sign",
			"--key-tag", keyTagArg(keyTag),
			"--message-hex", "not-hex",
			"--path", filepath.Join(t.TempDir(), "keystore.jks"),
			"--password", testKeystorePassword,
		)
		require.Error(t, err)
		require.Contains(t, output, "invalid message hex")
	})

	t.Run("empty message hex", func(t *testing.T) {
		t.Parallel()

		output, err := runUtilsCLI(t,
			"keys", "sign",
			"--key-tag", keyTagArg(keyTag),
			"--message-hex", "0x",
			"--path", filepath.Join(t.TempDir(), "keystore.jks"),
			"--password", testKeystorePassword,
		)
		require.Error(t, err)
		require.Contains(t, output, "message hex cannot be empty")
	})

	t.Run("missing relay key in keystore", func(t *testing.T) {
		t.Parallel()

		output, err := runUtilsCLI(t,
			"keys", "sign",
			"--key-tag", keyTagArg(keyTag),
			"--message-hex", "0x01",
			"--path", filepath.Join(t.TempDir(), "keystore.jks"),
			"--password", testKeystorePassword,
		)
		require.Error(t, err)
		require.Contains(t, output, "key not found")
	})
}

const testKeystorePassword = "password"

func keyTagArg(keyTag symbiotic.KeyTag) string {
	return strconv.FormatUint(uint64(keyTag), 10)
}

func createRelayKeyStore(t *testing.T, keyTag symbiotic.KeyTag, privateKey relaycrypto.PrivateKey) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "keystore.jks")
	keyStore, err := keyprovider.NewKeystoreProvider(path, testKeystorePassword)
	require.NoError(t, err)
	require.NoError(t, keyStore.AddKey(keyprovider.SYMBIOTIC_KEY_NAMESPACE, keyTag, privateKey, testKeystorePassword, false))

	return path
}

func runUtilsCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//nolint:gosec // Test helper runs the local CLI with test-controlled arguments.
	cmd := exec.CommandContext(ctx, "go", append([]string{"run", ".."}, args...)...)
	cmd.Dir = filepath.Dir(currentFile)

	output, err := cmd.CombinedOutput()
	return string(output), err
}
