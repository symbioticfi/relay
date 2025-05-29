package memory

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"github.com/samber/mo"

	"middleware-offchain/internal/entity"
)

type Repository struct {
	mu sync.Mutex

	extras       []entity.ValidatorSetExtra
	signed       []entity.ValidatorSetExtra
	signatures   map[common.Hash]entity.Signature
	signRequests map[common.Hash]entity.SignatureRequest
	aggProofs    map[common.Hash]entity.AggregationProof
}

func New() (*Repository, error) {
	return &Repository{
		mu: sync.Mutex{},
	}, nil
}

func (r *Repository) GetLatestValsetExtra(ctx context.Context) (mo.Option[entity.ValidatorSetExtra], error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.extras) == 0 {
		return mo.None[entity.ValidatorSetExtra](), nil
	}

	latestExtra := r.extras[len(r.extras)-1]
	return mo.Some(latestExtra), nil
}

func (r *Repository) SaveValsetExtra(ctx context.Context, extra entity.ValidatorSetExtra) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if the extra already exists
	for _, existingExtra := range r.extras {
		if existingExtra.Epoch == extra.Epoch {
			return errors.New("validator set extra for this epoch already exists")
		}
	}

	// Append the new extra to the slice
	r.extras = append(r.extras, extra)
	return nil
}

func (r *Repository) SaveLatestSignedValsetExtra(ctx context.Context, extra entity.ValidatorSetExtra) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if the extra already exists
	for _, existingExtra := range r.signed {
		if existingExtra.Epoch == extra.Epoch {
			return errors.New("signed validator set extra for this epoch already exists")
		}
	}

	// Append the new signed extra to the slice
	r.signed = append(r.signed, extra)
	return nil
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

func (r *Repository) GetValsetExtraByEpoch(ctx context.Context, epoch *big.Int) (entity.ValidatorSetExtra, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, extra := range r.extras {
		if extra.Epoch.Cmp(epoch) == 0 {
			return extra, nil
		}
	}

	return entity.ValidatorSetExtra{}, errors.New("validator set extra for this epoch not found")
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
	return crypto.Keccak256Hash([]byte{req.KeyTag}, req.RequiredEpoch.Bytes(), req.Message)
}
