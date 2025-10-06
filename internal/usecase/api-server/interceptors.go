package api_server

import (
	"context"
	"log/slog"

	"github.com/go-errors/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/symbioticfi/relay/core/entity"
)

// convertToGRPCError converts internal errors to gRPC status errors
func convertToGRPCError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	// If the error is already a gRPC status error, return it as-is
	if _, ok := status.FromError(err); ok {
		return err
	}

	slog.DebugContext(ctx, "Converting error to gRPC status", "error", err)

	// Handle known entity errors
	switch {
	case errors.Is(err, entity.ErrNotAnAggregator):
		return status.Error(codes.PermissionDenied, "Not an aggregator node")
	case errors.Is(err, entity.ErrEntityNotFound):
		return status.Error(codes.NotFound, "Entity not found")
	case errors.Is(err, entity.ErrChainNotFound):
		return status.Error(codes.NotFound, "Chain not found")
	case errors.Is(err, entity.ErrEntityAlreadyExist):
		return status.Error(codes.AlreadyExists, "Entity already exists")
	case errors.Is(err, entity.ErrNoPeers):
		return status.Error(codes.Unavailable, "No peers available")
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "Request cancelled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "Request timeout")
	default:
		// Log internal errors
		slog.ErrorContext(ctx, "Internal server error", "error", err)
		return status.Error(codes.Internal, "Internal server error")
	}
}

// ErrorHandlingInterceptor handles error conversion for unary RPCs
func ErrorHandlingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			return resp, convertToGRPCError(ctx, err)
		}
		return resp, nil
	}
}

// StreamErrorHandlingInterceptor handles error conversion for streaming RPCs
func StreamErrorHandlingInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, stream)
		if err != nil {
			return convertToGRPCError(stream.Context(), err)
		}
		return nil
	}
}
