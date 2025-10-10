package aggregationPolicyTypes

import (
	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type AggregationPolicy interface {
	ShouldAggregate(signatureMap entity.SignatureMap, validatorSet symbiotic.ValidatorSet) bool
}
