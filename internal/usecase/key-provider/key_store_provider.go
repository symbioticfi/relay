package keyprovider

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-errors/errors"
	"github.com/pavlo-v-chernykh/keystore-go/v4"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

type KeystoreProvider struct {
	mu            sync.RWMutex
	ks            keystore.KeyStore
	filePath      string
	storePassword string
}

func NewKeystoreProvider(filePath, password string) (*KeystoreProvider, error) {
	if password == "" {
		return nil, errors.New("keystore password cannot be empty")
	}
	k := &KeystoreProvider{ks: keystore.New(), filePath: filePath, storePassword: password}
	f, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return k, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := k.ks.Load(f, []byte(password)); err != nil {
		return nil, err
	}
	// Normalize every legacy entry before publishing the provider. Reading a
	// mounted secret must never write to disk; persistence is an explicit action.
	legacy := false
	for _, alias := range k.ks.Aliases() {
		if !k.ks.IsPrivateKeyEntry(alias) {
			continue
		}
		if _, err := k.ks.GetPrivateKeyEntry(alias, []byte(password)); err == nil {
			continue
		}
		entry, err := k.ks.GetPrivateKeyEntry(alias, nil)
		if err != nil {
			return nil, errors.Errorf("failed to decrypt entry %q: %w", alias, err)
		}
		if err := k.ks.SetPrivateKeyEntry(alias, entry, []byte(password)); err != nil {
			return nil, err
		}
		legacy = true
	}
	if legacy {
		slog.Warn("Legacy keystore entries detected; run relay_utils keys migrate on a writable copy to encrypt all entries at rest")
	}
	return k, nil
}

func (k *KeystoreProvider) GetAliases() []string {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.ks.Aliases()
}

func (k *KeystoreProvider) GetPrivateKey(keyTag symbiotic.KeyTag) (crypto.PrivateKey, error) {
	alias, err := KeyTagToAlias(keyTag)
	if err != nil {
		return nil, err
	}
	return k.GetPrivateKeyByAlias(alias)
}

func (k *KeystoreProvider) GetPrivateKeyByAlias(alias string) (crypto.PrivateKey, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	entry, err := k.ks.GetPrivateKeyEntry(alias, []byte(k.storePassword))
	if errors.Is(err, keystore.ErrEntryNotFound) {
		return nil, errors.New(entity.ErrKeyNotFound)
	}
	if err != nil {
		return nil, err
	}
	_, keyType, _, err := AliasToKeyTypeId(alias)
	if err != nil {
		return nil, err
	}
	return crypto.NewPrivateKey(keyType, entry.PrivateKey)
}

func (k *KeystoreProvider) GetPrivateKeyByNamespaceTypeId(namespace string, keyType symbiotic.KeyType, id int) (crypto.PrivateKey, error) {
	alias, err := ToAlias(namespace, keyType, id)
	if err != nil {
		return nil, err
	}
	key, err := k.GetPrivateKeyByAlias(alias)
	if errors.Is(err, entity.ErrKeyNotFound) && namespace == EVM_KEY_NAMESPACE && id != DEFAULT_EVM_CHAIN_ID {
		slog.Warn("Key not found, falling back to default EVM key", "alias", alias)
		return k.GetPrivateKeyByNamespaceTypeId(namespace, keyType, DEFAULT_EVM_CHAIN_ID)
	}
	return key, err
}

func (k *KeystoreProvider) HasKey(keyTag symbiotic.KeyTag) (bool, error) {
	alias, err := KeyTagToAlias(keyTag)
	if err != nil {
		return false, err
	}
	return k.HasKeyByAlias(alias)
}

func (k *KeystoreProvider) HasKeyByAlias(alias string) (bool, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.ks.IsPrivateKeyEntry(alias), nil
}

func (k *KeystoreProvider) HasKeyByNamespaceTypeId(ns string, tp symbiotic.KeyType, id int) (bool, error) {
	alias, err := ToAlias(ns, tp, id)
	if err != nil {
		return false, err
	}
	return k.HasKeyByAlias(alias)
}

func (k *KeystoreProvider) AddKey(ns string, tag symbiotic.KeyTag, key crypto.PrivateKey, password string, force bool) error {
	alias, err := KeyTagToAliasWithNS(ns, tag)
	if err != nil {
		return err
	}
	return k.add(alias, key, password, force)
}

func (k *KeystoreProvider) AddKeyByNamespaceTypeId(ns string, tp symbiotic.KeyType, id int, key crypto.PrivateKey, password string, force bool) error {
	alias, err := ToAlias(ns, tp, id)
	if err != nil {
		return err
	}
	return k.add(alias, key, password, force)
}

func (k *KeystoreProvider) add(alias string, key crypto.PrivateKey, password string, force bool) error {
	return k.mutate(password, func(ks keystore.KeyStore) error {
		if ks.IsPrivateKeyEntry(alias) && !force {
			return errors.New("key already exists")
		}
		return ks.SetPrivateKeyEntry(alias, keystore.PrivateKeyEntry{
			CreationTime: time.Now(), PrivateKey: key.Bytes(),
		}, []byte(password))
	})
}

func (k *KeystoreProvider) DeleteKey(tag symbiotic.KeyTag, password string) error {
	alias, err := KeyTagToAlias(tag)
	if err != nil {
		return err
	}
	return k.remove(alias, password)
}

func (k *KeystoreProvider) DeleteKeyByNamespaceTypeId(ns string, tp symbiotic.KeyType, id int, password string) error {
	alias, err := ToAlias(ns, tp, id)
	if err != nil {
		return err
	}
	return k.remove(alias, password)
}

func (k *KeystoreProvider) remove(alias, password string) error {
	return k.mutate(password, func(ks keystore.KeyStore) error {
		if !ks.IsPrivateKeyEntry(alias) {
			return errors.New("key does not exist")
		}
		ks.DeleteEntry(alias)
		return nil
	})
}

// Migrate persists all legacy entries, including keys not used by this node.
func (k *KeystoreProvider) Migrate(password string) error {
	return k.mutate(password, func(keystore.KeyStore) error { return nil })
}

// Mutations are serialized and become visible only after atomic persistence.
func (k *KeystoreProvider) mutate(password string, change func(keystore.KeyStore) error) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if password != k.storePassword {
		return errors.New("keystore password does not match the loaded keystore")
	}
	var data bytes.Buffer
	if err := k.ks.Store(&data, []byte(password)); err != nil {
		return err
	}
	next := keystore.New()
	if err := next.Load(&data, []byte(password)); err != nil {
		return err
	}
	if err := change(next); err != nil {
		return err
	}
	data.Reset()
	if err := next.Store(&data, []byte(password)); err != nil {
		return err
	}
	dir := filepath.Dir(k.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".keystore-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if _, err := f.Write(data.Bytes()); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(f.Name(), k.filePath); err != nil {
		return err
	}
	k.ks = next
	return nil
}
