package committer_app

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"middleware-offchain/internal/entity"
)

type valsetGenerator interface {
	GenerateValidatorSetHeader(ctx context.Context) (entity.ValidatorSetHeader, error)
}

type ethClient interface {
	CommitValsetHeader(ctx context.Context, valsetHeader entity.ValidatorSetHeader, proof []byte) error
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

	header, err := c.cfg.ValsetGenerator.GenerateValidatorSetHeader(ctx)
	if err != nil {
		return errors.Errorf("failed to generate valset header: %w", err)
	}

	slog.DebugContext(ctx, "generated valset header, committing", "header", header)

	err = c.cfg.EthClient.CommitValsetHeader(ctx, header, msg.Message.Proof)
	if err != nil {
		return errors.Errorf("failed to commit valset header: %w", err)
	}

	slog.DebugContext(ctx, "valset header committed successfully")

	return nil
}
