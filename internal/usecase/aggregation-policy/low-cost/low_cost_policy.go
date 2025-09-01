package lowCostPolicy

import "github.com/symbioticfi/relay/core/entity"

type LowCostPolicy struct {
	maxUnsigners uint64
}

func NewLowCostPolicy(maxUnsigners uint64) *LowCostPolicy {
	return &LowCostPolicy{maxUnsigners: maxUnsigners}
}

func (lcp *LowCostPolicy) ShouldAggregate(signatureMap entity.SignatureMap) bool {
	if !signatureMap.ThresholdReached() {
		return false
	}

	total := len(signatureMap.ActiveValidatorsMap)
	signers := len(signatureMap.IsPresent)

	return uint64(total-signers) <= lcp.maxUnsigners
}
