package proof

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"middleware-offchain/pkg/bls"

	"github.com/consensys/gnark-crypto/ecc/bn254"
)

func genValset(numValidators int, nonSigners []int) []ValidatorData {
	valset := make([]ValidatorData, numValidators)
	for i := 0; i < numValidators; i++ {
		pk := big.NewInt(int64(i + 10))
		valset[i].PrivateKey = pk
		valset[i].Key = getPubkeyG1(pk)
		valset[i].KeyG2 = getPubkeyG2(pk)
		valset[i].VotingPower = big.NewInt(100)
		valset[i].IsNonSigner = false
	}

	for _, nonSigner := range nonSigners {
		valset[nonSigner].IsNonSigner = true
	}

	return valset
}

func TestProof(t *testing.T) {
	t.Skipf("it works too long, so set skip here. For local debugging can remove this skip")
	// generate valset
	valset := genValset(10, []int{0, 1, 2})

	validatorData := normalizeValset(valset)

	message := big.NewInt(101).Bytes()
	messageG1, err := bls.HashToG1(message)
	if err != nil {
		t.Fatal(err)
	}
	messageG1Bn254 := bn254.G1Affine{X: messageG1.X, Y: messageG1.Y}

	aggSignature, aggKeyG2, _ := getAggSignature(messageG1Bn254, &validatorData)

	proveInput := ProveInput{
		ValidatorData:   validatorData,
		Message:         message,
		Signature:       *aggSignature,
		SignersAggKeyG2: *aggKeyG2,
	}

	proofData, err := Prove(proveInput)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Proof:", hex.EncodeToString(proofData.Proof))
	fmt.Println("Commitments:", hex.EncodeToString(proofData.Commitments))
	fmt.Println("CommitmentPok:", hex.EncodeToString(proofData.CommitmentPok))
	fmt.Println("NonSignersAggVotingPower:", proofData.NonSignersAggVotingPower.String())
}
