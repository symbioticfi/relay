package signer_app

import (
	"context"
	"log/slog"
	"time"

	"github.com/symbioticfi/relay/core/usecase/crypto"
	"github.com/symbioticfi/relay/pkg/log"

	"github.com/ethereum/go-ethereum/common"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbioticfi/relay/core/entity"
)

//go:generate mockgen -source=signer_app.go -destination=mocks/signer_app.go -package=mocks

type repo interface {
	GetSignatureRequest(ctx context.Context, reqHash common.Hash) (entity.SignatureRequest, error)
	GetAggregationProof(ctx context.Context, reqHash common.Hash) (entity.AggregationProof, error)
	SaveAggregationProof(ctx context.Context, reqHash common.Hash, ap entity.AggregationProof) error
	GetValidatorSetByEpoch(ctx context.Context, epoch uint64) (entity.ValidatorSet, error)
	GetValidatorByKey(ctx context.Context, epoch uint64, keyTag entity.KeyTag, publicKey []byte) (entity.Validator, int, error)
	SaveSignature(ctx context.Context, reqHash common.Hash, key entity.RawPublicKey, sig entity.SignatureExtended) error
	SaveSignatureRequest(_ context.Context, req entity.SignatureRequest) error
	UpdateSignatureStat(_ context.Context, reqHash common.Hash, s entity.SignatureStatStage, t time.Time) (entity.SignatureStat, error)
}

type p2pService interface {
	BroadcastSignatureGeneratedMessage(ctx context.Context, msg entity.SignatureMessage) error
}

type keyProvider interface {
	GetPrivateKey(keyTag entity.KeyTag) (crypto.PrivateKey, error)
}

type aggProofSignal interface {
	Emit(ctx context.Context, payload entity.AggregatedSignatureMessage) error
}

type aggregator interface {
	Verify(valset entity.ValidatorSet, keyTag entity.KeyTag, aggregationProof entity.AggregationProof) (bool, error)
}

type metrics interface {
	ObservePKSignDuration(d time.Duration)
	ObserveAppSignDuration(d time.Duration)
	ObserveAggReceived(stat entity.SignatureStat)
}

type Config struct {
	P2PService     p2pService     `validate:"required"`
	KeyProvider    keyProvider    `validate:"required"`
	Repo           repo           `validate:"required"`
	AggProofSignal aggProofSignal `validate:"required"`
	Aggregator     aggregator     `validate:"required"`
	Metrics        metrics        `validate:"required"`
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
	timeAppSignStart := time.Now()

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
		slog.DebugContext(ctx, "Aggregation proof already exists", "request", req)
		return nil
	}

	if _, err := s.cfg.Repo.UpdateSignatureStat(ctx, req.Hash(), entity.SignatureStatStageSignRequestReceived, timeAppSignStart); err != nil {
		slog.WarnContext(ctx, "Failed to update signature stat", "error", err)
	}

	private, err := s.cfg.KeyProvider.GetPrivateKey(req.KeyTag)
	if err != nil {
		return errors.Errorf("failed to get private key: %w", err)
	}

	public := private.PublicKey()
	_, _, err = s.cfg.Repo.GetValidatorByKey(ctx, uint64(req.RequiredEpoch), req.KeyTag, public.OnChain())
	if err != nil {
		return errors.Errorf("validator not found in epoch valset for public key: %w", err)
	}

	pkSignStart := time.Now()
	signature, hash, err := private.Sign(req.Message)
	if err != nil {
		return errors.Errorf("failed to sign valset header hash: %w", err)
	}
	s.cfg.Metrics.ObservePKSignDuration(time.Since(pkSignStart))

	extendedSignature := entity.SignatureExtended{
		MessageHash: hash,
		Signature:   signature,
		PublicKey:   public.Raw(),
	}

	if err := s.cfg.Repo.SaveSignature(ctx, req.Hash(), public.Raw(), extendedSignature); err != nil {
		return errors.Errorf("failed to save signature: %w", err)
	}

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

	slog.InfoContext(ctx, "Message signed", "hash", hash, "signature", signature, "duration", time.Since(timeAppSignStart))
	s.cfg.Metrics.ObserveAppSignDuration(time.Since(timeAppSignStart))
	if _, err := s.cfg.Repo.UpdateSignatureStat(ctx, req.Hash(), entity.SignatureStatStageSignCompleted, time.Now()); err != nil {
		slog.WarnContext(ctx, "Failed to update signature stat", "error", err)
	}

	return nil
}
