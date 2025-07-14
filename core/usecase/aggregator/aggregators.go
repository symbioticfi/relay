package aggregator

import (
	"middleware-offchain/core/entity"
	types "middleware-offchain/core/usecase/aggregator/aggregator-types"
	"middleware-offchain/core/usecase/aggregator/simple"
	"middleware-offchain/core/usecase/aggregator/zk"

	"github.com/go-errors/errors"
)

func NewAggregator(verificationType entity.VerificationType, prover types.Prover) (types.Aggregator, error) {
	switch verificationType {
	case entity.VerificationTypeZK:
		return zk.NewAggregator(prover), nil
	case entity.VerificationTypeSimple:
		return simple.NewAggregator(), nil
	}

	return nil, errors.New("unsupported verification type")
}
