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
		return nil, errors.New("key not found")
	}
	return entry, nil
}

func (k *SimpleKeystoreProvider) GetPrivateKeyByNamespaceTypeId(namespace string, keyType entity.KeyType, id int) (crypto.PrivateKey, error) {
	alias, err := ToAlias(namespace, keyType, id)
	if err != nil {
		return nil, err
	}
	return k.GetPrivateKeyByAlias(alias)
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
		return false, errors.New("key not found")
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
