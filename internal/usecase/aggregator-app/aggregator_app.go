package aggregator_app

import (
	"context"
	"log/slog"
	"time"

	keyprovider "github.com/symbioticfi/relay/core/usecase/key-provider"
	aggregationPolicyTypes "github.com/symbioticfi/relay/internal/usecase/aggregation-policy/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/crypto"
	"github.com/symbioticfi/relay/pkg/log"
)

//go:generate mockgen -source=aggregator_app.go -destination=mocks/aggregator_app.go -package=mocks
type repository interface {
	GetValidatorSetByEpoch(ctx context.Context, epoch entity.Epoch) (entity.ValidatorSet, error)
	GetAggregationProof(ctx context.Context, requestID common.Hash) (entity.AggregationProof, error)
	GetSignatureRequest(_ context.Context, requestID common.Hash) (entity.SignatureRequest, error)
	GetAllSignatures(ctx context.Context, requestID common.Hash) ([]entity.SignatureExtended, error)
	GetConfigByEpoch(ctx context.Context, epoch entity.Epoch) (entity.NetworkConfig, error)
	GetSignatureMap(ctx context.Context, requestID common.Hash) (entity.SignatureMap, error)
}

type p2pClient interface {
	BroadcastSignatureAggregatedMessage(ctx context.Context, proof entity.AggregationProof) error
}

type metrics interface {
	ObserveOnlyAggregateDuration(d time.Duration)
	ObserveAppAggregateDuration(d time.Duration)
}

type aggregator interface {
	Aggregate(valset entity.ValidatorSet, keyTag entity.KeyTag, messageHash []byte, signatures []entity.SignatureExtended) (entity.AggregationProof, error)
}

type aggregatorPolicy = aggregationPolicyTypes.AggregationPolicy

type Config struct {
	Repo              repository              `validate:"required"`
	P2PClient         p2pClient               `validate:"required"`
	Aggregator        aggregator              `validate:"required"`
	Metrics           metrics                 `validate:"required"`
	AggregationPolicy aggregatorPolicy        `validate:"required"`
	KeyProvider       keyprovider.KeyProvider `validate:"required"`
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

func (s *AggregatorApp) HandleSignatureProcessedMessage(ctx context.Context, msg entity.SignatureExtended) error {
	ctx = log.WithComponent(ctx, "aggregator")
	ctx = log.WithAttrs(ctx, slog.Uint64("epoch", uint64(msg.Epoch)))
	slog.DebugContext(ctx, "Received HandleSignatureProcessedMessage", "message", msg)

	_, err := s.cfg.Repo.GetAggregationProof(ctx, msg.RequestID())
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		return errors.Errorf("failed to get aggregation proof: %w", err)
	}
	if err == nil {
		slog.DebugContext(ctx, "Aggregation proof already exists", "request", msg)
		return nil
	}

	signatureMap, err := s.cfg.Repo.GetSignatureMap(ctx, msg.RequestID())
	if err != nil {
		return errors.Errorf("failed to get valset signature map: %w", err)
	}

	if signatureMap.RequestID != msg.RequestID() || signatureMap.Epoch != msg.Epoch {
		return errors.Errorf("signature map context mismatch: map %s/%d vs msg %s/%d",
			signatureMap.RequestID.Hex(), signatureMap.Epoch,
			msg.RequestID().Hex(), msg.Epoch,
		)
	}

	// Get validator set for quorum threshold checks
	// todo load only valset header when totalVotingPower is added to it
	validatorSet, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, msg.Epoch)
	if err != nil {
		return errors.Errorf("failed to get validator set: %w", err)
	}

	privKey, err := s.cfg.KeyProvider.GetPrivateKey(validatorSet.RequiredKeyTag)
	if err != nil {
		if errors.Is(err, keyprovider.ErrKeyNotFound) {
			slog.DebugContext(ctx, "No key for required key tag, skipping proof aggregation", "keyTag", validatorSet.RequiredKeyTag)
			return nil
		}
		return errors.Errorf("failed to get private key for required key tag %s: %w", validatorSet.RequiredKeyTag, err)
	}

	if !validatorSet.IsAggregator(privKey.PublicKey().OnChain()) {
		slog.DebugContext(ctx, "Not an Aggregator for this valset, skipping proof aggregation",
			"key", privKey.PublicKey().OnChain(),
			"epoch", msg.Epoch,
			"aggIndices", validatorSet.AggregatorIndices,
		)
		return nil
	}

	slog.DebugContext(ctx, "Is an Aggregator for this valset, checking quorum", "key", privKey.PublicKey().OnChain(), "epoch", msg.Epoch)

	totalActiveVotingPower := validatorSet.GetTotalActiveVotingPower()

	if !s.cfg.AggregationPolicy.ShouldAggregate(signatureMap, validatorSet) {
		slog.DebugContext(ctx, "Quorum not reached yet",
			"currentVotingPower", signatureMap.CurrentVotingPower.String(),
			"quorumThreshold", validatorSet.QuorumThreshold.String(),
			"totalActiveVotingPower", totalActiveVotingPower.String(),
		)
		return nil
	}

	slog.InfoContext(ctx, "Quorum reached, aggregating signatures and creating proof",
		"currentVotingPower", signatureMap.CurrentVotingPower.String(),
		"quorumThreshold", validatorSet.QuorumThreshold.String(),
		"totalActiveVotingPower", totalActiveVotingPower.String(),
	)

	appAggregationStart := time.Now()

	sigs, err := s.cfg.Repo.GetAllSignatures(ctx, msg.RequestID())
	slog.DebugContext(ctx, "Total received signatures", "sigs", len(sigs))
	if err != nil {
		return errors.Errorf("failed to get signature aggregated message: %w", err)
	}

	networkConfig, err := s.cfg.Repo.GetConfigByEpoch(ctx, msg.Epoch)
	if err != nil {
		return errors.Errorf("failed to get network config: %w", err)
	}

	slog.DebugContext(ctx, "Received network config", "networkConfig", networkConfig)

	onlyAggregateStart := time.Now()
	proofData, err := s.cfg.Aggregator.Aggregate(
		validatorSet,
		msg.KeyTag,
		msg.MessageHash,
		sigs,
	)
	if err != nil {
		return errors.Errorf("failed to prove: %w", err)
	}
	s.cfg.Metrics.ObserveOnlyAggregateDuration(time.Since(onlyAggregateStart))

	slog.InfoContext(ctx, "Proof created, trying to send aggregated signature message",
		"duration", time.Since(appAggregationStart).String(),
		"request_id", msg.RequestID().Hex(),
	)

	err = s.cfg.P2PClient.BroadcastSignatureAggregatedMessage(ctx, proofData)
	if err != nil {
		return errors.Errorf("failed to broadcast signature aggregated message: %w", err)
	}
	s.cfg.Metrics.ObserveAppAggregateDuration(time.Since(appAggregationStart))

	slog.InfoContext(ctx, "Proof sent via p2p",
		"totalAggDuration", time.Since(appAggregationStart).String())

	return nil
}

