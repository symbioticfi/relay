package sync_provider

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
)

type repo interface {
	GetSignatureRequestsByEpochPending(_ context.Context, epoch entity.Epoch, limit int, lastHash common.Hash) ([]entity.SignatureRequestWithTargetID, error)
	GetSignatureMap(ctx context.Context, signatureTargetID common.Hash) (entity.SignatureMap, error)
	GetLatestValidatorSetEpoch(ctx context.Context) (uint64, error)
	GetSignatureRequest(ctx context.Context, signatureTargetID common.Hash) (entity.SignatureRequest, error)
	GetValidatorByKey(ctx context.Context, epoch uint64, keyTag entity.KeyTag, publicKey []byte) (entity.Validator, uint32, error)
	GetAllSignatures(ctx context.Context, signatureTargetID common.Hash) ([]entity.SignatureExtended, error)
	GetSignatureByIndex(ctx context.Context, signatureTargetID common.Hash, validatorIndex uint32) (entity.SignatureExtended, error)
	GetSignatureRequestsWithoutAggregationProof(ctx context.Context, epoch entity.Epoch, limit int, lastHash common.Hash) ([]entity.SignatureRequestWithTargetID, error)
	GetAggregationProof(ctx context.Context, signatureTargetID common.Hash) (entity.AggregationProof, error)
}

type entityProcessor interface {
	ProcessSignature(ctx context.Context, param entity.SaveSignatureParam) error
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
