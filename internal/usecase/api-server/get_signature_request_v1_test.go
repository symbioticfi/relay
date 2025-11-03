package api_server

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/symbioticfi/relay/internal/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestGetSignatureRequest_Success(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestID := common.HexToHash("0xabcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234")
	requestIDStr := requestID.Hex()

	expectedRequest := symbiotic.SignatureRequest{
		KeyTag:        15,
		RequiredEpoch: 5,
		Message:       []byte("test message"),
	}

	setup.mockRepo.EXPECT().GetSignatureRequest(ctx, requestID).Return(expectedRequest, nil)

	req := &apiv1.GetSignatureRequestRequest{
		RequestId: requestIDStr,
	}

	response, err := setup.handler.GetSignatureRequest(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.GetSignatureRequest())

	// Verify request ID is included
	require.Equal(t, requestIDStr, response.GetSignatureRequest().GetRequestId())
	require.Equal(t, uint32(15), response.GetSignatureRequest().GetKeyTag())
	require.Equal(t, []byte("test message"), response.GetSignatureRequest().GetMessage())
	require.Equal(t, uint64(5), response.GetSignatureRequest().GetRequiredEpoch())
}

func TestGetSignatureRequest_NotFound(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestID := common.HexToHash("0xabcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234")
	requestIDStr := requestID.Hex()

	setup.mockRepo.EXPECT().GetSignatureRequest(ctx, requestID).Return(symbiotic.SignatureRequest{}, entity.ErrEntityNotFound)

	req := &apiv1.GetSignatureRequestRequest{
		RequestId: requestIDStr,
	}

	response, err := setup.handler.GetSignatureRequest(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)

	// Check that it's a NotFound error
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
	require.Contains(t, st.Message(), "not found")
}

func TestGetSignatureRequest_RepositoryError(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestID := common.HexToHash("0xabcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234")
	requestIDStr := requestID.Hex()

	expectedError := status.Error(codes.Internal, "internal error")
	setup.mockRepo.EXPECT().GetSignatureRequest(ctx, requestID).Return(symbiotic.SignatureRequest{}, expectedError)

	req := &apiv1.GetSignatureRequestRequest{
		RequestId: requestIDStr,
	}

	response, err := setup.handler.GetSignatureRequest(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	require.Equal(t, expectedError, err)
}
