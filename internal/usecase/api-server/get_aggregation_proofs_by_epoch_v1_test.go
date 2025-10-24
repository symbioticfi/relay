package api_server

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestGetAggregationProofsByEpoch_Success(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestedEpoch := symbiotic.Epoch(3)

	expectedProofs := []symbiotic.AggregationProof{
		{
			MessageHash: common.Hex2Bytes("hash1"),
			KeyTag:      15,
			Epoch:       requestedEpoch,
			Proof:       common.Hex2Bytes("proof1"),
		},
		{
			MessageHash: common.Hex2Bytes("hash2"),
			KeyTag:      15,
			Epoch:       requestedEpoch,
			Proof:       common.Hex2Bytes("proof2"),
		},
	}

	setup.mockRepo.EXPECT().GetAggregationProofsByEpoch(ctx, requestedEpoch).Return(expectedProofs, nil)

	req := &apiv1.GetAggregationProofsByEpochRequest{
		Epoch: uint64(requestedEpoch),
	}

	response, err := setup.handler.GetAggregationProofsByEpoch(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Len(t, response.GetAggregationProofs(), 2)
	require.Equal(t, []byte(expectedProofs[0].MessageHash), response.GetAggregationProofs()[0].GetMessageHash())
	require.Equal(t, []byte(expectedProofs[0].Proof), response.GetAggregationProofs()[0].GetProof())
	require.Equal(t, []byte(expectedProofs[1].MessageHash), response.GetAggregationProofs()[1].GetMessageHash())
	require.Equal(t, []byte(expectedProofs[1].Proof), response.GetAggregationProofs()[1].GetProof())
}

func TestGetAggregationProofsByEpoch_EmptyResult(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestedEpoch := symbiotic.Epoch(15)

	setup.mockRepo.EXPECT().GetAggregationProofsByEpoch(ctx, requestedEpoch).Return([]symbiotic.AggregationProof{}, nil)

	req := &apiv1.GetAggregationProofsByEpochRequest{
		Epoch: uint64(requestedEpoch),
	}

	response, err := setup.handler.GetAggregationProofsByEpoch(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Empty(t, response.GetAggregationProofs())
}

func TestGetAggregationProofsByEpoch_RepositoryError(t *testing.T) {
	setup := newTestSetup(t)
	ctx := context.Background()

	requestedEpoch := symbiotic.Epoch(3)
	expectedError := errors.New("storage unavailable")

	setup.mockRepo.EXPECT().GetAggregationProofsByEpoch(ctx, requestedEpoch).Return(nil, expectedError)

	req := &apiv1.GetAggregationProofsByEpochRequest{
		Epoch: uint64(requestedEpoch),
	}

	response, err := setup.handler.GetAggregationProofsByEpoch(ctx, req)

	require.Error(t, err)
	require.Nil(t, response)
	require.Contains(t, err.Error(), "failed to get aggregation proofs by epoch")
}
