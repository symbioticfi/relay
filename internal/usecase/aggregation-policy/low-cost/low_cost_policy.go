package lowCostPolicy

import "github.com/symbioticfi/relay/core/entity"

type LowCostPolicy struct {
	maxUnsigners uint64
}

func NewLowCostPolicy(maxUnsigners uint64) *LowCostPolicy {
	return &LowCostPolicy{maxUnsigners: maxUnsigners}
}

func (lcp *LowCostPolicy) ShouldAggregate(validatorMap entity.ValidatorMap) bool {
	if !validatorMap.ThresholdReached() {
		return false
	}

	total := len(validatorMap.ActiveValidatorsMap)
	signers := len(validatorMap.IsPresent)

	return uint64(total-signers) <= lcp.maxUnsigners
}
