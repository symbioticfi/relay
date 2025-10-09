package aggregationPolicyTypes

import "github.com/symbioticfi/relay/symbiotic/entity"

type AggregationPolicy interface {
	ShouldAggregate(signatureMap entity.SignatureMap, validatorSet entity.ValidatorSet) bool
}
