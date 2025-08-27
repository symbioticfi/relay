package api_server

import (
	"context"
	"testing"

	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/core/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
)

func TestGetValidatorSetHeader_ValidatorSetFoundInRepo(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := uint64(10)
	requestedEpoch := entity.Epoch(uint64(8))

	// Create test data
	validatorSet := createTestValidatorSet(requestedEpoch)
	expectedHeader, err := validatorSet.GetHeader() // Use the real GetHeader method
	require.NoError(t, err)

	// Setup mocks - validator set found in repository
	setup.mockEvmClient.EXPECT().GetCurrentEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(validatorSet, nil)

	// Execute the method under test
	req := &apiv1.GetValidatorSetHeaderRequest{
		Epoch: (*uint64)(&requestedEpoch),
	}

	response, err := setup.handler.GetValidatorSetHeader(ctx, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, uint32(expectedHeader.Version), response.GetVersion())
	require.Equal(t, uint32(expectedHeader.RequiredKeyTag), response.GetRequiredKeyTag())
	require.Equal(t, expectedHeader.Epoch, response.GetEpoch())
	require.Equal(t, expectedHeader.QuorumThreshold.String(), response.GetQuorumThreshold())
	require.Equal(t, expectedHeader.ValidatorsSszMRoot.Hex(), response.GetValidatorsSszMroot())
	require.Equal(t, expectedHeader.PreviousHeaderHash.Hex(), response.GetPreviousHeaderHash())
}

func TestGetValidatorSetHeader_ValidatorSetNotInRepo_DerivedSuccessfully(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := uint64(10)
	requestedEpoch := uint64(8)
	epochStart := uint64(1640995000)

	// Create test data
	validatorSet := createTestValidatorSet(entity.Epoch(requestedEpoch))
	expectedHeader, err := validatorSet.GetHeader() // Use the real GetHeader method
	require.NoError(t, err)

	networkConfig := entity.NetworkConfig{
		RequiredHeaderKeyTag: entity.KeyTag(15),
	}

	// Setup mocks - validator set not in repository, needs to be derived
	setup.mockEvmClient.EXPECT().GetCurrentEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(entity.ValidatorSet{}, entity.ErrEntityNotFound)
	setup.mockEvmClient.EXPECT().GetEpochStart(ctx, requestedEpoch).Return(epochStart, nil)
	setup.mockEvmClient.EXPECT().GetConfig(ctx, epochStart).Return(networkConfig, nil)
	setup.mockDeriver.EXPECT().GetValidatorSet(ctx, requestedEpoch, networkConfig).Return(validatorSet, nil)

	// Execute the method under test
	req := &apiv1.GetValidatorSetHeaderRequest{
		Epoch: &requestedEpoch,
	}

	response, err := setup.handler.GetValidatorSetHeader(ctx, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, uint32(expectedHeader.Version), response.GetVersion())
	require.Equal(t, uint32(expectedHeader.RequiredKeyTag), response.GetRequiredKeyTag())
	require.Equal(t, expectedHeader.Epoch, response.GetEpoch())
	require.Equal(t, expectedHeader.QuorumThreshold.String(), response.GetQuorumThreshold())
	require.Equal(t, expectedHeader.ValidatorsSszMRoot.Hex(), response.GetValidatorsSszMroot())
	require.Equal(t, expectedHeader.PreviousHeaderHash.Hex(), response.GetPreviousHeaderHash())
}

func TestGetValidatorSetHeader_UseCurrentEpoch_WhenNoEpochSpecified(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := uint64(10)

	// Create test data
	validatorSet := createTestValidatorSet(entity.Epoch(currentEpoch))

	// Setup mocks - no epoch specified, should use current epoch
	setup.mockEvmClient.EXPECT().GetCurrentEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, currentEpoch).Return(validatorSet, nil)

	// Execute the method under test - no epoch specified
	req := &apiv1.GetValidatorSetHeaderRequest{}

	response, err := setup.handler.GetValidatorSetHeader(ctx, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, currentEpoch, response.GetEpoch())
}

func TestGetValidatorSetHeader_ErrorWhenEpochFromFuture(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := uint64(10)
	futureEpoch := uint64(15)

	// Setup mocks
	setup.mockEvmClient.EXPECT().GetCurrentEpoch(ctx).Return(currentEpoch, nil)

	// Execute the method under test
	req := &apiv1.GetValidatorSetHeaderRequest{
		Epoch: &futureEpoch,
	}

	response, err := setup.handler.GetValidatorSetHeader(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "epoch requested is greater than latest epoch")
}

func TestGetValidatorSetHeader_ErrorWhenGetCurrentEpochFails(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	expectedError := errors.New("failed to get current epoch")

	// Setup mocks
	setup.mockEvmClient.EXPECT().GetCurrentEpoch(ctx).Return(uint64(0), expectedError)

	// Execute the method under test
	req := &apiv1.GetValidatorSetHeaderRequest{}

	response, err := setup.handler.GetValidatorSetHeader(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Equal(t, expectedError, err)
}

func TestGetValidatorSetHeader_ErrorWhenRepositoryFails(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := uint64(10)
	requestedEpoch := uint64(8)
	expectedError := errors.New("repository connection failed")

	// Setup mocks
	setup.mockEvmClient.EXPECT().GetCurrentEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(entity.ValidatorSet{}, expectedError)

	// Execute the method under test
	req := &apiv1.GetValidatorSetHeaderRequest{
		Epoch: &requestedEpoch,
	}

	response, err := setup.handler.GetValidatorSetHeader(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "failed to get validator set for epoch")
	require.Contains(t, err.Error(), expectedError.Error())
}

func TestGetValidatorSetHeader_ErrorWhenDeriverFails(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := uint64(10)
	requestedEpoch := uint64(8)
	epochStart := uint64(1640995000)
	networkConfig := entity.NetworkConfig{}
	expectedError := errors.New("derivation failed")

	// Setup mocks
	setup.mockEvmClient.EXPECT().GetCurrentEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(entity.ValidatorSet{}, entity.ErrEntityNotFound)
	setup.mockEvmClient.EXPECT().GetEpochStart(ctx, requestedEpoch).Return(epochStart, nil)
	setup.mockEvmClient.EXPECT().GetConfig(ctx, epochStart).Return(networkConfig, nil)
	setup.mockDeriver.EXPECT().GetValidatorSet(ctx, requestedEpoch, networkConfig).Return(entity.ValidatorSet{}, expectedError)

	// Execute the method under test
	req := &apiv1.GetValidatorSetHeaderRequest{
		Epoch: &requestedEpoch,
	}

	response, err := setup.handler.GetValidatorSetHeader(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Equal(t, expectedError, err)
}
