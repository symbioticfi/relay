package api_server

import (
	"context"
	"testing"

	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestGetValidatorByKey_Success(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestedEpoch := symbiotic.Epoch(5)
	currentEpoch := symbiotic.Epoch(10)
	keyTag := uint32(15)
	onChainKey := []byte("test-public-key")

	validatorSet := createTestValidatorSetWithMultipleValidators(requestedEpoch)
	validatorSet.Validators[0].Keys[0].Payload = onChainKey
	expectedValidator := validatorSet.Validators[0]

	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(validatorSet, nil)

	req := &apiv1.GetValidatorByKeyRequest{
		Epoch:      (*uint64)(&requestedEpoch),
		KeyTag:     keyTag,
		OnChainKey: onChainKey,
	}

	response, err := setup.handler.GetValidatorByKey(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.GetValidator())
	require.Equal(t, expectedValidator.Operator.Hex(), response.GetValidator().GetOperator())
	require.Equal(t, expectedValidator.VotingPower.String(), response.GetValidator().GetVotingPower())
}

func TestGetValidatorByKey_UseCurrentEpoch(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := symbiotic.Epoch(10)
	keyTag := uint32(15)
	onChainKey := []byte("test-public-key")

	validatorSet := createTestValidatorSetWithMultipleValidators(currentEpoch)
	validatorSet.Validators[0].Keys[0].Payload = onChainKey

	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, currentEpoch).Return(validatorSet, nil)

	req := &apiv1.GetValidatorByKeyRequest{
		KeyTag:     keyTag,
		OnChainKey: onChainKey,
	}

	response, err := setup.handler.GetValidatorByKey(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.GetValidator())
}

func TestGetValidatorByKey_ErrorWhenEpochFromFuture(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := symbiotic.Epoch(10)
	futureEpoch := uint64(15)
	keyTag := uint32(15)
	onChainKey := []byte("test-public-key")

	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)

	req := &apiv1.GetValidatorByKeyRequest{
		Epoch:      &futureEpoch,
		KeyTag:     keyTag,
		OnChainKey: onChainKey,
	}

	response, err := setup.handler.GetValidatorByKey(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "is greater than latest epoch")
}

func TestGetValidatorByKey_ErrorWhenKeyTagIsZero(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := symbiotic.Epoch(10)
	requestedEpoch := uint64(5)
	keyTag := uint32(0)
	onChainKey := []byte("test-public-key")

	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)

	req := &apiv1.GetValidatorByKeyRequest{
		Epoch:      &requestedEpoch,
		KeyTag:     keyTag,
		OnChainKey: onChainKey,
	}

	response, err := setup.handler.GetValidatorByKey(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "key tag must be positive")
}

func TestGetValidatorByKey_ErrorWhenOnChainKeyIsEmpty(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	currentEpoch := symbiotic.Epoch(10)
	requestedEpoch := uint64(5)
	keyTag := uint32(15)

	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)

	req := &apiv1.GetValidatorByKeyRequest{
		Epoch:      &requestedEpoch,
		KeyTag:     keyTag,
		OnChainKey: []byte{},
	}

	response, err := setup.handler.GetValidatorByKey(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "on chain key is empty")
}

func TestGetValidatorByKey_ErrorWhenValidatorNotFound(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestedEpoch := symbiotic.Epoch(5)
	currentEpoch := symbiotic.Epoch(10)
	keyTag := uint32(15)
	onChainKey := []byte("non-existent-key")

	validatorSet := createTestValidatorSetWithMultipleValidators(requestedEpoch)

	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(validatorSet, nil)

	req := &apiv1.GetValidatorByKeyRequest{
		Epoch:      (*uint64)(&requestedEpoch),
		KeyTag:     keyTag,
		OnChainKey: onChainKey,
	}

	response, err := setup.handler.GetValidatorByKey(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "validator not found")
}

func TestGetValidatorByKey_ErrorWhenRepositoryFails(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestedEpoch := symbiotic.Epoch(5)
	currentEpoch := symbiotic.Epoch(10)
	keyTag := uint32(15)
	onChainKey := []byte("test-public-key")
	expectedError := errors.New("database connection failed")

	setup.mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
	setup.mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(symbiotic.ValidatorSet{}, expectedError)

	req := &apiv1.GetValidatorByKeyRequest{
		Epoch:      (*uint64)(&requestedEpoch),
		KeyTag:     keyTag,
		OnChainKey: onChainKey,
	}

	response, err := setup.handler.GetValidatorByKey(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
}
