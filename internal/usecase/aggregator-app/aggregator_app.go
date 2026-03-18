package aggregator_app

import (
	"context"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/internal/entity"
	aggregationPolicyTypes "github.com/symbioticfi/relay/internal/usecase/aggregation-policy/types"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/tracing"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

//go:generate mockgen -source=aggregator_app.go -destination=mocks/aggregator_app.go -package=mocks
type repository interface {
	GetValidatorSetByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSet, error)
	GetAggregationProof(ctx context.Context, requestID common.Hash) (symbiotic.AggregationProof, error)
	GetSignatureRequest(_ context.Context, requestID common.Hash) (symbiotic.SignatureRequest, error)
	GetAllSignatures(ctx context.Context, requestID common.Hash) ([]symbiotic.Signature, error)
	GetConfigByEpoch(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.NetworkConfig, error)
	GetSignatureMap(ctx context.Context, requestID common.Hash) (entity.SignatureMap, error)
	GetSignatureRequestsWithoutAggregationProof(ctx context.Context, epoch symbiotic.Epoch, limit int, lastHash common.Hash) ([]symbiotic.SignatureRequestWithID, error)
	GetLatestValidatorSetEpoch(ctx context.Context) (symbiotic.Epoch, error)
	RemoveAggregationProofPending(ctx context.Context, epoch symbiotic.Epoch, requestID common.Hash) error
}

type p2pClient interface {
	BroadcastSignatureAggregatedMessage(ctx context.Context, proof symbiotic.AggregationProof) error
}

type metrics interface {
	ObserveOnlyAggregateDuration(d time.Duration)
	ObserveAppAggregateDuration(d time.Duration)
}

type aggregator interface {
	Aggregate(ctx context.Context, valset symbiotic.ValidatorSet, signatures []symbiotic.Signature) (symbiotic.AggregationProof, error)
}

type keyProvider interface {
	GetPrivateKey(keyTag symbiotic.KeyTag) (crypto.PrivateKey, error)
	GetOnchainKeyFromCache(keyTag symbiotic.KeyTag) (symbiotic.CompactPublicKey, error)
}

type aggregatorPolicy = aggregationPolicyTypes.AggregationPolicy

type ProofCatchupConfig struct {
	Enabled             bool
	Interval            time.Duration `validate:"gte=0"`
	EpochsToCheck       int           `validate:"gte=0"`
	EpochsOffset        int           `validate:"gte=0"`
	MaxRequestsPerCycle int           `validate:"gte=0"`
	MaxProofsPerCycle   int           `validate:"gte=0"`
}

type Config struct {
	Repo              repository       `validate:"required"`
	P2PClient         p2pClient        `validate:"required"`
	Aggregator        aggregator       `validate:"required"`
	Metrics           metrics          `validate:"required"`
	AggregationPolicy aggregatorPolicy `validate:"required"`
	KeyProvider       keyProvider      `validate:"required"`
	ForceAggregator   bool
	ProofCatchup      ProofCatchupConfig
}

func (c Config) Validate() error {
	if err := validate.New().Struct(c); err != nil {
		return errors.Errorf("failed to validate config: %w", err)
	}

	if c.ProofCatchup.Enabled {
		if c.ProofCatchup.Interval <= 0 {
			return errors.New("proof catchup interval must be greater than zero when enabled")
		}
		if c.ProofCatchup.EpochsToCheck <= 0 {
			return errors.New("proof catchup epochs-to-check must be greater than zero when enabled")
		}
	}

	return nil
}

type AggregatorApp struct {
	cfg Config
}

func NewAggregatorApp(cfg Config) (*AggregatorApp, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	app := &AggregatorApp{
		cfg: cfg,
	}

	return app, nil
}

func (s *AggregatorApp) HandleSignatureProcessedMessage(ctx context.Context, msg symbiotic.Signature) error {
	ctx, span := tracing.StartConsumerSpan(ctx, "aggregator.HandleSignatureProcessed",
		tracing.AttrRequestID.String(msg.RequestID().Hex()),
		tracing.AttrEpoch.Int64(int64(msg.Epoch)),
		tracing.AttrKeyTag.String(msg.KeyTag.String()),
	)
	defer span.End()

	ctx = log.WithComponent(ctx, "aggregator")
	if !msg.KeyTag.Type().AggregationKey() {
		slog.DebugContext(ctx, "Skipped processing signature processed message, key tag is not for aggregation",
			"message", msg,
			"epoch", msg.Epoch,
			"requestId", msg.RequestID().Hex(),
			"keyTag", msg.KeyTag.String(),
		)
		return nil
	}
	slog.DebugContext(ctx, "Received signature processed message",
		"message", msg,
		"epoch", msg.Epoch,
		"requestId", msg.RequestID().Hex(),
	)

	_, err := s.TryAggregateProofForRequestID(ctx, msg.RequestID())
	return err
}

