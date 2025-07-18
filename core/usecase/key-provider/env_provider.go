package keyprovider

import (
	"encoding/base64"
	"os"
	"strings"
	"sync"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/crypto"

	"github.com/go-errors/errors"
)

type EnvKeyProvider struct {
	cache map[string]crypto.PrivateKey
	mu    sync.RWMutex
}

func NewEnvKeyProvider() *EnvKeyProvider {
	return &EnvKeyProvider{
		cache: make(map[string]crypto.PrivateKey),
	}
}

func (e *EnvKeyProvider) GetPrivateKey(keyTag entity.KeyTag) (crypto.PrivateKey, error) {
	alias, err := KeyTagToAlias(keyTag)
	if err != nil {
		return nil, err
	}

	return e.GetPrivateKeyByAlias(alias)
}

func (e *EnvKeyProvider) GetPrivateKeyByNamespaceTypeId(
	namespace string,
	keyType entity.KeyType,
	keyId int,
) (crypto.PrivateKey, error) {
	alias, err := ToAlias(namespace, keyType, keyId)
	if err != nil {
		return nil, err
	}
	return e.GetPrivateKeyByAlias(alias)
}

func (e *EnvKeyProvider) GetPrivateKeyByAlias(alias string) (crypto.PrivateKey, error) {
	e.mu.RLock()
	key, ok := e.cache[alias]
	if ok {
		e.mu.RUnlock()
		return key, nil
	}
	e.mu.RUnlock()

	val := os.Getenv(strings.ToUpper(alias)) // todo ilya: research if it's safe to read private keys from environment variables
	if val == "" {
		return nil, errors.New("key not found in environment")
	}

	keyType, _, err := AliasToKeyTypeId(alias)
	if err != nil {
		return nil, err
	}
	decoded, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return nil, err
	}

	key, err = crypto.NewPrivateKey(keyType, decoded)
	if err != nil {
		return nil, errors.Errorf("failed to create private key: %s", err)
	}

	e.mu.Lock()
	e.cache[alias] = key
	e.mu.Unlock()

	return key, nil
}

func (e *EnvKeyProvider) HasKey(keyTag entity.KeyTag) (bool, error) {
	alias, err := KeyTagToAlias(keyTag)
	if err != nil {
		return false, err
	}

	return e.HasKeyByAlias(alias)
}

func (e *EnvKeyProvider) HasKeyByAlias(alias string) (bool, error) {
	e.mu.RLock()
	_, ok := e.cache[alias]
	e.mu.RUnlock()

	if ok {
		return true, nil
	}

	_, err := e.GetPrivateKeyByAlias(alias)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (e *EnvKeyProvider) HasKeyByNamespaceTypeId(namespace string, keyType entity.KeyType, id int) (bool, error) {
	alias, err := ToAlias(namespace, keyType, id)
	if err != nil {
		return false, err
	}

	return e.HasKeyByAlias(alias)
}
