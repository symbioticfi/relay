package types

import (
	"fmt"
	"math/big"
	"testing"
)

func TestValidatorSetHeader(t *testing.T) {
	v := ValidatorSetHeader{
		Version: 1,
		ActiveAggregatedKeys: []Key{{
			Tag:         15,
			PayloadHash: [32]byte{123},
		}},
		TotalActiveVotingPower: new(big.Int).SetInt64(123),
		ValidatorsSszMRoot:     [32]byte{99, 123},
		ExtraData:              []byte{1, 2, 3},
	}

	fmt.Println(v.Encode())
}
