package api_server

import (
	"context"
	"math/big"
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

func TestGetAggregationStatus_Success_ReturnsStatusWithSortedOperators(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAggregator := mocks.NewMockaggregator(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Aggregator: mockAggregator,
		},
	}

	ctx := context.Background()
	requestID := common.HexToHash("0x1234")
	requestIDStr := requestID.Hex()

	aggregationStatus := symbiotic.AggregationStatus{
		VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
		Validators: []symbiotic.Validator{
			{Operator: common.HexToAddress("0x789")},
			{Operator: common.HexToAddress("0x123")},
			{Operator: common.HexToAddress("0x456")},
		},
	}

	mockAggregator.EXPECT().
		GetAggregationStatus(ctx, requestID).
		Return(aggregationStatus, nil)

	req := &apiv1.GetAggregationStatusRequest{
		RequestId: requestIDStr,
	}

	response, err := handler.GetAggregationStatus(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "1000", response.GetCurrentVotingPower())
	assert.Equal(t, []string{
		common.HexToAddress("0x123").Hex(),
		common.HexToAddress("0x456").Hex(),
		common.HexToAddress("0x789").Hex(),
	}, response.GetSignerOperators())
}

func TestGetAggregationStatus_NoAggregator_ReturnsError(t *testing.T) {
	handler := &grpcHandler{
		cfg: Config{
			Aggregator: nil,
		},
	}

	ctx := context.Background()
	req := &apiv1.GetAggregationStatusRequest{
		RequestId: "0x1234",
	}

	response, err := handler.GetAggregationStatus(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	assert.Equal(t, entity.ErrNotAnAggregator, err)
}

func TestGetAggregationStatus_NotFound_ReturnsNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAggregator := mocks.NewMockaggregator(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Aggregator: mockAggregator,
		},
	}

	ctx := context.Background()
	requestID := common.HexToHash("0xnonexistent")
	requestIDStr := requestID.Hex()

	mockAggregator.EXPECT().
		GetAggregationStatus(ctx, requestID).
		Return(symbiotic.AggregationStatus{}, entity.ErrEntityNotFound)

	req := &apiv1.GetAggregationStatusRequest{
		RequestId: requestIDStr,
	}

	response, err := handler.GetAggregationStatus(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Contains(t, st.Message(), "not found")
}

func TestGetAggregationStatus_AggregatorError_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAggregator := mocks.NewMockaggregator(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Aggregator: mockAggregator,
		},
	}

	ctx := context.Background()
	requestID := common.HexToHash("0x1234")
	requestIDStr := requestID.Hex()
	aggregatorError := assert.AnError

	mockAggregator.EXPECT().
		GetAggregationStatus(ctx, requestID).
		Return(symbiotic.AggregationStatus{}, aggregatorError)

	req := &apiv1.GetAggregationStatusRequest{
		RequestId: requestIDStr,
	}

	response, err := handler.GetAggregationStatus(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	assert.Equal(t, aggregatorError, err)
}
