package committer_app

import (
	"context"
	"log/slog"
	"math/big"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"middleware-offchain/internal/client/valset"
	"middleware-offchain/internal/entity"
)

type valsetGenerator interface {
	GenerateCurrentValidatorSetHeader(ctx context.Context) (entity.ValidatorSetHeader, error)
	GenerateExtraData(ctx context.Context, valsetHeader entity.ValidatorSetHeader, verificationType uint32) ([]entity.ExtraData, error)
}

type ethClient interface {
	CommitValsetHeader(ctx context.Context, valsetHeader entity.ValidatorSetHeader, extraData []entity.ExtraData, proof []byte, hint []byte) error
	VerifyQuorumSig(ctx context.Context, epoch *big.Int, message []byte, keyTag uint8, threshold *big.Int, proof []byte, hint []byte) (bool, error)
}

type p2pClient interface {
	SetSignaturesAggregatedMessageHandler(mh func(ctx context.Context, msg entity.P2PSignaturesAggregatedMessage) error)
}

type Config struct {
	ValsetGenerator valsetGenerator `validate:"required"`
	EthClient       ethClient       `validate:"required"`
	P2PClient       p2pClient       `validate:"required"`
}

func (c Config) Validate() error {
	return validator.New().Struct(c)
}

type CommitterApp struct {
	cfg Config
}

func NewCommitterApp(cfg Config) (*CommitterApp, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	app := &CommitterApp{
		cfg: cfg,
	}

	cfg.P2PClient.SetSignaturesAggregatedMessageHandler(app.HandleSignaturesAggregatedMessage)

	return app, nil
}

func (c *CommitterApp) HandleSignaturesAggregatedMessage(ctx context.Context, msg entity.P2PSignaturesAggregatedMessage) error {
	slog.DebugContext(ctx, "got signatures aggregated message", "message", msg)

	switch msg.Message.HashType {
	case entity.HashTypeValsetHeader:
		return c.commitValsetHeader(ctx, msg)
	case entity.HashTypeMessage:
		return c.verifyQuorumSig(ctx, msg)
	}

	return errors.Errorf("unsupported hash type: %s", msg.Message.HashType)
}

func (c *CommitterApp) commitValsetHeader(ctx context.Context, msg entity.P2PSignaturesAggregatedMessage) error {
	header, err := c.cfg.ValsetGenerator.GenerateCurrentValidatorSetHeader(ctx)
	if err != nil {
		return errors.Errorf("failed to generate valset header: %w", err)
	}

	extraData, err := c.cfg.ValsetGenerator.GenerateExtraData(ctx, header, valset.ZkVerificationType)
	if err != nil {
		return errors.Errorf("failed to generate extra data: %w", err)
	}

	slog.DebugContext(ctx, "generated valset header, committing", "header", header)

	err = c.cfg.EthClient.CommitValsetHeader(ctx, header, extraData, msg.Message.Proof, []byte{})
	if err != nil {
		return errors.Errorf("failed to commit valset header: %w", err)
	}

	slog.DebugContext(ctx, "valset header committed successfully")

	return nil
}

func (c *CommitterApp) verifyQuorumSig(ctx context.Context, msg entity.P2PSignaturesAggregatedMessage) error {
	epoch := new(big.Int).SetInt64(10) // todo ilya pass from signer
	isOK, err := c.cfg.EthClient.VerifyQuorumSig(ctx, epoch, msg.Message.Message, 15, new(big.Int).SetInt64(1e18) /*1%*/, msg.Message.Proof, []byte{})
	if err != nil {
		return errors.Errorf("failed to verify quorum signature: %w", err)
	}
	if !isOK {
		return errors.New("quorum signature verification failed")
	}

	slog.DebugContext(ctx, "quorum signature verified successfully")
	return nil
}
