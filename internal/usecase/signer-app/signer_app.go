package signer_app

import (
	"context"
	"log/slog"

	key_types "middleware-offchain/core/usecase/crypto/key-types"
	"middleware-offchain/pkg/log"

	"github.com/ethereum/go-ethereum/common"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"middleware-offchain/core/entity"
)

type repo interface {
	GetSignatureRequest(ctx context.Context, reqHash common.Hash) (entity.SignatureRequest, error)
	GetAggregationProof(ctx context.Context, reqHash common.Hash) (entity.AggregationProof, error)
	SaveAggregationProof(ctx context.Context, reqHash common.Hash, ap entity.AggregationProof) error
	GetValidatorSetByEpoch(ctx context.Context, epoch uint64) (entity.ValidatorSet, error)
	SaveSignature(ctx context.Context, reqHash common.Hash, key []byte, sig entity.SignatureExtended) error
	SaveSignatureRequest(_ context.Context, req entity.SignatureRequest) error
}

type p2pService interface {
	BroadcastSignatureGeneratedMessage(ctx context.Context, msg entity.SignatureMessage) error
}

type keyProvider interface {
	GetPrivateKey(keyTag entity.KeyTag) (key_types.PrivateKey, error)
}

type aggProofSignal interface {
	Emit(ctx context.Context, payload entity.AggregatedSignatureMessage)
}

type aggregator interface {
	Verify(valset entity.ValidatorSet, keyTag entity.KeyTag, aggregationProof entity.AggregationProof) (bool, error)
}

type Config struct {
	P2PService     p2pService     `validate:"required"`
	KeyProvider    keyProvider    `validate:"required"`
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

	return app, nil
}

func (s *SignerApp) Sign(ctx context.Context, req entity.SignatureRequest) error {
	ctx = log.WithComponent(ctx, "signer")

	_, err := s.cfg.Repo.GetSignatureRequest(ctx, req.Hash())
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		return errors.Errorf("failed to get signature request: %w", err)
	}
	if entityFound := !errors.Is(err, entity.ErrEntityNotFound); entityFound {
		slog.DebugContext(ctx, "Signature request already exists", "request", req)
		return nil
	}

	_, err = s.cfg.Repo.GetAggregationProof(ctx, req.Hash())
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		return errors.Errorf("failed to get aggregation proof: %w", err)
	}
	if err == nil {
		return errors.New("aggregation proof already exists for this request")
	}

	valset, err := s.cfg.Repo.GetValidatorSetByEpoch(ctx, uint64(req.RequiredEpoch))
	if err != nil {
		return errors.Errorf("failed to get valset by epoch %d: %w", req.RequiredEpoch, err)
	}

	private, err := s.cfg.KeyProvider.GetPrivateKey(req.KeyTag)
	if err != nil {
		return errors.Errorf("failed to get private key: %w", err)
	}

	public := private.PublicKey()
	_, found := valset.FindValidatorByKey(req.KeyTag, public.OnChain())
	if !found {
		return errors.Errorf("validator not found in epoch valset for public key")
	}

	signature, hash, err := private.Sign(req.Message)
	if err != nil {
		return errors.Errorf("failed to sign valset header hash: %w", err)
	}

	extendedSignature := entity.SignatureExtended{
		MessageHash: hash,
		Signature:   signature,
		PublicKey:   public.Raw(),
	}

	if err := s.cfg.Repo.SaveSignature(ctx, req.Hash(), public.Raw(), extendedSignature); err != nil {
		return errors.Errorf("failed to save signature: %w", err)
	}

	slog.InfoContext(ctx, "Message signed, sending via p2p", "hash", hash, "signature", signature)

	err = s.cfg.P2PService.BroadcastSignatureGeneratedMessage(ctx, entity.SignatureMessage{
		RequestHash: req.Hash(),
		KeyTag:      req.KeyTag,
		Epoch:       req.RequiredEpoch,
		Signature:   extendedSignature,
	})
	if err != nil {
		return errors.Errorf("failed to broadcast signature: %w", err)
	}

	if err := s.cfg.Repo.SaveSignatureRequest(ctx, req); err != nil {
		return errors.Errorf("failed to save signature request: %w", err)
	}

	return nil
}
