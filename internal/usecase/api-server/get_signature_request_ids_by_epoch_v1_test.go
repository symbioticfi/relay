package api_server

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestGetSignatureRequestIDsByEpoch_Success(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestedEpoch := symbiotic.Epoch(5)

	expectedIDs := []common.Hash{
		common.HexToHash("0xabcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"),
		common.HexToHash("0xefgh5678efgh5678efgh5678efgh5678efgh5678efgh5678efgh5678efgh5678"),
		common.HexToHash("0x12345678123456781234567812345678123456781234567812345678123456789"),
	}

	setup.mockRepo.EXPECT().GetSignatureRequestIDsByEpoch(ctx, requestedEpoch).Return(expectedIDs, nil)

	req := &apiv1.GetSignatureRequestIDsByEpochRequest{
		Epoch: uint64(requestedEpoch),
	}

	response, err := setup.handler.GetSignatureRequestIDsByEpoch(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Len(t, response.GetRequestIds(), 3)
	require.Equal(t, expectedIDs[0].Hex(), response.GetRequestIds()[0])
	require.Equal(t, expectedIDs[1].Hex(), response.GetRequestIds()[1])
	require.Equal(t, expectedIDs[2].Hex(), response.GetRequestIds()[2])
}

func TestGetSignatureRequestIDsByEpoch_EmptyResult(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestedEpoch := symbiotic.Epoch(10)

	setup.mockRepo.EXPECT().GetSignatureRequestIDsByEpoch(ctx, requestedEpoch).Return([]common.Hash{}, nil)

	req := &apiv1.GetSignatureRequestIDsByEpochRequest{
		Epoch: uint64(requestedEpoch),
	}

	response, err := setup.handler.GetSignatureRequestIDsByEpoch(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Empty(t, response.GetRequestIds())
}

func TestGetSignatureRequestIDsByEpoch_RepositoryError(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestedEpoch := symbiotic.Epoch(5)
	expectedError := errors.New("database connection failed")

	setup.mockRepo.EXPECT().GetSignatureRequestIDsByEpoch(ctx, requestedEpoch).Return(nil, expectedError)

	req := &apiv1.GetSignatureRequestIDsByEpochRequest{
		Epoch: uint64(requestedEpoch),
	}

	response, err := setup.handler.GetSignatureRequestIDsByEpoch(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "failed to get signature request IDs by epoch")
}
