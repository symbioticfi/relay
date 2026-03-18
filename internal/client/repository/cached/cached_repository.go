package cached

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/client/repository/cache"
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type Repository interface {
	Close() error

	// Signatures
	SaveSignature(ctx context.Context, signature symbiotic.Signature, validator symbiotic.Validator, activeIndex uint32) error
	GetAllSignatures(ctx context.Context, requestID common.Hash) ([]symbiotic.Signature, error)
	GetSignatureByIndex(ctx context.Context, requestID common.Hash, validatorIndex uint32) (symbiotic.Signature, error)
	GetSignaturesStartingFromEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]symbiotic.Signature, error)
	GetSignaturesByEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]symbiotic.Signature, error)

	// Signature Maps
	UpdateSignatureMap(ctx context.Context, vm entity.SignatureMap) error
	GetSignatureMap(ctx context.Context, requestID common.Hash) (entity.SignatureMap, error)

	// Signature Requests
	SaveSignatureRequest(ctx context.Context, requestID common.Hash, req symbiotic.SignatureRequest) error
	GetSignatureRequest(ctx context.Context, requestID common.Hash) (symbiotic.SignatureRequest, error)
	GetSignatureRequestsByEpoch(ctx context.Context, epoch symbiotic.Epoch, limit int, lastHash common.Hash) ([]symbiotic.SignatureRequest, error)
	GetSignatureRequestsWithIDByEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]entity.SignatureRequestWithID, error)
	GetSignatureRequestIDsByEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]common.Hash, error)
	GetSignaturePending(ctx context.Context, limit int) ([]common.Hash, error)
	RemoveSignaturePending(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) error

	// Aggregation Proofs
	SaveProof(ctx context.Context, aggregationProof symbiotic.AggregationProof) error
	GetAggregationProof(ctx context.Context, requestID common.Hash) (symbiotic.AggregationProof, error)
	GetAggregationProofsByEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]symbiotic.AggregationProof, error)
	GetAggregationProofsStartingFromEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]symbiotic.AggregationProof, error)
	GetSignatureRequestsWithoutAggregationProof(ctx context.Context, epoch symbiotic.Epoch, limit int, lastHash common.Hash) ([]symbiotic.SignatureRequestWithID, error)
	RemoveAggregationProofPending(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) error

	// Validator Sets
	GetValidatorSetByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSet, error)
	GetValidatorSetHeaderByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSetHeader, error)
	GetValidatorSetsStartingFromEpoch(ctx context.Context, startEpoch symbiotic.Epoch) ([]symbiotic.ValidatorSet, error)
	GetValidatorByKey(ctx context.Context, epoch symbiotic.Epoch, keyTag symbiotic.KeyTag, publicKey []byte) (symbiotic.Validator, uint32, error)
	GetActiveValidatorCountByEpoch(ctx context.Context, epoch symbiotic.Epoch) (uint32, error)
	GetLatestValidatorSetHeader(ctx context.Context) (symbiotic.ValidatorSetHeader, error)
	GetLatestValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error)
	GetOldestValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error)
	GetLatestAggregatedValsetHeader(ctx context.Context) (symbiotic.ValidatorSetHeader, error)
	UpdateValidatorSetStatus(ctx context.Context, epoch symbiotic.Epoch, status symbiotic.ValidatorSetStatus) error
	UpdateValidatorSetStatusAndRemovePendingProof(ctx context.Context, valset symbiotic.ValidatorSet) error
	SaveFirstUncommittedValidatorSetEpoch(ctx context.Context, epoch symbiotic.Epoch) error
	GetFirstUncommittedValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error)

	// Validator Set Metadata
	GetValidatorSetMetadata(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSetMetadata, error)

	// Network Config
	SaveConfig(ctx context.Context, config symbiotic.NetworkConfig, epoch symbiotic.Epoch) error
	GetConfigByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.NetworkConfig, error)

	// Proof Commits
	GetPendingProofCommitsSinceEpoch(ctx context.Context, epoch symbiotic.Epoch, limit int) ([]symbiotic.ProofCommitKey, error)

	// Composite Operations
	SaveNextValsetData(ctx context.Context, data entity.NextValsetData) error

	// Pruning
	PruneValsetEntities(ctx context.Context, epoch symbiotic.Epoch) error
	PruneProofEntities(ctx context.Context, epoch symbiotic.Epoch) error
	PruneSignatureEntitiesForEpoch(ctx context.Context, epoch symbiotic.Epoch) error
	PruneRequestIDEpochIndices(ctx context.Context, epoch symbiotic.Epoch) error
}

