package aggregationPolicyTypes

import "github.com/symbioticfi/relay/core/entity"

type AggregationPolicy interface {
	ShouldAggregate(signatureMap entity.SignatureMap) bool
}
