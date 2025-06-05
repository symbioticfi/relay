package memory

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/samber/lo"
	"github.com/samber/mo"

	"middleware-offchain/internal/entity"
)

type Repository struct {
	mu sync.Mutex

	networkConfigs map[uint64]entity.NetworkConfig
	validatorSets  map[uint64]entity.ValidatorSet
	signed         []entity.ValidatorSet
	signatures     map[common.Hash]entity.Signature
	signRequests   map[common.Hash]entity.SignatureRequest
	aggProofs      map[common.Hash]entity.AggregationProof
}

func New() (*Repository, error) {
	return &Repository{
		mu: sync.Mutex{},
	}, nil
}

func (r *Repository) GetLatestValset(ctx context.Context) (entity.ValidatorSet, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.networkConfigs) == 0 {
		return entity.ValidatorSet{}, errors.New(entity.ErrEntityNotFound)
	}

	latestValset := lo.MaxBy(lo.Values(r.validatorSets), func(a entity.ValidatorSet, b entity.ValidatorSet) bool {
		return a.Epoch < b.Epoch
	})

	return r.validatorSets[latestValset.Epoch], nil
}

func (r *Repository) SaveConfig(ctx context.Context, config entity.NetworkConfig, epoch uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if the config already exists
	if _, ok := r.networkConfigs[epoch]; ok {
		return errors.New("validator set config for this epoch already exists")
	}

	// Append the new config to the slice
	r.networkConfigs[epoch] = config
	return nil
}

func (r *Repository) SaveValidatorSet(ctx context.Context, valset entity.ValidatorSet) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if the config already exists
	if _, ok := r.validatorSets[valset.Epoch]; ok {
		return errors.New("validator set for this epoch already exists")
	}

	// Append the new config to the slice
	r.validatorSets[valset.Epoch] = valset
	return nil
}

func (r *Repository) GetLatestSignedValset(_ context.Context) (entity.ValidatorSet, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.signed) == 0 {
		return entity.ValidatorSet{}, errors.New(entity.ErrEntityNotFound)
	}

	latestSignedExtra := r.signed[len(r.signed)-1]
	return latestSignedExtra, nil
}

// todo ilya get rid of mo.Option in favor of returning error
func (r *Repository) GetSignatureRequest(_ context.Context, req entity.SignatureRequest) (mo.Option[entity.SignatureRequest], error) {
	hash := signRequestHash(req)
	r.mu.Lock()
	defer r.mu.Unlock()

	if existingReq, exists := r.signRequests[hash]; exists {
		return mo.Some(existingReq), nil
	}

	return mo.None[entity.SignatureRequest](), nil
}

func (r *Repository) GetAggregationProof(ctx context.Context, req entity.SignatureRequest) (mo.Option[entity.AggregationProof], error) {
	hash := signRequestHash(req)
	r.mu.Lock()
	defer r.mu.Unlock()

	if proof, exists := r.aggProofs[hash]; exists {
		return mo.Some(proof), nil
	}

	return mo.None[entity.AggregationProof](), nil
}

func (r *Repository) GetValsetByEpoch(ctx context.Context, epoch uint64) (entity.ValidatorSet, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	valset, ok := r.validatorSets[epoch]
	if !ok {
		return entity.ValidatorSet{}, errors.New(entity.ErrEntityNotFound)
	}

	return valset, nil
}

func (r *Repository) SaveSignature(ctx context.Context, req entity.SignatureRequest, sig entity.Signature) error {
	hash := signRequestHash(req)
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.signatures[hash]; exists {
		return errors.New("signature for this request already exists")
	}

	r.signatures[hash] = sig

	return nil
}

func signRequestHash(req entity.SignatureRequest) common.Hash {
	return crypto.Keccak256Hash([]byte{uint8(req.KeyTag)}, new(big.Int).SetInt64(int64(req.RequiredEpoch)).Bytes(), req.Message)
}
