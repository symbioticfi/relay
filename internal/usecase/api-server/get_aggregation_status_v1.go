package api_server

import (
	"context"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetAggregationStatus handles the gRPC GetAggregationStatus request
func (h *grpcHandler) GetAggregationStatus(ctx context.Context, req *v1.GetAggregationStatusRequest) (*v1.GetAggregationStatusResponse, error) {
	if h.cfg.Aggregator == nil {
		return nil, entity.ErrNotAnAggregator
	}

	aggregationStatus, err := h.cfg.Aggregator.GetAggregationStatus(ctx, common.HexToHash(req.GetRequestHash()))
	if err != nil {
		return nil, err
	}

	operators := lo.Map(aggregationStatus.Validators, func(v entity.Validator, _ int) string {
		return v.Operator.Hex()
	})
	sort.Strings(operators)

	return &v1.GetAggregationStatusResponse{
		CurrentVotingPower: aggregationStatus.VotingPower.String(),
		SignerOperators:    operators,
	}, nil
}
