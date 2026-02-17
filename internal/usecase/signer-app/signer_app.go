package signer_app

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/tracing"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"
	"k8s.io/client-go/util/workqueue"
)

//go:generate mockgen -source=signer_app.go -destination=mocks/signer_app.go -package=mocks

type repo interface {
	SaveSignatureRequest(ctx context.Context, requestID common.Hash, req symbiotic.SignatureRequest) error
	RemoveSignaturePending(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) error
	GetSignaturePending(ctx context.Context, limit int) ([]common.Hash, error)
	GetSignatureRequest(ctx context.Context, requestID common.Hash) (symbiotic.SignatureRequest, error)
	GetValidatorSetByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSet, error)
}

type p2pService interface {
	BroadcastSignatureGeneratedMessage(ctx context.Context, msg symbiotic.Signature) error
}

type keyProvider interface {
	GetPrivateKey(keyTag symbiotic.KeyTag) (crypto.PrivateKey, error)
	GetOnchainKeyFromCache(keyTag symbiotic.KeyTag) (symbiotic.CompactPublicKey, error)
}

type metrics interface {
	ObservePKSignDuration(d time.Duration)
	ObserveAppSignDuration(d time.Duration)
}

type entityProcessor interface {
	ProcessSignature(ctx context.Context, signature symbiotic.Signature, self bool) error
	ProcessAggregationProof(ctx context.Context, proof symbiotic.AggregationProof) error
}

type Config struct {
	KeyProvider     keyProvider     `validate:"required"`
	Repo            repo            `validate:"required"`
	EntityProcessor entityProcessor `validate:"required"`
	Metrics         metrics         `validate:"required"`
}

func (c Config) Validate() error {
	if err := validate.New().Struct(c); err != nil {
		return errors.Errorf("failed to validate config: %w", err)
	}

	return nil
}

type SignerApp struct {
	cfg   Config
	queue *workqueue.Typed[common.Hash]
}

func NewSignerApp(cfg Config) (*SignerApp, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	app := &SignerApp{
		cfg:   cfg,
		queue: workqueue.NewTyped[common.Hash](),
	}

	return app, nil
}

// RequestSignature creates a signature request and queues it for signing, returns requestID
// The actual signing is done in the background by workers
func (s *SignerApp) RequestSignature(ctx context.Context, req symbiotic.SignatureRequest) (common.Hash, error) {
	ctx, span := tracing.StartServerSpan(ctx, "signer.RequestSignature",
		tracing.AttrEpoch.Int64(int64(req.RequiredEpoch)),
		tracing.AttrKeyTag.String(req.KeyTag.String()),
	)
	defer span.End()
	ctx = log.WithComponent(ctx, "signer")
	ctx = log.WithAttrs(ctx, slog.Uint64("epoch", uint64(req.RequiredEpoch)))

	if !req.KeyTag.Type().SignerKey() {
		err := errors.Errorf("key tag %s is not a signing key", req.KeyTag)
		tracing.RecordError(span, err)
		return common.Hash{}, err
	}

	msgHash, err := crypto.HashMessage(req.KeyTag.Type(), req.Message)
	if err != nil {
		tracing.RecordError(span, err)
		return common.Hash{}, errors.Errorf("failed to hash message: %w", err)
	}

	extendedSignature := symbiotic.Signature{
		MessageHash: msgHash,
		KeyTag:      req.KeyTag,
		Epoch:       req.RequiredEpoch,
	}

	requestId := extendedSignature.RequestID()
	tracing.SetAttributes(span, tracing.AttrRequestID.String(requestId.Hex()))

	err = s.cfg.Repo.SaveSignatureRequest(ctx, requestId, req)
	if err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
		tracing.RecordError(span, err)
		return common.Hash{}, errors.Errorf("failed to get signature request: %w", err)
	}

	s.queue.Add(requestId)

	tracing.AddEvent(span, "signature_requested")
	// does not return the actual signature yet
	return requestId, nil
}

func (s *SignerApp) EnqueueRequestID(ctx context.Context, requestID common.Hash) {
	s.queue.Add(requestID)
	slog.DebugContext(ctx, "Enqueued signature request", "requestId", requestID.Hex())
}

