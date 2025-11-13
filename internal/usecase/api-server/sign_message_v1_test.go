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
)

func TestSignMessage_WithRequiredEpoch_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSigner := mocks.NewMocksigner(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Signer: mockSigner,
			Repo:   mockRepo,
		},
	}

	ctx := context.Background()
	keyTag := uint32(15)
	message := []byte("test message")
	requiredEpoch := uint64(10)
	expectedRequestID := common.HexToHash("0x1234")

	expectedSignReq := symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(keyTag),
		Message:       message,
		RequiredEpoch: symbiotic.Epoch(requiredEpoch),
	}

	mockSigner.EXPECT().
		RequestSignature(ctx, expectedSignReq).
		Return(expectedRequestID, nil)

	req := &apiv1.SignMessageRequest{
		KeyTag:        keyTag,
		Message:       message,
		RequiredEpoch: &requiredEpoch,
	}

	response, err := handler.SignMessage(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, expectedRequestID.Hex(), response.GetRequestId())
	assert.Equal(t, requiredEpoch, response.GetEpoch())
}

func TestSignMessage_WithoutRequiredEpoch_UsesLatestEpoch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSigner := mocks.NewMocksigner(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Signer: mockSigner,
			Repo:   mockRepo,
		},
	}

	ctx := context.Background()
	keyTag := uint32(20)
	message := []byte("another message")
	latestEpoch := symbiotic.Epoch(25)
	expectedRequestID := common.HexToHash("0xabcd")

	mockRepo.EXPECT().
		GetLatestValidatorSetEpoch(ctx).
		Return(latestEpoch, nil)

	expectedSignReq := symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(keyTag),
		Message:       message,
		RequiredEpoch: latestEpoch,
	}

	mockSigner.EXPECT().
		RequestSignature(ctx, expectedSignReq).
		Return(expectedRequestID, nil)

	req := &apiv1.SignMessageRequest{
		KeyTag:        keyTag,
		Message:       message,
		RequiredEpoch: nil,
	}

	response, err := handler.SignMessage(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, expectedRequestID.Hex(), response.GetRequestId())
	assert.Equal(t, uint64(latestEpoch), response.GetEpoch())
}

func TestSignMessage_GetLatestEpochFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSigner := mocks.NewMocksigner(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Signer: mockSigner,
			Repo:   mockRepo,
		},
	}

	ctx := context.Background()
	repoError := assert.AnError

	mockRepo.EXPECT().
		GetLatestValidatorSetEpoch(ctx).
		Return(symbiotic.Epoch(0), repoError)

	req := &apiv1.SignMessageRequest{
		KeyTag:        10,
		Message:       []byte("test"),
		RequiredEpoch: nil,
	}

	response, err := handler.SignMessage(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	assert.Equal(t, repoError, err)
}

func TestSignMessage_RequestSignatureFails_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSigner := mocks.NewMocksigner(ctrl)
	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Signer: mockSigner,
			Repo:   mockRepo,
		},
	}

	ctx := context.Background()
	keyTag := uint32(5)
	message := []byte("failing message")
	requiredEpoch := uint64(8)
	signerError := assert.AnError

	expectedSignReq := symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(keyTag),
		Message:       message,
		RequiredEpoch: symbiotic.Epoch(requiredEpoch),
	}

	mockSigner.EXPECT().
		RequestSignature(ctx, expectedSignReq).
		Return(common.Hash{}, signerError)

	req := &apiv1.SignMessageRequest{
		KeyTag:        keyTag,
		Message:       message,
		RequiredEpoch: &requiredEpoch,
	}

	response, err := handler.SignMessage(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	assert.Equal(t, signerError, err)
}