func (s *AggregatorApp) TryAggregateProofForRequestID(ctx context.Context, requestID common.Hash) (symbiotic.AggregationProof, error) {
	ctx, span := tracing.StartSpan(ctx, "aggregator.TryAggregateProofForRequestID",
		tracing.AttrRequestID.String(requestID.Hex()),
	)
	defer span.End()

	ctx = log.WithComponent(ctx, "aggregator")
	ctx = log.WithAttrs(ctx,
		slog.String("requestId", requestID.Hex()),
	)

	slog.DebugContext(ctx, "Started proof aggregation for request")

	_, err := s.cfg.Repo.GetAggregationProof(ctx, requestID)
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		tracing.RecordError(span, err)
		return symbiotic.AggregationProof{}, errors.Errorf("failed to get aggregation proof: %w", err)
	}
	if err == nil {
		tracing.AddEvent(span, "proof_already_exists")
		slog.DebugContext(ctx, "Skipped aggregation, proof already exists")
		return symbiotic.AggregationProof{}, nil
	}

	signatureMap, err := s.cfg.Repo.GetSignatureMap(ctx, requestID)
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.AggregationProof{}, errors.Errorf("failed to get valset signature map: %w", err)
	}

	ctx = log.WithAttrs(ctx, slog.Uint64("epoch", uint64(signatureMap.Epoch)))
	tracing.SetAttributes(span, tracing.AttrEpoch.Int64(int64(signatureMap.Epoch)))

	// Get validator set for quorum threshold checks
	validatorSet, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, signatureMap.Epoch)
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.AggregationProof{}, errors.Errorf("failed to get validator set: %w", err)
	}

	tracing.SetAttributes(span, tracing.AttrValidatorCount.Int(len(validatorSet.Validators)))

	if s.cfg.ForceAggregator {
		slog.DebugContext(ctx, "Force aggregator mode enabled")
	} else {
		onchainKey, err := s.cfg.KeyProvider.GetOnchainKeyFromCache(validatorSet.RequiredKeyTag)
		if err != nil {
			if errors.Is(err, entity.ErrKeyNotFound) {
				tracing.AddEvent(span, "skipped_not_key_not_found")
				slog.DebugContext(ctx, "Skipped aggregation, no onchain key for required key tag", "keyTag", validatorSet.RequiredKeyTag)
				return symbiotic.AggregationProof{}, nil
			}
			tracing.RecordError(span, err)
			return symbiotic.AggregationProof{}, errors.Errorf("failed to get private key for required key tag %s: %w", validatorSet.RequiredKeyTag, err)
		}

		if !validatorSet.IsAggregator(onchainKey) {
			tracing.AddEvent(span, "skipped_not_aggregator")
			slog.DebugContext(ctx, "Skipped aggregation, not an aggregator for this validator set",
				"key", onchainKey,
				"epoch", signatureMap.Epoch,
				"aggIndices", validatorSet.AggregatorIndices,
				"ourIndex", validatorSet.ValidatorIndex(onchainKey),
			)
			return symbiotic.AggregationProof{}, nil
		}
	}

	slog.DebugContext(ctx, "Confirmed as aggregator for this validator set")

	totalActiveVotingPower := validatorSet.GetTotalActiveVotingPower()

	if !s.cfg.AggregationPolicy.ShouldAggregate(signatureMap, validatorSet) {
		tracing.AddEvent(span, "quorum_not_reached")
		tracing.SetAttributes(span,
			tracing.AttrQuorumThreshold.Int(int(validatorSet.QuorumThreshold.Uint64())),
		)
		slog.DebugContext(ctx, "Skipped aggregation, quorum not reached",
			"currentVotingPower", signatureMap.CurrentVotingPower.String(),
			"quorumThreshold", validatorSet.QuorumThreshold.String(),
			"totalActiveVotingPower", totalActiveVotingPower.String(),
		)
		return symbiotic.AggregationProof{}, nil
	}

	tracing.AddEvent(span, "quorum_reached")
	tracing.SetAttributes(span,
		tracing.AttrQuorumThreshold.Int(int(validatorSet.QuorumThreshold.Uint64())),
	)
	slog.InfoContext(ctx, "Quorum reached, starting aggregation",
		"currentVotingPower", signatureMap.CurrentVotingPower.String(),
		"quorumThreshold", validatorSet.QuorumThreshold.String(),
		"totalActiveVotingPower", totalActiveVotingPower.String(),
	)

	appAggregationStart := time.Now()

	sigs, err := s.cfg.Repo.GetAllSignatures(ctx, requestID)
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.AggregationProof{}, errors.Errorf("failed to get signature aggregated message: %w", err)
	}
	tracing.SetAttributes(span, tracing.AttrSignatureCount.Int(len(sigs)))
	slog.DebugContext(ctx, "Loaded signatures for aggregation", "count", len(sigs))

	networkConfig, err := s.cfg.Repo.GetConfigByEpoch(ctx, signatureMap.Epoch)
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.AggregationProof{}, errors.Errorf("failed to get network config: %w", err)
	}

	slog.DebugContext(ctx, "Loaded network config", "networkConfig", networkConfig)

	onlyAggregateStart := time.Now()
	proofData, err := s.cfg.Aggregator.Aggregate(ctx, validatorSet, sigs)
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.AggregationProof{}, errors.Errorf("failed to prove: %w", err)
	}
	s.cfg.Metrics.ObserveOnlyAggregateDuration(time.Since(onlyAggregateStart))

	tracing.AddEvent(span, "proof_created")
	tracing.SetAttributes(span, tracing.AttrProofSize.Int(len(proofData.Proof)))
	slog.InfoContext(ctx, "Aggregation proof created",
		"duration", time.Since(appAggregationStart).String(),
	)

	err = s.cfg.P2PClient.BroadcastSignatureAggregatedMessage(ctx, proofData)
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.AggregationProof{}, errors.Errorf("failed to broadcast signature aggregated message: %w", err)
	}
	s.cfg.Metrics.ObserveAppAggregateDuration(time.Since(appAggregationStart))

	tracing.AddEvent(span, "aggregation_completed")
	slog.InfoContext(ctx, "Aggregation completed, proof broadcast via p2p",
		"totalDuration", time.Since(appAggregationStart).String())

	return proofData, nil
}

