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
	Aggregate(valset symbiotic.ValidatorSet, signatures []symbiotic.Signature) (symbiotic.AggregationProof, error)
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
	ctx = log.WithComponent(ctx, "aggregator")
	slog.DebugContext(ctx, "Received signature processed message",
		"message", msg,
		"epoch", msg.Epoch,
		"requestId", msg.RequestID().Hex(),
	)

	return s.TryAggregateProofForRequestID(ctx, msg.RequestID())
}

func (s *AggregatorApp) TryAggregateProofForRequestID(ctx context.Context, requestID common.Hash) error {
	ctx = log.WithComponent(ctx, "aggregator")
	ctx = log.WithAttrs(ctx,
		slog.String("requestId", requestID.Hex()),
	)

	_, err := s.cfg.Repo.GetAggregationProof(ctx, requestID)
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		return errors.Errorf("failed to get aggregation proof: %w", err)
	}
	if err == nil {
		slog.DebugContext(ctx, "Skipped aggregation, proof already exists")
		return nil
	}

	signatureMap, err := s.cfg.Repo.GetSignatureMap(ctx, requestID)
	if err != nil {
		return errors.Errorf("failed to get valset signature map: %w", err)
	}

	ctx = log.WithAttrs(ctx, slog.Uint64("epoch", uint64(signatureMap.Epoch)))

	// Get validator set for quorum threshold checks
	validatorSet, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, signatureMap.Epoch)
	if err != nil {
		return errors.Errorf("failed to get validator set: %w", err)
	}

	if s.cfg.ForceAggregator {
		slog.DebugContext(ctx, "Force aggregator mode enabled")
	} else {
		onchainKey, err := s.cfg.KeyProvider.GetOnchainKeyFromCache(validatorSet.RequiredKeyTag)
		if err != nil {
			if errors.Is(err, entity.ErrKeyNotFound) {
				slog.DebugContext(ctx, "Skipped aggregation, no onchain key for required key tag", "keyTag", validatorSet.RequiredKeyTag)
				return nil
			}
			return errors.Errorf("failed to get private key for required key tag %s: %w", validatorSet.RequiredKeyTag, err)
		}

		if !validatorSet.IsAggregator(onchainKey) {
			slog.DebugContext(ctx, "Skipped aggregation, not an aggregator for this validator set",
				"key", onchainKey,
				"epoch", signatureMap.Epoch,
				"aggIndices", validatorSet.AggregatorIndices,
			)
			return nil
		}
	}

	slog.DebugContext(ctx, "Confirmed as aggregator for this validator set")

	totalActiveVotingPower := validatorSet.GetTotalActiveVotingPower()

	if !s.cfg.AggregationPolicy.ShouldAggregate(signatureMap, validatorSet) {
		slog.DebugContext(ctx, "Skipped aggregation, quorum not reached",
			"currentVotingPower", signatureMap.CurrentVotingPower.String(),
			"quorumThreshold", validatorSet.QuorumThreshold.String(),
			"totalActiveVotingPower", totalActiveVotingPower.String(),
		)
		return nil
	}

	slog.InfoContext(ctx, "Quorum reached, starting aggregation",
		"currentVotingPower", signatureMap.CurrentVotingPower.String(),
		"quorumThreshold", validatorSet.QuorumThreshold.String(),
		"totalActiveVotingPower", totalActiveVotingPower.String(),
	)

	appAggregationStart := time.Now()

	sigs, err := s.cfg.Repo.GetAllSignatures(ctx, requestID)
	slog.DebugContext(ctx, "Loaded signatures for aggregation", "count", len(sigs))
	if err != nil {
		return errors.Errorf("failed to get signature aggregated message: %w", err)
	}

	networkConfig, err := s.cfg.Repo.GetConfigByEpoch(ctx, signatureMap.Epoch)
	if err != nil {
		return errors.Errorf("failed to get network config: %w", err)
	}

	slog.DebugContext(ctx, "Loaded network config", "networkConfig", networkConfig)

	onlyAggregateStart := time.Now()
	proofData, err := s.cfg.Aggregator.Aggregate(validatorSet, sigs)
	if err != nil {
		return errors.Errorf("failed to prove: %w", err)
	}
	s.cfg.Metrics.ObserveOnlyAggregateDuration(time.Since(onlyAggregateStart))

	slog.InfoContext(ctx, "Aggregation proof created",
		"duration", time.Since(appAggregationStart).String(),
	)

	err = s.cfg.P2PClient.BroadcastSignatureAggregatedMessage(ctx, proofData)
	if err != nil {
		return errors.Errorf("failed to broadcast signature aggregated message: %w", err)
	}
	s.cfg.Metrics.ObserveAppAggregateDuration(time.Since(appAggregationStart))

	slog.InfoContext(ctx, "Aggregation completed, proof broadcast via p2p",
		"totalDuration", time.Since(appAggregationStart).String())

	return nil
}

