package api_server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/symbioticfi/relay/internal/usecase/broadcaster"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/internal/usecase/metrics"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/server"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

//go:generate mockgen -source=app.go -destination=mocks/app_mock.go -package=mocks
type signer interface {
	RequestSignature(ctx context.Context, req symbiotic.SignatureRequest) (common.Hash, error)
}

type repo interface {
	GetAggregationProof(ctx context.Context, requestID common.Hash) (symbiotic.AggregationProof, error)
	GetValidatorSetByEpoch(_ context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSet, error)
	GetAllSignatures(ctx context.Context, requestID common.Hash) ([]symbiotic.Signature, error)
	GetSignatureRequest(ctx context.Context, requestID common.Hash) (symbiotic.SignatureRequest, error)
	GetLatestValidatorSetHeader(_ context.Context) (symbiotic.ValidatorSetHeader, error)
	GetLatestValidatorSetEpoch(_ context.Context) (symbiotic.Epoch, error)
	GetValidatorSetMetadata(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.ValidatorSetMetadata, error)
	GetSignaturesStartingFromEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]symbiotic.Signature, error)
	GetSignaturesByEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]symbiotic.Signature, error)
	GetAggregationProofsStartingFromEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]symbiotic.AggregationProof, error)
	GetAggregationProofsByEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]symbiotic.AggregationProof, error)
	GetValidatorSetsStartingFromEpoch(ctx context.Context, epoch symbiotic.Epoch) ([]symbiotic.ValidatorSet, error)
}
type evmClient interface {
	GetCurrentEpoch(ctx context.Context) (symbiotic.Epoch, error)
	GetEpochStart(ctx context.Context, epoch symbiotic.Epoch) (symbiotic.Timestamp, error)
	GetConfig(ctx context.Context, timestamp symbiotic.Timestamp, epoch symbiotic.Epoch) (symbiotic.NetworkConfig, error)
	GetLastCommittedHeaderEpoch(ctx context.Context, addr symbiotic.CrossChainAddress) (_ symbiotic.Epoch, err error)
}

type aggregator interface {
	GetAggregationStatus(ctx context.Context, requestID common.Hash) (symbiotic.AggregationStatus, error)
}

type deriver interface {
	GetValidatorSet(ctx context.Context, epoch symbiotic.Epoch, config symbiotic.NetworkConfig) (symbiotic.ValidatorSet, error)
}

type keyProvider interface {
	GetOnchainKeyFromCache(keyTag symbiotic.KeyTag) (symbiotic.CompactPublicKey, error)
}

