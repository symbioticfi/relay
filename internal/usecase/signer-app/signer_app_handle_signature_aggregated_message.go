package signer_app

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
	p2pEntity "github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
)

func (s *SignerApp) HandleSignaturesAggregatedMessage(ctx context.Context, p2pMsg p2pEntity.P2PMessage[entity.AggregatedSignatureMessage]) error {
	ctx = log.WithComponent(ctx, "signer")
	msg := p2pMsg.Message

	err := s.cfg.EntityProcessor.ProcessAggregationProof(ctx, msg)
	if err != nil {
		// if the aggregation proof already exists, we have already seen the message and broadcasted it so short-circuit
		if errors.Is(err, entity.ErrEntityAlreadyExist) {
			slog.DebugContext(ctx, "Aggregation proof already exists, skipping", "requestHash", msg.RequestHash)
			return nil
		}
		return err
	}

	stat, err := s.cfg.Repo.UpdateSignatureStat(ctx, msg.RequestHash, entity.SignatureStatStageAggProofReceived, time.Now())
	if err != nil {
		slog.WarnContext(ctx, "Failed to update signature stat", "error", err)
	}
	s.cfg.Metrics.ObserveAggReceived(stat)

	return nil
}
