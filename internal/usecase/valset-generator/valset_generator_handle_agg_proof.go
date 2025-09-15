package valset_generator

import (
	"context"
	"fmt"
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
		valset, err = s.cfg.Repo.GetValidatorSetByEpoch(ctx, uint64(msg.Epoch))
		if err != nil {
			if errors.Is(err, entity.ErrEntityNotFound) && !retryAttempted { // TODO: do i need to check if there is a local signature for request? it's still possible to commit
				if err = s.process(ctx); err != nil {
					slog.ErrorContext(ctx, "failed to process epochs, on demand from committer", "error", err)
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

	err = s.cfg.Repo.SaveAggregationProof(ctx, msg.RequestHash, msg.AggregationProof)
	if err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
		return err
	}

	fmt.Println(valset.Status)
	if valset.Status == entity.HeaderCommitted {
		slog.DebugContext(ctx, "Valset is already committed", "epoch", msg.Epoch)
		return nil
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
	errs := make([]error, len(config.Settlements))
	for i, settlement := range config.Settlements {
		slog.DebugContext(ctx, "Trying to commit valset header to settlement", "settlement", settlement)

		// todo replace it with tx check instead of call to contract
		// if commit tx was sent but still not finalized this check will
		// return false positive and trigger one more commitment tx
		committed, err := s.cfg.EvmClient.IsValsetHeaderCommittedAt(ctx, settlement, header.Epoch)
		if err != nil {
			errs[i] = errors.Errorf("failed to check if header is committed at epoch %d: %w", header.Epoch, err)
			break
		}

		if committed {
			continue
		}

		lastCommittedEpoch, err := s.cfg.EvmClient.GetLastCommittedHeaderEpoch(ctx, settlement)
		if err != nil {
			errs[i] = errors.Errorf("failed to get last committed header epoch: %w", err)
			break
		}

		if header.Epoch != lastCommittedEpoch+1 {
			errs[i] = errors.Errorf("commits should be consequent: %w", err)
			break
		}

		result, err := s.cfg.EvmClient.CommitValsetHeader(ctx, settlement, header, extraData, proof)
		if err != nil {
			errs[i] = errors.Errorf("failed to commit valset header to settlement %s: %w", settlement.Address.Hex(), err)
			break
		}

		slog.InfoContext(ctx, "Validator set header committed",
			"epoch", header.Epoch,
			"settlement", settlement,
			"txHash", result.TxHash,
		)
	}

	return errors.Join(errs...)
}
