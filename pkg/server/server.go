package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	metricsMw "github.com/slok/go-http-metrics/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Config struct {
	Address           string        `validate:"required"`
	ReadHeaderTimeout time.Duration `validate:"required,gt=0"`
	ShutdownTimeout   time.Duration `validate:"required,gt=0"`
	Prefix            string        `validate:"required"`
	APIHandler        http.Handler  `validate:"required"`
	MetricsRegistry   prometheus.Registerer
	SwaggerHandler    http.HandlerFunc
}

func (c Config) Validate() error {
	return validator.New().Struct(c)
}

type Server struct {
	cfg Config
	srv *http.Server
}

func New(cfg Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	return &Server{
		cfg: cfg,
		srv: &http.Server{
			Addr:              cfg.Address,
			Handler:           initHandler(cfg),
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		},
	}, nil
}

func initHandler(cfg Config) http.Handler {
	r := chi.NewRouter()
	r.Use(
		middleware.Recoverer,
		handlerProvider("", metricsMw.New(metricsMw.Config{ //nolint:exhaustruct
			Recorder: metrics.NewRecorder(metrics.Config{ //nolint:exhaustruct
				Registry:        cfg.MetricsRegistry,
				DurationBuckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 5, 10, 15, 20, 40, 45, 60},
			}),
		})),
	)

	r.Handle("/metrics", promhttp.Handler())

	if cfg.SwaggerHandler != nil {
		r.Get("/swagger.yaml", cfg.SwaggerHandler)
		r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, r.RequestURI+"/index.html", http.StatusMovedPermanently)
		})
		r.Get("/docs/*", httpSwagger.Handler(
			httpSwagger.URL("/swagger.yaml"),
		))
	}

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequestID)
		r.Use(RequestToSlog)
		r.Use(middleware.RequestLogger(logFormatter{}))
		r.Mount(cfg.Prefix, cfg.APIHandler)
	})

	return r
}

func (s *Server) Serve(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		ctxShutdown, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()

		if err := s.srv.Shutdown(ctxShutdown); err != nil { //nolint:contextcheck // we must use separate context for shutdown
			slog.WarnContext(ctx, "Failed to shutdown server", "error", err)
		}
	}()

	slog.InfoContext(ctx, "Server started", "address", s.cfg.Address)

	if err := s.srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return errors.Errorf("failed to listen and serve: %w", err)
	}

	slog.InfoContext(ctx, "Server stopped")

	return nil
}
