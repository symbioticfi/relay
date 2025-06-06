package signer_app

import (
	"context"

	"middleware-offchain/internal/entity"
)

func (s *SignerApp) HandleSignaturesAggregatedMessage(ctx context.Context, msg entity.P2PSignaturesAggregatedMessage) error {
	// todo ilya validate proof before saving
	err := s.cfg.Repo.SaveAggregationProof(ctx, msg.Message.Request, msg.Message.Proof)
	if err != nil {
		return err
	}

	return nil
}
