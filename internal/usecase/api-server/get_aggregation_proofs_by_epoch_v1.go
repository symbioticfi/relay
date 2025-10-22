package api_server

import (
	"context"

	"github.com/go-errors/errors"
	"github.com/samber/lo"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/symbiotic/entity"
)

// GetAggregationProofsByEpoch handles the gRPC GetAggregationProofsByEpoch request
func (h *grpcHandler) GetAggregationProofsByEpoch(ctx context.Context, req *apiv1.GetAggregationProofsByEpochRequest) (*apiv1.GetAggregationProofsByEpochResponse, error) {
	epoch := req.GetEpoch()

	proofs, err := h.cfg.Repo.GetAggregationProofsByEpoch(ctx, entity.Epoch(epoch))
	if err != nil {
		return nil, errors.Errorf("failed to get aggregation proofs by epoch: %w", err)
	}

	return &apiv1.GetAggregationProofsByEpochResponse{
		AggregationProofs: lo.Map(proofs, func(proof entity.AggregationProof, _ int) *apiv1.AggregationProof {
			return convertAggregationProofToPB(proof)
		}),
	}, nil
}
