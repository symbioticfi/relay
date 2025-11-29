package valset_listener

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-errors/errors"
	"go.opentelemetry.io/otel/attribute"

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/tracing"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

const (
	minCommitterPollIntervalSeconds = uint64(5)
	commitCheckBatchSize            = 5
)

func (s *Service) StartCommitterLoop(ctx context.Context) error {
	ctx = log.WithComponent(ctx, "valset_committer_loop")
	// get the latest epoch and try to find schedule of committers and start committing
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

		tickInterval, err := s.tryToCommitPendingProofs(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "Error in valset committer loop", "error", err)
		}

		if tickInterval > 0 {
			ticker.Reset(time.Duration(tickInterval) * time.Second)
		}
	}
}

func (s *Service) tryToCommitPendingProofs(ctx context.Context) (uint64, error) {
	ctx, span := tracing.StartSpan(ctx, "valset_listener.TryToCommitPendingProofs")
	defer span.End()

	tracing.AddEvent(span, "loading_latest_validator_set")
	valsetHeader, err := s.cfg.Repo.GetLatestValidatorSetHeader(ctx)
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		tracing.RecordError(span, err)
		return 0, errors.Errorf("failed to get latest signed epoch: %w", err)
	}
	if errors.Is(err, entity.ErrEntityNotFound) {
		tracing.AddEvent(span, "no_validator_sets_found")
		return 0, nil
	}

	tracing.SetAttributes(span, tracing.AttrEpoch.Int64(int64(valsetHeader.Epoch)))

	valset, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, valsetHeader.Epoch)
	if err != nil {
		tracing.RecordError(span, err)
		return 0, errors.Errorf("failed to get validator set by epoch %d: %w", valsetHeader.Epoch, err)
	}

	if valset.Status == symbiotic.HeaderCommitted {
		tracing.AddEvent(span, "already_committed")
		slog.DebugContext(ctx, "Latest validator set already committed. skipping commit loop", "epoch", valset.Epoch)
		return 0, nil
	}

	tracing.AddEvent(span, "loading_network_config")
	nwCfg, err := s.cfg.Repo.GetConfigByEpoch(ctx, valsetHeader.Epoch)
	if err != nil {
		tracing.RecordError(span, err)
		return 0, errors.Errorf("failed to get network config by epoch %d: %w", valsetHeader.Epoch, err)
	}

	tickInterval := nwCfg.CommitterSlotDuration / 2
	if tickInterval < minCommitterPollIntervalSeconds {
		tickInterval = minCommitterPollIntervalSeconds
	}

	tracing.SetAttributes(span,
		tracing.AttrValidatorCount.Int(len(valset.Validators)),
		attribute.Int64("tick_interval", int64(tickInterval)),
	)

	tracing.AddEvent(span, "checking_committer_role")
	if s.cfg.ForceCommitter {
		tracing.SetAttributes(span, attribute.Bool("force_committer", true))
		slog.DebugContext(ctx, "Force committer mode enabled", "epoch", valsetHeader.Epoch)
	} else {
		onchainKey, err := s.cfg.KeyProvider.GetOnchainKeyFromCache(valset.RequiredKeyTag)
		if err != nil {
			if errors.Is(err, entity.ErrKeyNotFound) {
				tracing.AddEvent(span, "skipped_no_key")
				slog.DebugContext(ctx, "Skipped proof commitment, no onchain key for required key tag", "keyTag", valset.RequiredKeyTag, "epoch", valsetHeader.Epoch)
				return tickInterval, nil
			}
			tracing.RecordError(span, err)
			return tickInterval, errors.Errorf("failed to get onchain key for required key tag %s: %w", valset.RequiredKeyTag, err)
		}

		now := symbiotic.Timestamp(uint64(time.Now().Unix()))

		if !valset.IsActiveCommitter(ctx, nwCfg.CommitterSlotDuration, now, minCommitterPollIntervalSeconds, onchainKey) {
			tracing.AddEvent(span, "skipped_not_active_committer")
			slog.DebugContext(ctx, "Skipped proof commitment, not a committer for this validator set",
				"key", onchainKey,
				"epoch", valset.Epoch,
				"committerSlotDuration", nwCfg.CommitterSlotDuration,
				"now", now,
				"minPollInterval", minCommitterPollIntervalSeconds,
				"committerIndices", valset.CommitterIndices,
			)
			return tickInterval, nil
		}

		tracing.AddEvent(span, "confirmed_active_committer")
	}

	tracing.AddEvent(span, "detecting_last_committed_epoch_from_db")
	lastCommittedEpoch := s.detectLastCommittedEpochFromDB(ctx)

	tracing.SetAttributes(span, attribute.Int64("last_committed_epoch", int64(lastCommittedEpoch)))

	if lastCommittedEpoch >= valset.Epoch {
		tracing.AddEvent(span, "all_epochs_committed")
		slog.DebugContext(ctx, "No pending proofs to commit, all epochs committed", "lastCommittedEpoch", lastCommittedEpoch, "knownValsetEpoch", valset.Epoch)
		return tickInterval, nil
	}

	slog.DebugContext(ctx, "Detected last committed epoch", "lastCommittedEpoch", lastCommittedEpoch, "knownValsetEpoch", valset.Epoch)

	uncommittedCount := valset.Epoch - lastCommittedEpoch
	tracing.SetAttributes(span, attribute.Int64("uncommitted_epochs", int64(uncommittedCount)))

	if uncommittedCount > symbiotic.Epoch(commitCheckBatchSize) {
		tracing.AddEvent(span, "detecting_last_committed_epoch_from_chain")
		newLastCommit := s.detectLastCommittedEpochFromChain(ctx, nwCfg)
		if newLastCommit > lastCommittedEpoch {
			tracing.SetAttributes(span,
				attribute.Int64("last_committed_epoch_updated", int64(newLastCommit)),
			)
			slog.InfoContext(ctx, "Number of uncommitted epochs exceeds batch size, updated last committed epoch based on settlement chain data",
				"oldLastCommittedEpoch", lastCommittedEpoch,
				"newLastCommittedEpoch", newLastCommit,
			)
			lastCommittedEpoch = newLastCommit
		}
	}

	tracing.AddEvent(span, "loading_pending_proofs")
	pendingProofs, err := s.cfg.Repo.GetPendingProofCommitsSinceEpoch(ctx, lastCommittedEpoch+1, commitCheckBatchSize)
	if err != nil {
		tracing.RecordError(span, err)
		return tickInterval, errors.Errorf("failed to get pending proof commits since epoch %d: %w", valset.Epoch, err)
	}

	tracing.SetAttributes(span, attribute.Int("pending_proofs_count", len(pendingProofs)))

	if len(pendingProofs) == 0 {
		tracing.AddEvent(span, "no_pending_proofs")
		slog.DebugContext(ctx, "No pending proof commits found")
		return tickInterval, nil
	}

	tracing.AddEvent(span, "processing_pending_proofs")
	processedCount := 0
	for _, proofKey := range pendingProofs {
		err = s.processPendingProof(ctx, proofKey)
		if err != nil {
			tracing.RecordError(span, err)
			slog.ErrorContext(ctx, "Error processing pending proof",
				slog.String("requestId", proofKey.RequestID.Hex()),
				slog.Uint64("epoch", uint64(proofKey.Epoch)),
				slog.String("error", err.Error()),
			)
			break
		}
		processedCount++
	}

	tracing.SetAttributes(span, attribute.Int("processed_count", processedCount))

	return tickInterval, nil
}

