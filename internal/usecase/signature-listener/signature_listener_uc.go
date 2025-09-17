package signature_listener

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
	intEntity "github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/signals"
)

//go:generate mockgen -source=signature_listener_uc.go -destination=mocks/signature_listener_uc.go -package=mocks

type repo interface {
	GetValidatorByKey(ctx context.Context, epoch uint64, keyTag entity.KeyTag, publicKey []byte) (entity.Validator, uint32, error)
	GetValidatorSetByEpoch(ctx context.Context, epoch uint64) (entity.ValidatorSet, error)
}

type entityProcessor interface {
	ProcessSignature(ctx context.Context, param entity.SaveSignatureParam) error
}

type Config struct {
	Repo            repo            `validate:"required"`
	EntityProcessor entityProcessor `validate:"required"`
	SignalCfg       signals.Config  `validate:"required"`
	SelfP2PID       string          `validate:"required"`
}

type SignatureListenerUseCase struct {
	cfg Config
}

func New(cfg Config) (*SignatureListenerUseCase, error) {
	if err := validate.New().Struct(cfg); err != nil {
		return nil, errors.Errorf("invalid config: %w", err)
	}

	return &SignatureListenerUseCase{
		cfg: cfg,
	}, nil
}

func (s *SignatureListenerUseCase) HandleSignatureReceivedMessage(ctx context.Context, p2pMsg intEntity.P2PMessage[entity.SignatureMessage]) error {
	ctx = log.WithComponent(ctx, "sign_listener")

	msg := p2pMsg.Message

	slog.DebugContext(ctx, "Received signature hash generated message", "message", msg, "sender", p2pMsg.SenderInfo.Sender)

	if p2pMsg.SenderInfo.Sender == s.cfg.SelfP2PID {
		slog.DebugContext(ctx, "Ignoring signature message from self, because it's already stored in signer")
		return nil
	}

	param := entity.SaveSignatureParam{
		KeyTag:           msg.KeyTag,
		RequestHash:      msg.RequestHash,
		Signature:        msg.Signature,
		Epoch:            msg.Epoch,
		SignatureRequest: nil,
	}

	err := s.cfg.EntityProcessor.ProcessSignature(ctx, param)
	if err != nil {
		return errors.Errorf("failed to process signature: %w", err)
	}

	slog.InfoContext(ctx, "Listener processed received signature",
		"request_hash", msg.RequestHash.Hex(),
		"epoch", msg.Epoch,
	)
	return nil
}
