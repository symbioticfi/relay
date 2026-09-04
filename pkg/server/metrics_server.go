package server

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsConfig struct {
	Address           string
	ReadHeaderTimeout time.Duration
	ServePprof        bool
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

func initMetricsHandler(cfg MetricsConfig) http.Handler {
	r := chi.NewRouter()
	r.Handle("/metrics", promhttp.Handler())
	if cfg.ServePprof {
		r.HandleFunc("/pprof", func(w http.ResponseWriter, req *http.Request) {
			http.Redirect(w, req, req.RequestURI+"/", http.StatusMovedPermanently)
		})
		r.HandleFunc("/pprof/", pprof.Index)
		r.HandleFunc("/pprof/cmdline", pprof.Cmdline)
		r.HandleFunc("/pprof/profile", pprof.Profile)
		r.HandleFunc("/pprof/symbol", pprof.Symbol)
		r.HandleFunc("/pprof/trace", pprof.Trace)
		r.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
		r.Handle("/pprof/threadcreate", pprof.Handler("threadcreate"))
		r.Handle("/pprof/mutex", pprof.Handler("mutex"))
		r.Handle("/pprof/heap", pprof.Handler("heap"))
		r.Handle("/pprof/block", pprof.Handler("block"))
		r.Handle("/pprof/allocs", pprof.Handler("allocs"))
	}

	return r
}

func (s *MetricsServer) Serve(ctx context.Context) error {
	go func() { //nolint:gosec // we must use separate context for shutdown
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
