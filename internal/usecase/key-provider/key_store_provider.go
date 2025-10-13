package keyprovider

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"

	"github.com/go-errors/errors"
	"github.com/pavlo-v-chernykh/keystore-go/v4"
)

type KeystoreProvider struct {
	ks       keystore.KeyStore
	filePath string
}

func NewKeystoreProvider(filePath, password string) (*KeystoreProvider, error) {
	ks := keystore.New()

	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &KeystoreProvider{ks: ks, filePath: filePath}, nil
		}
		return nil, err
	}
	defer f.Close()

	err = ks.Load(f, []byte(password))
	if err != nil {
		return nil, err
	}
	return &KeystoreProvider{ks: ks, filePath: filePath}, nil
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
	entry, err := k.ks.GetPrivateKeyEntry(alias, []byte{})
	if err != nil {
		if errors.Is(err, keystore.ErrEntryNotFound) {
			return nil, errors.New(entity.ErrKeyNotFound)
		}
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
	if err != nil {
		if errors.Is(err, entity.ErrKeyNotFound) && namespace == EVM_KEY_NAMESPACE {
			// For EVM keys, we check for default key with chain ID 0 if the requested chain id is absent
			slog.Warn("Key not found, checking for default EVM key", "alias", alias)
			defaultAlias, err := ToAlias(EVM_KEY_NAMESPACE, keyType, DEFAULT_EVM_CHAIN_ID)
			if err != nil {
				return nil, err
			}
			return k.GetPrivateKeyByAlias(defaultAlias)
		}
		return nil, err
	}
	return key, nil
}

func (k *KeystoreProvider) HasKey(keyTag symbiotic.KeyTag) (bool, error) {
	alias, err := KeyTagToAlias(keyTag)
	if err != nil {
		return false, err
	}
	return k.ks.IsPrivateKeyEntry(alias), nil
}

func (k *KeystoreProvider) HasKeyByAlias(alias string) (bool, error) {
	return k.ks.IsPrivateKeyEntry(alias), nil
}

func (k *KeystoreProvider) HasKeyByNamespaceTypeId(namespace string, keyType symbiotic.KeyType, id int) (bool, error) {
	alias, err := ToAlias(namespace, keyType, id)
	if err != nil {
		return false, err
	}
	return k.ks.IsPrivateKeyEntry(alias), nil
}

func (k *KeystoreProvider) AddKey(namespace string, keyTag symbiotic.KeyTag, privateKey crypto.PrivateKey, password string, force bool) error {
	exists, err := k.HasKey(keyTag)
	if err != nil {
		return err
	}

	if exists && !force {
		return errors.New("key already exists")
	}

	alias, err := KeyTagToAliasWithNS(namespace, keyTag)
	if err != nil {
		return err
	}

	err = k.ks.SetPrivateKeyEntry(alias, keystore.PrivateKeyEntry{
		CreationTime:     time.Now(),
		PrivateKey:       privateKey.Bytes(),
		CertificateChain: nil,
	}, []byte{})
	if err != nil {
		return err
	}

	err = k.dump(password)
	if err != nil {
		return err
	}

	if exists {
		slog.Info("Key was updated!")
	}

	return nil
}

func (k *KeystoreProvider) DeleteKey(keyTag symbiotic.KeyTag, password string) error {
	exists, err := k.HasKey(keyTag)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("key does not exist")
	}

	alias, err := KeyTagToAlias(keyTag)
	if err != nil {
		return err
	}

	k.ks.DeleteEntry(alias)

	err = k.dump(password)
	if err != nil {
		return err
	}

	return nil
}

func (k *KeystoreProvider) AddKeyByNamespaceTypeId(ns string, tp symbiotic.KeyType, id int, privateKey crypto.PrivateKey, password string, force bool) error {
	exists, err := k.HasKeyByNamespaceTypeId(ns, tp, id)
	if err != nil {
		return err
	}

	if exists && !force {
		return errors.New("key already exists")
	}

	alias, err := ToAlias(ns, tp, id)
	if err != nil {
		return err
	}

	err = k.ks.SetPrivateKeyEntry(alias, keystore.PrivateKeyEntry{
		CreationTime:     time.Now(),
		PrivateKey:       privateKey.Bytes(),
		CertificateChain: nil,
	}, []byte{})
	if err != nil {
		return err
	}

	err = k.dump(password)
	if err != nil {
		return err
	}

	if exists {
		slog.Info("Key was updated!")
	}

	return nil
}

func (k *KeystoreProvider) DeleteKeyByNamespaceTypeId(ns string, tp symbiotic.KeyType, id int, password string) error {
	exists, err := k.HasKeyByNamespaceTypeId(ns, tp, id)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("key does not exist")
	}

	alias, err := ToAlias(ns, tp, id)
	if err != nil {
		return err
	}

	k.ks.DeleteEntry(alias)

	err = k.dump(password)
	if err != nil {
		return err
	}

	return nil
}

func (k *KeystoreProvider) dump(password string) error {
	dir := filepath.Dir(k.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(k.filePath)
	if err != nil {
		slog.Error("Failed to create file", "err", err.Error(), "path", k.filePath)
		return err
	}

	defer func() {
		if err := f.Close(); err != nil {
			slog.Warn("Failed to close file", "err", err.Error())
		}
	}()

	err = k.ks.Store(f, []byte(password))
	if err != nil {
		return err
	}

	return nil
}
