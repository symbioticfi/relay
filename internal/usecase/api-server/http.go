package api_server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-errors/errors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// sseResponseWriter wraps http.ResponseWriter to convert newline-delimited JSON to SSE format
type sseResponseWriter struct {
	http.ResponseWriter

	flusher        http.Flusher
	headersWritten bool
}

// WriteHeader intercepts header writes to set SSE headers
func (s *sseResponseWriter) WriteHeader(statusCode int) {
	if !s.headersWritten {
		// Override content-type for SSE
		s.ResponseWriter.Header().Set("Content-Type", "text/event-stream")
		s.ResponseWriter.Header().Set("Cache-Control", "no-cache")
		s.ResponseWriter.Header().Set("Connection", "keep-alive")
		s.ResponseWriter.Header().Set("X-Accel-Buffering", "no")
		s.headersWritten = true
	}
	s.ResponseWriter.WriteHeader(statusCode)
}

func (s *sseResponseWriter) Write(b []byte) (int, error) {
	// Ensure headers are written
	if !s.headersWritten {
		s.WriteHeader(http.StatusOK)
	}

	// Convert each line to SSE format: data: {json}\n\n
	lines := strings.Split(string(b), "\n")
	totalWritten := 0

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Write in SSE format
		sseData := fmt.Sprintf("data: %s\n\n", line)
		n, err := s.ResponseWriter.Write([]byte(sseData))
		totalWritten += n
		if err != nil {
			return totalWritten, err
		}
		s.flusher.Flush()
	}

	return len(b), nil
}

func (s *sseResponseWriter) Flush() {
	s.flusher.Flush()
}

// setupHttpProxy configures the HTTP-to-gRPC gateway proxy
// Returns a start function that should be called after the gRPC server starts listening
func setupHttpProxy(ctx context.Context, grpcAddr string, httpMux *http.ServeMux) func() error {
	gwMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
	)

	// Create gRPC client connection to the actual gRPC server via TCP
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024), // 10MB
			grpc.MaxCallSendMsgSize(10*1024*1024), // 10MB
		),
	}

	var conn *grpc.ClientConn

	// Start function that will be called after gRPC server is listening
	startFn := func() error {
		var err error
		conn, err = grpc.NewClient(grpcAddr, opts...)
		if err != nil {
			return errors.Errorf("failed to create gRPC client for gateway: %w", err)
		}

		if err := apiv1.RegisterSymbioticAPIServiceHandler(ctx, gwMux, conn); err != nil {
			return errors.Errorf("failed to register gateway handler: %w", err)
		}

		slog.InfoContext(ctx, "HTTP Gateway connected to gRPC server",
			"grpc_address", grpcAddr,
			"gateway_path", "/api/v1/")
		return nil
	}

	// Mount the gateway under /api prefix with CORS and streaming support
	httpMux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		// Check if gateway is initialized
		if conn == nil {
			http.Error(w, "Gateway not ready", http.StatusServiceUnavailable)
			return
		}

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")

		// For streaming endpoints, wrap the response writer to intercept and convert to SSE
		if strings.HasPrefix(r.URL.Path, "/v1/stream/") {
			if flusher, ok := w.(http.Flusher); ok {
				// Wrap with SSE writer that will handle headers and format conversion
				w = &sseResponseWriter{ResponseWriter: w, flusher: flusher}
			}
		}

		gwMux.ServeHTTP(w, r)
	})

	return startFn
}
