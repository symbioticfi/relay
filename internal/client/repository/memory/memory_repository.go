package memory

import (
	"context"
	"maps"
	"slices"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"middleware-offchain/core/entity"
)

type Repository struct {
	mu sync.Mutex

	networkConfigs map[uint64]entity.NetworkConfig
	validatorSets  map[uint64]entity.ValidatorSet
	signatures     map[common.Hash]map[common.Hash]entity.Signature
	signRequests   map[common.Hash]entity.SignatureRequest
	aggProofs      map[common.Hash]entity.AggregationProof
	pendingValsets map[common.Hash]entity.ValidatorSet
}

func New() (*Repository, error) {
	return &Repository{
		mu:             sync.Mutex{},
		networkConfigs: make(map[uint64]entity.NetworkConfig),
		validatorSets:  make(map[uint64]entity.ValidatorSet),
		signatures:     make(map[common.Hash]map[common.Hash]entity.Signature),
		signRequests:   make(map[common.Hash]entity.SignatureRequest),
		aggProofs:      make(map[common.Hash]entity.AggregationProof),
		pendingValsets: make(map[common.Hash]entity.ValidatorSet),
	}, nil
}

func (r *Repository) SaveConfig(_ context.Context, config entity.NetworkConfig, epoch uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.networkConfigs[epoch]; ok {
		return errors.New("validator set config for this epoch already exists")
	}

	r.networkConfigs[epoch] = config
	return nil
}

func (r *Repository) GetConfigByEpoch(_ context.Context, epoch uint64) (entity.NetworkConfig, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	config, ok := r.networkConfigs[epoch]
	if !ok {
		return entity.NetworkConfig{}, errors.New(entity.ErrEntityNotFound)
	}

	return config, nil
}

func (r *Repository) GetLatestValidatorSet(_ context.Context) (entity.ValidatorSet, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.validatorSets) == 0 {
		return entity.ValidatorSet{}, errors.New(entity.ErrEntityNotFound)
	}

	latestValset := lo.MaxBy(lo.Values(r.validatorSets), func(a entity.ValidatorSet, b entity.ValidatorSet) bool {
		return a.Epoch > b.Epoch
	})

	return r.validatorSets[latestValset.Epoch], nil
}

func (r *Repository) SaveValidatorSet(_ context.Context, valset entity.ValidatorSet) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.validatorSets[valset.Epoch]; ok {
		return errors.New("validator set for this epoch already exists")
	}

	r.validatorSets[valset.Epoch] = valset
	return nil
}

func (r *Repository) GetValidatorSetByEpoch(_ context.Context, epoch uint64) (entity.ValidatorSet, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	valset, ok := r.validatorSets[epoch]
	if !ok {
		return entity.ValidatorSet{}, errors.New(entity.ErrEntityNotFound)
	}

	return valset, nil
}

func (r *Repository) GetSignatureRequest(_ context.Context, reqHash common.Hash) (entity.SignatureRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existingReq, exists := r.signRequests[reqHash]
	if !exists {
		return entity.SignatureRequest{}, errors.New(entity.ErrEntityNotFound)
	}

	return existingReq, nil
}

func (r *Repository) SaveSignatureRequest(_ context.Context, req entity.SignatureRequest) error {
	hash := req.Hash()
	r.mu.Lock()
	defer r.mu.Unlock()

	r.signRequests[hash] = req

	return nil
}

func (r *Repository) GetAggregationProof(_ context.Context, reqHash common.Hash) (entity.AggregationProof, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	proof, exists := r.aggProofs[reqHash]
	if exists {
		return proof, nil
	}

	return entity.AggregationProof{}, errors.New(entity.ErrEntityNotFound)
}

func (r *Repository) SaveAggregationProof(_ context.Context, reqHash common.Hash, ap entity.AggregationProof) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.aggProofs[reqHash]; exists {
		return errors.Errorf("aggregation proof for this request already exists: %w", entity.ErrEntityAlreadyExist)
	}

	r.aggProofs[reqHash] = ap

	return nil
}

func (r *Repository) SaveSignature(_ context.Context, reqHash common.Hash, key []byte, sig entity.Signature) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	keyHash := crypto.Keccak256Hash(key)

	_, exists := r.signatures[reqHash]
	if !exists {
		r.signatures[reqHash] = make(map[common.Hash]entity.Signature)
	}

	if _, exists = r.signatures[reqHash][keyHash]; exists {
		return nil
	}

	r.signatures[reqHash][keyHash] = sig

	return nil
}

func (r *Repository) GetAllSignatures(_ context.Context, reqHash common.Hash) ([]entity.Signature, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.signatures[reqHash]
	if !exists {
		return []entity.Signature{}, nil
	}

	return slices.Collect(maps.Values(r.signatures[reqHash])), nil
}

func (r *Repository) SavePendingValidatorSet(_ context.Context, reqHash common.Hash, valset entity.ValidatorSet) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.pendingValsets[reqHash]
	if exists {
		return errors.Errorf("pending valset for this request already exists: %w", entity.ErrEntityAlreadyExist)
	}

	r.pendingValsets[reqHash] = valset
	return nil
}

func (r *Repository) GetPendingValidatorSet(_ context.Context, reqHash common.Hash) (entity.ValidatorSet, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	valset, ok := r.pendingValsets[reqHash]
	if !ok {
		return entity.ValidatorSet{}, errors.New(entity.ErrEntityNotFound)
	}
	return valset, nil
}
