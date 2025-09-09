package aggregationPolicy

import (
	"errors"

	"github.com/symbioticfi/relay/core/entity"
	lowCostPolicy "github.com/symbioticfi/relay/internal/usecase/aggregation-policy/low-cost"
	lowLatencyPolicy "github.com/symbioticfi/relay/internal/usecase/aggregation-policy/low-latency"
	aggregationPolicyTypes "github.com/symbioticfi/relay/internal/usecase/aggregation-policy/types"
)

func NewAggregationPolicy(aggregationPolicyType entity.AggregationPolicyType, maxUnsigners uint64) (aggregationPolicyTypes.AggregationPolicy, error) {
	switch aggregationPolicyType {
	case entity.AggregationPolicyLowLatency:
		return lowLatencyPolicy.NewLowLatencyPolicy(), nil
	case entity.AggregationPolicyLowCost:
		return lowCostPolicy.NewLowCostPolicy(maxUnsigners), nil
	}

	return nil, errors.New("unknown aggregation policy type")
}
