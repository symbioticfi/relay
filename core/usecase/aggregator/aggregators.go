package aggregator

import (
	"github.com/symbioticfi/relay/core/entity"
	types "github.com/symbioticfi/relay/core/usecase/aggregator/aggregator-types"
	"github.com/symbioticfi/relay/core/usecase/aggregator/simple"
	"github.com/symbioticfi/relay/core/usecase/aggregator/zk"

	"github.com/go-errors/errors"
)

type Aggregator = types.Aggregator
type Prover = types.Prover

func NewAggregator(verificationType entity.VerificationType, prover Prover) (Aggregator, error) {
	switch verificationType {
	case entity.VerificationTypeZK:
		return zk.NewAggregator(prover), nil
	case entity.VerificationTypeSimple:
		return simple.NewAggregator(), nil
	}

	return nil, errors.New("unsupported verification type")
}
