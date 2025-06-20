package valset_generator

import (
	"context"
	"encoding/hex"
	"log/slog"

	"github.com/go-errors/errors"

	"middleware-offchain/core/entity"
)

func (s *Service) HandleProofAggregated(ctx context.Context, msg entity.AggregatedSignatureMessage) error {
	if !s.cfg.IsCommitter {
		slog.DebugContext(ctx, "not a committer, skipping proof commitment")
		return nil
	}

	valset, err := s.cfg.Repo.GetPendingValidatorSet(ctx, msg.RequestHash)
	if err != nil {
		slog.DebugContext(ctx, "no pending valset, skipping proof commitment")
		return nil //nolint:nilerr // if no pending valset, nothing to commit
	}

	slog.DebugContext(ctx, "proof data", "proof", hex.EncodeToString(msg.AggregationProof.Proof))

	config, err := s.cfg.Eth.GetConfig(ctx, valset.CaptureTimestamp)
	if err != nil {
		return errors.Errorf("failed to get config for epoch %d: %w", msg.Epoch, err)
	}

	extraData, err := s.cfg.Aggregator.GenerateExtraData(valset, config)
	if err != nil {
		return errors.Errorf("failed to generate extra data for validator set: %w", err)
	}

	header, err := valset.GetHeader()
	slog.DebugContext(ctx, "On commit header", "header", header)
	slog.DebugContext(ctx, "On commit extra data", "extraData", extraData)
	if err != nil {
		return errors.Errorf("failed to get validator set header: %w", err)
	}

	err = s.commitValsetToAllSettlements(ctx, config, header, extraData, msg.AggregationProof.Proof)
	if err != nil {
		return errors.Errorf("failed to commit valset header: %w", err)
	}

	return nil
}

func (s *Service) commitValsetToAllSettlements(ctx context.Context, config entity.NetworkConfig, header entity.ValidatorSetHeader, extraData []entity.ExtraData, proof []byte) error {
	errs := make([]error, len(config.Replicas))
	for i, replica := range config.Replicas {
		slog.DebugContext(ctx, "trying to commit valset header to settlement", "replica", replica)

		result, err := s.cfg.Eth.CommitValsetHeader(ctx, replica, header, extraData, proof)
		if err != nil {
			errs[i] = errors.Errorf("failed to commit valset header to settlement %s: %w", replica.Address.Hex(), err)
		}

		slog.DebugContext(ctx, "Validator set header committed",
			"epoch", header.Epoch,
			"replica", replica,
			"txHash", result.TxHash,
		)
	}

	return errors.Join(errs...)
}
