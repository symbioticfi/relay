package lowLatencyPolicy

import "github.com/symbioticfi/relay/core/entity"

type LowLatencyPolicy struct {
}

func NewLowLatencyPolicy() *LowLatencyPolicy {
	return &LowLatencyPolicy{}
}

func (llp *LowLatencyPolicy) ShouldAggregate(signatureMap entity.SignatureMap) bool {
	return signatureMap.ThresholdReached()
}
