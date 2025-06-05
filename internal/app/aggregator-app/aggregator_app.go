package aggregator_app

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"

	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
	"middleware-offchain/pkg/log"
	"middleware-offchain/pkg/proof"
)

//go:generate mockgen -source=aggregator_app.go -destination=mocks/aggregator_app.go -package=mocks
type ethClient interface {
	GetQuorumThreshold(ctx context.Context, timestamp *big.Int, keyTag entity.KeyTag) (*big.Int, error)
}

type valsetDeriver interface {
	GetValidatorSet(ctx context.Context, timestamp *big.Int) (entity.ValidatorSet, error)
}

type p2pClient interface {
	BroadcastSignatureAggregatedMessage(ctx context.Context, msg entity.SignaturesAggregatedMessage) error
	SetSignatureHashMessageHandler(mh func(ctx context.Context, msg entity.P2PSignatureHashMessage) error)
}

type Config struct {
	EthClient     ethClient     `validate:"required"`
	ValsetDeriver valsetDeriver `validate:"required"`
	P2PClient     p2pClient     `validate:"required"`
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

	validatorSet, err := s.cfg.ValsetDeriver.GetValidatorSet(ctx /*msg.Message.ValsetHeaderTimestamp*/, nil) // todo ilya
	if err != nil {
		return fmt.Errorf("failed to get validator set: %w", err)
	}

	g1, _, err := bls.UnpackPublicG1G2(msg.Message.PublicKey) // todo ilya discuss how to get rid of dependency on bls package here
	if err != nil {
		return errors.Errorf("failed to unpack public key: %w", err)
	}

	validator, found := validatorSet.FindValidatorByKey(msg.Message.KeyTag, g1.Marshal())
	if !found {
		return errors.Errorf("validator not found for public key: %x", msg.Message.PublicKey)
	}

	slog.DebugContext(ctx, "found validator", "validator", validator)

	current, err := s.hashStore.PutHash(msg.Message, validator)
	if err != nil {
		return errors.Errorf("failed to put signature: %w", err)
	}

	slog.DebugContext(ctx, "total voting power", "currentVotingPower", current.votingPower.String())

	quorumThreshold, err := s.cfg.EthClient.GetQuorumThreshold(ctx, big.NewInt(time.Now().Unix()), msg.Message.KeyTag)
	if err != nil {
		return errors.Errorf("failed to get quorum threshold: %w", err)
	}

	slog.DebugContext(ctx, "got quorum threshold", "quorumThreshold", quorumThreshold.String())

	coef1e18 := big.NewInt(1e18)

	vpMul1e18 := new(big.Int).Mul(current.votingPower, coef1e18)
	percent1e18 := new(big.Int).Div(vpMul1e18, validatorSet.TotalActiveVotingPower)

	thresholdReached := percent1e18.Cmp(quorumThreshold) >= 0
	if !thresholdReached {
		slog.InfoContext(ctx, "quorum not reached yet",
			"percentReached", decimal.NewFromBigInt(percent1e18, 0).Div(decimal.NewFromBigInt(coef1e18, 0)).String(),
			"percentQuorumThreshold", decimal.NewFromBigInt(quorumThreshold, 0).Div(decimal.NewFromBigInt(coef1e18, 0)).String(),
			"currentVotingPower", current.votingPower,
			"quorumThreshold", quorumThreshold,
			"totalActiveVotingPower", validatorSet.TotalActiveVotingPower,
			"aggSignature", current.aggSignature,
			"aggPublicKeyG1", current.aggPublicKeyG1,
			"aggPublicKeyG2", current.aggPublicKeyG2,
		)
		return nil
	}

	slog.InfoContext(ctx, "quorum reached, aggregating signatures and creating proof",
		"percentReached", decimal.NewFromBigInt(percent1e18, 0).Div(decimal.NewFromBigInt(coef1e18, 0)).String(),
		"percentQuorumThreshold", decimal.NewFromBigInt(quorumThreshold, 0).Div(decimal.NewFromBigInt(coef1e18, 0)).String(),
		"currentVotingPower", current.votingPower,
		"quorumThreshold", quorumThreshold,
		"totalActiveVotingPower", validatorSet.TotalActiveVotingPower,
	)

	// todo ilya, make proof only once when threshold is reached
	start := time.Now()
	proofData, err := proof.DoProve(proof.RawProveInput{
		SignerValidators: current.validators,
		AllValidators:    validatorSet.Validators,
		RequiredKeyTag:   msg.Message.KeyTag,
		Message:          msg.Message.MessageHash,
		Signature:        *current.aggSignature,
		SignersAggKeyG2:  *current.aggPublicKeyG2,
	})
	if err != nil {
		return fmt.Errorf("failed to prove: %w", err)
	}

	slog.InfoContext(ctx, "proof created, trying to send aggregated signature message",
		"duration", time.Since(start).String(),
	)
	err = s.cfg.P2PClient.BroadcastSignatureAggregatedMessage(ctx, entity.SignaturesAggregatedMessage{
		PublicKeyG1: current.aggPublicKeyG1,
		Proof:       proofData.Marshall(),
		Message:     msg.Message.MessageHash,
		HashType:    msg.Message.HashType,
		Epoch:       msg.Message.Epoch,
	})
	if err != nil {
		return errors.Errorf("failed to broadcast signature aggregated message: %w", err)
	}

	slog.InfoContext(ctx, "proof sent via p2p")

	return nil
}
