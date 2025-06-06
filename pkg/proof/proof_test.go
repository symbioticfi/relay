package proof

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

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

//nolint:unused // will be used later
func mockValset() []ValidatorData {
	pks := []string{
		"87191036493798670866484781455694320176667203290824056510541300741498740913410",
		"26972876870930381973856869753776124637336739336929668162870464864826929175089",
		"11008377096554045051122023680185802911050337017631086444859313200352654461863",
	}

	valset := make([]ValidatorData, len(pks))
	for i := 0; i < len(pks); i++ {
		pk, ok := new(big.Int).SetString(pks[i], 10)
		if !ok {
			panic(errors.New("failed to convert pk to big.Int"))
		}
		valset[i].PrivateKey = pk
		valset[i].Key = getPubkeyG1(pk)
		valset[i].KeyG2 = getPubkeyG2(pk)
		valset[i].VotingPower = big.NewInt(10000000000000)
		valset[i].IsNonSigner = false
	}

	return valset
}

func TestProof(t *testing.T) {
	//t.Skipf("it works too long, so set skip here. For local debugging can remove this skip")

	startTime := time.Now()
	prover := NewZkProver()
	fmt.Printf("prover initialation took %v\n", time.Since(startTime))

	// generate valset
	valset := genValset(11, []int{})
	// valset := mockValset()

	validatorData := NormalizeValset(valset)

	messageString := "658bc250cfe17f8ad77a5f5d92afb6e9316088b5c89c6df2db63785116b22948"
	message, err := hex.DecodeString(messageString)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("message:", hex.EncodeToString(message))
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

	startTime = time.Now()
	proofData, pubInpHash, err := prover.Prove(proveInput)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("proving took %v\n", time.Since(startTime))

	fmt.Println("Proof:", hex.EncodeToString(proofData.Proof))
	fmt.Println("Commitments:", hex.EncodeToString(proofData.Commitments))
	fmt.Println("CommitmentPok:", hex.EncodeToString(proofData.CommitmentPok))
	fmt.Println("SignersAggVotingPower:", proofData.SignersAggVotingPower.String())

	startTime = time.Now()
	res, err := prover.Verify(len(validatorData), pubInpHash, proofData.Marshall())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("verification took %v\n", time.Since(startTime))

	if !res {
		t.Fatal("failed to verify")
	}
}