func (s *AggregatorApp) tryAggregateRequestsWithoutProof(ctx context.Context) error {
	latestEpoch, err := s.cfg.Repo.GetLatestValidatorSetEpoch(ctx)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			slog.DebugContext(ctx, "No validator sets synced yet, skipping aggregation catch-up")
			return nil
		}
		return errors.Errorf("failed to get latest epoch: %w", err)
	}

	catchupCfg := s.cfg.ProofCatchup

	var scanFrom symbiotic.Epoch
	if symbiotic.Epoch(catchupCfg.EpochsOffset) <= latestEpoch {
		scanFrom = latestEpoch - symbiotic.Epoch(catchupCfg.EpochsOffset)
	}

	startEpoch := symbiotic.Epoch(0)
	if scanFrom >= symbiotic.Epoch(catchupCfg.EpochsToCheck) {
		startEpoch = scanFrom - symbiotic.Epoch(catchupCfg.EpochsToCheck) + 1
	}

	slog.InfoContext(ctx, "Started aggregation catch-up for requests without proof",
		"latestEpoch", latestEpoch,
		"scanFrom", scanFrom,
		"startEpoch", startEpoch,
	)

	requestsChecked := 0
	proofsGenerated := 0

	for epoch := scanFrom; ; epoch-- {
		var lastHash common.Hash
		for {
			requests, err := s.cfg.Repo.GetSignatureRequestsWithoutAggregationProof(ctx, epoch, 10, lastHash)
			if err != nil {
				return errors.Errorf("failed to get signature requests without aggregation proof for epoch %d: %w", epoch, err)
			}

			if len(requests) == 0 {
				break
			}

			for _, req := range requests {
				if !req.KeyTag.Type().AggregationKey() {
					continue
				}

				if catchupCfg.MaxRequestsPerCycle > 0 && requestsChecked >= catchupCfg.MaxRequestsPerCycle {
					slog.InfoContext(ctx, "Aggregation catch-up reached max requests per cycle",
						"requestsChecked", requestsChecked,
						"proofsGenerated", proofsGenerated,
					)
					return nil
				}

				proof, err := s.TryAggregateProofForRequestID(ctx, req.RequestID)
				requestsChecked++
				if err != nil {
					if ctx.Err() != nil {
						return ctx.Err()
					}
					slog.ErrorContext(ctx, "Failed to aggregate proof for request, skipping",
						"requestId", req.RequestID.Hex(),
						"epoch", epoch,
						"error", err,
					)
					continue
				}

				if len(proof.Proof) > 0 {
					proofsGenerated++
					if catchupCfg.MaxProofsPerCycle > 0 && proofsGenerated >= catchupCfg.MaxProofsPerCycle {
						slog.InfoContext(ctx, "Aggregation catch-up reached max proofs per cycle",
							"requestsChecked", requestsChecked,
							"proofsGenerated", proofsGenerated,
						)
						return nil
					}
				}
			}

			lastHash = requests[len(requests)-1].RequestID
		}

		if epoch == startEpoch {
			break // Prevent underflow when decrementing unsigned epoch
		}
	}

	slog.InfoContext(ctx, "Aggregation catch-up completed",
		"requestsChecked", requestsChecked,
		"proofsGenerated", proofsGenerated,
	)

	return nil
}

