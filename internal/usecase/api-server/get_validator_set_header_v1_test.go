package api_server

import (
	"context"
	"testing"

	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/symbiotic/entity"
)

func TestGetValidatorSetHeader_ValidatorSetFoundInRepo(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := entity.Epoch(10)
	requestedEpoch := entity.Epoch(8)

	// Create test data
	validatorSet := createTestValidatorSet(requestedEpoch)
	expectedHeader, err := validatorSet.GetHeader() // Use the real GetHeader method
	require.NoError(t, err)

	// Setup mocks - validator set found in repository
	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
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
	require.Equal(t, expectedHeader.Epoch, entity.Epoch(response.GetEpoch()))
	require.Equal(t, expectedHeader.QuorumThreshold.String(), response.GetQuorumThreshold())
	require.Equal(t, expectedHeader.ValidatorsSszMRoot.Hex(), response.GetValidatorsSszMroot())
}

func TestGetValidatorSetHeader_UseCurrentEpoch_WhenNoEpochSpecified(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := entity.Epoch(10)
	// Create test data
	validatorSet := createTestValidatorSet(currentEpoch)

	// Setup mocks - no epoch specified, should use current epoch
	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, currentEpoch).Return(validatorSet, nil)

	// Execute the method under test - no epoch specified
	req := &apiv1.GetValidatorSetHeaderRequest{}

	response, err := setup.handler.GetValidatorSetHeader(ctx, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, currentEpoch, entity.Epoch(response.GetEpoch()))
}

func TestGetValidatorSetHeader_ErrorWhenEpochFromFuture(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := entity.Epoch(10)
	futureEpoch := uint64(15)

	// Setup mocks
	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)

	// Execute the method under test
	req := &apiv1.GetValidatorSetHeaderRequest{
		Epoch: &futureEpoch,
	}

	response, err := setup.handler.GetValidatorSetHeader(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "is greater than latest epoch")
}

func TestGetValidatorSetHeader_ErrorWhenGetCurrentEpochFails(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	expectedError := errors.New("failed to get current epoch")

	// Setup mocks
	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(entity.Epoch(0), expectedError)

	// Execute the method under test
	req := &apiv1.GetValidatorSetHeaderRequest{}

	response, err := setup.handler.GetValidatorSetHeader(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "failed to get latest validator set epoch")
}

func TestGetValidatorSetHeader_ErrorWhenRepositoryFails(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := entity.Epoch(10)
	requestedEpoch := entity.Epoch(8)
	expectedError := errors.New("repository connection failed")

	// Setup mocks
	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(entity.ValidatorSet{}, expectedError)

	// Execute the method under test
	req := &apiv1.GetValidatorSetHeaderRequest{
		Epoch: (*uint64)(&requestedEpoch),
	}

	response, err := setup.handler.GetValidatorSetHeader(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "failed to get validator set for epoch")
	require.Contains(t, err.Error(), expectedError.Error())
}
