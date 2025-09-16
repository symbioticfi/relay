package api_server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	keyprovider "github.com/symbioticfi/relay/core/usecase/key-provider"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/symbioticfi/relay/core/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/internal/usecase/metrics"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/server"
)

//go:generate mockgen -source=app.go -destination=mocks/app_mock.go -package=mocks
type signer interface {
	Sign(ctx context.Context, req entity.SignatureRequest) error
}

type repo interface {
	GetAggregationProof(ctx context.Context, reqHash common.Hash) (entity.AggregationProof, error)
	GetValidatorSetByEpoch(_ context.Context, epoch uint64) (entity.ValidatorSet, error)
	GetAllSignatures(_ context.Context, reqHash common.Hash) ([]entity.SignatureExtended, error)
	GetSignatureRequest(_ context.Context, reqHash common.Hash) (entity.SignatureRequest, error)
	GetLatestValidatorSetHeader(_ context.Context) (entity.ValidatorSetHeader, error)
	GetLatestValidatorSetEpoch(_ context.Context) (uint64, error)
}

type evmClient interface {
	GetCurrentEpoch(ctx context.Context) (uint64, error)
	GetEpochStart(ctx context.Context, epoch uint64) (uint64, error)
	GetConfig(ctx context.Context, timestamp uint64) (entity.NetworkConfig, error)
}

type aggregator interface {
	GetAggregationStatus(ctx context.Context, requestHash common.Hash) (entity.AggregationStatus, error)
}

type deriver interface {
	GetValidatorSet(ctx context.Context, epoch uint64, config entity.NetworkConfig) (entity.ValidatorSet, error)
}

type Config struct {
	Address           string        `validate:"required"`
	ReadHeaderTimeout time.Duration `validate:"required,gt=0"`
	ShutdownTimeout   time.Duration `validate:"required,gt=0"`

	Signer       signer    `validate:"required"`
	Repo         repo      `validate:"required"`
	EvmClient    evmClient `validate:"required"`
	Deriver      deriver   `validate:"required"`
	Aggregator   aggregator
	ServeMetrics bool
	Metrics      *metrics.Metrics `validate:"required"`
	KeyProvider  keyprovider.KeyProvider
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("invalid config: %w", err)
	}
	return nil
}

// grpcHandler implements the gRPC service interface
type grpcHandler struct {
	apiv1.SymbioticAPIServiceServer

	cfg Config
}

type SymbioticServer struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	listener   net.Listener
	cfg        Config
}

func NewSymbioticServer(ctx context.Context, cfg Config) (*SymbioticServer, error) {
	if err := validator.New().Struct(cfg); err != nil {
		return nil, errors.Errorf("failed to validate config: %w", err)
	}

	// Create listener
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", cfg.Address)
	if err != nil {
		return nil, errors.Errorf("failed to listen on %s: %w", cfg.Address, err)
	}

	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			server.PanicRecoveryInterceptor(),
			cfg.Metrics.UnaryServerInterceptor(),
			server.LoggingInterceptor(),
			ErrorHandlingInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			server.StreamPanicRecoveryInterceptor(),
			cfg.Metrics.StreamServerInterceptor(),
			server.StreamLoggingInterceptor(),
			StreamErrorHandlingInterceptor(),
		),
	)

	// Create and register the handler
	handler := &grpcHandler{
		cfg: cfg,
	}

	apiv1.RegisterSymbioticAPIServiceServer(grpcServer, handler)

	// Register health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection service for development
	reflection.Register(grpcServer)

	// Create HTTP server for documentation with panic recovery
	httpMux := http.NewServeMux()

	// Wrap the entire mux with panic recovery
	recoveredMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(ctx context.Context) {
			if err := recover(); err != nil {
				slog.ErrorContext(ctx, "HTTP handler panic recovered",
					"error", err,
					"path", r.URL.Path,
					"method", r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Internal server error","status":500}`))
			}
		}(r.Context())
		httpMux.ServeHTTP(w, r)
	})

	// Root redirect to docs
	httpMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/docs/", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})

	// Health check endpoint
	httpMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Serve API documentation
	docFS := http.FileServer(http.Dir("api/docs/v1"))
	httpMux.Handle("/docs/", http.StripPrefix("/docs/", docFS))

	// Serve metrics endpoint if enabled
	if cfg.ServeMetrics {
		httpMux.Handle("/metrics", promhttp.Handler())
		slog.InfoContext(ctx, "Metrics endpoint enabled", "path", "/metrics")
	}

	// Create HTTP/2 server that can handle both HTTP and gRPC
	httpServer := &http.Server{
		Handler:           createMuxHandler(grpcServer, recoveredMux),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}

	return &SymbioticServer{
		grpcServer: grpcServer,
		httpServer: httpServer,
		listener:   listener,
		cfg:        cfg,
	}, nil
}

// createMuxHandler creates a handler that multiplexes between gRPC and HTTP
func createMuxHandler(grpcServer *grpc.Server, httpHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a gRPC request
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			// Handle gRPC request
			grpcServer.ServeHTTP(w, r)
			return
		}

		// Handle as HTTP request (documentation, health checks, etc.)
		httpHandler.ServeHTTP(w, r)
	}), &http2.Server{})
}

func (a *SymbioticServer) Start(ctx context.Context) error {
	logCtx := log.WithComponent(ctx, "api")

	slog.InfoContext(logCtx, "Starting gRPC/HTTP multiplexed server",
		"grpc_address", a.cfg.Address,
		"docs_path", "/docs/",
		"metrics_path", "/metrics",
		"metrics_enabled", a.cfg.ServeMetrics)

	// Start serving in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := a.httpServer.Serve(a.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- errors.Errorf("failed to serve HTTP/gRPC multiplexed server: %w", err)
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		slog.InfoContext(logCtx, "Shutting down gRPC/HTTP server...")

		// Graceful shutdown with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
		defer cancel()

		// Shutdown HTTP server
		//nolint:contextcheck // we need to use background context here as the original context is already cancelled
		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			slog.WarnContext(logCtx, "Failed to shutdown HTTP server gracefully", "error", err)

			// Force stop gRPC server
			a.grpcServer.Stop()
		} else {
			// Graceful stop for gRPC server
			done := make(chan struct{})
			go func() {
				a.grpcServer.GracefulStop()
				close(done)
			}()

			select {
			case <-shutdownCtx.Done():
				slog.WarnContext(logCtx, "Force stopping gRPC server due to timeout")
				a.grpcServer.Stop()
			case <-done:
				slog.InfoContext(logCtx, "gRPC server stopped gracefully")
			}
		}

		slog.InfoContext(logCtx, "Server shutdown complete")
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}
