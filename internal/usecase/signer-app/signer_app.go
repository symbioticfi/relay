package signer_app

import (
	"context"
	"log/slog"
	"time"

	"github.com/symbioticfi/relay/core/usecase/crypto"
	"github.com/symbioticfi/relay/pkg/log"

	"github.com/ethereum/go-ethereum/common"

	"github.com/go-errors/errors"
	validate "github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
)

//go:generate mockgen -source=signer_app.go -destination=mocks/signer_app.go -package=mocks

type repo interface {
	GetSignatureRequest(ctx context.Context, signatureTargetID common.Hash) (entity.SignatureRequest, error)
	GetAggregationProof(ctx context.Context, signatureTargetID common.Hash) (entity.AggregationProof, error)
	SaveAggregationProof(ctx context.Context, signatureTargetID common.Hash, ap entity.AggregationProof) error
	GetValidatorSetByEpoch(ctx context.Context, epoch uint64) (entity.ValidatorSet, error)
	GetValidatorByKey(ctx context.Context, epoch uint64, keyTag entity.KeyTag, publicKey []byte) (entity.Validator, uint32, error)
}

type p2pService interface {
	BroadcastSignatureGeneratedMessage(ctx context.Context, msg entity.SignatureExtended) error
}

type keyProvider interface {
	GetPrivateKey(keyTag entity.KeyTag) (crypto.PrivateKey, error)
}

type aggProofSignal interface {
	Emit(payload entity.AggregationProof) error
}

type aggregator interface {
	Verify(valset entity.ValidatorSet, keyTag entity.KeyTag, aggregationProof entity.AggregationProof) (bool, error)
}

type metrics interface {
	ObservePKSignDuration(d time.Duration)
	ObserveAppSignDuration(d time.Duration)
}

type entityProcessor interface {
	ProcessSignature(ctx context.Context, param entity.SaveSignatureParam) error
	ProcessAggregationProof(ctx context.Context, proof entity.AggregationProof) error
}

type Config struct {
	P2PService      p2pService      `validate:"required"`
	KeyProvider     keyProvider     `validate:"required"`
	Repo            repo            `validate:"required"`
	EntityProcessor entityProcessor `validate:"required"`
	AggProofSignal  aggProofSignal  `validate:"required"`
	Aggregator      aggregator      `validate:"required"`
	Metrics         metrics         `validate:"required"`
}

func (c Config) Validate() error {
	if err := validate.New().Struct(c); err != nil {
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

func (s *SignerApp) Sign(ctx context.Context, req entity.SignatureRequest) (entity.SignatureExtended, error) {
	ctx = log.WithComponent(ctx, "signer")
	ctx = log.WithAttrs(ctx, slog.Uint64("epoch", uint64(req.RequiredEpoch)))
	timeAppSignStart := time.Now()

	if !req.KeyTag.Type().SignerKey() {
		return entity.SignatureExtended{}, errors.Errorf("key tag %s is not a signing key", req.KeyTag)
	}

	private, err := s.cfg.KeyProvider.GetPrivateKey(req.KeyTag)
	if err != nil {
		return entity.SignatureExtended{}, errors.Errorf("failed to get private key: %w", err)
	}

	pkSignStart := time.Now()
	signature, hash, err := private.Sign(req.Message)
	if err != nil {
		return entity.SignatureExtended{}, errors.Errorf("failed to sign valset header hash: %w", err)
	}
	s.cfg.Metrics.ObservePKSignDuration(time.Since(pkSignStart))

	extendedSignature := entity.SignatureExtended{
		MessageHash: hash,
		KeyTag:      req.KeyTag,
		Epoch:       req.RequiredEpoch,
		PublicKey:   private.PublicKey().Raw(),
		Signature:   signature,
	}

	signatureTargetId := extendedSignature.SignatureTargetID()
	_, err = s.cfg.Repo.GetSignatureRequest(ctx, signatureTargetId)
	if err != nil && !errors.Is(err, entity.ErrEntityNotFound) {
		return entity.SignatureExtended{}, errors.Errorf("failed to get signature request: %w", err)
	}
	if entityFound := !errors.Is(err, entity.ErrEntityNotFound); entityFound {
		slog.DebugContext(ctx, "Signature request already exists", "request", req)
		return entity.SignatureExtended{}, errors.Errorf("signature request already exists: %w", entity.ErrEntityAlreadyExist)
	}

	param := entity.SaveSignatureParam{
		KeyTag:           req.KeyTag,
		Signature:        extendedSignature,
		Epoch:            req.RequiredEpoch,
		SignatureRequest: &req,
	}

	if err := s.cfg.EntityProcessor.ProcessSignature(ctx, param); err != nil {
		return entity.SignatureExtended{}, errors.Errorf("failed to process signature: %w", err)
	}

	err = s.cfg.P2PService.BroadcastSignatureGeneratedMessage(ctx, extendedSignature)
	if err != nil {
		return entity.SignatureExtended{}, errors.Errorf("failed to broadcast signature: %w", err)
	}

	slog.InfoContext(ctx, "Message signed",
		"hash", hash,
		"signature", signature,
		"key_tag", req.KeyTag,
		"duration", time.Since(timeAppSignStart),
	)
	s.cfg.Metrics.ObserveAppSignDuration(time.Since(timeAppSignStart))

	return extendedSignature, nil
}