type Config struct {
	NetworkConfigCacheSize int
	ValidatorSetCacheSize  int
}

type CachedRepository struct {
	Repository

	networkConfigCache        cache.Cache[symbiotic.Epoch, symbiotic.NetworkConfig]
	validatorSetCache         cache.Cache[symbiotic.Epoch, symbiotic.ValidatorSet]
	validatorSetMetadataCache cache.Cache[symbiotic.Epoch, symbiotic.ValidatorSetMetadata]
}

func NewCached(repo Repository, cfg Config) (*CachedRepository, error) {
	networkConfigCache, err := cache.NewCache[symbiotic.Epoch, symbiotic.NetworkConfig](
		cache.Config{Size: cfg.NetworkConfigCacheSize},
		func(epoch symbiotic.Epoch) uint32 { return uint32(epoch) },
	)
	if err != nil {
		return nil, errors.Errorf("failed to create network config cache: %w", err)
	}

	validatorSetCache, err := cache.NewCache[symbiotic.Epoch, symbiotic.ValidatorSet](
		cache.Config{Size: cfg.ValidatorSetCacheSize},
		func(epoch symbiotic.Epoch) uint32 { return uint32(epoch) },
	)
	if err != nil {
		return nil, errors.Errorf("failed to create validator set cache: %w", err)
	}

	validatorSetMetadataCache, err := cache.NewCache[symbiotic.Epoch, symbiotic.ValidatorSetMetadata](
		cache.Config{Size: cfg.ValidatorSetCacheSize},
		func(epoch symbiotic.Epoch) uint32 { return uint32(epoch) },
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
	if config, ok := r.networkConfigCache.Get(epoch); ok {
		return config, nil
	}

	config, err := r.Repository.GetConfigByEpoch(ctx, epoch)
	if err != nil {
		return symbiotic.NetworkConfig{}, err
	}

	r.networkConfigCache.Add(epoch, config)
	return config, nil
}

func (r *CachedRepository) SaveConfig(ctx context.Context, config symbiotic.NetworkConfig, epoch symbiotic.Epoch) error {
	if err := r.Repository.SaveConfig(ctx, config, epoch); err != nil {
		return err
	}
	r.networkConfigCache.Add(epoch, config)
	return nil
}

func (r *CachedRepository) GetValidatorSetByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSet, error) {
	if validatorSet, ok := r.validatorSetCache.Get(epoch); ok {
		return validatorSet, nil
	}

	validatorSet, err := r.Repository.GetValidatorSetByEpoch(ctx, epoch)
	if err != nil {
		return symbiotic.ValidatorSet{}, err
	}

	r.validatorSetCache.Add(epoch, validatorSet)
	return validatorSet, nil
}

func (r *CachedRepository) GetValidatorSetMetadata(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSetMetadata, error) {
	if metadata, ok := r.validatorSetMetadataCache.Get(epoch); ok {
		return metadata, nil
	}

	metadata, err := r.Repository.GetValidatorSetMetadata(ctx, epoch)
	if err != nil {
		return symbiotic.ValidatorSetMetadata{}, err
	}

	r.validatorSetMetadataCache.Add(epoch, metadata)
	return metadata, nil
}

func (r *CachedRepository) PruneValsetEntities(ctx context.Context, epoch symbiotic.Epoch) error {
	if err := r.Repository.PruneValsetEntities(ctx, epoch); err != nil {
		return err
	}
	r.evictValsetCaches(epoch)
	return nil
}

func (r *CachedRepository) SaveNextValsetData(ctx context.Context, data entity.NextValsetData) error {
	if err := r.Repository.SaveNextValsetData(ctx, data); err != nil {
		return err
	}

	r.validatorSetCache.Add(data.PrevValidatorSet.Epoch, data.PrevValidatorSet)
	r.networkConfigCache.Add(data.PrevValidatorSet.Epoch, data.PrevNetworkConfig)

	r.validatorSetCache.Add(data.NextValidatorSet.Epoch, data.NextValidatorSet)
	r.networkConfigCache.Add(data.NextValidatorSet.Epoch, data.NextNetworkConfig)

	r.validatorSetMetadataCache.Add(data.ValidatorSetMetadata.Epoch, data.ValidatorSetMetadata)
	return nil
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
