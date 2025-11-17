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

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/internal/usecase/api-server/mocks"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	deriverMocks "github.com/symbioticfi/relay/symbiotic/usecase/valset-deriver/mocks"
)

func TestGetLastCommitted_Success_ReturnsLastCommittedEpoch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	mockEvmClient := deriverMocks.NewMockEvmClient(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo:      mockRepo,
			EvmClient: mockEvmClient,
		},
	}

	ctx := context.Background()
	chainID := uint64(1)
	currentEpoch := symbiotic.Epoch(10)
	lastCommittedEpoch := symbiotic.Epoch(8)
	epochStart := symbiotic.Timestamp(1640995200)

	settlement := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x123"),
	}

	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement},
	}

	mockRepo.EXPECT().
		GetLatestValidatorSetEpoch(ctx).
		Return(currentEpoch, nil)

	mockEvmClient.EXPECT().
		GetConfig(ctx, gomock.Any(), currentEpoch).
		Return(config, nil)

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement).
		Return(lastCommittedEpoch, nil)

	mockEvmClient.EXPECT().
		GetEpochStart(ctx, lastCommittedEpoch).
		Return(epochStart, nil)

	req := &apiv1.GetLastCommittedRequest{
		SettlementChainId: chainID,
	}

	response, err := handler.GetLastCommitted(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, chainID, response.GetSettlementChainId())
	assert.Equal(t, uint64(lastCommittedEpoch), response.GetEpochInfo().GetLastCommittedEpoch())
	assert.NotNil(t, response.GetEpochInfo().GetStartTime())
}

func TestGetLastCommitted_ZeroChainID_ReturnsInvalidArgumentError(t *testing.T) {
	handler := &grpcHandler{
		cfg: Config{},
	}

	ctx := context.Background()
	req := &apiv1.GetLastCommittedRequest{
		SettlementChainId: 0,
	}

	response, err := handler.GetLastCommitted(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "cannot be 0")
}

func TestGetLastCommitted_ChainNotFound_ReturnsNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	mockEvmClient := deriverMocks.NewMockEvmClient(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo:      mockRepo,
			EvmClient: mockEvmClient,
		},
	}

	ctx := context.Background()
	requestedChainID := uint64(999)
	currentEpoch := symbiotic.Epoch(10)

	settlement := symbiotic.CrossChainAddress{
		ChainId: 1,
		Address: common.HexToAddress("0x123"),
	}

	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement},
	}

	mockRepo.EXPECT().
		GetLatestValidatorSetEpoch(ctx).
		Return(currentEpoch, nil)

	mockEvmClient.EXPECT().
		GetConfig(ctx, gomock.Any(), currentEpoch).
		Return(config, nil)

	req := &apiv1.GetLastCommittedRequest{
		SettlementChainId: requestedChainID,
	}

	response, err := handler.GetLastCommitted(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Contains(t, st.Message(), "not found in network config")
}

func TestGetLastCommitted_GetLastCommittedHeaderEpochFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	mockEvmClient := deriverMocks.NewMockEvmClient(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo:      mockRepo,
			EvmClient: mockEvmClient,
		},
	}

	ctx := context.Background()
	chainID := uint64(1)
	currentEpoch := symbiotic.Epoch(10)
	evmError := assert.AnError

	settlement := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x123"),
	}

	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement},
	}

	mockRepo.EXPECT().
		GetLatestValidatorSetEpoch(ctx).
		Return(currentEpoch, nil)

	mockEvmClient.EXPECT().
		GetConfig(ctx, gomock.Any(), currentEpoch).
		Return(config, nil)

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement).
		Return(symbiotic.Epoch(0), evmError)

	req := &apiv1.GetLastCommittedRequest{
		SettlementChainId: chainID,
	}

	response, err := handler.GetLastCommitted(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get last committed epoch")
}

func TestGetLastCommitted_GetEpochStartFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	mockEvmClient := deriverMocks.NewMockEvmClient(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo:      mockRepo,
			EvmClient: mockEvmClient,
		},
	}

	ctx := context.Background()
	chainID := uint64(1)
	currentEpoch := symbiotic.Epoch(10)
	lastCommittedEpoch := symbiotic.Epoch(8)
	evmError := assert.AnError

	settlement := symbiotic.CrossChainAddress{
		ChainId: chainID,
		Address: common.HexToAddress("0x123"),
	}

	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement},
	}

	mockRepo.EXPECT().
		GetLatestValidatorSetEpoch(ctx).
		Return(currentEpoch, nil)

	mockEvmClient.EXPECT().
		GetConfig(ctx, gomock.Any(), currentEpoch).
		Return(config, nil)

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement).
		Return(lastCommittedEpoch, nil)

	mockEvmClient.EXPECT().
		GetEpochStart(ctx, lastCommittedEpoch).
		Return(symbiotic.Timestamp(0), evmError)

	req := &apiv1.GetLastCommittedRequest{
		SettlementChainId: chainID,
	}

	response, err := handler.GetLastCommitted(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get epoch start")
}

func TestGetLastCommitted_GetConfigFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	mockEvmClient := deriverMocks.NewMockEvmClient(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo:      mockRepo,
			EvmClient: mockEvmClient,
		},
	}

	ctx := context.Background()
	chainID := uint64(1)
	currentEpoch := symbiotic.Epoch(10)
	evmError := assert.AnError

	mockRepo.EXPECT().
		GetLatestValidatorSetEpoch(ctx).
		Return(currentEpoch, nil)

	mockEvmClient.EXPECT().
		GetConfig(ctx, gomock.Any(), currentEpoch).
		Return(symbiotic.NetworkConfig{}, evmError)

	req := &apiv1.GetLastCommittedRequest{
		SettlementChainId: chainID,
	}

	response, err := handler.GetLastCommitted(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get config")
}