type Config struct {
	Address           string        `validate:"required"`
	ReadHeaderTimeout time.Duration `validate:"required,gt=0"`
	ShutdownTimeout   time.Duration `validate:"required,gt=0"`

	Signer                 signer      `validate:"required"`
	Repo                   repo        `validate:"required"`
	EvmClient              evmClient   `validate:"required"`
	Deriver                deriver     `validate:"required"`
	KeyProvider            keyProvider `validate:"required"`
	Aggregator             aggregator
	ServeMetrics           bool
	ServePprof             bool
	ServeHTTPGateway       bool
	Metrics                *metrics.Metrics `validate:"required"`
	VerboseLogging         bool
	MaxAllowedStreamsCount int `validate:"required,gt=0"`
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

	proofsHub        *broadcaster.Hub[symbiotic.AggregationProof]
	signatureHub     *broadcaster.Hub[symbiotic.Signature]
	validatorSetsHub *broadcaster.Hub[symbiotic.ValidatorSet]
}
type SymbioticServer struct {
	grpcServer       *grpc.Server
	httpServer       *http.Server
	listener         net.Listener
	cfg              Config
	handler          *grpcHandler
	startGatewayFunc func() error
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
			server.LoggingInterceptor(cfg.VerboseLogging),
			ErrorHandlingInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			server.StreamPanicRecoveryInterceptor(),
			cfg.Metrics.StreamServerInterceptor(),
			//nolint:contextcheck // the context comes from th stream
			server.StreamLoggingInterceptor(cfg.VerboseLogging),
			StreamErrorHandlingInterceptor(),
		),
	)

	// Create and register the handler
	handler := &grpcHandler{
		cfg: cfg,
		proofsHub: broadcaster.NewHub[symbiotic.AggregationProof](
			broadcaster.WithBufferSize[symbiotic.AggregationProof](cfg.MaxAllowedStreamsCount),
		),
		signatureHub: broadcaster.NewHub[symbiotic.Signature](
			broadcaster.WithBufferSize[symbiotic.Signature](cfg.MaxAllowedStreamsCount),
		),
		validatorSetsHub: broadcaster.NewHub[symbiotic.ValidatorSet](
			broadcaster.WithBufferSize[symbiotic.ValidatorSet](cfg.MaxAllowedStreamsCount),
		),
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

	// Register HTTP gateway if enabled
	var startGatewayFunc func() error
	if cfg.ServeHTTPGateway {
		startGatewayFunc = setupHttpProxy(ctx, cfg.Address, httpMux)
	}

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

	// Debug pprof endpoints if enabled
	if cfg.ServePprof {
		httpMux.HandleFunc("/pprof", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, r.RequestURI+"/", http.StatusMovedPermanently)
		})
		httpMux.HandleFunc("/pprof/", pprof.Index)
		httpMux.HandleFunc("/pprof/cmdline", pprof.Cmdline)
		httpMux.HandleFunc("/pprof/profile", pprof.Profile)
		httpMux.HandleFunc("/pprof/symbol", pprof.Symbol)
		httpMux.HandleFunc("/pprof/trace", pprof.Trace)
		httpMux.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
		httpMux.Handle("/pprof/threadcreate", pprof.Handler("threadcreate"))
		httpMux.Handle("/pprof/mutex", pprof.Handler("mutex"))
		httpMux.Handle("/pprof/heap", pprof.Handler("heap"))
		httpMux.Handle("/pprof/block", pprof.Handler("block"))
		httpMux.Handle("/pprof/allocs", pprof.Handler("allocs"))
		slog.InfoContext(ctx, "Pprof debug endpoints enabled", "path", "/pprof/")
	}

	// Serve API documentation
	docFS := http.FileServer(http.Dir("docs/api/v1"))
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
		grpcServer:       grpcServer,
		httpServer:       httpServer,
		listener:         listener,
		cfg:              cfg,
		handler:          handler,
		startGatewayFunc: startGatewayFunc,
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
		"http_gateway_enabled", a.cfg.ServeHTTPGateway,
		"http_gateway_path", "/api/v1/",
		"docs_path", "/docs/",
		"metrics_path", "/metrics",
		"metrics_enabled", a.cfg.ServeMetrics,
		"pprof_enabled", a.cfg.ServePprof)

	// Start serving in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := a.httpServer.Serve(a.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- errors.Errorf("failed to serve HTTP/gRPC multiplexed server: %w", err)
		}
	}()

	// Initialize HTTP gateway connection after server starts
	if a.startGatewayFunc != nil {
		// Retry connection to gRPC server with exponential backoff
		const maxRetries = 5
		var lastErr error

		for i := 0; i < maxRetries; i++ {
			// Wait before attempting (exponential backoff)
			if i > 0 {
				backoff := time.Duration(50*(1<<uint(i-1))) * time.Millisecond
				slog.DebugContext(logCtx, "Retrying HTTP gateway connection",
					"attempt", i+1,
					"max_retries", maxRetries,
					"backoff", backoff)
				time.Sleep(backoff)
			}

			if err := a.startGatewayFunc(); err != nil {
				lastErr = err
				slog.WarnContext(logCtx, "Failed to connect HTTP gateway to gRPC server",
					"attempt", i+1,
					"error", err)
				continue
			}

			// Success
			lastErr = nil
			break
		}

		if lastErr != nil {
			return errors.Errorf("failed to start HTTP gateway after %d attempts: %w", maxRetries, lastErr)
		}
	}

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

func (a *SymbioticServer) HandleProofAggregated() func(context.Context, symbiotic.AggregationProof) error {
	return func(ctx context.Context, proof symbiotic.AggregationProof) error {
		a.handler.proofsHub.Broadcast(proof)
		return nil
	}
}

func (a *SymbioticServer) HandleSignatureProcessed() func(context.Context, symbiotic.Signature) error {
	return func(ctx context.Context, signature symbiotic.Signature) error {
		a.handler.signatureHub.Broadcast(signature)
		return nil
	}
}

func (a *SymbioticServer) HandleValidatorSet() func(context.Context, symbiotic.ValidatorSet) error {
	return func(ctx context.Context, validatorSet symbiotic.ValidatorSet) error {
		a.handler.validatorSetsHub.Broadcast(validatorSet)
		return nil
	}
}