const epochsToCheckForMissingProofs = 20

func (s *AggregatorApp) TryAggregateRequestsWithoutProof(ctx context.Context) error {
	latestEpoch, err := s.cfg.Repo.GetLatestValidatorSetEpoch(ctx)
	if err != nil {
		return errors.Errorf("failed to get latest epoch: %w", err)
	}

	startEpoch := symbiotic.Epoch(0)
	if latestEpoch >= symbiotic.Epoch(epochsToCheckForMissingProofs) {
		startEpoch = latestEpoch - symbiotic.Epoch(epochsToCheckForMissingProofs)
	}

	for epoch := latestEpoch; epoch >= startEpoch; epoch-- {
		var lastHash common.Hash
		requests, err := s.cfg.Repo.GetSignatureRequestsWithoutAggregationProof(ctx, epoch, 10, lastHash)
		if err != nil {
			return errors.Errorf("failed to get signature requests without aggregation proof for epoch %d: %w", epoch, err)
		}

		if len(requests) == 0 {
			continue // No more requests for this epoch
		}

		// Collect request ids
		for _, req := range requests {
			if !req.KeyTag.Type().AggregationKey() {
				continue // Skip non-aggregation requests
			}

			err := s.TryAggregateProofForRequestID(ctx, req.RequestID)
			if err != nil {
				return errors.Errorf("failed to try aggregate proof for request ID %s: %w", req.RequestID.Hex(), err)
			}
			// remove pending from db
			err = s.cfg.Repo.RemoveAggregationProofPending(ctx, req.RequiredEpoch, req.RequestID)
			// ignore not found and tx conflict errors, as they indicate the proof was already processed or is being processed
			if err != nil && !errors.Is(err, entity.ErrEntityNotFound) && !errors.Is(err, entity.ErrTxConflict) {
				return errors.Errorf("failed to remove aggregation proof from pending collection: %w", err)
			}

			lastHash = req.RequestID
		}

	}

	return nil
}

func (s *AggregatorApp) GetAggregationStatus(ctx context.Context, requestID common.Hash) (symbiotic.AggregationStatus, error) {
	signatureRequest, err := s.cfg.Repo.GetSignatureRequest(ctx, requestID)
	if err != nil {
		return symbiotic.AggregationStatus{}, errors.Errorf("failed to get signature request: %w", err)
	}

	if !signatureRequest.KeyTag.Type().AggregationKey() {
		return symbiotic.AggregationStatus{}, errors.Errorf("key tag %s is not an aggregation key", signatureRequest.KeyTag)
	}
	signatures, err := s.cfg.Repo.GetAllSignatures(ctx, requestID)
	if err != nil {
		return symbiotic.AggregationStatus{}, errors.Errorf("failed to get all signatures: %w", err)
	}

	// Get validator set for quorum threshold checks and aggregation
	validatorSet, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, signatureRequest.RequiredEpoch)
	if err != nil {
		return symbiotic.AggregationStatus{}, errors.Errorf("failed to get validator set: %w", err)
	}

	validators, err := validatorSet.FindValidatorsByKeys(signatureRequest.KeyTag, extractPublicKeys(signatures))
	if err != nil {
		return symbiotic.AggregationStatus{}, errors.Errorf("failed to find validators by keys: %w", err)
	}

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
