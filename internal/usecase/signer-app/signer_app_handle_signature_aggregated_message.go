package signer_app

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
	p2pEntity "github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
)

func (s *SignerApp) HandleSignaturesAggregatedMessage(ctx context.Context, p2pMsg p2pEntity.P2PMessage[entity.AggregatedSignatureMessage]) error {
	ctx = log.WithComponent(ctx, "signer")
	msg := p2pMsg.Message
	validatorSet, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, uint64(msg.Epoch))
	if err != nil {
		return errors.Errorf("failed to get validator set: %w", err)
	}

	ok, err := s.cfg.Aggregator.Verify(validatorSet, msg.KeyTag, msg.AggregationProof)
	if err != nil {
		return errors.Errorf("failed to verify aggregation proof: %w", err)
	}
	if !ok {
		return errors.Errorf("aggregation proof invalid")
	}

	err = s.cfg.Repo.SaveAggregationProof(ctx, msg.RequestHash, msg.AggregationProof)
	if err != nil {
		// if the aggregation proof already exists, we have already seen the message and broadcasted it so short-circuit
		if errors.Is(err, entity.ErrEntityAlreadyExist) {
			slog.DebugContext(ctx, "Aggregation proof already exists, skipping", "requestHash", msg.RequestHash)
			return nil
		}
		return err
	}

	s.cfg.AggProofSignal.Emit(ctx, msg)

	return nil
}
