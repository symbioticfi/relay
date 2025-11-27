package sync_runner

import (
	"context"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/tracing"
)

type p2pService interface {
	SendWantSignaturesRequest(ctx context.Context, request entity.WantSignaturesRequest) (entity.WantSignaturesResponse, error)
	SendWantAggregationProofsRequest(ctx context.Context, request entity.WantAggregationProofsRequest) (entity.WantAggregationProofsResponse, error)
}

type provider interface {
	BuildWantSignaturesRequest(ctx context.Context) (entity.WantSignaturesRequest, error)
	ProcessReceivedSignatures(ctx context.Context, response entity.WantSignaturesResponse, wantSignatures map[common.Hash]entity.Bitmap) entity.SignatureProcessingStats
	BuildWantAggregationProofsRequest(ctx context.Context) (entity.WantAggregationProofsRequest, error)
	ProcessReceivedAggregationProofs(ctx context.Context, response entity.WantAggregationProofsResponse) (entity.AggregationProofProcessingStats, error)
}

type metrics interface {
	ObserveP2PSyncSignaturesProcessed(resultType string, count int)
	ObserveP2PSyncRequestedHashes(count int)
	ObserveP2PSyncAggregationProofsProcessed(resultType string, count int)
	ObserveP2PSyncRequestedAggregationProofs(count int)
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
			slog.DebugContext(ctx, "Sync cycle started")

			// Run signature sync
			if err := s.runSignatureSync(ctx); err != nil {
				slog.ErrorContext(ctx, "Failed to sync signatures", "error", err)
			}

			// Run aggregation proof sync independently
			if err := s.runAggregationProofSync(ctx); err != nil {
				slog.ErrorContext(ctx, "Failed to sync aggregation proofs", "error", err)
			}

			slog.DebugContext(ctx, "Sync cycle completed")
			timer.Reset(s.cfg.SyncPeriod)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Runner) runSignatureSync(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx, "sync_runner.SyncSignatures")
	defer span.End()

	// Create context with timeout for signature sync
	ctx, cancel := context.WithTimeout(ctx, s.cfg.SyncTimeout)
	defer cancel()

	tracing.AddEvent(span, "building_request")
	request, err := s.cfg.Provider.BuildWantSignaturesRequest(ctx)
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to build want signatures request: %w", err)
	}

	tracing.SetAttributes(span, tracing.AttrSignatureCount.Int(len(request.WantSignatures)))
	s.cfg.Metrics.ObserveP2PSyncRequestedHashes(len(request.WantSignatures))

	if len(request.WantSignatures) == 0 {
		slog.InfoContext(ctx, "No pending signature requests found")
		return nil
	}

	tracing.AddEvent(span, "requesting_from_peers")
	response, err := s.cfg.P2PService.SendWantSignaturesRequest(ctx, request)
	if err != nil {
		if errors.Is(err, entity.ErrNoPeers) {
			slog.DebugContext(ctx, "No peers available to request signatures from")
			return nil
		}
		tracing.RecordError(span, err)
		return errors.Errorf("failed to send want signatures request: %w", err)
	}

	tracing.AddEvent(span, "processing_response")
	slog.DebugContext(ctx, "Received signature response", "signaturesCount", len(response.Signatures))

	stats := s.cfg.Provider.ProcessReceivedSignatures(ctx, response, request.WantSignatures)

	tracing.SetAttributes(span, tracing.AttrSignatureCount.Int(stats.ProcessedCount))
	slog.InfoContext(ctx, "Signature sync completed",
		"processed", stats.ProcessedCount,
		"totalFails", stats.TotalErrors(),
		"unrequestedSignatures", stats.UnrequestedSignatureCount,
		"unrequestedHashes", stats.UnrequestedHashCount,
		"signatureRequestFails", stats.SignatureRequestFailCount,
		"processingFails", stats.ProcessingFailCount,
		"alreadyExist", stats.AlreadyExistCount,
	)

	s.cfg.Metrics.ObserveP2PSyncSignaturesProcessed("processed", stats.ProcessedCount)
	s.cfg.Metrics.ObserveP2PSyncSignaturesProcessed("unrequested_signatures", stats.UnrequestedSignatureCount)
	s.cfg.Metrics.ObserveP2PSyncSignaturesProcessed("unrequested_hashes", stats.UnrequestedHashCount)
	s.cfg.Metrics.ObserveP2PSyncSignaturesProcessed("signature_request_fails", stats.SignatureRequestFailCount)
	s.cfg.Metrics.ObserveP2PSyncSignaturesProcessed("processing_fails", stats.ProcessingFailCount)
	s.cfg.Metrics.ObserveP2PSyncSignaturesProcessed("already_exist", stats.AlreadyExistCount)

	return nil
}

func (s *Runner) runAggregationProofSync(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx, "sync_runner.SyncAggregationProofs")
	defer span.End()

	// Create context with timeout for aggregation proof sync
	ctx, cancel := context.WithTimeout(ctx, s.cfg.SyncTimeout)
	defer cancel()

	tracing.AddEvent(span, "building_request")
	request, err := s.cfg.Provider.BuildWantAggregationProofsRequest(ctx)
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to build want aggregation proofs request: %w", err)
	}

	s.cfg.Metrics.ObserveP2PSyncRequestedAggregationProofs(len(request.RequestIDs))

	if len(request.RequestIDs) == 0 {
		slog.InfoContext(ctx, "No pending aggregation proof requests found")
		return nil
	}

	tracing.AddEvent(span, "requesting_from_peers")
	response, err := s.cfg.P2PService.SendWantAggregationProofsRequest(ctx, request)
	if err != nil {
		if errors.Is(err, entity.ErrNoPeers) {
			slog.DebugContext(ctx, "No peers available to request aggregation proofs from")
			return nil
		}
		tracing.RecordError(span, err)
		return errors.Errorf("failed to send want aggregation proofs request: %w", err)
	}

	tracing.AddEvent(span, "processing_response")
	slog.DebugContext(ctx, "Received aggregation proof response", "proofsCount", len(response.Proofs))

	stats, err := s.cfg.Provider.ProcessReceivedAggregationProofs(ctx, response)
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to process received aggregation proofs: %w", err)
	}

	tracing.SetAttributes(span, tracing.AttrProofSize.Int(stats.ProcessedCount))
	slog.InfoContext(ctx, "Aggregation proof sync completed",
		"processed", stats.ProcessedCount,
		"totalFails", stats.TotalErrors(),
		"unrequestedProofs", stats.UnrequestedProofCount,
		"verificationFails", stats.VerificationFailCount,
		"processingFails", stats.ProcessingFailCount,
		"alreadyExist", stats.AlreadyExistCount,
	)

	s.cfg.Metrics.ObserveP2PSyncAggregationProofsProcessed("processed", stats.ProcessedCount)
	s.cfg.Metrics.ObserveP2PSyncAggregationProofsProcessed("unrequested_proofs", stats.UnrequestedProofCount)
	s.cfg.Metrics.ObserveP2PSyncAggregationProofsProcessed("verification_fails", stats.VerificationFailCount)
	s.cfg.Metrics.ObserveP2PSyncAggregationProofsProcessed("processing_fails", stats.ProcessingFailCount)
	s.cfg.Metrics.ObserveP2PSyncAggregationProofsProcessed("already_exist", stats.AlreadyExistCount)

	return nil
}
