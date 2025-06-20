package keyprovider

import (
	"errors"
	"middleware-offchain/core/usecase/crypto"
	"sync"

	"middleware-offchain/core/entity"
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
	k.mu.RLock()
	defer k.mu.RUnlock()

	alias, err := getAlias(keyTag)
	if err != nil {
		return nil, err
	}

	entry, ok := k.keys[alias]
	if !ok {
		return nil, errors.New("key not found")
	}

	return entry, nil
}

func (k *SimpleKeystoreProvider) HasKey(keyTag entity.KeyTag) (bool, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	alias, err := getAlias(keyTag)
	if err != nil {
		return false, err
	}
	_, ok := k.keys[alias]
	if !ok {
		return false, errors.New("key not found")
	}
	return ok, nil
}

func (k *SimpleKeystoreProvider) AddKey(keyTag entity.KeyTag, privateKey crypto.PrivateKey) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	alias, err := getAlias(keyTag)
	if err != nil {
		return err
	}

	k.keys[alias] = privateKey

	return nil
}

func (k *SimpleKeystoreProvider) DeleteKey(keyTag entity.KeyTag) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	alias, err := getAlias(keyTag)
	if err != nil {
		return err
	}

	delete(k.keys, alias)

	return nil
}
