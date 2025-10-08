package api_server

import (
	"context"
	"testing"

	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/symbiotic/entity"
)

func TestGetValidatorByAddress_ValidatorFoundInRepo(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestedEpoch := entity.Epoch(8)
	currentEpoch := entity.Epoch(10)
	validatorAddress := "0x0000000000000000000000000000000000000123"

	// Create test data
	validatorSet := createTestValidatorSetWithMultipleValidators(requestedEpoch)
	expectedValidator := validatorSet.Validators[0] // First validator has address 0x123

	// Setup mocks - validator set found in repository
	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(validatorSet, nil)

	// Execute the method under test
	req := &apiv1.GetValidatorByAddressRequest{
		Address: validatorAddress,
		Epoch:   (*uint64)(&requestedEpoch),
	}

	response, err := setup.handler.GetValidatorByAddress(ctx, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.GetValidator())
	require.Equal(t, expectedValidator.Operator.Hex(), response.GetValidator().GetOperator())
	require.Equal(t, expectedValidator.VotingPower.String(), response.GetValidator().GetVotingPower())
	require.Equal(t, expectedValidator.IsActive, response.GetValidator().GetIsActive())
	require.Len(t, response.GetValidator().GetKeys(), len(expectedValidator.Keys))
	require.Equal(t, uint32(expectedValidator.Keys[0].Tag), response.GetValidator().GetKeys()[0].GetTag())
	require.Equal(t, []byte(expectedValidator.Keys[0].Payload), response.GetValidator().GetKeys()[0].GetPayload())
	require.Len(t, response.GetValidator().GetVaults(), len(expectedValidator.Vaults))
}

func TestGetValidatorByAddress_ValidatorSetNotInRepo_DerivedFail(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := entity.Epoch(10)
	requestedEpoch := entity.Epoch(8)

	validatorAddress := "0x0000000000000000000000000000000000000abc"

	// Setup mocks - validator set not in repository, needs to be derived
	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(entity.ValidatorSet{}, entity.ErrEntityNotFound)

	// Execute the method under test
	req := &apiv1.GetValidatorByAddressRequest{
		Address: validatorAddress,
		Epoch:   (*uint64)(&requestedEpoch),
	}

	response, err := setup.handler.GetValidatorByAddress(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
}

func TestGetValidatorByAddress_UseCurrentEpoch_WhenNoEpochSpecified(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := entity.Epoch(10)
	validatorAddress := "0x0000000000000000000000000000000000000123"

	// Create test data
	validatorSet := createTestValidatorSetWithMultipleValidators(currentEpoch)

	// Setup mocks - no epoch specified, should use current epoch
	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, currentEpoch).Return(validatorSet, nil)

	// Execute the method under test - no epoch specified
	req := &apiv1.GetValidatorByAddressRequest{
		Address: validatorAddress,
	}

	response, err := setup.handler.GetValidatorByAddress(ctx, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.GetValidator())
	require.Equal(t, validatorAddress, response.GetValidator().GetOperator())
}

func TestGetValidatorByAddress_ErrorWhenEpochFromFuture(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := entity.Epoch(10)
	futureEpoch := uint64(15)
	validatorAddress := "0x0000000000000000000000000000000000000123"

	// Setup mocks
	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)

	// Execute the method under test
	req := &apiv1.GetValidatorByAddressRequest{
		Address: validatorAddress,
		Epoch:   &futureEpoch,
	}

	response, err := setup.handler.GetValidatorByAddress(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "is greater than latest epoch")
}

func TestGetValidatorByAddress_ErrorWhenInvalidAddress(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := entity.Epoch(10)
	requestedEpoch := uint64(8)
	invalidAddress := "not-a-valid-address"

	// Setup mocks
	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)

	// Execute the method under test
	req := &apiv1.GetValidatorByAddressRequest{
		Address: invalidAddress,
		Epoch:   &requestedEpoch,
	}

	response, err := setup.handler.GetValidatorByAddress(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "invalid validator address format")
}

func TestGetValidatorByAddress_ErrorWhenValidatorNotFound(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := entity.Epoch(10)
	requestedEpoch := entity.Epoch(8)
	nonExistentAddress := "0x0000000000000000000000000000000000000999"

	// Create test data without the requested validator
	validatorSet := createTestValidatorSetWithMultipleValidators(requestedEpoch)

	// Setup mocks - validator set found in repository
	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(requestedEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(entity.ValidatorSet{}, entity.ErrEntityNotFound)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(validatorSet, nil)

	// Execute the method under test
	req := &apiv1.GetValidatorByAddressRequest{
		Address: nonExistentAddress,
		Epoch:   (*uint64)(&requestedEpoch),
	}

	response, err := setup.handler.GetValidatorByAddress(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "validator set for epoch 8 not found")

	response, err = setup.handler.GetValidatorByAddress(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "not found for epoch")
}

func TestGetValidatorByAddress_ErrorWhenGetCurrentEpochFails(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	validatorAddress := "0x0000000000000000000000000000000000000123"
	expectedError := errors.New("failed to get current epoch")

	// Setup mocks
	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(entity.Epoch(0), expectedError)

	// Execute the method under test
	req := &apiv1.GetValidatorByAddressRequest{
		Address: validatorAddress,
	}

	response, err := setup.handler.GetValidatorByAddress(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "failed to get latest validator set epoch")
}

func TestGetValidatorByAddress_ErrorWhenRepositoryFails(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := entity.Epoch(10)
	requestedEpoch := entity.Epoch(8)
	validatorAddress := "0x0000000000000000000000000000000000000123"
	expectedError := errors.New("repository connection failed")

	// Setup mocks
	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(entity.ValidatorSet{}, expectedError)

	// Execute the method under test
	req := &apiv1.GetValidatorByAddressRequest{
		Address: validatorAddress,
		Epoch:   (*uint64)(&requestedEpoch),
	}

	response, err := setup.handler.GetValidatorByAddress(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "failed to get validator set for epoch")
	require.Contains(t, err.Error(), expectedError.Error())
}
