package keyprovider

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/symbioticfi/relay/core/usecase/crypto"

	"github.com/symbioticfi/relay/core/entity"

	"github.com/pavlo-v-chernykh/keystore-go/v4"
)

type SimpleKeystoreProvider struct {
	keys map[string]crypto.PrivateKey
	mu   sync.RWMutex
}

func NewSimpleKeystoreProvider() (*SimpleKeystoreProvider, error) {
	return &SimpleKeystoreProvider{
		keys: make(map[string]crypto.PrivateKey),
	}, nil
}

func (k *SimpleKeystoreProvider) GetPrivateKey(keyTag entity.KeyTag) (crypto.PrivateKey, error) {
	alias, err := KeyTagToAlias(keyTag)
	if err != nil {
		return nil, err
	}

	return k.GetPrivateKeyByAlias(alias)
}

func (k *SimpleKeystoreProvider) GetPrivateKeyByAlias(alias string) (crypto.PrivateKey, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	entry, ok := k.keys[alias]
	if !ok {
		return nil, keystore.ErrEntryNotFound
	}
	return entry, nil
}

func (k *SimpleKeystoreProvider) GetPrivateKeyByNamespaceTypeId(namespace string, keyType entity.KeyType, id int) (crypto.PrivateKey, error) {
	alias, err := ToAlias(namespace, keyType, id)
	if err != nil {
		return nil, err
	}
	key, err := k.GetPrivateKeyByAlias(alias)
	if err != nil {
		if errors.Is(err, keystore.ErrEntryNotFound) && namespace == EVM_KEY_NAMESPACE {
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

func (k *SimpleKeystoreProvider) HasKey(keyTag entity.KeyTag) (bool, error) {
	alias, err := KeyTagToAlias(keyTag)
	if err != nil {
		return false, err
	}
	return k.HasKeyByAlias(alias)
}

func (k *SimpleKeystoreProvider) HasKeyByAlias(alias string) (bool, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	_, ok := k.keys[alias]
	if !ok {
		return false, nil
	}
	return ok, nil
}

func (k *SimpleKeystoreProvider) HasKeyByNamespaceTypeId(namespace string, keyType entity.KeyType, id int) (bool, error) {
	alias, err := ToAlias(namespace, keyType, id)
	if err != nil {
		return false, err
	}
	return k.HasKeyByAlias(alias)
}

func (k *SimpleKeystoreProvider) AddKey(keyTag entity.KeyTag, privateKey crypto.PrivateKey) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	alias, err := KeyTagToAlias(keyTag)
	if err != nil {
		return err
	}

	k.keys[alias] = privateKey

	return nil
}

func (k *SimpleKeystoreProvider) AddKeyByNamespaceTypeId(ns string, tp entity.KeyType, id int, privateKey crypto.PrivateKey) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	alias, err := ToAlias(ns, tp, id)
	if err != nil {
		return err
	}

	k.keys[alias] = privateKey

	return nil
}

func (k *SimpleKeystoreProvider) DeleteKey(keyTag entity.KeyTag) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	alias, err := KeyTagToAlias(keyTag)
	if err != nil {
		return err
	}

	delete(k.keys, alias)

	return nil
}
