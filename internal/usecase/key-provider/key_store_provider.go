package keyprovider

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/go-errors/errors"
	"github.com/pavlo-v-chernykh/keystore-go/v4"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

// KeystoreProvider supports concurrent reads after construction. Mutations
// require exclusive access. Construction migrates legacy files before use.
type KeystoreProvider struct {
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
	err = k.ks.Load(f, []byte(password))
	closeErr := f.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	// Authenticate the store before detecting and migrating legacy entries.
	legacy, err := reencryptEntries(k.ks, password, password)
	if err != nil {
		return nil, err
	}

	if legacy {
		if err := writeKeystore(filePath, k.ks, password); err != nil {
			return nil, errors.Errorf("failed to migrate legacy keystore: %w", err)
		}
	}
	return k, nil
}

func (k *KeystoreProvider) GetAliases() []string {
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

// The outer store password must be validated before legacy entry fallback.
// Re-encryption retains the entire entry, including certificates and timestamps.
func reencryptEntries(ks keystore.KeyStore, oldPassword, newPassword string) (bool, error) {
	legacy := false
	for _, alias := range ks.Aliases() {
		if !ks.IsPrivateKeyEntry(alias) {
			continue
		}
		entry, err := ks.GetPrivateKeyEntry(alias, []byte(oldPassword))
		if err == nil && oldPassword == newPassword {
			continue
		}
		if err != nil {
			entry, err = ks.GetPrivateKeyEntry(alias, nil)
			if err != nil {
				return false, errors.Errorf("failed to decrypt entry %q: %w", alias, err)
			}
			legacy = true
		}
		if err := ks.SetPrivateKeyEntry(alias, entry, []byte(newPassword)); err != nil {
			return false, err
		}
	}
	return legacy, nil
}

// Mutations become visible only after successful atomic persistence.
func (k *KeystoreProvider) mutate(password string, change func(keystore.KeyStore) error) error {
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
	if err := writeKeystore(k.filePath, next, password); err != nil {
		return err
	}
	k.ks = next
	return nil
}

func writeKeystore(filePath string, ks keystore.KeyStore, password string) error {
	var data bytes.Buffer
	if err := ks.Store(&data, []byte(password)); err != nil {
		return err
	}
	dir := filepath.Dir(filePath)
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
	if err := os.Rename(f.Name(), filePath); err != nil {
		return err
	}
	return nil
}
