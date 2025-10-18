package sync_provider

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type repo interface {
	GetSignatureMap(ctx context.Context, requestID common.Hash) (entity.SignatureMap, error)
	GetLatestValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error)
	GetSignatureRequest(ctx context.Context, requestID common.Hash) (symbiotic.SignatureRequest, error)
	GetValidatorByKey(ctx context.Context, epoch symbiotic.Epoch, keyTag symbiotic.KeyTag, publicKey []byte) (symbiotic.Validator, uint32, error)
	GetAllSignatures(ctx context.Context, requestID common.Hash) ([]symbiotic.Signature, error)
	GetSignatureByIndex(ctx context.Context, requestID common.Hash, validatorIndex uint32) (symbiotic.Signature, error)
	GetSignatureRequestsWithoutAggregationProof(ctx context.Context, epoch symbiotic.Epoch, limit int, lastHash common.Hash) ([]symbiotic.SignatureRequestWithID, error)
	GetAggregationProof(ctx context.Context, requestID common.Hash) (symbiotic.AggregationProof, error)
	RemoveAggregationProofPending(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) error
}

type entityProcessor interface {
	ProcessSignature(ctx context.Context, signature symbiotic.Signature, self bool) error
	ProcessAggregationProof(ctx context.Context, proof symbiotic.AggregationProof) error
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
