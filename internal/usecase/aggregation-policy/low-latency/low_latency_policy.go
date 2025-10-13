package lowLatencyPolicy

import (
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type LowLatencyPolicy struct {
}

func NewLowLatencyPolicy() *LowLatencyPolicy {
	return &LowLatencyPolicy{}
}

func (llp *LowLatencyPolicy) ShouldAggregate(signatureMap entity.SignatureMap, validatorSet symbiotic.ValidatorSet) bool {
	return signatureMap.ThresholdReached(validatorSet.QuorumThreshold)
}
