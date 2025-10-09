package valset_listener

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/symbiotic/entity"
)

const (
	minCommitterPollIntervalSeconds = uint64(5)
)

func (s *Service) HandleProofAggregated(ctx context.Context, msg entity.AggregationProof) error {
	ctx = log.WithComponent(ctx, "generator")

	slog.DebugContext(ctx, "Handling proof aggregated message", "msg", msg)

	// we always try to save the proof, even if we aren't a committer
	if err := s.cfg.Repo.SaveProof(ctx, msg); err != nil && !errors.Is(err, entity.ErrEntityAlreadyExist) {
		return err
	}

	// the epoch for the valset is always signature request epoch + 1
	valsetEpoch := msg.Epoch + 1

	var (
		valset entity.ValidatorSet
		err    error
	)
	retryAttempted := false
	for {
		valset, err = s.cfg.Repo.GetValidatorSetByEpoch(ctx, valsetEpoch)
		if err != nil {
			if !errors.Is(err, entity.ErrEntityNotFound) {
				return errors.Errorf("failed to get validator set by epoch %d: %w", valsetEpoch, err)
			}
			// if not found, try to load missing epochs and retry once
			// this can happen if we receive proof before valset is processed from signature requests
			if !retryAttempted {
				if err = s.tryLoadMissingEpochs(ctx); err != nil {
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

	if valset.Status == entity.HeaderCommitted {
		slog.DebugContext(ctx, "Valset is already committed", "epoch", valsetEpoch)
		return nil
	}

	// check if proof is a valset proof, only then commit
	valsetMeta, err := s.cfg.Repo.GetValidatorSetMetadata(ctx, valsetEpoch)
	if err != nil {
		return errors.Errorf("failed to get validator set metadata: %w", err)
	}
	if valsetMeta.RequestID != msg.RequestID() {
		slog.DebugContext(ctx, "Aggregation proof is not for valset, skipping proof commitment", "epoch", valsetEpoch, "requestId", msg.RequestID().Hex(), "valsetRequestId", valsetMeta.RequestID.Hex())
		return nil
	}

	// we store pending commit request for all nodes and not just current commiters because
	// if committers of this epoch fail then commiters for next epoch should still try to commit old proofs
	if err := s.cfg.Repo.SaveProofCommitPending(ctx, valsetEpoch, msg.RequestID()); err != nil {
		if !errors.Is(err, entity.ErrEntityAlreadyExist) {
			return errors.Errorf("failed to mark proof commit as pending: %w", err)
		}
		slog.DebugContext(ctx, "Proof commit is already pending, skipping", "epoch", valsetEpoch)
		return nil
	}

	if s.cfg.Metrics != nil {
		s.cfg.Metrics.ObserveAggregationProofSize(len(msg.Proof), len(valset.Validators))
	}

	slog.DebugContext(ctx, "Marked proof commit as pending", "epoch", valsetEpoch, "request_id", msg.RequestID().Hex())
	return nil
}

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

		privKey, err := s.cfg.KeyProvider.GetPrivateKey(valset.RequiredKeyTag)
		if err != nil {
			if errors.Is(err, entity.ErrKeyNotFound) {
				slog.DebugContext(ctx, "No key for required key tag, skipping proof commitment", "keyTag", valset.RequiredKeyTag)
				continue
			}
			return errors.Errorf("failed to get private key for required key tag %s: %w", valset.RequiredKeyTag, err)
		}

		now := entity.Timestamp(uint64(time.Now().Unix()))
		if !valset.IsActiveCommitter(ctx, nwCfg.CommitterSlotDuration, now, minCommitterPollIntervalSeconds, privKey.PublicKey().OnChain()) {
			slog.DebugContext(ctx, "Not a committer for this valset, skipping proof commitment",
				"key", privKey.PublicKey().OnChain(),
				"epoch", valset.Epoch,
				"committerSlotDuration", nwCfg.CommitterSlotDuration,
				"now", now,
				"minPollInterval", minCommitterPollIntervalSeconds,
				"committerIndices", valset.CommitterIndices,
			)
			continue
		}

		// get lat committed epoch
		// TODO(oxsteins): if this is too slow, might have to update status-tracker to catchup quickly and store last committed epoch in db
		// currently it is polling one by one and asynchronously so might not catchup early enough for committer to work
		lastCommittedEpoch := s.detectLastCommittedEpoch(ctx, nwCfg)

		slog.DebugContext(ctx, "Detected last committed epoch", "epoch", lastCommittedEpoch, "knownValsetEpoch", valset.Epoch)

		pendingProofs, err := s.cfg.Repo.GetPendingProofCommitsSinceEpoch(ctx, lastCommittedEpoch+1, 5)
		if err != nil {
			return errors.Errorf("failed to get pending proof commits since epoch %d: %w", valset.Epoch, err)
		}

		if len(pendingProofs) == 0 {
			slog.DebugContext(ctx, "No pending proof commits found")
			continue
		}

		for _, proofKey := range pendingProofs {
			slog.DebugContext(ctx, "Found pending proof commit", "epoch", proofKey.Epoch, "requestId", proofKey.RequestID.Hex())

			// get proof
			proof, err := s.cfg.Repo.GetAggregationProof(ctx, proofKey.RequestID)
			if err != nil {
				if errors.Is(err, entity.ErrEntityNotFound) {
					// should not happen, but if it does just remove pending state
					slog.WarnContext(ctx, "no aggregation proof found for pending proof commit, removing pending state", "epoch", proofKey.Epoch, "requestId", proofKey.RequestID.Hex())
					if err := s.cfg.Repo.RemoveProofCommitPending(ctx, proofKey.Epoch, proofKey.RequestID); err != nil {
						slog.ErrorContext(ctx, "failed to remove proof commit pending state", "error", err, "epoch", proofKey.Epoch, "requestId", proofKey.RequestID.Hex())
					}
					break
				}
				slog.ErrorContext(ctx, "failed to get aggregation proof", "error", err, "epoch", proofKey.Epoch, "requestId", proofKey.RequestID.Hex())
				break
			}

			targetValset, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, proofKey.Epoch)
			if err != nil {
				slog.ErrorContext(ctx, "failed to get validator set by epoch", "error", err, "epoch", proofKey.Epoch, "requestId", proofKey.RequestID.Hex())
				break
			}

			config, err := s.cfg.EvmClient.GetConfig(ctx, targetValset.CaptureTimestamp)
			if err != nil {
				return errors.Errorf("failed to get config for epoch %d: %w", proofKey.Epoch, err)
			}

			extraData, err := s.cfg.Aggregator.GenerateExtraData(targetValset, config.RequiredKeyTags)
			if err != nil {
				return errors.Errorf("failed to generate extra data for validator set: %w", err)
			}

			header, err := targetValset.GetHeader()
			if err != nil {
				slog.ErrorContext(ctx, "failed to get validator set header", "error", err, "epoch", proofKey.Epoch, "requestId", proofKey.RequestID.Hex())
				break
			}

			slog.DebugContext(ctx, "On commit proof", "header", header, "extraData", extraData)

			err = s.commitValsetToAllSettlements(ctx, config, header, extraData, proof.Proof)
			if err != nil {
				slog.ErrorContext(ctx, "failed to commit valset to all settlements", "error", err, "epoch", proofKey.Epoch, "requestId", proofKey.RequestID.Hex())
				break
			}

			if err := s.cfg.Repo.RemoveProofCommitPending(ctx, proofKey.Epoch, proofKey.RequestID); err != nil {
				slog.ErrorContext(ctx, "failed to remove proof commit pending state", "error", err, "epoch", proofKey.Epoch, "requestId", proofKey.RequestID.Hex())
				break
			}
		}
	}
}

func (s *Service) detectLastCommittedEpoch(ctx context.Context, config entity.NetworkConfig) entity.Epoch {
	minVal := entity.Epoch(0)
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

func (s *Service) commitValsetToAllSettlements(ctx context.Context, config entity.NetworkConfig, header entity.ValidatorSetHeader, extraData []entity.ExtraData, proof []byte) error {
	errs := make([]error, len(config.Settlements))
	for i, settlement := range config.Settlements {
		slog.DebugContext(ctx, "Trying to commit valset header to settlement", "settlement", settlement)

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
			"epoch", header.Epoch,
			"settlement", settlement,
			"txHash", result.TxHash,
		)
	}

	return errors.Join(errs...)
}
