package keyprovider

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/pavlo-v-chernykh/keystore-go/v4"
	"github.com/stretchr/testify/require"
)

func regressionLegacyStore(t *testing.T) (string, []string) {
	t.Helper()
	path := t.TempDir() + "/legacy.jks"
	ks := keystore.New()
	relayAlias, err := KeyTagToAlias(15)
	require.NoError(t, err)
	evmAlias, err := ToAlias(EVM_KEY_NAMESPACE, 0, 0)
	require.NoError(t, err)
	aliases := []string{relayAlias, evmAlias}
	for _, alias := range aliases {
		require.NoError(t, ks.SetPrivateKeyEntry(alias, keystore.PrivateKeyEntry{CreationTime: time.Now(), PrivateKey: []byte{1}}, nil))
	}
	f, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, ks.Store(f, []byte("password")))
	require.NoError(t, f.Close())
	return path, aliases
}

func TestRegressionLegacyReadOnly(t *testing.T) {
	path, aliases := regressionLegacyStore(t)
	require.NoError(t, os.Chmod(path, 0400))
	kp, err := NewKeystoreProvider(path, "password")
	require.NoError(t, err)
	_, err = kp.GetPrivateKeyByAlias(aliases[0])
	t.Logf("legacy read-only key lookup: %v", err)
	require.NoError(t, err)
}

func TestRegressionLegacyConcurrentReads(t *testing.T) {
	path, aliases := regressionLegacyStore(t)
	kp, err := NewKeystoreProvider(path, "password")
	require.NoError(t, err)
	start := make(chan struct{})
	results := make(chan error, 8)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			_, err := kp.GetPrivateKeyByAlias(aliases[i%len(aliases)])
			results <- err
		}(i)
	}
	close(start)
	wg.Wait()
	close(results)
	for err := range results {
		require.NoError(t, err)
	}
}

func TestLegacyMigrationEncryptsAllEntries(t *testing.T) {
	path, aliases := regressionLegacyStore(t)
	before, err := os.ReadFile(path)
	require.NoError(t, err)
	kp, err := NewKeystoreProvider(path, "password")
	require.NoError(t, err)
	_, err = kp.GetPrivateKeyByAlias(aliases[0])
	require.NoError(t, err)
	afterRead, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, before, afterRead)
	require.NoError(t, kp.Migrate("password"))
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	ks := keystore.New()
	require.NoError(t, ks.Load(f, []byte("password")))
	for _, alias := range aliases {
		_, err := ks.GetPrivateKeyEntry(alias, []byte("password"))
		require.NoError(t, err)
		_, err = ks.GetPrivateKeyEntry(alias, nil)
		require.Error(t, err)
	}
	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestFailedKeystorePersistencePreservesState(t *testing.T) {
	path, aliases := regressionLegacyStore(t)
	kp, err := NewKeystoreProvider(path, "password")
	require.NoError(t, err)
	before, err := os.ReadFile(path)
	require.NoError(t, err)
	// Renaming a file over a directory fails even when tests run as root.
	kp.filePath = t.TempDir()
	err = kp.remove(aliases[0], "password")
	require.Error(t, err)
	exists, err := kp.HasKeyByAlias(aliases[0])
	require.NoError(t, err)
	require.True(t, exists)
	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, before, after)
}
