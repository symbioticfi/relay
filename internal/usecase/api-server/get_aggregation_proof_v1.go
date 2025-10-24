package api_server

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/symbioticfi/relay/internal/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetAggregationProof handles the gRPC GetAggregationProof request
func (h *grpcHandler) GetAggregationProof(ctx context.Context, req *apiv1.GetAggregationProofRequest) (*apiv1.GetAggregationProofResponse, error) {
	proof, err := h.cfg.Repo.GetAggregationProof(ctx, common.HexToHash(req.GetRequestId()))
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil, status.Errorf(codes.NotFound, "aggregation proof for request %s not found", req.GetRequestId())
		}
		return nil, errors.Errorf("failed to get aggregation proof: %w", err)
	}

	return &apiv1.GetAggregationProofResponse{AggregationProof: convertAggregationProofToPB(proof)}, nil
}

func convertAggregationProofToPB(proof symbiotic.AggregationProof) *apiv1.AggregationProof {
	return &apiv1.AggregationProof{
		MessageHash: proof.MessageHash,
		Proof:       proof.Proof,
	}
}
