package api_server

import (
	"context"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/symbioticfi/relay/internal/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// GetAggregationStatus handles the gRPC GetAggregationStatus request
func (h *grpcHandler) GetAggregationStatus(ctx context.Context, req *apiv1.GetAggregationStatusRequest) (*apiv1.GetAggregationStatusResponse, error) {
	if h.cfg.Aggregator == nil {
		return nil, entity.ErrNotAnAggregator
	}

	aggregationStatus, err := h.cfg.Aggregator.GetAggregationStatus(ctx, common.HexToHash(req.GetRequestId()))
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil, status.Errorf(codes.NotFound, "aggregation status for request %s not found", req.GetRequestId())
		}
		return nil, err
	}

	operators := lo.Map(aggregationStatus.Validators, func(v symbiotic.Validator, _ int) string {
		return v.Operator.Hex()
	})
	sort.Strings(operators)

	return &apiv1.GetAggregationStatusResponse{
		CurrentVotingPower: aggregationStatus.VotingPower.String(),
		SignerOperators:    operators,
	}, nil
}
