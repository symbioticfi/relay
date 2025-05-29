package memory

import (
	"context"
	"sync"

	"github.com/go-errors/errors"
	"github.com/samber/mo"

	"middleware-offchain/internal/entity"
)

type Repository struct {
	mu sync.Mutex

	extras []entity.ValidatorSetExtra
	signed []entity.ValidatorSetExtra
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
