package keyprovider

import (
	"testing"

	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/crypto"

	"github.com/pavlo-v-chernykh/keystore-go/v4"
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
	key, err := crypto.NewPrivateKey(entity.KeyTypeBlsBn254, pk)
	require.NoError(t, err)

	err = kp.AddKey(15, key, password, false)
	require.NoError(t, err)
}

func TestForceAddKey(t *testing.T) {
	path := t.TempDir() + "/TMP-keystore"
	password := "password"

	kp, err := NewKeystoreProvider(path, password)
	require.NoError(t, err)

	pk := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'}
	key, err := crypto.NewPrivateKey(entity.KeyTypeBlsBn254, pk)
	require.NoError(t, err)

	err = kp.AddKey(15, key, password, false)
	require.NoError(t, err)

	err = kp.AddKey(15, key, password, false)
	require.Error(t, err)

	err = kp.AddKey(15, key, password, true)
	require.NoError(t, err)
}

func TestCreateAndReopen(t *testing.T) {
	path := t.TempDir() + "/TMP-keystore"
	password := "password"

	kp, err := NewKeystoreProvider(path, password)
	require.NoError(t, err)

	pk := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'}
	key, err := crypto.NewPrivateKey(entity.KeyTypeBlsBn254, pk)
	require.NoError(t, err)

	err = kp.AddKey(15, key, password, false)
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
	key, err := crypto.NewPrivateKey(entity.KeyTypeBlsBn254, pk)
	require.NoError(t, err)

	_, err = kp.GetPrivateKeyByNamespaceTypeId(EVM_KEY_NAMESPACE, entity.KeyTypeBlsBn254, 111)
	require.ErrorAs(t, err, &keystore.ErrEntryNotFound, "expected entry not found error for non-existing key")

	err = kp.AddKeyByNamespaceTypeId(EVM_KEY_NAMESPACE, entity.KeyTypeBlsBn254, DEFAULT_EVM_CHAIN_ID, key, password, false)
	require.NoError(t, err)

	storedPk, err := kp.GetPrivateKeyByNamespaceTypeId(EVM_KEY_NAMESPACE, entity.KeyTypeBlsBn254, 111)
	require.NoError(t, err)
	require.Equal(t, storedPk.Bytes(), pk)

	// shouldn't work for other chains
	_, err = kp.GetPrivateKeyByNamespaceTypeId(SYMBIOTIC_KEY_NAMESPACE, entity.KeyTypeBlsBn254, 111)
	require.ErrorAs(t, err, &keystore.ErrEntryNotFound, "expected entry not found error for non-existing key")

}
