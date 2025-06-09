package signer_app

import (
	"context"

	"github.com/go-errors/errors"

	"middleware-offchain/internal/entity"
)

func (s *SignerApp) HandleSignaturesAggregatedMessage(ctx context.Context, msg entity.P2PSignaturesAggregatedMessage) error {
	validatorSet, err := s.cfg.Repo.GetValsetByEpoch(ctx, msg.Message.Epoch)
	if err != nil {
		return errors.Errorf("failed to get validator set: %w", err)
	}

	ok, err := s.cfg.Aggregator.Verify(&validatorSet, msg.Message.KeyTag, &msg.Message.AggregationProof)
	if err != nil {
		return errors.Errorf("failed to verify aggregation proof: %w", err)
	}
	if !ok {
		return errors.Errorf("aggregation proof invalid")
	}

	err = s.cfg.Repo.SaveAggregationProof(ctx, msg.Message.RequestHash, msg.Message.AggregationProof)
	if err != nil {
		return err
	}

	s.cfg.AggProofSignal.Emit(ctx, msg.Message)

	return nil
}
