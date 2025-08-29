package aggregator_app

import (
	"context"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/crypto"
	"github.com/symbioticfi/relay/pkg/log"
)

//go:generate mockgen -source=aggregator_app.go -destination=mocks/aggregator_app.go -package=mocks
type repository interface {
	GetValidatorSetByEpoch(ctx context.Context, epoch uint64) (entity.ValidatorSet, error)
	GetSignatureRequest(_ context.Context, reqHash common.Hash) (entity.SignatureRequest, error)
	GetAllSignatures(ctx context.Context, reqHash common.Hash) ([]entity.SignatureExtended, error)
	GetConfigByEpoch(ctx context.Context, epoch uint64) (entity.NetworkConfig, error)
	UpdateSignatureStat(_ context.Context, reqHash common.Hash, s entity.SignatureStatStage, t time.Time) (entity.SignatureStat, error)
	GetValidatorMap(_ context.Context, reqHash common.Hash) (entity.ValidatorMap, error)
}

type p2pClient interface {
	BroadcastSignatureAggregatedMessage(ctx context.Context, msg entity.AggregatedSignatureMessage) error
}

type metrics interface {
	ObserveOnlyAggregateDuration(d time.Duration)
	ObserveAppAggregateDuration(d time.Duration)
	ObserveAggCompleted(stat entity.SignatureStat)
}

type aggregator interface {
	Aggregate(valset entity.ValidatorSet, keyTag entity.KeyTag, messageHash []byte, signatures []entity.SignatureExtended) (entity.AggregationProof, error)
}

type Config struct {
	Repo       repository `validate:"required"`
	P2PClient  p2pClient  `validate:"required"`
	Aggregator aggregator `validate:"required"`
	Metrics    metrics    `validate:"required"`
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

func (s *AggregatorApp) HandleSignatureGeneratedMessage(ctx context.Context, msg entity.SignatureMessage) error {
	ctx = log.WithComponent(ctx, "aggregator")

	validatorMap, err := s.cfg.Repo.GetValidatorMap(ctx, msg.RequestHash)
	if err != nil {
		return errors.Errorf("failed to get valset validator map: %w", err)
	}

	if validatorMap.RequestHash != msg.RequestHash || validatorMap.Epoch != uint64(msg.Epoch) {
		return errors.Errorf("validator map context mismatch: map %s/%d vs msg %s/%d",
			validatorMap.RequestHash.Hex(), validatorMap.Epoch,
			msg.RequestHash.Hex(), msg.Epoch,
		)
	}

	if !validatorMap.ThresholdReached() {
		slog.DebugContext(ctx, "Quorum not reached yet",
			"currentVotingPower", validatorMap.CurrentVotingPower.String(),
			"quorumThreshold", validatorMap.QuorumThreshold.String(),
			"totalActiveVotingPower", validatorMap.TotalVotingPower.String(),
		)
		return nil
	}

	slog.InfoContext(ctx, "Quorum reached, aggregating signatures and creating proof",
		"currentVotingPower", validatorMap.CurrentVotingPower.String(),
		"quorumThreshold", validatorMap.QuorumThreshold.String(),
		"totalActiveVotingPower", validatorMap.TotalVotingPower.String(),
	)

	if _, err := s.cfg.Repo.UpdateSignatureStat(ctx, msg.RequestHash, entity.SignatureStatStageAggQuorumReached, time.Now()); err != nil {
		slog.WarnContext(ctx, "Failed to update signature stat: %s", "error", err)
	}

	appAggregationStart := time.Now()

	// Get validator set for quorum threshold checks and aggregation
	validatorSet, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, uint64(msg.Epoch))
	if err != nil {
		return errors.Errorf("failed to get validator set: %w", err)
	}

	sigs, err := s.cfg.Repo.GetAllSignatures(ctx, msg.RequestHash)
	slog.DebugContext(ctx, "Total received signatures", "sigs", len(sigs))
	if err != nil {
		return errors.Errorf("failed to get signature aggregated message: %w", err)
	}

	networkConfig, err := s.cfg.Repo.GetConfigByEpoch(ctx, uint64(msg.Epoch))
	if err != nil {
		return errors.Errorf("failed to get network config: %w", err)
	}

	slog.DebugContext(ctx, "Received network config", "networkConfig", networkConfig)

	onlyAggregateStart := time.Now()
	proofData, err := s.cfg.Aggregator.Aggregate(
		validatorSet,
		msg.KeyTag,
		msg.Signature.MessageHash,
		sigs,
	)
	if err != nil {
		return errors.Errorf("failed to prove: %w", err)
	}
	s.cfg.Metrics.ObserveOnlyAggregateDuration(time.Since(onlyAggregateStart))

	slog.InfoContext(ctx, "Proof created, trying to send aggregated signature message",
		"duration", time.Since(appAggregationStart).String(),
	)
	err = s.cfg.P2PClient.BroadcastSignatureAggregatedMessage(ctx, entity.AggregatedSignatureMessage{
		RequestHash:      msg.RequestHash,
		KeyTag:           msg.KeyTag,
		Epoch:            msg.Epoch,
		AggregationProof: proofData,
	})
	if err != nil {
		return errors.Errorf("failed to broadcast signature aggregated message: %w", err)
	}

	stat, err := s.cfg.Repo.UpdateSignatureStat(ctx, msg.RequestHash, entity.SignatureStatStageAggCompleted, time.Now())
	if err != nil {
		slog.WarnContext(ctx, "Failed to update signature stat: %s", "error", err)
	}

	s.cfg.Metrics.ObserveAppAggregateDuration(time.Since(appAggregationStart))
	s.cfg.Metrics.ObserveAggCompleted(stat)

	slog.InfoContext(ctx, "Proof sent via p2p",
		"totalAggDuration", time.Since(appAggregationStart).String())

	return nil
}

func (s *AggregatorApp) GetAggregationStatus(ctx context.Context, requestHash common.Hash) (entity.AggregationStatus, error) {
	signatureRequest, err := s.cfg.Repo.GetSignatureRequest(ctx, requestHash)
	if err != nil {
		return entity.AggregationStatus{}, errors.Errorf("failed to get signature request: %w", err)
	}

	signatures, err := s.cfg.Repo.GetAllSignatures(ctx, requestHash)
	if err != nil {
		return entity.AggregationStatus{}, errors.Errorf("failed to get all signatures: %w", err)
	}

	publicKeys, err := extractPublicKeys(signatureRequest.KeyTag, signatures)
	if err != nil {
		return entity.AggregationStatus{}, errors.Errorf("failed to extract public keys: %w", err)
	}

	// Get validator set for quorum threshold checks and aggregation
	validatorSet, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, uint64(signatureRequest.RequiredEpoch))
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
