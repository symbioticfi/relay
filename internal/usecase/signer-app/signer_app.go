package signer_app

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"middleware-offchain/core/entity"
	p2pEntity "middleware-offchain/internal/entity"
)

type repo interface {
	GetSignatureRequest(ctx context.Context, reqHash common.Hash) (entity.SignatureRequest, error)
	GetAggregationProof(ctx context.Context, reqHash common.Hash) (entity.AggregationProof, error)
	SaveAggregationProof(ctx context.Context, reqHash common.Hash, ap entity.AggregationProof) error
	GetValsetByEpoch(ctx context.Context, epoch uint64) (entity.ValidatorSet, error)
	SaveSignature(ctx context.Context, reqHash common.Hash, key []byte, sig entity.Signature) error
	SaveSignatureRequest(_ context.Context, req entity.SignatureRequest) error
}

type p2pService interface {
	BroadcastSignatureGeneratedMessage(ctx context.Context, msg entity.SignatureMessage) error
	SetSignaturesAggregatedMessageHandler(mh func(ctx context.Context, si p2pEntity.SenderInfo, msg entity.AggregatedSignatureMessage) error)
}

type signer interface {
	Sign(keyTag entity.KeyTag, message []byte) (entity.Signature, error)
	Hash(keyTag entity.KeyTag, message []byte) ([]byte, error)
	GetPublicKey(keyTag entity.KeyTag) ([]byte, error)
}

type aggProofSignal interface {
	Emit(ctx context.Context, payload entity.AggregatedSignatureMessage)
}

type aggregator interface {
	Verify(
		valset *entity.ValidatorSet,
		keyTag entity.KeyTag,
		aggregationProof *entity.AggregationProof,
	) (bool, error)
}

type Config struct {
	P2PService     p2pService     `validate:"required"`
	Signer         signer         `validate:"required"`
	Repo           repo           `validate:"required"`
	AggProofSignal aggProofSignal `validate:"required"`
	Aggregator     aggregator     `validate:"required"`
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("failed to validate config: %w", err)
	}

	return nil
}

type SignerApp struct {
	cfg Config
}

func NewSignerApp(cfg Config) (*SignerApp, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	app := &SignerApp{
		cfg: cfg,
	}

	cfg.P2PService.SetSignaturesAggregatedMessageHandler(app.HandleSignaturesAggregatedMessage)

	return app, nil
}

func (s *SignerApp) Sign(ctx context.Context, req entity.SignatureRequest) error {
	_, err := s.cfg.Repo.GetSignatureRequest(ctx, req.Hash())
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		return errors.Errorf("failed to get signature request: %w", err)
	}
	if entityFound := !errors.Is(err, entity.ErrEntityNotFound); entityFound {
		slog.DebugContext(ctx, "signature request already exists", "request", req)
		return nil
	}

	_, err = s.cfg.Repo.GetAggregationProof(ctx, req.Hash())
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		return errors.Errorf("failed to get aggregation proof: %w", err)
	}
	if !errors.Is(err, entity.ErrEntityNotFound) {
		return errors.New("aggregation proof already exists for this request")
	}

	valset, err := s.cfg.Repo.GetValsetByEpoch(ctx, req.RequiredEpoch)
	if err != nil {
		return errors.Errorf("failed to get valset by epoch %d: %w", req.RequiredEpoch, err)
	}

	public, err := s.cfg.Signer.GetPublicKey(req.KeyTag)
	if err != nil {
		return errors.Errorf("failed to get public key for key tag %d: %w", req.KeyTag, err)
	}

	_, found := valset.FindValidatorByKey(req.KeyTag, public)
	if !found {
		return errors.Errorf("validator not found in epoch valset for public key")
	}

	signature, err := s.cfg.Signer.Sign(req.KeyTag, req.Message)
	if err != nil {
		return errors.Errorf("failed to sign valset header hash: %w", err)
	}

	if err := s.cfg.Repo.SaveSignature(ctx, req.Hash(), public, signature); err != nil {
		return errors.Errorf("failed to save signature: %w", err)
	}

	slog.InfoContext(ctx, "valset header hash signed, sending via p2p", "headerSignature", signature)

	err = s.cfg.P2PService.BroadcastSignatureGeneratedMessage(ctx, entity.SignatureMessage{
		RequestHash: req.Hash(),
		KeyTag:      req.KeyTag,
		Epoch:       req.RequiredEpoch,
		Signature:   signature,
	})
	if err != nil {
		return errors.Errorf("failed to broadcast valset header: %w", err)
	}

	if err := s.cfg.Repo.SaveSignatureRequest(ctx, req); err != nil {
		return errors.Errorf("failed to save signature request: %w", err)
	}

	return nil
}
