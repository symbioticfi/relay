package signer_app

import (
	"context"
	"log/slog"
	"math/big"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/samber/mo"

	"middleware-offchain/internal/entity"
)

type repo interface {
	GetSignatureRequest(ctx context.Context, req entity.SignatureRequest) (mo.Option[entity.SignatureRequest], error)
	GetAggregationProof(ctx context.Context, req entity.SignatureRequest) (mo.Option[entity.AggregationProof], error)
	GetLatestValsetExtra(ctx context.Context) (mo.Option[entity.ValidatorSetExtra], error)
	GetValsetExtraByEpoch(ctx context.Context, epoch *big.Int) (entity.ValidatorSetExtra, error)
	SaveSignature(ctx context.Context, req entity.SignatureRequest, sig entity.Signature) error
}

type p2pService interface {
	BroadcastSignatureGeneratedMessage(ctx context.Context, msg entity.SignatureHashMessage) error
}

type signer interface {
	Sign(keyTag entity.KeyTag, message []byte) (entity.Signature, error)
	Hash(keyTag entity.KeyTag, message []byte) ([]byte, error)
}

type keyProvider interface {
	GetPublic(keyTag entity.KeyTag) ([]byte, error)
}

type Config struct {
	P2PService  p2pService  `validate:"required"`
	Signer      signer      `validate:"required"`
	Repo        repo        `validate:"required"`
	KeyProvider keyProvider `validate:"required"`
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

	return &SignerApp{
		cfg: cfg,
	}, nil
}

func (s *SignerApp) Sign(ctx context.Context, req entity.SignatureRequest) error {
	existed, err := s.cfg.Repo.GetSignatureRequest(ctx, req)
	if err != nil {
		return errors.Errorf("failed to get signature request: %w", err)
	}
	if existed.IsPresent() {
		return errors.New(entity.ErrSignatureRequestExists)
	}

	existedProof, err := s.cfg.Repo.GetAggregationProof(ctx, req)
	if err != nil {
		return errors.Errorf("failed to get aggregation proof: %w", err)
	}
	if existedProof.IsPresent() {
		return errors.New("aggregation proof already exists for this request")
	}

	latestValsetExtra, err := s.cfg.Repo.GetLatestValsetExtra(ctx)
	if err != nil {
		return errors.Errorf("failed to get latest valset extra: %w", err)
	}
	if !latestValsetExtra.IsPresent() {
		return errors.New("no latest valset extra found")
	}

	if !isRecentEpoch(latestValsetExtra.MustGet().Epoch, req.RequiredEpoch) {
		return errors.Errorf("epoch difference is too large: max allowed: %d", entity.MaxSavedEpochs)
	}

	epochValsetExtra, err := s.cfg.Repo.GetValsetExtraByEpoch(ctx, req.RequiredEpoch)
	if err != nil {
		return errors.Errorf("failed to get valset extra by epoch %s: %w", req.RequiredEpoch, err)
	}
	epochValset := epochValsetExtra.MakeValidatorSet()

	public, err := s.cfg.KeyProvider.GetPublic(req.KeyTag)
	if err != nil {
		return errors.Errorf("failed to get public key for key tag %d: %w", req.KeyTag, err)
	}
	_, found := epochValset.FindValidatorByKey(req.KeyTag, public)
	if !found {
		return errors.Errorf("validator not found in epoch valset for public key")
	}

	signature, err := s.cfg.Signer.Sign(req.KeyTag, req.Message)
	if err != nil {
		return errors.Errorf("failed to sign valset header hash: %w", err)
	}

	if err := s.cfg.Repo.SaveSignature(ctx, req, signature); err != nil {
		return errors.Errorf("failed to save signature: %w", err)
	}

	slog.InfoContext(ctx, "valset header hash signed, sending via p2p", "headerSignature", signature)

	err = s.cfg.P2PService.BroadcastSignatureGeneratedMessage(ctx, entity.SignatureHashMessage{
		MessageHash: req.Message,
		KeyTag:      req.KeyTag,
		Signature:   signature.Signature,
		PublicKey:   signature.PublicKey,
		HashType:    entity.HashTypeValsetHeader,
		Epoch:       req.RequiredEpoch,
	})
	if err != nil {
		return errors.Errorf("failed to broadcast valset header: %w", err)
	}

	return nil
}

func isRecentEpoch(latestValsetEpoch, requiredEpoch *big.Int) bool {
	diffEpochs := new(big.Int).Sub(latestValsetEpoch, requiredEpoch)
	return diffEpochs.Cmp(new(big.Int).SetInt64(entity.MaxSavedEpochs)) > 0
}
