package aggregationPolicy

import (
	"errors"

	lowCostPolicy "github.com/symbioticfi/relay/internal/usecase/aggregation-policy/low-cost"
	lowLatencyPolicy "github.com/symbioticfi/relay/internal/usecase/aggregation-policy/low-latency"
	aggregationPolicyTypes "github.com/symbioticfi/relay/internal/usecase/aggregation-policy/types"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func NewAggregationPolicy(aggregationPolicyType symbiotic.AggregationPolicyType, maxUnsigners uint64) (aggregationPolicyTypes.AggregationPolicy, error) {
	switch aggregationPolicyType {
	case symbiotic.AggregationPolicyLowLatency:
		return lowLatencyPolicy.NewLowLatencyPolicy(), nil
	case symbiotic.AggregationPolicyLowCost:
		return lowCostPolicy.NewLowCostPolicy(maxUnsigners), nil
	}

	return nil, errors.New("unknown aggregation policy type")
}
