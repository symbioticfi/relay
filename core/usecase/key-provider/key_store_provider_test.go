package keyprovider

import (
	"middleware-offchain/core/usecase/crypto"
	"testing"

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
	key, err := crypto.NewPrivateKey(15, pk)
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
	key, err := crypto.NewPrivateKey(15, pk)
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
	key, err := crypto.NewPrivateKey(15, pk)
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

	require.Equal(t, storedPk, pk)
}
