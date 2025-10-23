package signer_app

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func (s *SignerApp) HandleSignaturesAggregatedMessage(ctx context.Context, p2pMsg entity.P2PMessage[symbiotic.AggregationProof]) error {
	ctx = log.WithComponent(ctx, "signer")

	ctx = log.WithAttrs(ctx,
		slog.Uint64("epoch", uint64(p2pMsg.Message.Epoch)),
		slog.String("requestId", p2pMsg.Message.RequestID().Hex()),
	)

	msg := p2pMsg.Message

	err := s.cfg.EntityProcessor.ProcessAggregationProof(ctx, msg)
	if err != nil {
		// if the aggregation proof already exists, we have already seen the message and broadcasted it so short-circuit
		if errors.Is(err, entity.ErrEntityAlreadyExist) {
			slog.DebugContext(ctx, "Aggregation proof already exists, skipping")
			return nil
		}
		return err
	}

	slog.DebugContext(ctx, "Proof saved")

	return nil
}
