package keyprovider

import (
	"testing"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"

	"github.com/stretchr/testify/require"
)

func TestNewKeystore(t *testing.T) {
	path := t.TempDir() + "/TMP-keystore"
	password := "password"

	_, err := NewKeystoreProvider(path, password)
	require.NoError(t, err)
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
