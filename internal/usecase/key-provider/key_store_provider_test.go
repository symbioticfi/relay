package keyprovider

import (
	"os"
	"testing"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"

	"github.com/pavlo-v-chernykh/keystore-go/v4"
	"github.com/stretchr/testify/require"
)

func TestNewKeystore(t *testing.T) {
	path := t.TempDir() + "/TMP-keystore"
	password := "password"

	_, err := NewKeystoreProvider(path, password)
	require.NoError(t, err)
}

func TestNewKeystoreRejectsEmptyPassword(t *testing.T) {
	_, err := NewKeystoreProvider(t.TempDir()+"/TMP-keystore", "")
	require.ErrorContains(t, err, "password cannot be empty")
}

func TestAddKey(t *testing.T) {
	path := t.TempDir() + "/TMP-keystore"
	password := "password"

	kp, err := NewKeystoreProvider(path, password)
	require.NoError(t, err)

	pk := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'}
	key, err := crypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, pk)
	require.NoError(t, err)

	err = kp.AddKey(SYMBIOTIC_KEY_NAMESPACE, 15, key, password, false)
	require.NoError(t, err)
}

func TestForceAddKey(t *testing.T) {
	path := t.TempDir() + "/TMP-keystore"
	password := "password"

	kp, err := NewKeystoreProvider(path, password)
	require.NoError(t, err)

	pk := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'}
	key, err := crypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, pk)
	require.NoError(t, err)

	err = kp.AddKey(SYMBIOTIC_KEY_NAMESPACE, 15, key, password, false)
	require.NoError(t, err)

	err = kp.AddKey(SYMBIOTIC_KEY_NAMESPACE, 15, key, password, false)
	require.Error(t, err)

	err = kp.AddKey(SYMBIOTIC_KEY_NAMESPACE, 15, key, password, true)
	require.NoError(t, err)
}

func TestCreateAndReopen(t *testing.T) {
	path := t.TempDir() + "/TMP-keystore"
	password := "password"

	kp, err := NewKeystoreProvider(path, password)
	require.NoError(t, err)

	pk := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'}
	key, err := crypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, pk)
	require.NoError(t, err)

	err = kp.AddKey(SYMBIOTIC_KEY_NAMESPACE, 15, key, password, false)
	require.NoError(t, err)

	kp, err = NewKeystoreProvider(path, password)
	require.NoError(t, err)

	exists, err := kp.HasKey(15)
	require.NoError(t, err)

	require.Truef(t, exists, "key should exist in keystore after reopening")

	storedPk, err := kp.GetPrivateKey(15)
	require.NoError(t, err)

	require.Equal(t, storedPk.Bytes(), pk)

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	rawKeystore := keystore.New()
	require.NoError(t, rawKeystore.Load(f, []byte(password)))
	alias, err := KeyTagToAlias(15)
	require.NoError(t, err)
	_, err = rawKeystore.GetPrivateKeyEntry(alias, []byte{})
	require.Error(t, err)
	entry, err := rawKeystore.GetPrivateKeyEntry(alias, []byte(password))
	require.NoError(t, err)
	require.Equal(t, pk, entry.PrivateKey)
}

func TestDefaultEVMKey(t *testing.T) {
	path := t.TempDir() + "/TMP-keystore"
	password := "password"

	kp, err := NewKeystoreProvider(path, password)
	require.NoError(t, err)

	pk := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'}
	key, err := crypto.NewPrivateKey(symbiotic.KeyTypeBlsBn254, pk)
	require.NoError(t, err)

	_, err = kp.GetPrivateKeyByNamespaceTypeId(EVM_KEY_NAMESPACE, symbiotic.KeyTypeBlsBn254, 11)
	require.ErrorIs(t, err, entity.ErrKeyNotFound, "expected entry not found error for non-existing key")

	err = kp.AddKeyByNamespaceTypeId(EVM_KEY_NAMESPACE, symbiotic.KeyTypeBlsBn254, DEFAULT_EVM_CHAIN_ID, key, password, false)
	require.NoError(t, err)

	storedPk, err := kp.GetPrivateKeyByNamespaceTypeId(EVM_KEY_NAMESPACE, symbiotic.KeyTypeBlsBn254, 11)
	require.NoError(t, err)
	require.Equal(t, storedPk.Bytes(), pk)

	// shouldn't work for other chains
	_, err = kp.GetPrivateKeyByNamespaceTypeId(SYMBIOTIC_KEY_NAMESPACE, symbiotic.KeyTypeBlsBn254, 11)
	require.ErrorIs(t, err, entity.ErrKeyNotFound, "expected entry not found error for non-existing key")
}
