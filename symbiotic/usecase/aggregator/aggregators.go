package aggregator

import (
	"github.com/symbioticfi/relay/symbiotic/entity"
	types "github.com/symbioticfi/relay/symbiotic/usecase/aggregator/aggregator-types"
	"github.com/symbioticfi/relay/symbiotic/usecase/aggregator/blsBn254Simple"
	"github.com/symbioticfi/relay/symbiotic/usecase/aggregator/blsBn254ZK"

	"github.com/go-errors/errors"
)

type Aggregator = types.Aggregator
type Prover = types.Prover

func NewAggregator(verificationType entity.VerificationType, prover Prover) (Aggregator, error) {
	switch verificationType {
	case entity.VerificationTypeBlsBn254ZK:
		return blsBn254ZK.NewAggregator(prover)
	case entity.VerificationTypeBlsBn254Simple:
		return blsBn254Simple.NewAggregator()
	}

	return nil, errors.New("unsupported verification type")
}
