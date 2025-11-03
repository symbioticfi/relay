package api_server

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestGetSignatureRequestsByEpoch_Success(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestedEpoch := symbiotic.Epoch(5)

	requestID1 := common.HexToHash("0xabcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234")
	requestID2 := common.HexToHash("0xefab5678efab5678efab5678efab5678efab5678efab5678efab5678efab5678")

	expectedRequests := []entity.SignatureRequestWithID{
		{
			RequestID: requestID1,
			SignatureRequest: symbiotic.SignatureRequest{
				KeyTag:        15,
				RequiredEpoch: requestedEpoch,
				Message:       []byte("message1"),
			},
		},
		{
			RequestID: requestID2,
			SignatureRequest: symbiotic.SignatureRequest{
				KeyTag:        20,
				RequiredEpoch: requestedEpoch,
				Message:       []byte("message2"),
			},
		},
	}

	setup.mockRepo.EXPECT().GetSignatureRequestsWithIDByEpoch(ctx, requestedEpoch).Return(expectedRequests, nil)

	req := &apiv1.GetSignatureRequestsByEpochRequest{
		Epoch: uint64(requestedEpoch),
	}

	response, err := setup.handler.GetSignatureRequestsByEpoch(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Len(t, response.GetSignatureRequests(), 2)

	// Verify first request
	require.Equal(t, requestID1.Hex(), response.GetSignatureRequests()[0].GetRequestId())
	require.Equal(t, uint32(15), response.GetSignatureRequests()[0].GetKeyTag())
	require.Equal(t, []byte("message1"), response.GetSignatureRequests()[0].GetMessage())
	require.Equal(t, uint64(requestedEpoch), response.GetSignatureRequests()[0].GetRequiredEpoch())

	// Verify second request
	require.Equal(t, requestID2.Hex(), response.GetSignatureRequests()[1].GetRequestId())
	require.Equal(t, uint32(20), response.GetSignatureRequests()[1].GetKeyTag())
	require.Equal(t, []byte("message2"), response.GetSignatureRequests()[1].GetMessage())
	require.Equal(t, uint64(requestedEpoch), response.GetSignatureRequests()[1].GetRequiredEpoch())
}

func TestGetSignatureRequestsByEpoch_EmptyResult(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestedEpoch := symbiotic.Epoch(10)

	setup.mockRepo.EXPECT().GetSignatureRequestsWithIDByEpoch(ctx, requestedEpoch).Return([]entity.SignatureRequestWithID{}, nil)

	req := &apiv1.GetSignatureRequestsByEpochRequest{
		Epoch: uint64(requestedEpoch),
	}

	response, err := setup.handler.GetSignatureRequestsByEpoch(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Empty(t, response.GetSignatureRequests())
}

func TestGetSignatureRequestsByEpoch_RepositoryError(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestedEpoch := symbiotic.Epoch(5)
	expectedError := errors.New("database connection failed")

	setup.mockRepo.EXPECT().GetSignatureRequestsWithIDByEpoch(ctx, requestedEpoch).Return(nil, expectedError)

	req := &apiv1.GetSignatureRequestsByEpochRequest{
		Epoch: uint64(requestedEpoch),
	}

	response, err := setup.handler.GetSignatureRequestsByEpoch(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "failed to get signature requests by epoch")
}
