package valset_generator

import (
	"context"
	"encoding/hex"
	"log/slog"

	"github.com/go-errors/errors"

	"middleware-offchain/internal/entity"
)

func (s *Service) HandleProofAggregated(ctx context.Context, msg entity.AggregatedSignatureMessage) error {
	if !s.cfg.IsCommitter {
		slog.DebugContext(ctx, "not a committer, skipping proof commitment")
		return nil
	}

	aggProof, err := s.cfg.Repo.GetAggregationProof(ctx, msg.RequestHash)
	if err != nil {
		return errors.Errorf("failed to get aggregation proof: %w", err)
	}

	valset, err := s.cfg.Repo.GetPendingValset(ctx, msg.RequestHash)
	if err != nil {
		return errors.Errorf("failed to get pending valset: %w", err)
	}

	slog.DebugContext(ctx, "proof data", "proof", hex.EncodeToString(msg.AggregationProof.Proof))

	config, err := s.cfg.Eth.GetConfig(ctx, valset.CaptureTimestamp)
	if err != nil {
		return errors.Errorf("failed to get config for epoch %d: %w", msg.Epoch, err)
	}

	extraData, err := s.cfg.Deriver.GenerateExtraData(valset, config)
	if err != nil {
		return errors.Errorf("failed to generate extra data for validator set: %w", err)
	}

	header, err := valset.GetHeader()
	slog.DebugContext(ctx, "On commit header", "header", header)
	slog.DebugContext(ctx, "On commit extra data", "header", extraData)
	if err != nil {
		return errors.Errorf("failed to get validator set header: %w", err)
	}

	networkData, err := s.cfg.Deriver.GetNetworkData(ctx)
	if err != nil {
		return errors.Errorf("failed to get network data: %w", err)
	}

	hash, err := s.cfg.Deriver.HeaderCommitmentHash(networkData, header, extraData)
	if err != nil {
		return errors.Errorf("failed to get header commitment hash: %w", err)
	}
	slog.DebugContext(ctx, "committing commitment hash", "hash", hex.EncodeToString(hash))

	result, err := s.cfg.Eth.CommitValsetHeader(ctx, header, extraData, aggProof.Proof)
	if err != nil {
		return errors.Errorf("failed to commit valset header: %w", err)
	}

	slog.InfoContext(ctx, "valset header committed", "txHash", result.TxHash.String(), "epoch", msg.Epoch)

	return nil
}
