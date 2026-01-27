package api_server

import (
	"context"
	"testing"

	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/internal/usecase/api-server/mocks"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestGetLocalValidator_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)

	handler := &grpcHandler{
		cfg: Config{
			Repo:        mockRepo,
			KeyProvider: mockKeyProvider,
		},
	}

	ctx := context.Background()
	requestedEpoch := symbiotic.Epoch(5)
	currentEpoch := symbiotic.Epoch(10)
	localKey := symbiotic.CompactPublicKey("local-validator-key")

	validatorSet := createTestValidatorSetWithMultipleValidators(requestedEpoch)
	validatorSet.Validators[0].Keys[0].Payload = localKey
	expectedValidator := validatorSet.Validators[0]

	mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
	mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(validatorSet, nil)
	mockKeyProvider.EXPECT().GetOnchainKeyFromCache(symbiotic.KeyTag(15)).Return(localKey, nil)

	req := &apiv1.GetLocalValidatorRequest{
		Epoch: (*uint64)(&requestedEpoch),
	}

	response, err := handler.GetLocalValidator(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.GetValidator())
	require.Equal(t, expectedValidator.Operator.Hex(), response.GetValidator().GetOperator())
	require.Equal(t, expectedValidator.VotingPower.String(), response.GetValidator().GetVotingPower())
}

func TestGetLocalValidator_UseCurrentEpoch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)

	handler := &grpcHandler{
		cfg: Config{
			Repo:        mockRepo,
			KeyProvider: mockKeyProvider,
		},
	}

	ctx := context.Background()
	currentEpoch := symbiotic.Epoch(10)
	localKey := symbiotic.CompactPublicKey("local-validator-key")

	validatorSet := createTestValidatorSetWithMultipleValidators(currentEpoch)
	validatorSet.Validators[0].Keys[0].Payload = localKey

	mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
	mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, currentEpoch).Return(validatorSet, nil)
	mockKeyProvider.EXPECT().GetOnchainKeyFromCache(symbiotic.KeyTag(15)).Return(localKey, nil)

	req := &apiv1.GetLocalValidatorRequest{}

	response, err := handler.GetLocalValidator(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.GetValidator())
}

func TestGetLocalValidator_ErrorWhenEpochFromFuture(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)

	handler := &grpcHandler{
		cfg: Config{
			Repo:        mockRepo,
			KeyProvider: mockKeyProvider,
		},
	}

	ctx := context.Background()
	currentEpoch := symbiotic.Epoch(10)
	futureEpoch := uint64(15)

	mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)

	req := &apiv1.GetLocalValidatorRequest{
		Epoch: &futureEpoch,
	}

	response, err := handler.GetLocalValidator(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "is greater than latest epoch")
}

func TestGetLocalValidator_ErrorWhenKeyProviderFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)

	handler := &grpcHandler{
		cfg: Config{
			Repo:        mockRepo,
			KeyProvider: mockKeyProvider,
		},
	}

	ctx := context.Background()
	requestedEpoch := symbiotic.Epoch(5)
	currentEpoch := symbiotic.Epoch(10)
	expectedError := errors.New("key not found in cache")

	validatorSet := createTestValidatorSetWithMultipleValidators(requestedEpoch)

	mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
	mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(validatorSet, nil)
	mockKeyProvider.EXPECT().GetOnchainKeyFromCache(symbiotic.KeyTag(15)).Return(symbiotic.CompactPublicKey(""), expectedError)

	req := &apiv1.GetLocalValidatorRequest{
		Epoch: (*uint64)(&requestedEpoch),
	}

	response, err := handler.GetLocalValidator(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "failed to get onchain key from cache")
}

func TestGetLocalValidator_ErrorWhenValidatorNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)

	handler := &grpcHandler{
		cfg: Config{
			Repo:        mockRepo,
			KeyProvider: mockKeyProvider,
		},
	}

	ctx := context.Background()
	requestedEpoch := symbiotic.Epoch(5)
	currentEpoch := symbiotic.Epoch(10)
	localKey := symbiotic.CompactPublicKey("non-existent-key")

	validatorSet := createTestValidatorSetWithMultipleValidators(requestedEpoch)

	mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
	mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(validatorSet, nil)
	mockKeyProvider.EXPECT().GetOnchainKeyFromCache(symbiotic.KeyTag(15)).Return(localKey, nil)

	req := &apiv1.GetLocalValidatorRequest{
		Epoch: (*uint64)(&requestedEpoch),
	}

	response, err := handler.GetLocalValidator(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "local validator not found")
}

func TestGetLocalValidator_ErrorWhenRepositoryFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	mockKeyProvider := mocks.NewMockkeyProvider(ctrl)

	handler := &grpcHandler{
		cfg: Config{
			Repo:        mockRepo,
			KeyProvider: mockKeyProvider,
		},
	}

	ctx := context.Background()
	requestedEpoch := symbiotic.Epoch(5)
	currentEpoch := symbiotic.Epoch(10)
	expectedError := errors.New("database connection failed")

	mockRepo.EXPECT().GetLatestValidatorSetEpoch(ctx).Return(currentEpoch, nil)
	mockRepo.EXPECT().GetValidatorSetByEpoch(ctx, requestedEpoch).Return(symbiotic.ValidatorSet{}, expectedError)

	req := &apiv1.GetLocalValidatorRequest{
		Epoch: (*uint64)(&requestedEpoch),
	}

	response, err := handler.GetLocalValidator(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
}
