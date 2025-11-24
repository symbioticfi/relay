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
		// get latest known valset
		valsetHeader, err := s.cfg.Repo.GetLatestValidatorSetHeader(ctx)
		if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
			slog.ErrorContext(ctx, "Failed to get latest signed epoch", "error", err)
			continue
		}
		if errors.Is(err, entity.ErrEntityNotFound) {
			continue
		}

		valset, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, valsetHeader.Epoch)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get validator set by epoch", "error", err, "epoch", valsetHeader.Epoch)
			continue
		}

		// if latest valset is already committed, nothing to do
		if valset.Status == symbiotic.HeaderCommitted {
			slog.DebugContext(ctx, "Latest validator set already committed. skipping commit loop", "epoch", valset.Epoch)
			continue
		}

		nwCfg, err := s.cfg.Repo.GetConfigByEpoch(ctx, valsetHeader.Epoch)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get network config by epoch", "error", err, "epoch", valsetHeader.Epoch)
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
		lastCommittedEpoch := s.detectLastCommittedEpoch(ctx)

		if lastCommittedEpoch >= valset.Epoch {
			slog.DebugContext(ctx, "No pending proofs to commit, all epochs committed", "lastCommittedEpoch", lastCommittedEpoch, "knownValsetEpoch", valset.Epoch)
			continue
		}

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
			err = s.processPendingProof(ctx, proofKey)
			if err != nil {
				slog.ErrorContext(ctx, "Error processing pending proof",
					slog.String("requestId", proofKey.RequestID.Hex()),
					slog.Uint64("epoch", uint64(proofKey.Epoch)),
					slog.String("error", err.Error()),
				)
				break
			}
		}
	}
}

func (s *Service) processPendingProof(ctx context.Context, proofKey symbiotic.ProofCommitKey) error {
	ctx = log.WithAttrs(ctx,
		slog.String("requestId", proofKey.RequestID.Hex()),
		slog.Uint64("epoch", uint64(proofKey.Epoch)),
	)
	slog.DebugContext(ctx, "Found pending proof commit")

	// get proof
	proof, err := s.cfg.Repo.GetAggregationProof(ctx, proofKey.RequestID)
	if err != nil {
		return errors.Errorf("failed to get aggregation proof for request ID %s: %w", proofKey.RequestID.Hex(), err)
	}

	targetValset, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, proofKey.Epoch)
	if err != nil {
		return errors.Errorf("failed to get validator set by epoch %d: %w", proofKey.Epoch, err)
	}

	s.cfg.Metrics.ObserveAggregationProofSize(len(proof.Proof), len(targetValset.Validators))

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
		return errors.Errorf("failed to get validator set header: %w", err)
	}

	slog.DebugContext(ctx, "Committing proof to settlements", "header", header, "extraData", extraData)

	err = s.commitValsetToAllSettlements(ctx, config, header, extraData, proof.Proof)
	if err != nil {
		return errors.Errorf("failed to commit valset to all settlements: %w", err)
	}

	return nil
}

func (s *Service) detectLastCommittedEpoch(ctx context.Context) symbiotic.Epoch {
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
func (s *Service) commitValsetToAllSettlements(ctx context.Context, config symbiotic.NetworkConfig, header symbiotic.ValidatorSetHeader, extraData []symbiotic.ExtraData, proof []byte) error {
	errs := make([]error, len(config.Settlements))
	for i, settlement := range config.Settlements {
		slog.DebugContext(ctx, "Attempting to commit valset header to settlement", "settlement", settlement)

		// todo replace it with tx check instead of call to contract
		// if commit tx was sent but still not finalized this check will
		// return false positive and trigger one more commitment tx
		committed, err := s.cfg.EvmClient.IsValsetHeaderCommittedAt(ctx, settlement, header.Epoch, symbiotic.WithEVMBlockNumber(symbiotic.BlockNumberLatest))
		if err != nil {
			errs[i] = errors.Errorf("failed to check if header is committed at epoch %d: %v/%s: %w", header.Epoch, settlement.ChainId, settlement.Address.Hex(), err)
			continue
		}

		if committed {
			slog.DebugContext(ctx, "Valset header already committed at settlement", "settlement", settlement, "epoch", header.Epoch)
			continue
		}

		lastCommittedEpoch, err := s.cfg.EvmClient.GetLastCommittedHeaderEpoch(ctx, settlement, symbiotic.WithEVMBlockNumber(symbiotic.BlockNumberLatest))
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
