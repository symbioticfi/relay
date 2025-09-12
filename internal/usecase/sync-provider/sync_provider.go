package sync_provider

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/pkg/signals"
)

type repo interface {
	GetSignatureRequestsByEpochPending(_ context.Context, epoch entity.Epoch, limit int, lastHash common.Hash) ([]entity.SignatureRequest, error)
	GetSignatureMap(ctx context.Context, reqHash common.Hash) (entity.SignatureMap, error)
	GetLatestValidatorSetEpoch(ctx context.Context) (uint64, error)
	GetSignatureRequest(ctx context.Context, reqHash common.Hash) (entity.SignatureRequest, error)
	GetValidatorByKey(ctx context.Context, epoch uint64, keyTag entity.KeyTag, publicKey []byte) (entity.Validator, uint32, error)
	GetAllSignatures(ctx context.Context, reqHash common.Hash) ([]entity.SignatureExtended, error)
}

type signatureProcessor interface {
	ProcessSignature(ctx context.Context, param entity.SaveSignatureParam) error
}

type Config struct {
	Repo                        repo                                     `validate:"required"`
	SignatureProcessor          signatureProcessor                       `validate:"required"`
	EpochsToSync                uint64                                   `validate:"gte=0"`
	MaxSignatureRequestsPerSync int                                      `validate:"gt=0"`
	MaxResponseSignatureCount   int                                      `validate:"gt=0"`
	SignatureReceivedSignal     *signals.Signal[entity.SignatureMessage] `validate:"required"`
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
