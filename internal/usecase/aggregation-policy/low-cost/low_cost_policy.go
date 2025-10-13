package lowCostPolicy

import (
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"

	"github.com/symbioticfi/relay/internal/entity"
)

type LowCostPolicy struct {
	maxUnsigners uint64
}

func NewLowCostPolicy(maxUnsigners uint64) *LowCostPolicy {
	return &LowCostPolicy{maxUnsigners: maxUnsigners}
}

func (lcp *LowCostPolicy) ShouldAggregate(signatureMap entity.SignatureMap, validatorSet symbiotic.ValidatorSet) bool {
	if !signatureMap.ThresholdReached(validatorSet.QuorumThreshold) {
		return false
	}

	total := validatorSet.GetTotalActiveValidators()
	signers := signatureMap.SignedValidatorsBitmap.GetCardinality()

	return uint64(total)-signers <= lcp.maxUnsigners
}
