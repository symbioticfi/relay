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
}

type p2pClient interface {
	BroadcastSignatureAggregatedMessage(ctx context.Context, proof symbiotic.AggregationProof) error
}

type metrics interface {
	ObserveOnlyAggregateDuration(d time.Duration)
	ObserveAppAggregateDuration(d time.Duration)
}

type aggregator interface {
	Aggregate(valset symbiotic.ValidatorSet, keyTag symbiotic.KeyTag, messageHash []byte, signatures []symbiotic.Signature) (symbiotic.AggregationProof, error)
}

type keyProvider interface {
	GetPrivateKey(keyTag symbiotic.KeyTag) (crypto.PrivateKey, error)
	GetOnchainKeyFromCache(keyTag symbiotic.KeyTag) (symbiotic.CompactPublicKey, error)
}

type aggregatorPolicy = aggregationPolicyTypes.AggregationPolicy

type Config struct {
	Repo              repository       `validate:"required"`
	P2PClient         p2pClient        `validate:"required"`
	Aggregator        aggregator       `validate:"required"`
	Metrics           metrics          `validate:"required"`
	AggregationPolicy aggregatorPolicy `validate:"required"`
	KeyProvider       keyProvider      `validate:"required"`
	ForceAggregator   bool
}

func (c Config) Validate() error {
	if err := validate.New().Struct(c); err != nil {
		return errors.Errorf("failed to validate config: %w", err)
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
	ctx = log.WithAttrs(ctx,
		slog.Uint64("epoch", uint64(msg.Epoch)),
		slog.String("requestId", msg.RequestID().Hex()),
	)
	slog.DebugContext(ctx, "Received signature processed message", "message", msg)

	tracing.AddEvent(span, "checking_existing_proof")
	_, err := s.cfg.Repo.GetAggregationProof(ctx, msg.RequestID())
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to get aggregation proof: %w", err)
	}
	if err == nil {
		tracing.AddEvent(span, "proof_already_exists")
		slog.DebugContext(ctx, "Skipped aggregation, proof already exists", "request", msg)
		return nil
	}

	tracing.AddEvent(span, "loading_signature_map")
	signatureMap, err := s.cfg.Repo.GetSignatureMap(ctx, msg.RequestID())
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to get valset signature map: %w", err)
	}

	if signatureMap.RequestID != msg.RequestID() || signatureMap.Epoch != msg.Epoch {
		err := errors.Errorf("signature map context mismatch: map %s/%d vs msg %s/%d",
			signatureMap.RequestID.Hex(), signatureMap.Epoch,
			msg.RequestID().Hex(), msg.Epoch,
		)
		tracing.RecordError(span, err)
		return err
	}

	tracing.AddEvent(span, "loading_validator_set")
	validatorSet, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, msg.Epoch)
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to get validator set: %w", err)
	}

	tracing.SetAttributes(span, tracing.AttrValidatorCount.Int(len(validatorSet.Validators)))

	tracing.AddEvent(span, "checking_aggregator_role")
	if s.cfg.ForceAggregator {
		slog.DebugContext(ctx, "Force aggregator mode enabled")
	} else {
		onchainKey, err := s.cfg.KeyProvider.GetOnchainKeyFromCache(validatorSet.RequiredKeyTag)
		if err != nil {
			if errors.Is(err, entity.ErrKeyNotFound) {
				tracing.AddEvent(span, "skipped_not_aggregator")
				slog.DebugContext(ctx, "Skipped aggregation, no onchain key for required key tag", "keyTag", validatorSet.RequiredKeyTag)
				return nil
			}
			tracing.RecordError(span, err)
			return errors.Errorf("failed to get private key for required key tag %s: %w", validatorSet.RequiredKeyTag, err)
		}

		if !validatorSet.IsAggregator(onchainKey) {
			tracing.AddEvent(span, "skipped_not_aggregator")
			slog.DebugContext(ctx, "Skipped aggregation, not an aggregator for this validator set",
				"key", onchainKey,
				"epoch", msg.Epoch,
				"aggIndices", validatorSet.AggregatorIndices,
			)
			return nil
		}
	}

	slog.DebugContext(ctx, "Confirmed as aggregator for this validator set")

	totalActiveVotingPower := validatorSet.GetTotalActiveVotingPower()

	tracing.AddEvent(span, "checking_quorum")
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
		return nil
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

	tracing.AddEvent(span, "loading_signatures")
	sigs, err := s.cfg.Repo.GetAllSignatures(ctx, msg.RequestID())
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to get signature aggregated message: %w", err)
	}
	tracing.SetAttributes(span, tracing.AttrSignatureCount.Int(len(sigs)))
	slog.DebugContext(ctx, "Loaded signatures for aggregation", "count", len(sigs))

	tracing.AddEvent(span, "loading_network_config")
	networkConfig, err := s.cfg.Repo.GetConfigByEpoch(ctx, msg.Epoch)
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to get network config: %w", err)
	}

	slog.DebugContext(ctx, "Loaded network config", "networkConfig", networkConfig)

	tracing.AddEvent(span, "aggregating_proof")
	onlyAggregateStart := time.Now()
	proofData, err := s.cfg.Aggregator.Aggregate(
		validatorSet,
		msg.KeyTag,
		msg.MessageHash,
		sigs,
	)
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to prove: %w", err)
	}
	s.cfg.Metrics.ObserveOnlyAggregateDuration(time.Since(onlyAggregateStart))

	tracing.AddEvent(span, "proof_created")
	tracing.SetAttributes(span, tracing.AttrProofSize.Int(len(proofData.Proof)))
	slog.InfoContext(ctx, "Aggregation proof created",
		"duration", time.Since(appAggregationStart).String(),
	)

	tracing.AddEvent(span, "broadcasting_proof")
	err = s.cfg.P2PClient.BroadcastSignatureAggregatedMessage(ctx, proofData)
	if err != nil {
		tracing.RecordError(span, err)
		return errors.Errorf("failed to broadcast signature aggregated message: %w", err)
	}
	s.cfg.Metrics.ObserveAppAggregateDuration(time.Since(appAggregationStart))

	tracing.AddEvent(span, "aggregation_completed")
	slog.InfoContext(ctx, "Aggregation completed, proof broadcast via p2p",
		"totalDuration", time.Since(appAggregationStart).String())

	return nil
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

	tracing.AddEvent(span, "loading_signatures")
	signatures, err := s.cfg.Repo.GetAllSignatures(ctx, requestID)
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.AggregationStatus{}, errors.Errorf("failed to get all signatures: %w", err)
	}

	tracing.SetAttributes(span, tracing.AttrSignatureCount.Int(len(signatures)))

	tracing.AddEvent(span, "loading_validator_set")
	validatorSet, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, signatureRequest.RequiredEpoch)
	if err != nil {
		tracing.RecordError(span, err)
		return symbiotic.AggregationStatus{}, errors.Errorf("failed to get validator set: %w", err)
	}

	tracing.AddEvent(span, "finding_validators")
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
