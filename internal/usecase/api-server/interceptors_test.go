package api_server

import (
	"context"
	"testing"

	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/symbioticfi/relay/internal/entity"
)

func TestConvertToGRPCError_NilError_ReturnsNil(t *testing.T) {
	ctx := context.Background()

	result := convertToGRPCError(ctx, nil)

	require.NoError(t, result)
}

func TestConvertToGRPCError_AlreadyGRPCError_ReturnsAsIs(t *testing.T) {
	ctx := context.Background()
	originalErr := status.Error(codes.InvalidArgument, "already a gRPC error")

	result := convertToGRPCError(ctx, originalErr)

	require.Equal(t, originalErr, result)
	st, ok := status.FromError(result)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Equal(t, "already a gRPC error", st.Message())
}

func TestConvertToGRPCError_ErrNotAnAggregator_ReturnsPermissionDenied(t *testing.T) {
	ctx := context.Background()
	err := entity.ErrNotAnAggregator

	result := convertToGRPCError(ctx, err)

	require.Error(t, result)
	st, ok := status.FromError(result)
	require.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, st.Code())
	assert.Contains(t, st.Message(), "Not an aggregator node")
}

func TestConvertToGRPCError_ErrEntityNotFound_ReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	err := entity.ErrEntityNotFound

	result := convertToGRPCError(ctx, err)

	require.Error(t, result)
	st, ok := status.FromError(result)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Contains(t, st.Message(), "Entity not found")
}

func TestConvertToGRPCError_ErrChainNotFound_ReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	err := entity.ErrChainNotFound

	result := convertToGRPCError(ctx, err)

	require.Error(t, result)
	st, ok := status.FromError(result)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Contains(t, st.Message(), "Chain not found")
}

func TestConvertToGRPCError_ErrEntityAlreadyExist_ReturnsAlreadyExists(t *testing.T) {
	ctx := context.Background()
	err := entity.ErrEntityAlreadyExist

	result := convertToGRPCError(ctx, err)

	require.Error(t, result)
	st, ok := status.FromError(result)
	require.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, st.Code())
	assert.Contains(t, st.Message(), "Entity already exists")
}

func TestConvertToGRPCError_ErrNoPeers_ReturnsUnavailable(t *testing.T) {
	ctx := context.Background()
	err := entity.ErrNoPeers

	result := convertToGRPCError(ctx, err)

	require.Error(t, result)
	st, ok := status.FromError(result)
	require.True(t, ok)
	assert.Equal(t, codes.Unavailable, st.Code())
	assert.Contains(t, st.Message(), "No peers available")
}

func TestConvertToGRPCError_ContextCanceled_ReturnsCanceled(t *testing.T) {
	ctx := context.Background()
	err := context.Canceled

	result := convertToGRPCError(ctx, err)

	require.Error(t, result)
	st, ok := status.FromError(result)
	require.True(t, ok)
	assert.Equal(t, codes.Canceled, st.Code())
	assert.Contains(t, st.Message(), "Request cancelled")
}

func TestConvertToGRPCError_ContextDeadlineExceeded_ReturnsDeadlineExceeded(t *testing.T) {
	ctx := context.Background()
	err := context.DeadlineExceeded

	result := convertToGRPCError(ctx, err)

	require.Error(t, result)
	st, ok := status.FromError(result)
	require.True(t, ok)
	assert.Equal(t, codes.DeadlineExceeded, st.Code())
	assert.Contains(t, st.Message(), "Request timeout")
}

func TestConvertToGRPCError_UnknownError_ReturnsInternal(t *testing.T) {
	ctx := context.Background()
	err := errors.New("some random internal error")

	result := convertToGRPCError(ctx, err)

	require.Error(t, result)
	st, ok := status.FromError(result)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "Internal server error")
}

func TestErrorHandlingInterceptor_HandlerReturnsNil_ReturnsSuccess(t *testing.T) {
	ctx := context.Background()
	interceptor := ErrorHandlingInterceptor()

	mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success response", nil
	}

	resp, err := interceptor(ctx, "test request", &grpc.UnaryServerInfo{}, mockHandler)

	require.NoError(t, err)
	assert.Equal(t, "success response", resp)
}

func TestErrorHandlingInterceptor_HandlerReturnsError_ConvertsError(t *testing.T) {
	ctx := context.Background()
	interceptor := ErrorHandlingInterceptor()

	mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, entity.ErrEntityNotFound
	}

	resp, err := interceptor(ctx, "test request", &grpc.UnaryServerInfo{}, mockHandler)

	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestErrorHandlingInterceptor_HandlerReturnsGRPCError_ReturnsAsIs(t *testing.T) {
	ctx := context.Background()
	interceptor := ErrorHandlingInterceptor()

	originalErr := status.Error(codes.InvalidArgument, "invalid input")
	mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, originalErr
	}

	resp, err := interceptor(ctx, "test request", &grpc.UnaryServerInfo{}, mockHandler)

	require.Error(t, err)
	require.Nil(t, resp)
	assert.Equal(t, originalErr, err)
}

type mockServerStream struct {
	grpc.ServerStream

	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

func TestStreamErrorHandlingInterceptor_HandlerReturnsNil_ReturnsSuccess(t *testing.T) {
	ctx := context.Background()
	interceptor := StreamErrorHandlingInterceptor()

	stream := &mockServerStream{ctx: ctx}

	mockHandler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}

	err := interceptor(nil, stream, &grpc.StreamServerInfo{}, mockHandler)

	require.NoError(t, err)
}

func TestStreamErrorHandlingInterceptor_HandlerReturnsError_ConvertsError(t *testing.T) {
	ctx := context.Background()
	interceptor := StreamErrorHandlingInterceptor()

	stream := &mockServerStream{ctx: ctx}

	mockHandler := func(srv interface{}, stream grpc.ServerStream) error {
		return entity.ErrNoPeers
	}

	err := interceptor(nil, stream, &grpc.StreamServerInfo{}, mockHandler)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unavailable, st.Code())
}

func TestStreamErrorHandlingInterceptor_HandlerReturnsGRPCError_ReturnsAsIs(t *testing.T) {
	ctx := context.Background()
	interceptor := StreamErrorHandlingInterceptor()

	stream := &mockServerStream{ctx: ctx}

	originalErr := status.Error(codes.Unauthenticated, "not authenticated")
	mockHandler := func(srv interface{}, stream grpc.ServerStream) error {
		return originalErr
	}

	err := interceptor(nil, stream, &grpc.StreamServerInfo{}, mockHandler)

	require.Error(t, err)
	assert.Equal(t, originalErr, err)
}
