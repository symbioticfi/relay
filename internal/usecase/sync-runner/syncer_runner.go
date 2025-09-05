package sync_runner

import (
	"context"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
)

type p2pService interface {
	SendWantSignaturesRequest(ctx context.Context, request entity.WantSignaturesRequest) (entity.WantSignaturesResponse, error)
}

type provider interface {
	BuildWantSignaturesRequest(ctx context.Context) (entity.WantSignaturesRequest, error)
	ProcessReceivedSignatures(ctx context.Context, response entity.WantSignaturesResponse, wantSignatures map[common.Hash]entity.SignatureBitmap) entity.SignatureProcessingStats
}

type Config struct {
	P2PService  p2pService    `validate:"required"`
	Provider    provider      `validate:"required"`
	SyncPeriod  time.Duration `validate:"gt=0"`
	SyncTimeout time.Duration `validate:"gt=0"`
}

type Runner struct {
	cfg Config
}

func New(cfg Config) (*Runner, error) {
	if err := validator.New().Struct(cfg); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}
	return &Runner{
		cfg: cfg,
	}, nil
}

func (s *Runner) Start(ctx context.Context) error {
	timer := time.NewTimer(s.cfg.SyncPeriod)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			if err := s.runSync(ctx); err != nil {
				slog.ErrorContext(ctx, "Failed to ask signatures", "error", err)
			}
			timer.Reset(s.cfg.SyncPeriod)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Runner) runSync(ctx context.Context) error {
	// Create context with timeout for the entire sync operation
	syncCtx, cancel := context.WithTimeout(ctx, s.cfg.SyncTimeout)
	defer cancel()

	slog.InfoContext(syncCtx, "Starting signature sync")

	request, err := s.cfg.Provider.BuildWantSignaturesRequest(ctx)
	if err != nil {
		return errors.Errorf("failed to build want signatures request: %w", err)
	}

	response, err := s.cfg.P2PService.SendWantSignaturesRequest(ctx, request)
	if err != nil {
		if errors.Is(err, entity.ErrNoPeers) {
			return nil
		}
		return errors.Errorf("failed to send want signatures request: %w", err)
	}

	slog.InfoContext(ctx, "Received signature response", "signatures_count", len(response.Signatures))

	stats := s.cfg.Provider.ProcessReceivedSignatures(ctx, response, request.WantSignatures)

	slog.InfoContext(ctx, "Signature sync completed",
		"processed", stats.ProcessedCount,
		"total_errors", stats.TotalErrors(),
		"unrequested_signatures", stats.UnrequestedSignatureCount,
		"unrequested_hashes", stats.UnrequestedHashCount,
		"signature_request_errors", stats.SignatureRequestErrorCount,
		"public_key_errors", stats.PublicKeyErrorCount,
		"validator_info_errors", stats.ValidatorInfoErrorCount,
		"processing_errors", stats.ProcessingErrorCount,
		"already_exist", stats.AlreadyExistCount,
	)

	return nil
}