func (s *Service) processPendingProof(ctx context.Context, proofKey symbiotic.ProofCommitKey) error {
	ctx, span := tracing.StartSpan(ctx, "valset_listener.ProcessPendingProof",
		tracing.AttrRequestID.String(proofKey.RequestID.Hex()),
		tracing.AttrEpoch.Int64(int64(proofKey.Epoch)),
	)
	defer span.End()

	ctx = log.WithAttrs(ctx,
		slog.String("requestId", proofKey.RequestID.Hex()),
		slog.Uint64("epoch", uint64(proofKey.Epoch)),
	)
	slog.DebugContext(ctx, "Found pending proof commit")

	tracing.AddEvent(span, "loading_proof")
	proof, err := s.cfg.Repo.GetAggregationProof(ctx, proofKey.RequestID)
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to get aggregation proof for request ID %s: %w", proofKey.RequestID.Hex(), err)
	}

	tracing.SetAttributes(span, tracing.AttrProofSize.Int(len(proof.Proof)))

	tracing.AddEvent(span, "loading_validator_set")
	targetValset, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, proofKey.Epoch)
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to get validator set by epoch %d: %w", proofKey.Epoch, err)
	}

	tracing.SetAttributes(span, tracing.AttrValidatorCount.Int(len(targetValset.Validators)))
	s.cfg.Metrics.ObserveAggregationProofSize(len(proof.Proof), len(targetValset.Validators))

	tracing.AddEvent(span, "loading_config")
	config, err := s.cfg.EvmClient.GetConfig(ctx, targetValset.CaptureTimestamp, proofKey.Epoch)
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to get config for epoch %d: %w", proofKey.Epoch, err)
	}

	tracing.AddEvent(span, "generating_extra_data")
	extraData, err := s.cfg.Aggregator.GenerateExtraData(targetValset, config.RequiredKeyTags)
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to generate extra data for validator set: %w", err)
	}

	header, err := targetValset.GetHeader()
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to get validator set header: %w", err)
	}

	pubkey, err := s.cfg.KeyProvider.GetOnchainKeyFromCache(header.RequiredKeyTag)
	if err != nil {
		return errors.Errorf("failed to get onchain key from cache: %w", err)
	}

	validator, found := targetValset.FindValidatorByKey(symbiotic.ValsetHeaderKeyTag, pubkey)
	if !found {
		return errors.Errorf("local validator not found")
	}

	ctx = log.WithAttrs(ctx,
		slog.String("validatorAddress", validator.Operator.Hex()),
	)
	tracing.SetAttributes(span, tracing.AttrValidatorAddress.String(validator.Operator.Hex()))

	slog.DebugContext(ctx, "Committing proof to settlements", "header", header, "extraData", extraData)

	tracing.AddEvent(span, "committing_to_settlements")
	ok, err := s.commitValsetToAllSettlements(ctx, config, header, extraData, proof.Proof)
	if !ok {
		_err := errors.Errorf("failed to commit valset to all settlements for epoch %d, error=%w", proofKey.Epoch, err)
		tracing.RecordError(span, _err)
		return _err
	} else if err != nil {
		// on partial failure just log error and continue, we will retry later
		slog.ErrorContext(ctx, "Failed to commit valset to some settlements", "error", err)
	}

	return nil
}

