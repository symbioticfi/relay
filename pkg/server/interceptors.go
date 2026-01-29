package server

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/symbioticfi/relay/pkg/log"
)

// LoggingInterceptor provides request logging for unary RPCs
func LoggingInterceptor(verboseLogging bool) grpc.UnaryServerInterceptor {
	errorLogger := slog.DebugContext
	infoLogger := slog.DebugContext
	if verboseLogging {
		errorLogger = slog.ErrorContext
		infoLogger = slog.InfoContext
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		logCtx := log.WithComponent(ctx, "grpc")
		slog.InfoContext(logCtx, "gRPC request started", "method", info.FullMethod)

		resp, err := handler(logCtx, req)

		duration := time.Since(start)
		if err != nil {
			errorLogger(logCtx, "gRPC request failed", "method", info.FullMethod, "duration", duration, "error", err)
		} else {
			infoLogger(logCtx, "gRPC request completed", "method", info.FullMethod, "duration", duration)
		}

		return resp, err
	}
}

// wrappedStream wraps grpc.ServerStream to inject context
type wrappedStream struct {
	grpc.ServerStream

	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// StreamLoggingInterceptor provides request logging for streaming RPCs
func StreamLoggingInterceptor(verboseLogging bool) grpc.StreamServerInterceptor {
	errorLogger := slog.DebugContext
	infoLogger := slog.DebugContext
	if verboseLogging {
		errorLogger = slog.ErrorContext
		infoLogger = slog.InfoContext
	}

	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		ctx := stream.Context()
		logCtx := log.WithComponent(ctx, "grpc-stream")
		slog.InfoContext(logCtx, "gRPC stream started", "method", info.FullMethod)

		err := handler(srv, &wrappedStream{stream, logCtx})

		duration := time.Since(start)
		if err != nil {
			errorLogger(logCtx, "gRPC stream failed", "method", info.FullMethod, "duration", duration, "error", err)
		} else {
			infoLogger(logCtx, "gRPC stream completed", "method", info.FullMethod, "duration", duration)
		}

		return err
	}
}

// PanicRecoveryInterceptor recovers from panics in unary gRPC handlers
func PanicRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				slog.ErrorContext(ctx, "gRPC unary handler panic recovered",
					"method", info.FullMethod,
					"error", r)
				err = status.Error(codes.Internal, "Internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

// StreamPanicRecoveryInterceptor recovers from panics in streaming gRPC handlers
func StreamPanicRecoveryInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				slog.ErrorContext(stream.Context(), "gRPC stream handler panic recovered",
					"method", info.FullMethod,
					"error", r)
				err = status.Error(codes.Internal, "Internal server error")
			}
		}()

		return handler(srv, stream)
	}
}

// TraceContextInterceptor enriches the context with trace information for structured logging
func TraceContextInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Enrich context with trace_id and span_id from OpenTelemetry span
		ctx = log.WithTraceContext(ctx)
		return handler(ctx, req)
	}
}

// StreamTraceContextInterceptor enriches the context with trace information for streaming RPCs
func StreamTraceContextInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Enrich context with trace_id and span_id from OpenTelemetry span
		ctx := log.WithTraceContext(stream.Context())
		return handler(srv, &wrappedStream{stream, ctx})
	}
}
