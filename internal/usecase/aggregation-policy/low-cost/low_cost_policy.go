package lowCostPolicy

import "github.com/symbioticfi/relay/core/entity"

type LowCostPolicy struct {
	maxUnsigners uint64
}

func NewLowCostPolicy(maxUnsigners uint64) *LowCostPolicy {
	return &LowCostPolicy{maxUnsigners: maxUnsigners}
}

func (lcp *LowCostPolicy) ShouldAggregate(signatureMap entity.SignatureMap, validatorSet entity.ValidatorSet) bool {
	if !signatureMap.ThresholdReached(validatorSet.QuorumThreshold) {
		return false
	}

	total := validatorSet.GetTotalActiveValidators()
	signers := signatureMap.SignedValidatorsBitmap.GetCardinality()

	return uint64(total)-signers <= lcp.maxUnsigners
}
