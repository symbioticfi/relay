package aggregator_app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"

	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
	"middleware-offchain/pkg/log"
)

//go:generate mockgen -source=aggregator_app.go -destination=mocks/aggregator_app.go -package=mocks
type repository interface {
	GetValsetByEpoch(ctx context.Context, epoch uint64) (entity.ValidatorSet, error)
	SaveSignature(ctx context.Context, reqHash common.Hash, key [32]byte, sig entity.Signature) error
	GetAllSignatures(ctx context.Context, reqHash common.Hash) ([]entity.Signature, error)
}

type p2pClient interface {
	BroadcastSignatureAggregatedMessage(ctx context.Context, msg entity.AggregatedSignatureMessage) error
	SetSignatureHashMessageHandler(mh func(ctx context.Context, msg entity.P2PSignatureHashMessage) error)
}

type aggregator interface {
	Aggregate(
		valset *entity.ValidatorSet,
		keyTag entity.KeyTag,
		verificationType entity.VerificationType,
		messageHash []byte,
		signatures []entity.Signature,
	) (*entity.AggregationProof, error)
}

type Config struct {
	Repo       repository `validate:"required"`
	P2PClient  p2pClient  `validate:"required"`
	Aggregator aggregator `validate:"required"`
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

	cfg.P2PClient.SetSignatureHashMessageHandler(app.HandleSignatureGeneratedMessage)

	return app, nil
}

func (s *AggregatorApp) HandleSignatureGeneratedMessage(ctx context.Context, msg entity.P2PSignatureHashMessage) error {
	ctx = log.WithComponent(ctx, "aggregator")

	slog.DebugContext(ctx, "received signature hash generated message", "message", msg)

	validatorSet, err := s.cfg.Repo.GetValsetByEpoch(ctx, msg.Message.Epoch)
	if err != nil {
		return fmt.Errorf("failed to get validator set: %w", err)
	}

	g1, _, err := bls.UnpackPublicG1G2(msg.Message.Signature.PublicKey) // todo ilya discuss how to get rid of dependency on bls package here
	if err != nil {
		return errors.Errorf("failed to unpack public key: %w", err)
	}

	validator, found := validatorSet.FindValidatorByKey(msg.Message.KeyTag, g1.Marshal())
	if !found {
		return errors.Errorf("validator not found for public key: %x", msg.Message.Signature.PublicKey)
	}

	err = s.cfg.Repo.SaveSignature(ctx, msg.Message.RequestHash, g1.Bytes(), msg.Message.Signature)
	if err != nil {
		return fmt.Errorf("failed to save signature: %w", err)
	}

	slog.DebugContext(ctx, "found validator", "validator", validator)

	current, err := s.hashStore.PutHash(msg.Message.Signature, validator)
	if err != nil {
		return errors.Errorf("failed to put signature: %w", err)
	}

	slog.DebugContext(ctx, "total voting power", "currentVotingPower", current.votingPower.String())

	thresholdReached := current.votingPower.Cmp(validatorSet.QuorumThreshold) >= 0
	if !thresholdReached {
		slog.InfoContext(ctx, "quorum not reached yet",
			"currentVotingPower", current.votingPower,
			"quorumThreshold", validatorSet.QuorumThreshold,
			"totalActiveVotingPower", validatorSet.GetTotalActiveVotingPower(),
			"aggSignature", current.aggSignature,
			"aggPublicKeyG1", current.aggPublicKeyG1,
			"aggPublicKeyG2", current.aggPublicKeyG2,
		)
		return nil
	}

	slog.InfoContext(ctx, "quorum reached, aggregating signatures and creating proof",
		"currentVotingPower", current.votingPower,
		"quorumThreshold", validatorSet.QuorumThreshold,
		"totalActiveVotingPower", validatorSet.GetTotalActiveVotingPower(),
	)

	sigs, err := s.cfg.Repo.GetAllSignatures(ctx, msg.Message.RequestHash)
	slog.DebugContext(ctx, "total received signatures", "sigs", len(sigs))
	if err != nil {
		return fmt.Errorf("failed to get signature aggregated message: %w", err)
	}

	// todo ilya, make proof only once when threshold is reached
	start := time.Now()
	// todo fix aggregation type
	proofData, err := s.cfg.Aggregator.Aggregate(&validatorSet, msg.Message.KeyTag, entity.VerificationTypeZK, msg.Message.Signature.MessageHash, sigs)
	if err != nil {
		return fmt.Errorf("failed to prove: %w", err)
	}

	slog.InfoContext(ctx, "proof created, trying to send aggregated signature message",
		"duration", time.Since(start).String(),
	)
	err = s.cfg.P2PClient.BroadcastSignatureAggregatedMessage(ctx, entity.AggregatedSignatureMessage{
		RequestHash:      msg.Message.RequestHash,
		KeyTag:           msg.Message.KeyTag,
		Epoch:            msg.Message.Epoch,
		AggregationProof: *proofData,
	})
	if err != nil {
		return errors.Errorf("failed to broadcast signature aggregated message: %w", err)
	}

	slog.InfoContext(ctx, "proof sent via p2p")

	return nil
}
