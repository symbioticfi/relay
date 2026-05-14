package api_server

import (
	"context"

	"github.com/go-errors/errors"
	"github.com/samber/lo"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/symbiotic/entity"
)

func (h *grpcHandler) GetAggregationProofsByEpoch(ctx context.Context, req *apiv1.GetAggregationProofsByEpochRequest) (*apiv1.GetAggregationProofsByEpochResponse, error) {
	pageSize := clampPageSize(int(req.GetPageSize()), defaultListPageSize, maxListPageSize)

	from, err := decodeCursor(req.GetFrom())
	if err != nil {
		return nil, err
	}

	proofs, next, err := h.cfg.Repo.GetAggregationProofsByEpoch(ctx, entity.Epoch(req.GetEpoch()), pageSize, from)
	if err != nil {
		if e := asCursorErr(err); e != nil {
			return nil, e
		}
		return nil, errors.Errorf("failed to get aggregation proofs by epoch: %w", err)
	}

	return &apiv1.GetAggregationProofsByEpochResponse{
		AggregationProofs: lo.Map(proofs, func(proof entity.AggregationProof, _ int) *apiv1.AggregationProof {
			return convertAggregationProofToPB(proof)
		}),
		NextFrom: encodeCursor(next),
	}, nil
}
