package syncer

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
)

type repo interface {
	GetSignatureRequestsByEpoch(_ context.Context, epoch entity.Epoch, limit int, lastHash common.Hash) ([]entity.SignatureRequest, error)
}

type Config struct {
	Repo repo `validate:"required"`
}

type Syncer struct {
	cfg Config
}

func New(cfg Config) (*Syncer, error) {
	if err := validator.New().Struct(cfg); err != nil {
		return nil, err
	}
	return &Syncer{
		cfg: cfg,
	}, nil
}

func (s *Syncer) Sync(ctx context.Context, epoch entity.Epoch) error {
	return nil
}
