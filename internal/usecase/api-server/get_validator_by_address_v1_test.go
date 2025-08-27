package api_server

import (
	"context"
	"testing"

	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/core/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
)

func TestGetValidatorByAddress_ValidatorFoundInRepo(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := uint64(10)
	requestedEpoch := uint64(8)
	validatorAddress := "0x0000000000000000000000000000000000000123"

	// Create test data
	validatorSet := createTestValidatorSetWithMultipleValidators(entity.Epoch(requestedEpoch))
	expectedValidator := validatorSet.Validators[0] // First validator has address 0x123

	// Setup mocks - validator set found in repository
	setup.mockEvmClient.EXPECT().GetCurrentEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(validatorSet, nil)

	// Execute the method under test
	req := &apiv1.GetValidatorByAddressRequest{
		Address: validatorAddress,
		Epoch:   &requestedEpoch,
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

func TestGetValidatorByAddress_ValidatorSetNotInRepo_DerivedSuccessfully(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := uint64(10)
	requestedEpoch := uint64(8)
	epochStart := uint64(1640995000)
	validatorAddress := "0x0000000000000000000000000000000000000abc"

	// Create test data
	validatorSet := createTestValidatorSetWithMultipleValidators(entity.Epoch(requestedEpoch))
	expectedValidator := validatorSet.Validators[1] // Second validator has address 0xabc

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
	req := &apiv1.GetValidatorByAddressRequest{
		Address: validatorAddress,
		Epoch:   &requestedEpoch,
	}

	response, err := setup.handler.GetValidatorByAddress(ctx, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.GetValidator())
	require.Equal(t, expectedValidator.Operator.Hex(), response.GetValidator().GetOperator())
	require.Equal(t, expectedValidator.VotingPower.String(), response.GetValidator().GetVotingPower())
	require.Equal(t, expectedValidator.IsActive, response.GetValidator().GetIsActive())
}

func TestGetValidatorByAddress_UseCurrentEpoch_WhenNoEpochSpecified(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := uint64(10)
	validatorAddress := "0x0000000000000000000000000000000000000123"

	// Create test data
	validatorSet := createTestValidatorSetWithMultipleValidators(entity.Epoch(currentEpoch))

	// Setup mocks - no epoch specified, should use current epoch
	setup.mockEvmClient.EXPECT().GetCurrentEpoch(ctx).Return(currentEpoch, nil)
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

	currentEpoch := uint64(10)
	futureEpoch := uint64(15)
	validatorAddress := "0x0000000000000000000000000000000000000123"

	// Setup mocks
	setup.mockEvmClient.EXPECT().GetCurrentEpoch(ctx).Return(currentEpoch, nil)

	// Execute the method under test
	req := &apiv1.GetValidatorByAddressRequest{
		Address: validatorAddress,
		Epoch:   &futureEpoch,
	}

	response, err := setup.handler.GetValidatorByAddress(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "epoch requested is greater than latest epoch")
}

func TestGetValidatorByAddress_ErrorWhenInvalidAddress(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := uint64(10)
	requestedEpoch := uint64(8)
	invalidAddress := "not-a-valid-address"

	// Setup mocks
	setup.mockEvmClient.EXPECT().GetCurrentEpoch(ctx).Return(currentEpoch, nil)

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

	currentEpoch := uint64(10)
	requestedEpoch := uint64(8)
	nonExistentAddress := "0x0000000000000000000000000000000000000999"

	// Create test data without the requested validator
	validatorSet := createTestValidatorSetWithMultipleValidators(entity.Epoch(requestedEpoch))

	// Setup mocks - validator set found in repository
	setup.mockEvmClient.EXPECT().GetCurrentEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(validatorSet, nil)

	// Execute the method under test
	req := &apiv1.GetValidatorByAddressRequest{
		Address: nonExistentAddress,
		Epoch:   &requestedEpoch,
	}

	response, err := setup.handler.GetValidatorByAddress(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "validator not found for the given address and epoch")
}

func TestGetValidatorByAddress_ErrorWhenGetCurrentEpochFails(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	validatorAddress := "0x0000000000000000000000000000000000000123"
	expectedError := errors.New("failed to get current epoch")

	// Setup mocks
	setup.mockEvmClient.EXPECT().GetCurrentEpoch(ctx).Return(uint64(0), expectedError)

	// Execute the method under test
	req := &apiv1.GetValidatorByAddressRequest{
		Address: validatorAddress,
	}

	response, err := setup.handler.GetValidatorByAddress(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Equal(t, expectedError, err)
}

func TestGetValidatorByAddress_ErrorWhenRepositoryFails(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := uint64(10)
	requestedEpoch := uint64(8)
	validatorAddress := "0x0000000000000000000000000000000000000123"
	expectedError := errors.New("repository connection failed")

	// Setup mocks
	setup.mockEvmClient.EXPECT().GetCurrentEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(entity.ValidatorSet{}, expectedError)

	// Execute the method under test
	req := &apiv1.GetValidatorByAddressRequest{
		Address: validatorAddress,
		Epoch:   &requestedEpoch,
	}

	response, err := setup.handler.GetValidatorByAddress(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "failed to get validator set for epoch")
	require.Contains(t, err.Error(), expectedError.Error())
}

func TestGetValidatorByAddress_ErrorWhenDeriverFails(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := uint64(10)
	requestedEpoch := uint64(8)
	epochStart := uint64(1640995000)
	validatorAddress := "0x0000000000000000000000000000000000000123"
	networkConfig := entity.NetworkConfig{}
	expectedError := errors.New("derivation failed")

	// Setup mocks
	setup.mockEvmClient.EXPECT().GetCurrentEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(entity.ValidatorSet{}, entity.ErrEntityNotFound)
	setup.mockEvmClient.EXPECT().GetEpochStart(ctx, requestedEpoch).Return(epochStart, nil)
	setup.mockEvmClient.EXPECT().GetConfig(ctx, epochStart).Return(networkConfig, nil)
	setup.mockDeriver.EXPECT().GetValidatorSet(ctx, requestedEpoch, networkConfig).Return(entity.ValidatorSet{}, expectedError)

	// Execute the method under test
	req := &apiv1.GetValidatorByAddressRequest{
		Address: validatorAddress,
		Epoch:   &requestedEpoch,
	}

	response, err := setup.handler.GetValidatorByAddress(ctx, req)

	// Assertions
	require.Error(t, err)
	require.Nil(t, response)
	require.Equal(t, expectedError, err)
}

func TestGetValidatorByAddress_InactiveValidator(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := uint64(10)
	requestedEpoch := uint64(8)
	validatorAddress := "0x0000000000000000000000000000000000000789" // This validator is inactive in our test data

	// Create test data
	validatorSet := createTestValidatorSetWithMultipleValidators(entity.Epoch(requestedEpoch))
	expectedValidator := validatorSet.Validators[2] // Third validator has address 0x789 and IsActive: false

	// Setup mocks - validator set found in repository
	setup.mockEvmClient.EXPECT().GetCurrentEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(validatorSet, nil)

	// Execute the method under test
	req := &apiv1.GetValidatorByAddressRequest{
		Address: validatorAddress,
		Epoch:   &requestedEpoch,
	}

	response, err := setup.handler.GetValidatorByAddress(ctx, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.GetValidator())
	require.Equal(t, expectedValidator.Operator.Hex(), response.GetValidator().GetOperator())
	require.Equal(t, expectedValidator.VotingPower.String(), response.GetValidator().GetVotingPower())
	require.False(t, response.GetValidator().GetIsActive()) // Should be inactive
	require.Empty(t, response.GetValidator().GetVaults())   // Inactive validator has no vaults in our test data
}
