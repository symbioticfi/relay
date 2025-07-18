package apiApp

import (
	"context"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"github.com/symbiotic/relay/core/entity"
	"github.com/symbiotic/relay/internal/gen/api"
)

func (h *handler) GetAggregationStatusGet(ctx context.Context, params api.GetAggregationStatusGetParams) (*api.AggregationStatus, error) {
	if h.cfg.Aggregator == nil {
		return nil, errors.New(entity.ErrNotAnAggregator)
	}

	aggregationStatus, err := h.cfg.Aggregator.GetAggregationStatus(ctx, common.HexToHash(params.RequestHash))
	if err != nil {
		return nil, err
	}

	operators := lo.Map(aggregationStatus.Validators, func(v entity.Validator, _ int) string {
		return v.Operator.Hex()
	})
	sort.Strings(operators)

	return &api.AggregationStatus{
		CurrentVotingPower: aggregationStatus.VotingPower.String(),
		SignerOperators:    operators,
	}, nil
}
