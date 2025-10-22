package api_server

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

func TestGetSignaturesByEpoch_Success(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestedEpoch := symbiotic.Epoch(5)

	priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	expectedSignatures := []symbiotic.Signature{
		{
			MessageHash: common.Hex2Bytes("abcd1234"),
			KeyTag:      15,
			Epoch:       requestedEpoch,
			Signature:   common.Hex2Bytes("sig1"),
			PublicKey:   priv.PublicKey(),
		},
		{
			MessageHash: common.Hex2Bytes("efgh5678"),
			KeyTag:      15,
			Epoch:       requestedEpoch,
			Signature:   common.Hex2Bytes("sig2"),
			PublicKey:   priv.PublicKey(),
		},
	}

	setup.mockRepo.EXPECT().GetSignaturesByEpoch(ctx, requestedEpoch).Return(expectedSignatures, nil)

	req := &apiv1.GetSignaturesByEpochRequest{
		Epoch: uint64(requestedEpoch),
	}

	response, err := setup.handler.GetSignaturesByEpoch(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Len(t, response.GetSignatures(), 2)
	require.Equal(t, []byte(expectedSignatures[0].MessageHash), response.GetSignatures()[0].GetMessageHash())
	require.Equal(t, []byte(expectedSignatures[1].MessageHash), response.GetSignatures()[1].GetMessageHash())
}

func TestGetSignaturesByEpoch_EmptyResult(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestedEpoch := symbiotic.Epoch(10)

	setup.mockRepo.EXPECT().GetSignaturesByEpoch(ctx, requestedEpoch).Return([]symbiotic.Signature{}, nil)

	req := &apiv1.GetSignaturesByEpochRequest{
		Epoch: uint64(requestedEpoch),
	}

	response, err := setup.handler.GetSignaturesByEpoch(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Empty(t, response.GetSignatures())
}

func TestGetSignaturesByEpoch_RepositoryError(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestedEpoch := symbiotic.Epoch(5)
	expectedError := errors.New("database connection failed")

	setup.mockRepo.EXPECT().GetSignaturesByEpoch(ctx, requestedEpoch).Return(nil, expectedError)

	req := &apiv1.GetSignaturesByEpochRequest{
		Epoch: uint64(requestedEpoch),
	}

	response, err := setup.handler.GetSignaturesByEpoch(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "failed to get signatures by epoch")
}
