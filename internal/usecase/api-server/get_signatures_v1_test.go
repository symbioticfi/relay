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
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

func TestGetSignatures_Success_ReturnsAllSignatures(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo: mockRepo,
		},
	}

	ctx := context.Background()
	requestID := common.HexToHash("0xabcd")
	requestIDStr := requestID.Hex()

	priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	signatures := []symbiotic.Signature{
		{
			Signature:   []byte("sig1"),
			MessageHash: []byte("hash1"),
			PublicKey:   priv.PublicKey(),
		},
		{
			Signature:   []byte("sig2"),
			MessageHash: []byte("hash2"),
			PublicKey:   priv.PublicKey(),
		},
	}

	mockRepo.EXPECT().
		GetAllSignatures(ctx, requestID).
		Return(signatures, nil)

	req := &apiv1.GetSignaturesRequest{
		RequestId: requestIDStr,
	}

	response, err := handler.GetSignatures(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Len(t, response.GetSignatures(), 2)
	assert.Equal(t, []byte("sig1"), response.GetSignatures()[0].GetSignature())
	assert.Equal(t, []byte("sig2"), response.GetSignatures()[1].GetSignature())
}

func TestGetSignatures_NotFound_ReturnsNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo: mockRepo,
		},
	}

	ctx := context.Background()
	requestID := common.HexToHash("0xnonexistent")
	requestIDStr := requestID.Hex()

	mockRepo.EXPECT().
		GetAllSignatures(ctx, requestID).
		Return(nil, entity.ErrEntityNotFound)

	req := &apiv1.GetSignaturesRequest{
		RequestId: requestIDStr,
	}

	response, err := handler.GetSignatures(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Contains(t, st.Message(), "not found")
}

func TestGetSignatures_RepositoryError_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockrepo(ctrl)
	handler := &grpcHandler{
		cfg: Config{
			Repo: mockRepo,
		},
	}

	ctx := context.Background()
	requestID := common.HexToHash("0x1234")
	requestIDStr := requestID.Hex()
	repoError := assert.AnError

	mockRepo.EXPECT().
		GetAllSignatures(ctx, requestID).
		Return(nil, repoError)

	req := &apiv1.GetSignaturesRequest{
		RequestId: requestIDStr,
	}

	response, err := handler.GetSignatures(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get signatures")
}