func (s *Service) detectLastCommittedEpochFromDB(ctx context.Context) symbiotic.Epoch {
	uncommitted, err := s.cfg.Repo.GetFirstUncommittedValidatorSetEpoch(ctx)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			slog.DebugContext(ctx, "No uncommitted validator sets found, assuming none committed yet")
			return symbiotic.Epoch(0)
		}
		slog.ErrorContext(ctx, "Failed to get first uncommitted validator set epoch", "error", err)
		return symbiotic.Epoch(0)
	}
	return uncommitted - 1
}

func (s *Service) detectLastCommittedEpochFromChain(ctx context.Context, config symbiotic.NetworkConfig) symbiotic.Epoch {
	minVal := symbiotic.Epoch(0)
	for _, settlement := range config.Settlements {
		lastCommittedEpoch, err := s.cfg.EvmClient.GetLastCommittedHeaderEpoch(ctx, settlement, symbiotic.WithEVMBlockNumber(symbiotic.BlockNumberLatest))
		if err != nil {
			slog.WarnContext(ctx, "Failed to get last committed epoch for settlement, skipping", "settlement", settlement, "error", err)
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

// commitValsetToAllSettlements commits the validator set header to all configured settlement chains.
//
// Performance Optimization: This method uses entity.BlockNumberLatest instead of finalized blocks
// when checking commitment status. This optimization reduces latency by ~12-15 seconds
// on Ethereum (approximately 2 finalization epochs), allowing faster detection of already-committed
// headers and reducing unnecessary duplicate commitment transactions.
//
// Safety: The pending proof cleanup happens separately in the status tracker, which uses finalized
// blocks to verify commitments before removing pending proofs. This ensures data consistency:
//   - This method: uses latest blocks for fast pre-flight checks (avoid duplicate tx submissions)
//   - Status tracker: uses finalized blocks for authoritative verification (safe pending proof removal)
//
// Trade-off: Using latest blocks for pre-flight checks introduces a small reorg risk, but this is
// acceptable because:
//  1. False positives (thinking a header is committed when it's not due to reorg) may trigger a
//     duplicate transaction, but the contract will reject it
//  2. False negatives (missing a commitment due to reorg) will be corrected in the next iteration
//  3. The performance benefit of reduced latency outweighs the minimal reorg risk
//  4. Final cleanup only happens after finalized block confirmation in the status tracker
//
// Returns a bool to indicate if at least once settlement commit worked and error if any commitment fails
func (s *Service) commitValsetToAllSettlements(ctx context.Context, config symbiotic.NetworkConfig, header symbiotic.ValidatorSetHeader, extraData []symbiotic.ExtraData, proof []byte) (bool, error) {
	errs := []error{}
	for _, settlement := range config.Settlements {
		slog.DebugContext(ctx, "Attempting to commit valset header to settlement", "settlement", settlement)

		// todo replace it with tx check instead of call to contract
		// if commit tx was sent but still not finalized this check will
		// return false positive and trigger one more commitment tx
		committed, err := s.cfg.EvmClient.IsValsetHeaderCommittedAt(ctx, settlement, header.Epoch, symbiotic.WithEVMBlockNumber(symbiotic.BlockNumberLatest))
		if err != nil {
			errs = append(errs, errors.Errorf("failed to check if header is committed at epoch %d: %v/%s: %w", header.Epoch, settlement.ChainId, settlement.Address.Hex(), err))
			continue
		}

		if committed {
			slog.DebugContext(ctx, "Valset header already committed at settlement", "settlement", settlement, "epoch", header.Epoch)
			continue
		}

		lastCommittedEpoch, err := s.cfg.EvmClient.GetLastCommittedHeaderEpoch(ctx, settlement, symbiotic.WithEVMBlockNumber(symbiotic.BlockNumberLatest))
		if err != nil {
			errs = append(errs, errors.Errorf("failed to get last committed header epoch: %v/%s: %w", settlement.ChainId, settlement.Address.Hex(), err))
			continue
		}

		if header.Epoch != lastCommittedEpoch+1 {
			errs = append(errs, errors.Errorf("commits should be consequent: %v/%s", settlement.ChainId, settlement.Address.Hex()))
			continue
		}

		result, err := s.cfg.EvmClient.CommitValsetHeader(ctx, settlement, header, extraData, proof)
		if err != nil {
			errs = append(errs, errors.Errorf("failed to commit valset header to settlement %v/%s: %w", settlement.ChainId, settlement.Address.Hex(), err))
			continue
		}

		slog.InfoContext(ctx, "Validator set header committed",
			"settlement", settlement,
			"txHash", result.TxHash,
		)
	}

	return len(config.Settlements) == 0 || len(errs) != len(config.Settlements), errors.Join(errs...)
}
