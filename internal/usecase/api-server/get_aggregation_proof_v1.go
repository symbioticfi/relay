package api_server

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetAggregationProof handles the gRPC GetAggregationProof request
func (h *grpcHandler) GetAggregationProof(ctx context.Context, req *apiv1.GetAggregationProofRequest) (*apiv1.GetAggregationProofResponse, error) {
	proof, err := h.cfg.Repo.GetAggregationProof(ctx, common.HexToHash(req.GetSignatureTargetId()))
	if err != nil {
		return nil, errors.Errorf("failed to get aggregation proof: %w", err)
	}

	return &apiv1.GetAggregationProofResponse{
		AggregationProof: &apiv1.AggregationProof{
			MessageHash: proof.MessageHash,
			Proof:       proof.Proof,
		},
	}, nil
}
