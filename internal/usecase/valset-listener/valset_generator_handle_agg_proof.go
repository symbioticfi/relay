package valset_listener

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

const (
	minCommitterPollIntervalSeconds = uint64(5)
)

func (s *Service) StartCommitterLoop(ctx context.Context) error {
	// get the latest epoch and try find schedule of committers and start committing

	slog.InfoContext(ctx, "Starting valset committer loop")

	// force to tick immediately as soon as it starts
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "Valset committer loop stopped")
			return nil
		case <-ticker.C:
			slog.DebugContext(ctx, "Valset committer tick")
		}
		// get latest known valset
		valsetHeader, err := s.cfg.Repo.GetLatestValidatorSetHeader(ctx)
		if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
			slog.ErrorContext(ctx, "failed to get latest signed epoch", "error", err)
			continue
		}
		if errors.Is(err, entity.ErrEntityNotFound) {
			continue
		}

		valset, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, valsetHeader.Epoch)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get validator set by epoch", "error", err, "epoch", valsetHeader.Epoch)
			continue
		}

		nwCfg, err := s.cfg.Repo.GetConfigByEpoch(ctx, valsetHeader.Epoch)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get network config by epoch", "error", err, "epoch", valsetHeader.Epoch)
			continue
		}

		tickInterval := nwCfg.CommitterSlotDuration / 2
		if tickInterval < minCommitterPollIntervalSeconds {
			tickInterval = minCommitterPollIntervalSeconds
		}

		ticker.Reset(time.Duration(tickInterval) * time.Second)

		if s.cfg.ForceCommitter {
			slog.DebugContext(ctx, "Force committer mode enabled", "epoch", valsetHeader.Epoch)
		} else {
			onchainKey, err := s.cfg.KeyProvider.GetOnchainKeyFromCache(valset.RequiredKeyTag)
			if err != nil {
				if errors.Is(err, entity.ErrKeyNotFound) {
					slog.DebugContext(ctx, "Skipped proof commitment, no onchain key for required key tag", "keyTag", valset.RequiredKeyTag, "epoch", valsetHeader.Epoch)
					continue
				}
				return errors.Errorf("failed to get onchain key for required key tag %s: %w", valset.RequiredKeyTag, err)
			}

			now := symbiotic.Timestamp(uint64(time.Now().Unix()))

			if !valset.IsActiveCommitter(ctx, nwCfg.CommitterSlotDuration, now, minCommitterPollIntervalSeconds, onchainKey) {
				slog.DebugContext(ctx, "Skipped proof commitment, not a committer for this validator set",
					"key", onchainKey,
					"epoch", valset.Epoch,
					"committerSlotDuration", nwCfg.CommitterSlotDuration,
					"now", now,
					"minPollInterval", minCommitterPollIntervalSeconds,
					"committerIndices", valset.CommitterIndices,
				)
				continue
			}
		}

		// get lat committed epoch
		// TODO(oxsteins): if this is too slow, might have to update status-tracker to catchup quickly and store last committed epoch in db
		// currently it is polling one by one and asynchronously so might not catchup early enough for committer to work
		// (alrxy) i think it's better to do in listener cuz listener provides valsets and configs, it will be more consistent
		lastCommittedEpoch := s.detectLastCommittedEpoch(ctx, nwCfg)

		slog.DebugContext(ctx, "Detected last committed epoch", "lastCommittedEpoch", lastCommittedEpoch, "knownValsetEpoch", valset.Epoch)

		pendingProofs, err := s.cfg.Repo.GetPendingProofCommitsSinceEpoch(ctx, lastCommittedEpoch+1, 5)
		if err != nil {
			return errors.Errorf("failed to get pending proof commits since epoch %d: %w", valset.Epoch, err)
		}

		if len(pendingProofs) == 0 {
			slog.DebugContext(ctx, "No pending proof commits found")
			continue
		}

		for _, proofKey := range pendingProofs {
			//nolint:govet // shadow is ok here, we need separate ctx for each iteration
			ctx := log.WithAttrs(ctx,
				slog.String("requestId", proofKey.RequestID.Hex()),
				slog.Uint64("epoch", uint64(proofKey.Epoch)),
			)
			slog.DebugContext(ctx, "Found pending proof commit")

			// get proof
			proof, err := s.cfg.Repo.GetAggregationProof(ctx, proofKey.RequestID)
			if err != nil {
				if errors.Is(err, entity.ErrEntityNotFound) {
					slog.WarnContext(ctx, "no aggregation proof found for pending proof commit, ending current commit attempt")
					break
				}
				slog.ErrorContext(ctx, "failed to get aggregation proof", "error", err)
				break
			}

			targetValset, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, proofKey.Epoch)
			if err != nil {
				slog.ErrorContext(ctx, "failed to get validator set by epoch", "error", err)
				break
			}

			if s.cfg.Metrics != nil {
				s.cfg.Metrics.ObserveAggregationProofSize(len(proof.Proof), len(targetValset.Validators))
			}

			config, err := s.cfg.EvmClient.GetConfig(ctx, targetValset.CaptureTimestamp, proofKey.Epoch)
			if err != nil {
				return errors.Errorf("failed to get config for epoch %d: %w", proofKey.Epoch, err)
			}

			extraData, err := s.cfg.Aggregator.GenerateExtraData(targetValset, config.RequiredKeyTags)
			if err != nil {
				return errors.Errorf("failed to generate extra data for validator set: %w", err)
			}

			header, err := targetValset.GetHeader()
			if err != nil {
				slog.ErrorContext(ctx, "failed to get validator set header", "error", err)
				break
			}

			slog.DebugContext(ctx, "On commit proof", "header", header, "extraData", extraData)

			err = s.commitValsetToAllSettlements(ctx, config, header, extraData, proof.Proof)
			if err != nil {
				slog.ErrorContext(ctx, "failed to commit valset to all settlements", "error", err)
				break
			}

			if err := s.cfg.Repo.RemoveProofCommitPending(ctx, proofKey.Epoch, proofKey.RequestID); err != nil {
				slog.ErrorContext(ctx, "failed to remove proof commit pending state", "error", err)
				break
			}
		}
	}
}

