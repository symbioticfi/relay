package syncer

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
)

type repo interface {
	GetSignatureRequestsByEpochPending(_ context.Context, epoch entity.Epoch, limit int, lastHash common.Hash) ([]entity.SignatureRequest, error)
	GetSignatureMap(ctx context.Context, reqHash common.Hash) (entity.SignatureMap, error)
	GetLatestValidatorSetEpoch(ctx context.Context) (uint64, error)
}

type p2pService interface {
	SendResyncRequest(ctx context.Context, request WantSignaturesRequest) (WantSignatureResponse, error)
}

type Config struct {
	Repo         repo          `validate:"required"`
	P2PService   p2pService    `validate:"required"`
	EpochsToSync int           `validate:"gte=0"`
	SyncPeriod   time.Duration `validate:"gt=0"`
}

type Syncer struct {
	cfg Config
}

func New(cfg Config) (*Syncer, error) {
	if err := validator.New().Struct(cfg); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}
	return &Syncer{
		cfg: cfg,
	}, nil
}

func (s *Syncer) Sync(ctx context.Context, epoch entity.Epoch) error {
	return nil
}
