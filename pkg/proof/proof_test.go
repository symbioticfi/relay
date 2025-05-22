package proof

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"middleware-offchain/pkg/bls"
	"testing"

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
	// t.Skipf("it works too long, so set skip here. For local debugging can remove this skip")
	// generate valset
	// valset := genValset(10, []int{0, 1, 2})
	valset := mockValset()

	validatorData := normalizeValset(valset)

	messageString := "204e0c470c62e2f8426b236c004b581084dd3aaa935ed3afe24dc37e0d040823"
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

	proofData, err := Prove(proveInput)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Proof:", hex.EncodeToString(proofData.Proof))
	fmt.Println("Commitments:", hex.EncodeToString(proofData.Commitments))
	fmt.Println("CommitmentPok:", hex.EncodeToString(proofData.CommitmentPok))
	fmt.Println("SignersAggVotingPower:", proofData.SignersAggVotingPower.String())
}
