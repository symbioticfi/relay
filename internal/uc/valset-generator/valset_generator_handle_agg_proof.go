package valset_generator

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"

	"middleware-offchain/internal/entity"
)

func (s *Service) HandleProofAggregated(ctx context.Context, msg entity.SignaturesAggregatedMessage) error {
	if msg.HashType != entity.HashTypeValsetHeader {
		slog.InfoContext(ctx, "ignoring non-valset header aggregated message", "hashType", msg.HashType)
		return nil
	}

	aggProof, err := s.cfg.Repo.GetAggregationProof(ctx, msg.Request)
	if err != nil {
		return errors.Errorf("failed to get aggregation proof: %w", err)
	}
	config, err := s.cfg.Repo.GetConfigByEpoch(ctx, msg.Request.RequiredEpoch)
	if err != nil {
		return errors.Errorf("failed to get config for epoch %d: %w", msg.Request.RequiredEpoch, err)
	}

	validatorSet, err := s.cfg.Repo.GetValsetByEpoch(ctx, msg.Request.RequiredEpoch)
	if err != nil {
		return errors.Errorf("failed to get validator set for epoch %d: %w", msg.Request.RequiredEpoch, err)
	}

	extraData, err := s.cfg.Deriver.GenerateExtraData(validatorSet, config)
	if err != nil {
		return errors.Errorf("failed to generate extra data for validator set: %w", err)
	}

	header, err := validatorSet.GetHeader()
	if err != nil {
		return errors.Errorf("failed to get validator set header: %w", err)
	}

	result, err := s.cfg.Eth.CommitValsetHeader(ctx, header, extraData, aggProof.Proof)
	if err != nil {
		return errors.Errorf("failed to commit valset header: %w", err)
	}

	slog.InfoContext(ctx, "valset header committed", "txHash", result.TxHash.String(), "epoch", msg.Request.RequiredEpoch)

	return nil
}
