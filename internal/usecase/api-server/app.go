package api_server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/internal/usecase/metrics"
	"github.com/symbioticfi/relay/pkg/log"
	"github.com/symbioticfi/relay/pkg/server"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
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

type Config struct {
	Address           string        `validate:"required"`
	ReadHeaderTimeout time.Duration `validate:"required,gt=0"`
	ShutdownTimeout   time.Duration `validate:"required,gt=0"`

	Signer         signer    `validate:"required"`
	Repo           repo      `validate:"required"`
	EvmClient      evmClient `validate:"required"`
	Deriver        deriver   `validate:"required"`
	Aggregator     aggregator
	ServeMetrics   bool
	ServePprof     bool
	Metrics        *metrics.Metrics `validate:"required"`
	VerboseLogging bool
}

func (c Config) Validate() error {
	if err := validator.New().Struct(c); err != nil {
		return errors.Errorf("invalid config: %w", err)
	}
	return nil
}

// broadcasterHandler manages subscriptions grouped by request ID for O(1) lookup
type broadcasterHandler struct {
	lock        sync.RWMutex
	subscribers map[string][]chan symbiotic.AggregationProof // map[requestID][]*subscriber
}

// newBroadcasterHandler creates a new broadcaster handler
func newBroadcasterHandler() *broadcasterHandler {
	return &broadcasterHandler{
		subscribers: make(map[string][]chan symbiotic.AggregationProof),
	}
}

// Subscribe registers a new subscriber for a specific request ID
// Returns an unsubscribe function that should be called to clean up the subscription
func (b *broadcasterHandler) Subscribe(requestID string, ch chan symbiotic.AggregationProof) func() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.subscribers[requestID] = append(b.subscribers[requestID], ch)

	// Return unsubscribe function
	return func() {
		b.lock.Lock()
		defer b.lock.Unlock()

		subs := b.subscribers[requestID]
		// Find and remove this subscriber
		for i, s := range subs {
			if s == ch {
				// Remove by replacing with last element and truncating
				subs[i] = subs[len(subs)-1]
				b.subscribers[requestID] = subs[:len(subs)-1]

				// Clean up empty slice to avoid memory leaks
				if len(b.subscribers[requestID]) == 0 {
					delete(b.subscribers, requestID)
				}
				break
			}
		}
	}
}

// Notify broadcasts an event to all subscribers for a specific request ID - O(1) lookup
func (b *broadcasterHandler) Notify(proof symbiotic.AggregationProof) {
	requestID := proof.RequestID().Hex()

	b.lock.RLock()
	subs := b.subscribers[requestID]
	if len(subs) == 0 {
		b.lock.RUnlock()
		return
	}

	// Make a copy to avoid holding lock during notification
	subscribers := make([]chan symbiotic.AggregationProof, len(subs))
	copy(subscribers, subs)
	b.lock.RUnlock()

	// Broadcast to all subscribers for this request ID
	for _, subCh := range subscribers {
		// Non-blocking send to avoid blocking the signal handler
		select {
		case subCh <- proof:
			// Successfully sent
		default:
			// Channel is full or closed, skip
		}
	}
}

// grpcHandler implements the gRPC service interface
type grpcHandler struct {
	apiv1.SymbioticAPIServiceServer

	cfg         Config
	broadcaster *broadcasterHandler
}

// handleProofAggregated processes aggregation proof events from the signal
func (h *grpcHandler) handleProofAggregated(_ context.Context, proof symbiotic.AggregationProof) error {
	h.broadcaster.Notify(proof)
	return nil
}

type SymbioticServer struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	listener   net.Listener
	cfg        Config
	handler    *grpcHandler
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
		cfg:         cfg,
		broadcaster: newBroadcasterHandler(),
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
		grpcServer: grpcServer,
		httpServer: httpServer,
		listener:   listener,
		cfg:        cfg,
		handler:    handler,
	}, nil
}

// HandleProofAggregated returns the handler function for aggregation proof events
func (a *SymbioticServer) HandleProofAggregated() func(context.Context, symbiotic.AggregationProof) error {
	return a.handler.handleProofAggregated
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
		"metrics_enabled", a.cfg.ServeMetrics,
		"pprof_enabled", a.cfg.ServePprof)

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