func (s *SignerApp) completeSign(ctx context.Context, requestID common.Hash, p2pService p2pService) error {
	ctx, span := tracing.StartSpan(ctx, "signer.completeSign",
		tracing.AttrRequestID.String(requestID.Hex()),
	)
	defer span.End()
	ctx = log.WithAttrs(ctx, slog.String("requestId", requestID.Hex()))

	req, err := s.cfg.Repo.GetSignatureRequest(ctx, requestID)
	if err != nil {
		return errors.Errorf("failed to get signature request for requestId %s : %w", requestID.Hex(), err)
	}
	ctx = log.WithAttrs(ctx,
		slog.Uint64("epoch", uint64(req.RequiredEpoch)),
		slog.Uint64("keyTag", uint64(req.KeyTag)),
	)

	span.SetAttributes(
		tracing.AttrEpoch.Int64(int64(req.RequiredEpoch)),
		tracing.AttrKeyTag.String(req.KeyTag.String()),
	)

	valset, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, req.RequiredEpoch)
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to get validator set: %w", err)
	}

	private, err := s.cfg.KeyProvider.GetPrivateKey(req.KeyTag)
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to get private key: %w", err)
	}

	onchainKey, err := s.cfg.KeyProvider.GetOnchainKeyFromCache(valset.RequiredKeyTag)
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to get onchain symb key from cache: %w", err)
	}

	if !valset.IsSigner(onchainKey) {
		tracing.AddEvent(span, "not_a_signer_skipping")
		slog.DebugContext(ctx, "Skipped signing, not a signer for this validator set",
			"key", onchainKey,
		)
		msgHash, err := crypto.HashMessage(req.KeyTag.Type(), req.Message)
		if err != nil {
			tracing.RecordError(span, err)
			return errors.Errorf("failed to hash message: %w", err)
		}

		extendedSignature := symbiotic.Signature{
			MessageHash: msgHash,
			KeyTag:      req.KeyTag,
			Epoch:       req.RequiredEpoch,
			PublicKey:   private.PublicKey(),
		}
		if err := s.cfg.Repo.RemoveSignaturePending(ctx, req.RequiredEpoch, extendedSignature.RequestID()); err != nil {
			tracing.RecordError(span, err)
			return errors.Errorf("failed to remove self signature request pending: %w", err)
		}
		return nil
	}

	timeAppSignStart := time.Now()

	pkSignStart := time.Now()
	signature, hash, err := private.Sign(req.Message)
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to sign valset header hash: %w", err)
	}
	s.cfg.Metrics.ObservePKSignDuration(time.Since(pkSignStart))

	extendedSignature := symbiotic.Signature{
		MessageHash: hash,
		KeyTag:      req.KeyTag,
		Epoch:       req.RequiredEpoch,
		PublicKey:   private.PublicKey(),
		Signature:   signature,
	}

	ctx = log.WithAttrs(ctx, slog.String("requestId", extendedSignature.RequestID().Hex()))

	if err := s.cfg.EntityProcessor.ProcessSignature(ctx, extendedSignature, true); err != nil {
		if errors.Is(err, entity.ErrEntityAlreadyExist) {
			slog.InfoContext(ctx, "Skipped signature, already exists")
			tracing.AddEvent(span, "signature_already_exists")
			return nil
		}
		tracing.RecordError(span, err)
		return errors.Errorf("failed to process signature: %w", err)
	}

	if err := s.cfg.Repo.RemoveSignaturePending(ctx, req.RequiredEpoch, extendedSignature.RequestID()); err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to remove self signature request pending: %w", err)
	}

	tracing.AddEvent(span, "broadcasting_signature")
	err = p2pService.BroadcastSignatureGeneratedMessage(ctx, extendedSignature)
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to broadcast signature: %w", err)
	}

	slog.InfoContext(ctx, "Message signed",
		"hash", hash,
		"signature", signature,
		"duration", time.Since(timeAppSignStart),
	)
	s.cfg.Metrics.ObserveAppSignDuration(time.Since(timeAppSignStart))

	return nil
}

func (s *SignerApp) HandleSignatureRequests(ctx context.Context, workerCount int, p2pService p2pService) error {
	ctx = log.WithComponent(ctx, "signer")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// start workers
	for i := range workerCount {
		go s.worker(ctx, i+1, p2pService)
	}

	for {
		select {
		case <-ctx.Done():
			s.queue.ShutDown()
			slog.InfoContext(ctx, "Stopping missed signatures handler")
			return nil
		case <-ticker.C:
			if err := s.handleMissedSignaturesOnce(ctx); err != nil {
				slog.ErrorContext(ctx, "Failed to handle missed signatures", "error", err)
			}
		}
	}
}

func (s *SignerApp) worker(ctx context.Context, id int, p2pService p2pService) {
	slog.InfoContext(ctx, "Signature worker started", "workerId", id)
	for {
		item, shutdown := s.queue.Get()
		if shutdown {
			slog.InfoContext(ctx, "Worker shutting down", "workerId", id)
			return
		}

		func() {
			defer s.queue.Done(item)

			if err := s.completeSign(ctx, item, p2pService); err != nil {
				slog.ErrorContext(ctx, "Failed to complete sign for request", "requestId", item.Hex(), "error", err)
				return
			}
		}()
	}
}

func (s *SignerApp) handleMissedSignaturesOnce(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx, "signer.handleMissedSignaturesOnce")
	defer span.End()

	tracing.AddEvent(span, "fetching_pending_requests")
	pendingRequests, err := s.cfg.Repo.GetSignaturePending(ctx, 10)
	if err != nil {
		return errors.Errorf("failed to get self signature requests pending: %w", err)
	}
	if len(pendingRequests) == 0 {
		tracing.AddEvent(span, "no_pending_requests")
		slog.DebugContext(ctx, "No pending self signature requests found")
		return nil
	}

	tracing.AddEvent(span, "queuing_pending_requests", attribute.Int("count", len(pendingRequests)))
	slog.InfoContext(ctx, "Found pending self signature requests", "count", len(pendingRequests))
	for _, reqID := range pendingRequests {
		slog.InfoContext(ctx, "Queued pending self signature request", "requestId", reqID.Hex())
		s.queue.Add(reqID)
	}
	return nil
}
