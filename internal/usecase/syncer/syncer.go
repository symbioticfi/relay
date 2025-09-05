package syncer

import (
	"context"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
)

// SignatureProcessingStats contains detailed statistics for processing received signatures
type SignatureProcessingStats struct {
	ProcessedCount             int // Successfully processed signatures
	UnrequestedSignatureCount  int // Signatures for validators we didn't request
	UnrequestedHashCount       int // Signatures for hashes we didn't request
	SignatureRequestErrorCount int // Failed to get signature request
	PublicKeyErrorCount        int // Failed to create public key from signature
	ValidatorInfoErrorCount    int // Failed to get validator info
	ProcessingErrorCount       int // Failed to process signature
	AlreadyExistCount          int // Signature already exists (ErrEntityAlreadyExist)
}

// TotalErrors returns the total number of errors encountered
func (s SignatureProcessingStats) TotalErrors() int {
	return s.UnrequestedSignatureCount + s.UnrequestedHashCount + s.SignatureRequestErrorCount +
		s.PublicKeyErrorCount + s.ValidatorInfoErrorCount + s.ProcessingErrorCount + s.AlreadyExistCount
}

type repo interface {
	GetSignatureRequestsByEpochPending(_ context.Context, epoch entity.Epoch, limit int, lastHash common.Hash) ([]entity.SignatureRequest, error)
	GetSignatureMap(ctx context.Context, reqHash common.Hash) (entity.SignatureMap, error)
	GetLatestValidatorSetEpoch(ctx context.Context) (uint64, error)
	GetActiveValidatorCountByEpoch(ctx context.Context, epoch uint64) (uint32, error)
	GetSignatureRequest(ctx context.Context, reqHash common.Hash) (entity.SignatureRequest, error)
	GetValidatorByKey(ctx context.Context, epoch uint64, keyTag entity.KeyTag, publicKey []byte) (entity.Validator, uint32, error)
	GetAllSignatures(ctx context.Context, reqHash common.Hash) ([]entity.SignatureExtended, error)
}

type p2pService interface {
	SendWantSignaturesRequest(ctx context.Context, request entity.WantSignaturesRequest) (entity.WantSignaturesResponse, error)
}

type signatureProcessor interface {
	ProcessSignature(ctx context.Context, param entity.SaveSignatureParam) error
}

type Config struct {
	Repo                        repo               `validate:"required"`
	P2PService                  p2pService         `validate:"required"`
	SignatureProcessor          signatureProcessor `validate:"required"`
	EpochsToSync                int                `validate:"gte=0"`
	SyncPeriod                  time.Duration      `validate:"gt=0"`
	SyncTimeout                 time.Duration      `validate:"gt=0"`
	MaxSignatureRequestsPerSync int                `validate:"gt=0"`
	MaxResponseSignatureCount   int                `validate:"gt=0"`
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

func (s *Syncer) Start(ctx context.Context) error {
	timer := time.NewTimer(s.cfg.SyncPeriod)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			if err := s.askSignatures(ctx); err != nil {
				slog.ErrorContext(ctx, "Failed to ask signatures", "error", err)
			}
			timer.Reset(s.cfg.SyncPeriod)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
