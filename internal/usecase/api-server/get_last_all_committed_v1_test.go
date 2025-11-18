package api_server

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/internal/usecase/api-server/mocks"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	deriverMocks "github.com/symbioticfi/relay/symbiotic/usecase/valset-deriver/mocks"
)

func TestGetLastAllCommitted_Success_ReturnsAllChainsWithMinimum(t *testing.T) {
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
	currentEpoch := symbiotic.Epoch(20)

	settlement1 := symbiotic.CrossChainAddress{
		ChainId: 1,
		Address: common.HexToAddress("0x123"),
	}
	settlement2 := symbiotic.CrossChainAddress{
		ChainId: 2,
		Address: common.HexToAddress("0x456"),
	}

	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement1, settlement2},
	}

	mockRepo.EXPECT().
		GetLatestValidatorSetEpoch(ctx).
		Return(currentEpoch, nil)

	mockEvmClient.EXPECT().
		GetConfig(ctx, gomock.Any(), currentEpoch).
		Return(config, nil)

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement1).
		Return(symbiotic.Epoch(15), nil)

	mockEvmClient.EXPECT().
		GetEpochStart(ctx, symbiotic.Epoch(15)).
		Return(symbiotic.Timestamp(1640995200), nil)

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement2).
		Return(symbiotic.Epoch(12), nil)

	mockEvmClient.EXPECT().
		GetEpochStart(ctx, symbiotic.Epoch(12)).
		Return(symbiotic.Timestamp(1640995100), nil)

	req := &apiv1.GetLastAllCommittedRequest{}

	response, err := handler.GetLastAllCommitted(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.GetEpochInfos(), 2)
	assert.Equal(t, uint64(15), response.GetEpochInfos()[1].GetLastCommittedEpoch())
	assert.Equal(t, uint64(12), response.GetEpochInfos()[2].GetLastCommittedEpoch())
	assert.Equal(t, uint64(12), response.GetSuggestedEpochInfo().GetLastCommittedEpoch())
}

func TestGetLastAllCommitted_GetConfigFails_ReturnsError(t *testing.T) {
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
	currentEpoch := symbiotic.Epoch(10)
	evmError := assert.AnError

	mockRepo.EXPECT().
		GetLatestValidatorSetEpoch(ctx).
		Return(currentEpoch, nil)

	mockEvmClient.EXPECT().
		GetConfig(ctx, gomock.Any(), currentEpoch).
		Return(symbiotic.NetworkConfig{}, evmError)

	req := &apiv1.GetLastAllCommittedRequest{}

	response, err := handler.GetLastAllCommitted(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get config")
}

func TestGetLastAllCommitted_GetLastCommittedEpochFails_ReturnsError(t *testing.T) {
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
	currentEpoch := symbiotic.Epoch(10)

	settlement := symbiotic.CrossChainAddress{
		ChainId: 1,
		Address: common.HexToAddress("0x123"),
	}

	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement},
	}

	evmError := assert.AnError

	mockRepo.EXPECT().
		GetLatestValidatorSetEpoch(ctx).
		Return(currentEpoch, nil)

	mockEvmClient.EXPECT().
		GetConfig(ctx, gomock.Any(), currentEpoch).
		Return(config, nil)

	mockEvmClient.EXPECT().
		GetLastCommittedHeaderEpoch(ctx, settlement).
		Return(symbiotic.Epoch(0), evmError)

	req := &apiv1.GetLastAllCommittedRequest{}

	response, err := handler.GetLastAllCommitted(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get last committed epoch")
}

func TestGetLastAllCommitted_GetEpochStartFails_ReturnsError(t *testing.T) {
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
	currentEpoch := symbiotic.Epoch(10)

	settlement := symbiotic.CrossChainAddress{
		ChainId: 1,
		Address: common.HexToAddress("0x123"),
	}

	config := symbiotic.NetworkConfig{
		Settlements: []symbiotic.CrossChainAddress{settlement},
	}

	lastCommittedEpoch := symbiotic.Epoch(8)
	evmError := assert.AnError

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

	req := &apiv1.GetLastAllCommittedRequest{}

	response, err := handler.GetLastAllCommitted(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get epoch start")
}
