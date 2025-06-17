package entity

import (
	"math/big"

	"middleware-offchain/core/entity"
)

type AggregationStatus struct {
	VotingPower *big.Int
	Validators  []entity.Validator
}
