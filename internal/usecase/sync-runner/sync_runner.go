package sync_runner

import (
	"context"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/pkg/log"
)

type p2pService interface {
	SendWantSignaturesRequest(ctx context.Context, request entity.WantSignaturesRequest) (entity.WantSignaturesResponse, error)
}

type provider interface {
	BuildWantSignaturesRequest(ctx context.Context) (entity.WantSignaturesRequest, error)
	ProcessReceivedSignatures(ctx context.Context, response entity.WantSignaturesResponse, wantSignatures map[common.Hash]entity.SignatureBitmap) entity.SignatureProcessingStats
}

type metrics interface {
	ObserveP2PSyncSignaturesProcessed(resultType string, count int)
	ObserveP2PSyncRequestedHashes(count int)
}

type Config struct {
	Enabled     bool
	P2PService  p2pService    `validate:"required"`
	Provider    provider      `validate:"required"`
	SyncPeriod  time.Duration `validate:"gt=0"`
	SyncTimeout time.Duration `validate:"gt=0"`
	Metrics     metrics       `validate:"required"`
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
	if !s.cfg.Enabled {
		slog.InfoContext(ctx, "Signature sync runner is disabled")
		return nil
	}

	ctx = log.WithComponent(ctx, "sync_runner")

	timer := time.NewTimer(s.cfg.SyncPeriod)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			slog.DebugContext(ctx, "Signature sync started")
			if err := s.runSync(ctx); err != nil {
				slog.ErrorContext(ctx, "Failed to sync signatures", "error", err)
			}
			slog.DebugContext(ctx, "Signature sync completed")
			timer.Reset(s.cfg.SyncPeriod)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Runner) runSync(ctx context.Context) error {
	// Create context with timeout for the entire sync operation
	ctx, cancel := context.WithTimeout(ctx, s.cfg.SyncTimeout)
	defer cancel()

	request, err := s.cfg.Provider.BuildWantSignaturesRequest(ctx)
	if err != nil {
		return errors.Errorf("failed to build want signatures request: %w", err)
	}
	s.cfg.Metrics.ObserveP2PSyncRequestedHashes(len(request.WantSignatures))

	if len(request.WantSignatures) == 0 {
		slog.InfoContext(ctx, "No pending signature requests found")
		return nil
	}

	response, err := s.cfg.P2PService.SendWantSignaturesRequest(ctx, request)
	if err != nil {
		if errors.Is(err, entity.ErrNoPeers) {
			slog.DebugContext(ctx, "No peers available to request signatures from")
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
		"validator_index_missmatch_count", stats.ValidatorIndexMismatchCount,
		"processing_errors", stats.ProcessingErrorCount,
		"already_exist", stats.AlreadyExistCount,
	)

	s.cfg.Metrics.ObserveP2PSyncSignaturesProcessed("processed", stats.ProcessedCount)
	s.cfg.Metrics.ObserveP2PSyncSignaturesProcessed("unrequested_signatures", stats.UnrequestedSignatureCount)
	s.cfg.Metrics.ObserveP2PSyncSignaturesProcessed("unrequested_hashes", stats.UnrequestedHashCount)
	s.cfg.Metrics.ObserveP2PSyncSignaturesProcessed("signature_request_errors", stats.SignatureRequestErrorCount)
	s.cfg.Metrics.ObserveP2PSyncSignaturesProcessed("public_key_errors", stats.PublicKeyErrorCount)
	s.cfg.Metrics.ObserveP2PSyncSignaturesProcessed("validator_info_errors", stats.ValidatorInfoErrorCount)
	s.cfg.Metrics.ObserveP2PSyncSignaturesProcessed("validator_index_missmatch_count", stats.ValidatorIndexMismatchCount)
	s.cfg.Metrics.ObserveP2PSyncSignaturesProcessed("processing_errors", stats.ProcessingErrorCount)
	s.cfg.Metrics.ObserveP2PSyncSignaturesProcessed("already_exist", stats.AlreadyExistCount)

	return nil
}
