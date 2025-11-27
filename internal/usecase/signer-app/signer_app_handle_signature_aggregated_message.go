package signer_app

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/tracing"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (s *SignerApp) HandleSignaturesAggregatedMessage(ctx context.Context, p2pMsg entity.P2PMessage[symbiotic.AggregationProof]) error {
	if len(p2pMsg.TraceContext) > 0 {
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(p2pMsg.TraceContext))
	}

	ctx, span := tracing.StartConsumerSpan(ctx, "signer.HandleAggregationProof",
		tracing.AttrPeerID.String(p2pMsg.SenderInfo.Sender),
		tracing.AttrRequestID.String(p2pMsg.Message.RequestID().Hex()),
		tracing.AttrEpoch.Int64(int64(p2pMsg.Message.Epoch)),
		tracing.AttrKeyTag.String(p2pMsg.Message.KeyTag.String()),
	)
	defer span.End()

	ctx = log.WithComponent(ctx, "signer")
	ctx = log.WithAttrs(ctx,
		slog.Uint64("epoch", uint64(p2pMsg.Message.Epoch)),
		slog.String("requestId", p2pMsg.Message.RequestID().Hex()),
	)

	msg := p2pMsg.Message

	tracing.AddEvent(span, "processing_aggregation_proof")
	err := s.cfg.EntityProcessor.ProcessAggregationProof(ctx, msg)
	if err != nil {
		if errors.Is(err, entity.ErrEntityAlreadyExist) {
			slog.DebugContext(ctx, "Skipped aggregation proof, already exists")
			tracing.AddEvent(span, "aggregation_proof_already_exists")
			return nil
		}
		tracing.RecordError(span, err)
		return err
	}

	tracing.AddEvent(span, "aggregation_proof_processed")
	return nil
}
