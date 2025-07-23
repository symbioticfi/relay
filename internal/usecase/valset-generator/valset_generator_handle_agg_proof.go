package valset_generator

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/pkg/log"
)

func (s *Service) HandleProofAggregated(ctx context.Context, msg entity.AggregatedSignatureMessage) error {
	ctx = log.WithComponent(ctx, "generator")

	slog.DebugContext(ctx, "Handling proof aggregated message", "msg", msg)
	if !s.cfg.IsCommitter {
		slog.DebugContext(ctx, "Not a committer, skipping proof commitment")
		return nil
	}

	var (
		valset entity.ValidatorSet
		err    error
	)
	retryAttempted := false
	for {
		valset, err = s.cfg.Repo.GetPendingValidatorSet(ctx, msg.RequestHash)
		if err != nil {
			if errors.Is(err, entity.ErrEntityNotFound) && !retryAttempted {
				if err = s.process(ctx); err != nil {
					slog.ErrorContext(ctx, "failed to process epochs, on demand from commiter", "error", err)
					return nil
				}
				retryAttempted = true
				continue // retry after processing
			}
			slog.DebugContext(ctx, "No pending valset, skipping proof commitment")
			return nil
		}
		break
	}

	config, err := s.cfg.EvmClient.GetConfig(ctx, valset.CaptureTimestamp)
	if err != nil {
		return errors.Errorf("failed to get config for epoch %d: %w", msg.Epoch, err)
	}

	extraData, err := s.cfg.Aggregator.GenerateExtraData(valset, config.RequiredKeyTags)
	if err != nil {
		return errors.Errorf("failed to generate extra data for validator set: %w", err)
	}

	header, err := valset.GetHeader()
	if err != nil {
		return errors.Errorf("failed to get validator set header: %w", err)
	}
	slog.DebugContext(ctx, "On commit proof", "header", header, "extraData", extraData)

	err = s.commitValsetToAllSettlements(ctx, config, header, extraData, msg.AggregationProof.Proof)
	if err != nil {
		return errors.Errorf("failed to commit valset header: %w", err)
	}

	return nil
}

func (s *Service) commitValsetToAllSettlements(ctx context.Context, config entity.NetworkConfig, header entity.ValidatorSetHeader, extraData []entity.ExtraData, proof []byte) error {
	errs := make([]error, len(config.Replicas))
	for i, replica := range config.Replicas {
		slog.DebugContext(ctx, "Trying to commit valset header to settlement", "replica", replica)

		result, err := s.cfg.EvmClient.CommitValsetHeader(ctx, replica, header, extraData, proof)
		if err != nil {
			errs[i] = errors.Errorf("failed to commit valset header to settlement %s: %w", replica.Address.Hex(), err)
			continue
		}

		slog.DebugContext(ctx, "Validator set header committed",
			"epoch", header.Epoch,
			"replica", replica,
			"txHash", result.TxHash,
		)
	}

	return errors.Join(errs...)
}
