package signer_app

import (
	"context"

	"github.com/go-errors/errors"

	"middleware-offchain/core/entity"
	p2pEntity "middleware-offchain/internal/entity"
	"middleware-offchain/pkg/log"
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
		return err
	}

	s.cfg.AggProofSignal.Emit(ctx, msg)

	return nil
}
