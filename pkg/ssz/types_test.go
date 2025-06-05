package ssz

//import (
//	"fmt"
//	"math/big"
//	"testing"
//
//	"middleware-offchain/internal/entity"
//)
//
//func TestValidatorSetHeader(t *testing.T) {
//	v := entity.ValidatorSetHeader{
//		Version:                1,
//		RequiredKeyTag:         15,
//		Epoch:                  new(big.Int).SetInt64(1),
//		CaptureTimestamp:       new(big.Int).SetInt64(1),
//		QuorumThreshold:        new(big.Int).SetInt64(123),
//		TotalActiveVotingPower: new(big.Int).SetInt64(123),
//		ValidatorsSszMRoot:     [32]byte{99, 123},
//		PreviousHeaderHash:     [32]byte{1, 2, 3},
//	}
//
//	fmt.Println(v.AbiEncode())
//
//	e := entity.ExtraData{
//		Key:   [32]byte{1, 2, 3},
//		Value: [32]byte{4, 5, 6},
//	}
//
//	fmt.Println(entity.ExtraDataList{e}.AbiEncode())
//}