func (s *AggregatorApp) StartCatchupLoop(ctx context.Context) error {
	ctx = log.WithComponent(ctx, "aggregator")

	if !s.cfg.ProofCatchup.Enabled {
		slog.InfoContext(ctx, "Proof catch-up loop disabled")
		return nil
	}

	slog.InfoContext(ctx, "Started proof catch-up loop",
		"interval", s.cfg.ProofCatchup.Interval,
		"epochsToCheck", s.cfg.ProofCatchup.EpochsToCheck,
		"epochsOffset", s.cfg.ProofCatchup.EpochsOffset,
		"maxRequestsPerCycle", s.cfg.ProofCatchup.MaxRequestsPerCycle,
		"maxProofsPerCycle", s.cfg.ProofCatchup.MaxProofsPerCycle,
	)

	timer := time.NewTimer(0) // Fire immediately on first tick
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			if err := s.tryAggregateRequestsWithoutProof(ctx); err != nil {
				slog.ErrorContext(ctx, "Proof catch-up cycle failed", "error", err)
			}
			timer.Reset(s.cfg.ProofCatchup.Interval)
		case <-ctx.Done():
			slog.InfoContext(ctx, "Proof catch-up loop stopped")
			return nil
		}
	}
}

func (s *AggregatorApp) GetAggregationStatus(ctx context.Context, requestID common.Hash) (symbiotic.AggregationStatus, error) {
	ctx, span := tracing.StartSpan(ctx, "aggregator.GetAggregationStatus",
		tracing.AttrRequestID.String(requestID.Hex()),
	)
	defer span.End()

	tracing.AddEvent(span, "loading_signature_request")
	signatureRequest, err := s.cfg.Repo.GetSignatureRequest(ctx, requestID)
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.AggregationStatus{}, errors.Errorf("failed to get signature request: %w", err)
	}

	tracing.SetAttributes(span,
		tracing.AttrEpoch.Int64(int64(signatureRequest.RequiredEpoch)),
		tracing.AttrKeyTag.String(signatureRequest.KeyTag.String()),
	)

	if !signatureRequest.KeyTag.Type().AggregationKey() {
		err := errors.Errorf("key tag %s is not an aggregation key", signatureRequest.KeyTag)
		tracing.RecordError(span, err)
		return symbiotic.AggregationStatus{}, err
	}

	signatures, err := s.cfg.Repo.GetAllSignatures(ctx, requestID)
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.AggregationStatus{}, errors.Errorf("failed to get all signatures: %w", err)
	}

	tracing.SetAttributes(span, tracing.AttrSignatureCount.Int(len(signatures)))

	validatorSet, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, signatureRequest.RequiredEpoch)
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.AggregationStatus{}, errors.Errorf("failed to get validator set: %w", err)
	}

	validators, err := validatorSet.FindValidatorsByKeys(signatureRequest.KeyTag, extractPublicKeys(signatures))
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.AggregationStatus{}, errors.Errorf("failed to find validators by keys: %w", err)
	}

	tracing.SetAttributes(span, tracing.AttrValidatorCount.Int(len(validators)))

	return symbiotic.AggregationStatus{
		VotingPower: validators.GetTotalActiveVotingPower(),
		Validators:  validators,
	}, nil
}

func extractPublicKeys(signatures []symbiotic.Signature) []symbiotic.CompactPublicKey {
	publicKeys := make([]symbiotic.CompactPublicKey, 0, len(signatures))
	for _, signature := range signatures {
		publicKeys = append(publicKeys, signature.PublicKey.OnChain())
	}
	return publicKeys
}
