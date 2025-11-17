package api_server

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/symbioticfi/relay/internal/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/internal/usecase/api-server/mocks"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestGetAggregationProof_Success_ReturnsProof(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo: mockRepo,
		},
	}

	ctx := t.Context()
	requestID := common.HexToHash("0xabcd1234")
	requestIDStr := requestID.Hex()

	expectedProof := symbiotic.AggregationProof{
		MessageHash: []byte("message hash"),
		Proof:       []byte("proof data"),
	}

	mockRepo.EXPECT().
		GetAggregationProof(ctx, requestID).
		Return(expectedProof, nil)

	req := &apiv1.GetAggregationProofRequest{
		RequestId: requestIDStr,
	}

	response, err := handler.GetAggregationProof(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.GetAggregationProof())
	assert.Equal(t, []byte(expectedProof.MessageHash), response.GetAggregationProof().GetMessageHash())
	assert.Equal(t, []byte(expectedProof.Proof), response.GetAggregationProof().GetProof())
}

func TestGetAggregationProof_NotFound_ReturnsNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo: mockRepo,
		},
	}

	ctx := t.Context()
	requestID := common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")
	requestIDStr := requestID.Hex()

	mockRepo.EXPECT().
		GetAggregationProof(ctx, requestID).
		Return(symbiotic.AggregationProof{}, entity.ErrEntityNotFound)

	req := &apiv1.GetAggregationProofRequest{
		RequestId: requestIDStr,
	}

	response, err := handler.GetAggregationProof(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Contains(t, st.Message(), "not found")
	assert.Contains(t, st.Message(), requestIDStr)
}

func TestGetAggregationProof_RepositoryError_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo: mockRepo,
		},
	}

	ctx := t.Context()
	requestID := common.HexToHash("0x1234")
	requestIDStr := requestID.Hex()
	repoError := assert.AnError

	mockRepo.EXPECT().
		GetAggregationProof(ctx, requestID).
		Return(symbiotic.AggregationProof{}, repoError)

	req := &apiv1.GetAggregationProofRequest{
		RequestId: requestIDStr,
	}

	response, err := handler.GetAggregationProof(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get aggregation proof")
}
