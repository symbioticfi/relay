package api_server

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetAggregationProof handles the gRPC GetAggregationProof request
func (h *grpcHandler) GetAggregationProof(ctx context.Context, req *v1.GetAggregationProofRequest) (*v1.GetAggregationProofResponse, error) {
	proof, err := h.cfg.Repo.GetAggregationProof(ctx, common.HexToHash(req.GetRequestHash()))
	if err != nil {
		return nil, errors.Errorf("failed to get aggregation proof: %w", err)
	}

	return &v1.GetAggregationProofResponse{
		AggregationProof: &v1.AggregationProof{
			VerificationType: uint32(proof.VerificationType),
			MessageHash:      proof.MessageHash,
			Proof:            proof.Proof,
		},
	}, nil
}
