package metrics

import (
	"context"
	"time"

	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"

	"github.com/symbiotic/relay/pkg/log"
	"github.com/symbiotic/relay/pkg/server"
)

type AppConfig struct {
	Address           string        `validate:"required"`
	ReadHeaderTimeout time.Duration `validate:"required,gt=0"`
}

func (c AppConfig) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("invalid metrics app config: %w", err)
	}

	return nil
}

type App struct {
	srv *server.MetricsServer
}

func NewApp(cfg AppConfig) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	srv, err := server.NewMetricsServer(server.MetricsConfig{
		Address:           cfg.Address,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	})
	if err != nil {
		return nil, errors.Errorf("failed to create metrics server: %w", err)
	}

	return &App{
		srv: srv,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	return a.srv.Serve(log.WithComponent(ctx, "metrics-api"))
}