func (s *Service) detectLastCommittedEpoch(ctx context.Context, config symbiotic.NetworkConfig) symbiotic.Epoch {
	minVal := symbiotic.Epoch(0)
	for _, settlement := range config.Settlements {
		lastCommittedEpoch, err := s.cfg.EvmClient.GetLastCommittedHeaderEpoch(ctx, settlement)
		if err != nil {
			// skip chain if networking issue, we will recheck again anyway and if the rpc/chain recovers we will detect issue later
			continue
		}
		if minVal == 0 {
			minVal = lastCommittedEpoch
		} else if lastCommittedEpoch < minVal {
			minVal = lastCommittedEpoch
		}
	}
	return minVal
}

func (s *Service) commitValsetToAllSettlements(ctx context.Context, config symbiotic.NetworkConfig, header symbiotic.ValidatorSetHeader, extraData []symbiotic.ExtraData, proof []byte) error {
	errs := make([]error, len(config.Settlements))
	for i, settlement := range config.Settlements {
		slog.DebugContext(ctx, "Attempting to commit valset header to settlement", "settlement", settlement)

		// todo replace it with tx check instead of call to contract
		// if commit tx was sent but still not finalized this check will
		// return false positive and trigger one more commitment tx
		committed, err := s.cfg.EvmClient.IsValsetHeaderCommittedAt(ctx, settlement, header.Epoch)
		if err != nil {
			errs[i] = errors.Errorf("failed to check if header is committed at epoch %d: %v/%s: %w", header.Epoch, settlement.ChainId, settlement.Address.Hex(), err)
			continue
		}

		if committed {
			continue
		}

		lastCommittedEpoch, err := s.cfg.EvmClient.GetLastCommittedHeaderEpoch(ctx, settlement)
		if err != nil {
			errs[i] = errors.Errorf("failed to get last committed header epoch: %v/%s: %w", settlement.ChainId, settlement.Address.Hex(), err)
			continue
		}

		if header.Epoch != lastCommittedEpoch+1 {
			errs[i] = errors.Errorf("commits should be consequent: %v/%s", settlement.ChainId, settlement.Address.Hex())
			continue
		}

		result, err := s.cfg.EvmClient.CommitValsetHeader(ctx, settlement, header, extraData, proof)
		if err != nil {
			errs[i] = errors.Errorf("failed to commit valset header to settlement %v/%s: %w", settlement.ChainId, settlement.Address.Hex(), err)
			continue
		}

		slog.InfoContext(ctx, "Validator set header committed",
			"settlement", settlement,
			"txHash", result.TxHash,
		)
	}

	return errors.Join(errs...)
}
