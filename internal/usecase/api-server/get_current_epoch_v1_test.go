package api_server

import (
	"context"
	"testing"
	"time"

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

func TestGetCurrentEpoch_Success_ReturnsEpochAndStartTime(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo: mockRepo,
		},
	}

	ctx := context.Background()
	expectedEpoch := symbiotic.Epoch(42)
	captureTimestamp := symbiotic.Timestamp(1640995200) // 2022-01-01 00:00:00 UTC

	validatorSetHeader := symbiotic.ValidatorSetHeader{
		Epoch:              expectedEpoch,
		CaptureTimestamp:   captureTimestamp,
		ValidatorsSszMRoot: common.HexToHash("0x123"),
	}

	mockRepo.EXPECT().
		GetLatestValidatorSetHeader(ctx).
		Return(validatorSetHeader, nil)

	req := &apiv1.GetCurrentEpochRequest{}

	response, err := handler.GetCurrentEpoch(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, uint64(expectedEpoch), response.GetEpoch())
	assert.NotNil(t, response.GetStartTime())

	expectedTime := time.Unix(int64(captureTimestamp), 0).UTC()
	assert.Equal(t, expectedTime.Unix(), response.GetStartTime().AsTime().Unix())
}

func TestGetCurrentEpoch_NotFound_ReturnsNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo: mockRepo,
		},
	}

	ctx := context.Background()

	mockRepo.EXPECT().
		GetLatestValidatorSetHeader(ctx).
		Return(symbiotic.ValidatorSetHeader{}, entity.ErrEntityNotFound)

	req := &apiv1.GetCurrentEpochRequest{}

	response, err := handler.GetCurrentEpoch(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Contains(t, st.Message(), "no validator set header found")
}

func TestGetCurrentEpoch_RepositoryError_ReturnsError(t *testing.T) {
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
		GetLatestValidatorSetHeader(ctx).
		Return(symbiotic.ValidatorSetHeader{}, repoError)

	req := &apiv1.GetCurrentEpochRequest{}

	response, err := handler.GetCurrentEpoch(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get latest validator set header")
}
