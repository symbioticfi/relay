package apiApp

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	swag "middleware-offchain/api"
	"middleware-offchain/core/entity"
	p2pEntity "middleware-offchain/internal/entity"
	"middleware-offchain/internal/gen/api"
	"middleware-offchain/pkg/server"
)

type signer interface {
	Sign(ctx context.Context, req entity.SignatureRequest) error
}

type repo interface {
	GetAggregationProof(ctx context.Context, reqHash common.Hash) (entity.AggregationProof, error)
	GetValidatorSetByEpoch(_ context.Context, epoch uint64) (entity.ValidatorSet, error)
	GetAllSignatures(_ context.Context, reqHash common.Hash) ([]entity.SignatureExtended, error)
	GetSignatureRequest(_ context.Context, reqHash common.Hash) (entity.SignatureRequest, error)
}

type evmClient interface {
	GetCurrentEpoch(ctx context.Context) (uint64, error)
	GetEpochStart(ctx context.Context, epoch uint64) (uint64, error)
}

type aggregator interface {
	GetAggregationStatus(ctx context.Context, requestHash common.Hash) (p2pEntity.AggregationStatus, error)
}

type Config struct {
	Address           string        `validate:"required"`
	Prefix            string        `validate:"required"`
	ReadHeaderTimeout time.Duration `validate:"required,gt=0"`
	ShutdownTimeout   time.Duration `validate:"required,gt=0"`

	Signer     signer    `validate:"required"`
	Repo       repo      `validate:"required"`
	EVMClient  evmClient `validate:"required"`
	Aggregator aggregator
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("invalid config: %w", err)
	}
	return nil
}

type APIApp struct {
	srv *server.Server
}

func NewAPIApp(cfg Config) (*APIApp, error) {
	if err := validator.New().Struct(cfg); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	h := &handler{
		cfg: cfg,
	}

	opts := []api.ServerOption{
		api.WithPathPrefix(cfg.Prefix),
		api.WithErrorHandler(errorHandler),
	}
	apiServer, err := api.NewServer(h, opts...)
	if err != nil {
		return nil, errors.Errorf("failed to init db: %w", err)
	}

	srv, err := server.New(server.Config{
		Address:           cfg.Address,
		Prefix:            cfg.Prefix,
		APIHandler:        apiServer,
		SwaggerHandler:    swag.OapiSchemaHandler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ShutdownTimeout:   cfg.ShutdownTimeout,
	})
	if err != nil {
		return nil, errors.Errorf("failed to create server: %w", err)
	}

	return &APIApp{
		srv: srv,
	}, nil
}

func (a APIApp) Start(ctx context.Context) error {
	return a.srv.Serve(ctx)
}
