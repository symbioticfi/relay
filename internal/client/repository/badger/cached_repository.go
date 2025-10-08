package badger

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/client/repository/cache"
	"github.com/symbioticfi/relay/symbiotic/entity"
)

type CachedConfig struct {
	NetworkConfigCacheSize int
	ValidatorSetCacheSize  int
}

type CachedRepository struct {
	*Repository

	networkConfigCache        cache.Cache[entity.Epoch, entity.NetworkConfig]
	validatorSetCache         cache.Cache[entity.Epoch, entity.ValidatorSet]
	validatorSetMetadataCache cache.Cache[entity.Epoch, entity.ValidatorSetMetadata]
}

func NewCached(repo *Repository, cfg CachedConfig) (*CachedRepository, error) {
	networkConfigCache, err := cache.NewCache[entity.Epoch, entity.NetworkConfig](
		cache.Config{Size: cfg.NetworkConfigCacheSize},
		func(epoch entity.Epoch) uint32 {
			return uint32(epoch)
		},
	)
	if err != nil {
		return nil, errors.Errorf("failed to create network config cache: %w", err)
	}

	validatorSetCache, err := cache.NewCache[entity.Epoch, entity.ValidatorSet](
		cache.Config{Size: cfg.ValidatorSetCacheSize},
		func(epoch entity.Epoch) uint32 {
			return uint32(epoch)
		},
	)
	if err != nil {
		return nil, errors.Errorf("failed to create validator set cache: %w", err)
	}

	validatorSetMetadataCache, err := cache.NewCache[entity.Epoch, entity.ValidatorSetMetadata](
		cache.Config{Size: cfg.ValidatorSetCacheSize},
		func(epoch entity.Epoch) uint32 {
			return uint32(epoch)
		},
	)
	if err != nil {
		return nil, errors.Errorf("failed to create validator set metadata cache: %w", err)
	}
	return &CachedRepository{
		Repository:                repo,
		networkConfigCache:        networkConfigCache,
		validatorSetCache:         validatorSetCache,
		validatorSetMetadataCache: validatorSetMetadataCache,
	}, nil
}

func (r *CachedRepository) GetConfigByEpoch(ctx context.Context, epoch entity.Epoch) (entity.NetworkConfig, error) {
	// Try cache first
	if config, ok := r.networkConfigCache.Get(epoch); ok {
		return config, nil
	}

	// Cache miss - load from underlying repository
	config, err := r.Repository.GetConfigByEpoch(ctx, epoch)
	if err != nil {
		return entity.NetworkConfig{}, err
	}

	// Store in cache for future use
	r.networkConfigCache.Add(epoch, config)
	return config, nil
}

func (r *CachedRepository) SaveConfig(ctx context.Context, config entity.NetworkConfig, epoch entity.Epoch) error {
	err := r.Repository.SaveConfig(ctx, config, epoch)
	if err != nil {
		return err
	}
	// Cache the newly saved config
	r.networkConfigCache.Add(epoch, config)
	return nil
}

func (r *CachedRepository) GetValidatorSetByEpoch(ctx context.Context, epoch entity.Epoch) (entity.ValidatorSet, error) {
	// Try cache first
	if validatorSet, ok := r.validatorSetCache.Get(epoch); ok {
		return validatorSet, nil
	}

	// Cache miss - load from underlying repository
	validatorSet, err := r.Repository.GetValidatorSetByEpoch(ctx, epoch)
	if err != nil {
		return entity.ValidatorSet{}, err
	}

	// Store in cache for future use
	r.validatorSetCache.Add(epoch, validatorSet)
	return validatorSet, nil
}

func (r *CachedRepository) GetValidatorSetMetadata(ctx context.Context, epoch entity.Epoch) (entity.ValidatorSetMetadata, error) {
	// Try cache first
	if validatorSetMetadata, ok := r.validatorSetMetadataCache.Get(epoch); ok {
		return validatorSetMetadata, nil
	}

	// Cache miss - load from underlying repository
	validatorSetMetadata, err := r.Repository.GetValidatorSetMetadata(ctx, epoch)
	if err != nil {
		return entity.ValidatorSetMetadata{}, err
	}

	// Store in cache for future use
	r.validatorSetMetadataCache.Add(epoch, validatorSetMetadata)
	return validatorSetMetadata, nil
}

func (r *CachedRepository) SaveValidatorSet(ctx context.Context, validatorSet entity.ValidatorSet) error {
	err := r.Repository.SaveValidatorSet(ctx, validatorSet)
	if err != nil {
		return err
	}
	// Cache the newly saved validator set
	r.validatorSetCache.Add(validatorSet.Epoch, validatorSet)
	return nil
}

func (r *CachedRepository) SaveValidatorSetMetadata(ctx context.Context, validatorSetMetadata entity.ValidatorSetMetadata) error {
	err := r.Repository.SaveValidatorSetMetadata(ctx, validatorSetMetadata)
	if err != nil {
		return err
	}
	// Cache the newly saved validator set
	r.validatorSetMetadataCache.Add(validatorSetMetadata.Epoch, validatorSetMetadata)
	return nil
}
