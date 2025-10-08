package lowLatencyPolicy

import "github.com/symbioticfi/relay/symbiotic/entity"

type LowLatencyPolicy struct {
}

func NewLowLatencyPolicy() *LowLatencyPolicy {
	return &LowLatencyPolicy{}
}

func (llp *LowLatencyPolicy) ShouldAggregate(signatureMap entity.SignatureMap, validatorSet entity.ValidatorSet) bool {
	return signatureMap.ThresholdReached(validatorSet.QuorumThreshold)
}
