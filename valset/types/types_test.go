package types

import (
	"fmt"
	"math/big"
	"testing"

	"middleware-offchain/internal/entity"
)

func TestValidatorSetHeader(t *testing.T) {
	v := entity.ValidatorSetHeader{
		Version: 1,
		ActiveAggregatedKeys: []entity.Key{{
			Tag:     15,
			Payload: []byte{1, 2, 3},
		}},
		TotalActiveVotingPower: new(big.Int).SetInt64(123),
		ValidatorsSszMRoot:     [32]byte{99, 123},
		ExtraData:              []byte{1, 2, 3},
	}

	fmt.Println(v.Encode())
}