func (s *AggregatorApp) GetAggregationStatus(ctx context.Context, requestID common.Hash) (entity.AggregationStatus, error) {
	signatureRequest, err := s.cfg.Repo.GetSignatureRequest(ctx, requestID)
	if err != nil {
		return entity.AggregationStatus{}, errors.Errorf("failed to get signature request: %w", err)
	}

	if !signatureRequest.KeyTag.Type().AggregationKey() {
		return entity.AggregationStatus{}, errors.Errorf("key tag %s is not an aggregation key", signatureRequest.KeyTag)
	}
	signatures, err := s.cfg.Repo.GetAllSignatures(ctx, requestID)
	if err != nil {
		return entity.AggregationStatus{}, errors.Errorf("failed to get all signatures: %w", err)
	}

	publicKeys, err := extractPublicKeys(signatureRequest.KeyTag, signatures)
	if err != nil {
		return entity.AggregationStatus{}, errors.Errorf("failed to extract public keys: %w", err)
	}

	// Get validator set for quorum threshold checks and aggregation
	validatorSet, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, signatureRequest.RequiredEpoch)
	if err != nil {
		return entity.AggregationStatus{}, errors.Errorf("failed to get validator set: %w", err)
	}

	validators, err := validatorSet.FindValidatorsByKeys(signatureRequest.KeyTag, publicKeys)
	if err != nil {
		return entity.AggregationStatus{}, errors.Errorf("failed to find validators by keys: %w", err)
	}

	return entity.AggregationStatus{
		VotingPower: validators.GetTotalActiveVotingPower(),
		Validators:  validators,
	}, nil
}

func extractPublicKeys(keyTag entity.KeyTag, signatures []entity.SignatureExtended) ([]entity.CompactPublicKey, error) {
	publicKeys := make([]entity.CompactPublicKey, 0, len(signatures))
	for _, signature := range signatures {
		pk, err := crypto.NewPublicKey(keyTag.Type(), signature.PublicKey)
		if err != nil {
			return nil, errors.Errorf("failed to get public key: %w", err)
		}
		publicKeys = append(publicKeys, pk.OnChain())
	}
	return publicKeys, nil
}
