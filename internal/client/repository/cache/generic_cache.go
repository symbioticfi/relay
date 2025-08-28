package cache

import (
	"github.com/elastic/go-freelru"
	"github.com/go-errors/errors"
)

// Cache provides a generic LRU cache interface
type Cache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Add(key K, value V)
}

// Config holds cache configuration
type Config struct {
	Size int
}

// lruCache implements Cache using freelru
type lruCache[K comparable, V any] struct {
	cache *freelru.ShardedLRU[K, V]
}

// NewCache creates a new generic LRU cache
func NewCache[K comparable, V any](cfg Config, hashFunc func(K) uint32) (Cache[K, V], error) {
	if cfg.Size <= 0 {
		cfg.Size = 50 // default cache size
	}

	cache, err := freelru.NewSharded[K, V](uint32(cfg.Size), hashFunc)
	if err != nil {
		return nil, errors.Errorf("failed to create sharded LRU cache: %w", err)
	}

	return &lruCache[K, V]{
		cache: cache,
	}, nil
}

func (c *lruCache[K, V]) Get(key K) (V, bool) {
	return c.cache.Get(key)
}

func (c *lruCache[K, V]) Add(key K, value V) {
	c.cache.Add(key, value)
}
