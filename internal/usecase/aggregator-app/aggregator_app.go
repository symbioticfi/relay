package aggregator_app

import (
	"context"
	"log/slog"
	"time"

	"github.com/symbioticfi/relay/core/usecase/aggregator"
	"github.com/symbioticfi/relay/core/usecase/crypto"

	"github.com/ethereum/go-ethereum/common"

	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
	p2pEntity "github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
)

//go:generate mockgen -source=aggregator_app.go -destination=mocks/aggregator_app.go -package=mocks
type repository interface {
	GetValidatorSetByEpoch(ctx context.Context, epoch uint64) (entity.ValidatorSet, error)
	GetValidatorByKey(ctx context.Context, epoch uint64, keyTag entity.KeyTag, publicKey []byte) (entity.Validator, error)
	SaveSignature(ctx context.Context, reqHash common.Hash, key []byte, sig entity.SignatureExtended) error
	GetAllSignatures(ctx context.Context, reqHash common.Hash) ([]entity.SignatureExtended, error)
	GetConfigByEpoch(ctx context.Context, epoch uint64) (entity.NetworkConfig, error)
	UpdateSignatureStat(_ context.Context, reqHash common.Hash, s entity.SignatureStatStage, t time.Time) (entity.SignatureStat, error)
}

type p2pClient interface {
	BroadcastSignatureAggregatedMessage(ctx context.Context, msg entity.AggregatedSignatureMessage) error
}

type metrics interface {
	ObserveOnlyAggregateDuration(d time.Duration)
	ObserveAppAggregateDuration(d time.Duration)
	ObserveAggCompleted(stat entity.SignatureStat)
}

type Config struct {
	Repo       repository            `validate:"required"`
	P2PClient  p2pClient             `validate:"required"`
	Aggregator aggregator.Aggregator `validate:"required"`
	Metrics    metrics               `validate:"required"`
}

func (c Config) Validate() error {
	if err := validate.New().Struct(c); err != nil {
		return errors.Errorf("failed to validate config: %w", err)
	}

	return nil
}

type AggregatorApp struct {
	cfg       Config
	hashStore *hashStore
}

func NewAggregatorApp(cfg Config) (*AggregatorApp, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	app := &AggregatorApp{
		cfg:       cfg,
		hashStore: newHashStore(),
	}

	return app, nil
}

func (s *AggregatorApp) HandleSignatureGeneratedMessage(ctx context.Context, p2pMsg p2pEntity.P2PMessage[entity.SignatureMessage]) error {
	ctx = log.WithComponent(ctx, "aggregator")
	appAggregationStart := time.Now()

	msg := p2pMsg.Message

	slog.DebugContext(ctx, "Received signature hash generated message", "message", msg, "sender", p2pMsg.SenderInfo.Sender)

	publicKey, err := crypto.NewPublicKey(msg.KeyTag.Type(), msg.Signature.PublicKey)
	if err != nil {
		return errors.Errorf("failed to get public key: %w", err)
	}
	err = publicKey.VerifyWithHash(msg.Signature.MessageHash, msg.Signature.Signature)
	if err != nil {
		return errors.Errorf("failed to verify signature: %w", err)
	}

	validator, err := s.cfg.Repo.GetValidatorByKey(ctx, uint64(msg.Epoch), msg.KeyTag, publicKey.OnChain())
	if err != nil {
		return errors.Errorf("validator not found for public key %x: %w", msg.Signature.PublicKey, err)
	}

	err = s.cfg.Repo.SaveSignature(ctx, msg.RequestHash, publicKey.Raw(), msg.Signature)
	if err != nil {
		return errors.Errorf("failed to save signature: %w", err)
	}

	slog.DebugContext(ctx, "Found validator", "validator", validator)

	current, err := s.hashStore.PutHash(msg.Signature, validator)
	if err != nil {
		return errors.Errorf("failed to put signature: %w", err)
	}

	slog.DebugContext(ctx, "Total voting power", "currentVotingPower", current.VotingPower.String())

	// Get validator set for quorum threshold checks and aggregation
	validatorSet, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, uint64(msg.Epoch))
	if err != nil {
		return errors.Errorf("failed to get validator set: %w", err)
	}

	totalActiveVotingPower := validatorSet.GetTotalActiveVotingPower()
	thresholdReached := current.VotingPower.Cmp(validatorSet.QuorumThreshold.Int) >= 0
	if !thresholdReached {
		slog.InfoContext(ctx, "Quorum not reached yet",
			"currentVotingPower", current.VotingPower.String(),
			"quorumThreshold", validatorSet.QuorumThreshold.String(),
			"totalActiveVotingPower", totalActiveVotingPower.String(),
		)
		return nil
	}

	slog.InfoContext(ctx, "Quorum reached, aggregating signatures and creating proof",
		"currentVotingPower", current.VotingPower.String(),
		"quorumThreshold", validatorSet.QuorumThreshold.String(),
		"totalActiveVotingPower", totalActiveVotingPower.String(),
	)

	if _, err := s.cfg.Repo.UpdateSignatureStat(ctx, msg.RequestHash, entity.SignatureStatStageAggQuorumReached, time.Now()); err != nil {
		slog.WarnContext(ctx, "Failed to update signature stat: %s", "error", err)
	}

	sigs, err := s.cfg.Repo.GetAllSignatures(ctx, msg.RequestHash)
	slog.DebugContext(ctx, "Total received signatures", "sigs", len(sigs))
	if err != nil {
		return errors.Errorf("failed to get signature aggregated message: %w", err)
	}

	start := time.Now()
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
		"duration", time.Since(start).String(),
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

func (s *AggregatorApp) GetAggregationStatus(ctx context.Context, requestHash common.Hash) (p2pEntity.AggregationStatus, error) {
	current, err := s.hashStore.GetStatus(requestHash)
	if err != nil {
		return p2pEntity.AggregationStatus{}, errors.Errorf("failed to get aggregation status: %w", err)
	}

	return current, nil
}
