package keyprovider

import (
	"encoding/base64"
	"middleware-offchain/core/usecase/crypto"
	"os"
	"strings"
	"sync"

	"github.com/go-errors/errors"

	"middleware-offchain/core/entity"
)

type EnvKeyProvider struct {
	cache map[entity.KeyTag]crypto.PrivateKey
	mu    sync.RWMutex
}

func NewEnvKeyProvider() *EnvKeyProvider {
	return &EnvKeyProvider{
		cache: make(map[entity.KeyTag]crypto.PrivateKey),
	}
}

func (e *EnvKeyProvider) GetPrivateKey(keyTag entity.KeyTag) (crypto.PrivateKey, error) {
	e.mu.RLock()

	key, ok := e.cache[keyTag]
	if ok {
		e.mu.RUnlock()
		return key, nil
	}
	e.mu.RUnlock()

	alias, err := getAlias(keyTag)
	if err != nil {
		return nil, err
	}

	val := os.Getenv(strings.ToUpper(alias)) // todo ilya: research if it's safe to read private keys from environment variables
	if val == "" {
		return nil, errors.New("key not found in environment")
	}

	decoded, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return nil, err
	}

	key, err = crypto.NewPrivateKey(keyTag, decoded)
	if err != nil {
		return nil, errors.Errorf("failed to create private key: %s", err)
	}

	e.mu.Lock()
	e.cache[keyTag] = key
	e.mu.Unlock()

	return key, nil
}

func (e *EnvKeyProvider) HasKey(keyTag entity.KeyTag) (bool, error) {
	e.mu.RLock()
	_, ok := e.cache[keyTag]
	e.mu.RUnlock()

	if ok {
		return true, nil
	}

	_, err := e.GetPrivateKey(keyTag)
	if err != nil {
		return false, err
	}

	return true, nil
}
