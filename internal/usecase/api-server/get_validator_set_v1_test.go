package api_server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/internal/usecase/api-server/mocks"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestGetValidatorSet_WithoutEpoch_ReturnsLatestValidatorSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo: mockRepo,
		},
	}

	ctx := context.Background()
	latestEpoch := symbiotic.Epoch(10)
	validatorSet := createTestValidatorSet(latestEpoch)

	mockRepo.EXPECT().
		GetLatestValidatorSetEpoch(ctx).
		Return(latestEpoch, nil)

	mockRepo.EXPECT().
		GetValidatorSetByEpoch(ctx, latestEpoch).
		Return(validatorSet, nil)

	req := &apiv1.GetValidatorSetRequest{
		Epoch: nil,
	}

	response, err := handler.GetValidatorSet(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.GetValidatorSet())
	assert.Equal(t, uint64(latestEpoch), response.GetValidatorSet().GetEpoch())
}

func TestGetValidatorSet_WithSpecificEpoch_ReturnsRequestedValidatorSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo: mockRepo,
		},
	}

	ctx := context.Background()
	latestEpoch := symbiotic.Epoch(15)
	requestedEpoch := symbiotic.Epoch(12)
	requestedEpochUint64 := uint64(requestedEpoch)
	validatorSet := createTestValidatorSet(requestedEpoch)

	mockRepo.EXPECT().
		GetLatestValidatorSetEpoch(ctx).
		Return(latestEpoch, nil)

	mockRepo.EXPECT().
		GetValidatorSetByEpoch(ctx, requestedEpoch).
		Return(validatorSet, nil)

	req := &apiv1.GetValidatorSetRequest{
		Epoch: &requestedEpochUint64,
	}

	response, err := handler.GetValidatorSet(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.GetValidatorSet())
	assert.Equal(t, uint64(requestedEpoch), response.GetValidatorSet().GetEpoch())
}

func TestGetValidatorSet_FutureEpoch_ReturnsInvalidArgumentError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo: mockRepo,
		},
	}

	ctx := context.Background()
	latestEpoch := symbiotic.Epoch(10)
	futureEpoch := uint64(15)

	mockRepo.EXPECT().
		GetLatestValidatorSetEpoch(ctx).
		Return(latestEpoch, nil)

	req := &apiv1.GetValidatorSetRequest{
		Epoch: &futureEpoch,
	}

	response, err := handler.GetValidatorSet(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "greater than latest epoch")
}

func TestGetValidatorSet_GetLatestEpochFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo: mockRepo,
		},
	}

	ctx := context.Background()
	repoError := assert.AnError

	mockRepo.EXPECT().
		GetLatestValidatorSetEpoch(ctx).
		Return(symbiotic.Epoch(0), repoError)

	req := &apiv1.GetValidatorSetRequest{}

	response, err := handler.GetValidatorSet(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get latest validator set epoch")
}

func TestConvertValidatorSetStatusToPB_AllStatuses(t *testing.T) {
	tests := []struct {
		name     string
		status   symbiotic.ValidatorSetStatus
		expected apiv1.ValidatorSetStatus
	}{
		{
			name:     "HeaderDerived",
			status:   symbiotic.ValidatorSetStatus(symbiotic.HeaderDerived),
			expected: apiv1.ValidatorSetStatus_VALIDATOR_SET_STATUS_DERIVED,
		},
		{
			name:     "HeaderAggregated",
			status:   symbiotic.ValidatorSetStatus(symbiotic.HeaderAggregated),
			expected: apiv1.ValidatorSetStatus_VALIDATOR_SET_STATUS_AGGREGATED,
		},
		{
			name:     "HeaderCommitted",
			status:   symbiotic.ValidatorSetStatus(symbiotic.HeaderCommitted),
			expected: apiv1.ValidatorSetStatus_VALIDATOR_SET_STATUS_COMMITTED,
		},
		{
			name:     "HeaderMissed",
			status:   symbiotic.ValidatorSetStatus(symbiotic.HeaderMissed),
			expected: apiv1.ValidatorSetStatus_VALIDATOR_SET_STATUS_MISSED,
		},
		{
			name:     "UnknownStatus",
			status:   symbiotic.ValidatorSetStatus(2 << 6),
			expected: apiv1.ValidatorSetStatus_VALIDATOR_SET_STATUS_UNSPECIFIED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertValidatorSetStatusToPB(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}
