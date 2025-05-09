package signer_app

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"

	"middleware-offchain/internal/entity"
	"middleware-offchain/valset/types"
)

type ethClient interface {
	GetQuorumThreshold(ctx context.Context, timestamp *big.Int, keyTag uint8) (*big.Int, error)
}

type valsetDeriver interface {
	GetValidatorSet(ctx context.Context, timestamp *big.Int) (types.ValidatorSet, error)
}

type Config struct {
	EthClient     ethClient     `validate:"required"`
	ValsetDeriver valsetDeriver `validate:"required"`
}

func (c Config) Validate() error {
	if err := validate.New().Struct(c); err != nil {
		return errors.Errorf("failed to validate config: %w", err)
	}

	return nil
}

type AggregatorApp struct {
	cfg          Config
	hashStore    *hashStore
	validatorSet types.ValidatorSet
}

func NewAggregatorApp(ctx context.Context, cfg Config) (*AggregatorApp, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	nowBig := big.NewInt(time.Now().Unix())

	validatorSet, err := cfg.ValsetDeriver.GetValidatorSet(ctx, nowBig)
	if err != nil {
		return nil, fmt.Errorf("failed to get validator set: %w", err)
	}

	return &AggregatorApp{
		cfg:          cfg,
		validatorSet: validatorSet,
		hashStore:    newHashStore(),
	}, nil
}

func (s *AggregatorApp) Start(ctx context.Context) error {
	return nil
}

func (s *AggregatorApp) HandleSignatureGeneratedMessage(ctx context.Context, msg entity.P2PSignatureHashMessage) error {
	if err := s.handleSignatureGeneratedMessage(ctx, msg); err != nil {
		slog.ErrorContext(ctx, "Failed to handle signature generated message", "error", err)
		return nil
	}
	slog.DebugContext(ctx, "Successfully handled signature generated message", "message", msg)

	return nil
}

func (s *AggregatorApp) handleSignatureGeneratedMessage(ctx context.Context, msg entity.P2PSignatureHashMessage) error {
	slog.InfoContext(ctx, "received signature hash generated message", "message", msg)

	validator, found := s.validatorSet.FindValidatorByKey(msg.Message.PublicKeyG1)
	if !found {
		return errors.Errorf("validator not found for public key: %x", msg.Message.PublicKeyG1)
	}

	totalVotingPower, err := s.hashStore.PutHash(msg.Message, validator)
	if err != nil {
		return errors.Errorf("failed to put hash: %w", err)
	}

	quorumThreshold, err := s.cfg.EthClient.GetQuorumThreshold(ctx, big.NewInt(time.Now().Unix()), msg.Message.KeyTag)
	if err != nil {
		return errors.Errorf("failed to get quorum threshold: %w", err)
	}

	slog.DebugContext(ctx, "quorum threshold", "quorumThreshold", quorumThreshold)

	if totalVotingPower.Cmp(quorumThreshold) < 0 {
		slog.DebugContext(ctx, "quorum not reached yet", "totalVotingPower", totalVotingPower)
		return nil
	}

	slog.InfoContext(ctx, "!!!!! quorum reached, generating signature hash")
	return nil
}
