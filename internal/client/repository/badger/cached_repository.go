package badger

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/client/repository/cache"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type CachedConfig struct {
	NetworkConfigCacheSize int
	ValidatorSetCacheSize  int
}

type CachedRepository struct {
	*Repository

	networkConfigCache        cache.Cache[symbiotic.Epoch, symbiotic.NetworkConfig]
	validatorSetCache         cache.Cache[symbiotic.Epoch, symbiotic.ValidatorSet]
	validatorSetMetadataCache cache.Cache[symbiotic.Epoch, symbiotic.ValidatorSetMetadata]
}

func NewCached(repo *Repository, cfg CachedConfig) (*CachedRepository, error) {
	networkConfigCache, err := cache.NewCache[symbiotic.Epoch, symbiotic.NetworkConfig](
		cache.Config{Size: cfg.NetworkConfigCacheSize},
		func(epoch symbiotic.Epoch) uint32 {
			return uint32(epoch)
		},
	)
	if err != nil {
		return nil, errors.Errorf("failed to create network config cache: %w", err)
	}

	validatorSetCache, err := cache.NewCache[symbiotic.Epoch, symbiotic.ValidatorSet](
		cache.Config{Size: cfg.ValidatorSetCacheSize},
		func(epoch symbiotic.Epoch) uint32 {
			return uint32(epoch)
		},
	)
	if err != nil {
		return nil, errors.Errorf("failed to create validator set cache: %w", err)
	}

	validatorSetMetadataCache, err := cache.NewCache[symbiotic.Epoch, symbiotic.ValidatorSetMetadata](
		cache.Config{Size: cfg.ValidatorSetCacheSize},
		func(epoch symbiotic.Epoch) uint32 {
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

func (r *CachedRepository) GetConfigByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.NetworkConfig, error) {
	// Try cache first
	if config, ok := r.networkConfigCache.Get(epoch); ok {
		return config, nil
	}

	// Cache miss - load from underlying repository
	config, err := r.Repository.GetConfigByEpoch(ctx, epoch)
	if err != nil {
		return symbiotic.NetworkConfig{}, err
	}

	// Store in cache for future use
	r.networkConfigCache.Add(epoch, config)
	return config, nil
}

func (r *CachedRepository) SaveConfig(ctx context.Context, config symbiotic.NetworkConfig, epoch symbiotic.Epoch) error {
	err := r.saveConfig(ctx, config, epoch)
	if err != nil {
		return err
	}
	// Cache the newly saved config
	r.networkConfigCache.Add(epoch, config)
	return nil
}

func (r *CachedRepository) GetValidatorSetByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSet, error) {
	// Try cache first
	if validatorSet, ok := r.validatorSetCache.Get(epoch); ok {
		return validatorSet, nil
	}

	// Cache miss - load from underlying repository
	validatorSet, err := r.Repository.GetValidatorSetByEpoch(ctx, epoch)
	if err != nil {
		return symbiotic.ValidatorSet{}, err
	}

	// Store in cache for future use
	r.validatorSetCache.Add(epoch, validatorSet)
	return validatorSet, nil
}

func (r *CachedRepository) GetValidatorSetMetadata(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSetMetadata, error) {
	// Try cache first
	if validatorSetMetadata, ok := r.validatorSetMetadataCache.Get(epoch); ok {
		return validatorSetMetadata, nil
	}

	// Cache miss - load from underlying repository
	validatorSetMetadata, err := r.Repository.GetValidatorSetMetadata(ctx, epoch)
	if err != nil {
		return symbiotic.ValidatorSetMetadata{}, err
	}

	// Store in cache for future use
	r.validatorSetMetadataCache.Add(epoch, validatorSetMetadata)
	return validatorSetMetadata, nil
}

// PruneValsetEntities delegates to the underlying repository and evicts validator set caches.
func (r *CachedRepository) PruneValsetEntities(ctx context.Context, epoch symbiotic.Epoch) error {
	if err := r.Repository.PruneValsetEntities(ctx, epoch); err != nil {
		return err
	}

	r.evictValsetCaches(epoch)
	return nil
}

func (r *CachedRepository) SaveNextValsetData(ctx context.Context, data entity.NextValsetData) error {
	err := r.Repository.SaveNextValsetData(ctx, data)
	if err != nil {
		return err
	}

	r.validatorSetCache.Add(data.PrevValidatorSet.Epoch, data.PrevValidatorSet)
	r.networkConfigCache.Add(data.PrevValidatorSet.Epoch, data.PrevNetworkConfig)

	r.validatorSetCache.Add(data.NextValidatorSet.Epoch, data.NextValidatorSet)
	r.networkConfigCache.Add(data.NextValidatorSet.Epoch, data.NextNetworkConfig)

	r.validatorSetMetadataCache.Add(data.ValidatorSetMetadata.Epoch, data.ValidatorSetMetadata)
	return nil
}

// PruneProofEntities delegates to the underlying repository.
// No cache eviction needed as proofs are not cached.
func (r *CachedRepository) PruneProofEntities(ctx context.Context, epoch symbiotic.Epoch) error {
	return r.Repository.PruneProofEntities(ctx, epoch)
}

// PruneSignatureEntitiesForEpoch delegates to the underlying repository.
// No cache eviction needed as signatures are not cached.
func (r *CachedRepository) PruneSignatureEntitiesForEpoch(ctx context.Context, epoch symbiotic.Epoch) error {
	return r.Repository.PruneSignatureEntitiesForEpoch(ctx, epoch)
}

func (r *CachedRepository) evictValsetCaches(epoch symbiotic.Epoch) {
	if r.networkConfigCache != nil {
		r.networkConfigCache.Delete(epoch)
	}
	if r.validatorSetCache != nil {
		r.validatorSetCache.Delete(epoch)
	}
	if r.validatorSetMetadataCache != nil {
		r.validatorSetMetadataCache.Delete(epoch)
	}
}
