package keyprovider

import (
	"errors"
	"sync"

	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
)

type SimpleKeystoreProvider struct {
	keys map[string][]byte
	mu   sync.RWMutex
}

func NewSimpleKeystoreProvider() (*SimpleKeystoreProvider, error) {
	return &SimpleKeystoreProvider{
		keys: make(map[string][]byte),
	}, nil
}

func (k *SimpleKeystoreProvider) GetPrivateKey(keyTag entity.KeyTag) ([]byte, error) {
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

func (k *SimpleKeystoreProvider) GetPublic(keyTag entity.KeyTag) ([]byte, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	alias, err := getAlias(keyTag)
	if err != nil {
		return nil, err
	}

	sk, ok := k.keys[alias]
	if !ok {
		return nil, errors.New("key not found")
	}

	switch keyTag.Type() {
	case entity.KeyTypeBlsBn254:
		kp := bls.ComputeKeyPair(sk)
		return kp.PackPublicG1G2(), nil
	case entity.KeyTypeEcdsaSecp256k1:
		return nil, errors.New("ECDSA key type not supported in this provider")
	}

	return sk, nil
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

func (k *SimpleKeystoreProvider) AddKey(keyTag entity.KeyTag, privateKey []byte) error {
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
