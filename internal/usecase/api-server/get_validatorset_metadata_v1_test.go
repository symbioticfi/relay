package api_server

import (
	"context"
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

func TestGetValidatorSetMetadata_WithoutEpoch_ReturnsLatestMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo: mockRepo,
		},
	}

	ctx := context.Background()
	latestEpoch := symbiotic.Epoch(20)
	requestID := common.HexToHash("0x1234")

	metadata := symbiotic.ValidatorSetMetadata{
		RequestID: requestID,
		ExtraData: []symbiotic.ExtraData{
			{
				Key:   common.HexToHash("0x1"),
				Value: common.HexToHash("0x64"),
			},
		},
		CommitmentData: []byte("commitment"),
	}

	mockRepo.EXPECT().
		GetLatestValidatorSetEpoch(ctx).
		Return(latestEpoch, nil)

	mockRepo.EXPECT().
		GetValidatorSetMetadata(ctx, latestEpoch).
		Return(metadata, nil)

	req := &apiv1.GetValidatorSetMetadataRequest{
		Epoch: nil,
	}

	response, err := handler.GetValidatorSetMetadata(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, requestID.Hex(), response.GetRequestId())
	assert.Len(t, response.GetExtraData(), 1)
	assert.Equal(t, []byte("commitment"), response.GetCommitmentData())
}

func TestGetValidatorSetMetadata_WithEpoch_ReturnsMetadataForEpoch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo: mockRepo,
		},
	}

	ctx := context.Background()
	requestedEpoch := symbiotic.Epoch(15)
	requestedEpochUint64 := uint64(requestedEpoch)
	requestID := common.HexToHash("0xabcd")

	metadata := symbiotic.ValidatorSetMetadata{
		RequestID:      requestID,
		ExtraData:      []symbiotic.ExtraData{},
		CommitmentData: []byte("data"),
	}

	mockRepo.EXPECT().
		GetValidatorSetMetadata(ctx, requestedEpoch).
		Return(metadata, nil)

	req := &apiv1.GetValidatorSetMetadataRequest{
		Epoch: &requestedEpochUint64,
	}

	response, err := handler.GetValidatorSetMetadata(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, requestID.Hex(), response.GetRequestId())
}

func TestGetValidatorSetMetadata_NotFound_ReturnsNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo: mockRepo,
		},
	}

	ctx := context.Background()
	epoch := symbiotic.Epoch(10)
	epochUint64 := uint64(epoch)

	mockRepo.EXPECT().
		GetValidatorSetMetadata(ctx, epoch).
		Return(symbiotic.ValidatorSetMetadata{}, entity.ErrEntityNotFound)

	req := &apiv1.GetValidatorSetMetadataRequest{
		Epoch: &epochUint64,
	}

	response, err := handler.GetValidatorSetMetadata(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Contains(t, st.Message(), "no metadata found")
}

func TestGetValidatorSetMetadata_RepositoryError_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo: mockRepo,
		},
	}

	ctx := context.Background()
	epoch := symbiotic.Epoch(5)
	epochUint64 := uint64(epoch)
	repoError := assert.AnError

	mockRepo.EXPECT().
		GetValidatorSetMetadata(ctx, epoch).
		Return(symbiotic.ValidatorSetMetadata{}, repoError)

	req := &apiv1.GetValidatorSetMetadataRequest{
		Epoch: &epochUint64,
	}

	response, err := handler.GetValidatorSetMetadata(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get validator set metadata")
}

func TestGetValidatorSetMetadata_GetLatestEpochFails_ReturnsError(t *testing.T) {
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

	req := &apiv1.GetValidatorSetMetadataRequest{
		Epoch: nil,
	}

	response, err := handler.GetValidatorSetMetadata(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get latest validator set epoch")
}
