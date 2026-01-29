package signature_listener

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/signals"
	"github.com/symbioticfi/relay/pkg/tracing"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

//go:generate mockgen -source=signature_listener_uc.go -destination=mocks/signature_listener_uc.go -package=mocks

type repo interface {
	GetValidatorByKey(ctx context.Context, epoch symbiotic.Epoch, keyTag symbiotic.KeyTag, publicKey []byte) (symbiotic.Validator, uint32, error)
	GetValidatorSetByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSet, error)
}

type entityProcessor interface {
	ProcessSignature(ctx context.Context, signature symbiotic.Signature, self bool) error
}

type Config struct {
	Repo            repo            `validate:"required"`
	EntityProcessor entityProcessor `validate:"required"`
	SignalCfg       signals.Config  `validate:"required"`
	SelfP2PID       string          `validate:"required"`
}

type SignatureListenerUseCase struct {
	cfg Config
}

func New(cfg Config) (*SignatureListenerUseCase, error) {
	if err := validate.New().Struct(cfg); err != nil {
		return nil, errors.Errorf("invalid config: %w", err)
	}

	return &SignatureListenerUseCase{
		cfg: cfg,
	}, nil
}

func (s *SignatureListenerUseCase) HandleSignatureReceivedMessage(ctx context.Context, p2pMsg entity.P2PMessage[symbiotic.Signature]) error {
	if len(p2pMsg.TraceContext) > 0 {
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(p2pMsg.TraceContext))
	}

	ctx, span := tracing.StartConsumerSpan(ctx, "signature_listener.HandleSignatureReceived",
		tracing.AttrPeerID.String(p2pMsg.SenderInfo.Sender),
		tracing.AttrRequestID.String(p2pMsg.Message.RequestID().Hex()),
		tracing.AttrEpoch.Int64(int64(p2pMsg.Message.Epoch)),
		tracing.AttrKeyTag.String(p2pMsg.Message.KeyTag.String()),
	)
	defer span.End()
	ctx = log.WithComponent(ctx, "sign_listener")
	ctx = log.WithAttrs(ctx,
		slog.Uint64("epoch", uint64(p2pMsg.Message.Epoch)),
		slog.String("requestId", p2pMsg.Message.RequestID().Hex()),
	)

	msg := p2pMsg.Message

	slog.DebugContext(ctx, "Received signature message", "message", msg, "sender", p2pMsg.SenderInfo.Sender)

	if p2pMsg.SenderInfo.Sender == s.cfg.SelfP2PID {
		slog.DebugContext(ctx, "Skipped signature from self, already stored by signer")
		tracing.AddEvent(span, "ignored_self_message")
		return nil
	}

	err := s.cfg.EntityProcessor.ProcessSignature(ctx, msg, false)
	if err != nil {
		if errors.Is(err, entity.ErrEntityAlreadyExist) {
			slog.DebugContext(ctx, "Skipped signature, already exists")
			return nil
		}
		tracing.RecordError(span, err)
		return errors.Errorf("failed to process signature: %w", err)
	}

	slog.InfoContext(ctx, "Listener processed received signature")
	return nil
}
