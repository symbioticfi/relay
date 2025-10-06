package sync_provider

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
)

type repo interface {
	GetSignaturePendingByEpoch(_ context.Context, epoch entity.Epoch, limit int, lastHash common.Hash) ([]entity.SignatureRequestWithID, error)
	GetSignatureMap(ctx context.Context, requestID common.Hash) (entity.SignatureMap, error)
	GetLatestValidatorSetEpoch(ctx context.Context) (entity.Epoch, error)
	GetSignatureRequest(ctx context.Context, requestID common.Hash) (entity.SignatureRequest, error)
	GetValidatorByKey(ctx context.Context, epoch entity.Epoch, keyTag entity.KeyTag, publicKey []byte) (entity.Validator, uint32, error)
	GetAllSignatures(ctx context.Context, requestID common.Hash) ([]entity.SignatureExtended, error)
	GetSignatureByIndex(ctx context.Context, requestID common.Hash, validatorIndex uint32) (entity.SignatureExtended, error)
	GetSignatureRequestsWithoutAggregationProof(ctx context.Context, epoch entity.Epoch, limit int, lastHash common.Hash) ([]entity.SignatureRequestWithID, error)
	GetAggregationProof(ctx context.Context, requestID common.Hash) (entity.AggregationProof, error)
	RemoveAggregationProofPending(ctx context.Context, epoch entity.Epoch, requestID common.Hash) error
}

type entityProcessor interface {
	ProcessSignature(ctx context.Context, signature entity.SignatureExtended) error
	ProcessAggregationProof(ctx context.Context, proof entity.AggregationProof) error
}

type Config struct {
	Repo                        repo            `validate:"required"`
	EntityProcessor             entityProcessor `validate:"required"`
	EpochsToSync                uint64          `validate:"gte=0"`
	MaxSignatureRequestsPerSync int             `validate:"gt=0"`
	MaxResponseSignatureCount   int             `validate:"gt=0"`
	MaxAggProofRequestsPerSync  int             `validate:"gt=0"`
	MaxResponseAggProofCount    int             `validate:"gt=0"`
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
