package signature_listener

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/crypto"
	intEntity "github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/signals"
)

//go:generate mockgen -source=signature_listener_uc.go -destination=mocks/signature_listener_uc.go -package=mocks

type repo interface {
	GetValidatorByKey(ctx context.Context, epoch uint64, keyTag entity.KeyTag, publicKey []byte) (entity.Validator, uint32, error)
	GetValidatorSetByEpoch(ctx context.Context, epoch uint64) (entity.ValidatorSet, error)
}

type signatureProcessor interface {
	ProcessSignature(ctx context.Context, param entity.SaveSignatureParam) error
}

type Config struct {
	Repo               repo               `validate:"required"`
	SignatureProcessor signatureProcessor `validate:"required"`
	SignalCfg          signals.Config     `validate:"required"`
	SelfP2PID          string             `validate:"required"`
}

type SignatureListenerUseCase struct {
	cfg                  Config
	signatureSavedSignal *signals.Signal[entity.SignatureMessage]
}

func New(cfg Config) (*SignatureListenerUseCase, error) {
	if err := validate.New().Struct(cfg); err != nil {
		return nil, errors.Errorf("invalid config: %w", err)
	}

	return &SignatureListenerUseCase{
		cfg:                  cfg,
		signatureSavedSignal: signals.New[entity.SignatureMessage](cfg.SignalCfg, "signatureReceive", nil),
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

	publicKey, err := crypto.NewPublicKey(msg.KeyTag.Type(), msg.Signature.PublicKey)
	if err != nil {
		return errors.Errorf("failed to get public key: %w", err)
	}
	err = publicKey.VerifyWithHash(msg.Signature.MessageHash, msg.Signature.Signature)
	if err != nil {
		return errors.Errorf("failed to verify signature: %w", err)
	}

	validator, activeIndex, err := s.cfg.Repo.GetValidatorByKey(ctx, uint64(msg.Epoch), msg.KeyTag, publicKey.OnChain())
	if err != nil {
		return errors.Errorf("validator not found for public key %x: %w", msg.Signature.PublicKey, err)
	}

	if !validator.IsActive {
		return errors.Errorf("validator %s is not active", validator.Operator.Hex())
	}

	slog.DebugContext(ctx, "Found validator", "validator", validator)

	param := entity.SaveSignatureParam{
		RequestHash:      msg.RequestHash,
		Key:              publicKey.Raw(),
		Signature:        msg.Signature,
		ActiveIndex:      activeIndex,
		VotingPower:      validator.VotingPower,
		Epoch:            msg.Epoch,
		SignatureRequest: nil,
	}

	err = s.cfg.SignatureProcessor.ProcessSignature(ctx, param)
	if err != nil {
		return errors.Errorf("failed to process signature: %w", err)
	}

	return s.signatureSavedSignal.Emit(ctx, msg)
}

func (s *SignatureListenerUseCase) StartSignatureSavedMessageListener(ctx context.Context, mh func(ctx context.Context, msg entity.SignatureMessage) error) error {
	if err := s.signatureSavedSignal.SetHandler(mh); err != nil {
		return errors.Errorf("failed to set signature received message handler: %w", err)
	}
	return s.signatureSavedSignal.StartWorkers(ctx)
}
