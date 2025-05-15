package proof

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
)

func genValset(numValidators int, nonSigners []int) []ValidatorData {
	valset := make([]ValidatorData, numValidators)
	for i := 0; i < numValidators; i++ {
		pk := big.NewInt(int64(i + 10))
		valset[i].PrivateKey = *pk
		valset[i].Key = getPubkeyG1(pk)
		valset[i].KeyG2 = getPubkeyG2(pk)
		valset[i].VotingPower = *big.NewInt(100)
		valset[i].IsNonSigner = false
	}

	for _, nonSigner := range nonSigners {
		valset[nonSigner].IsNonSigner = true
	}

	return valset
}

func TestProof(t *testing.T) {
	// t.Skipf("it works too long, so set skip here. For local debugging can remove this skip")
	// generate valset
	valset := genValset(10, []int{0, 1, 2, 3, 4})

	proof, err := Prove(normalizeValset(valset))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Proof:", hex.EncodeToString(proof))
}
