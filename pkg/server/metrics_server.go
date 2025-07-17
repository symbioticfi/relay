package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsConfig struct {
	Address           string
	ReadHeaderTimeout time.Duration
}

func (c MetricsConfig) Validate() error {
	return validator.New().Struct(c)
}

type MetricsServer struct {
	srv *http.Server
	cfg MetricsConfig
}

func NewMetricsServer(cfg MetricsConfig) (*MetricsServer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate metrics server config: %w", err)
	}

	return &MetricsServer{
		cfg: cfg,
		srv: &http.Server{
			Addr:              cfg.Address,
			Handler:           initMetricsHandler(cfg),
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		},
	}, nil
}

func initMetricsHandler(_ MetricsConfig) http.Handler {
	r := chi.NewRouter()
	r.Handle("/metrics", promhttp.Handler())

	return r
}

func (s *MetricsServer) Serve(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		ctxShutdown, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err := s.srv.Shutdown(ctxShutdown); err != nil { //nolint:contextcheck // we must use separate context for shutdown
			slog.WarnContext(ctx, "Failed to shutdown metrics server", "error", err)
		}
	}()

	slog.InfoContext(ctx, "Server started", "address", s.cfg.Address)

	if err := s.srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return errors.Errorf("failed to listen and serve: %w", err)
	}

	slog.InfoContext(ctx, "Metrics server stopped")

	return nil
}
