package aggregator

import (
	"math/big"
	"middleware-offchain/core/entity"
	"middleware-offchain/core/usecase/aggregator/simple"
	"middleware-offchain/core/usecase/aggregator/zk"
	"middleware-offchain/core/usecase/crypto"
	proof2 "middleware-offchain/pkg/proof"
	"testing"

	crypto2 "github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
)

func TestSimpleAggregator(t *testing.T) {
	agg := simple.NewAggregator()
	valset, signatures, keyTag := genCorrectTest(10, []int{1, 2, 3})

	proof, err := agg.Aggregate(valset, keyTag, signatures[0].MessageHash, signatures)
	if err != nil {
		panic(err)
	}

	success, err := agg.Verify(valset, keyTag, proof)
	if err != nil {
		panic(err)
	}
	if !success {
		t.Fatal(errors.New("verification failed"))
	}
}

func TestInvalidSimpleAggregator(t *testing.T) {
	agg := simple.NewAggregator()
	valset, signatures, keyTag := genCorrectTest(10, []int{1, 2, 3})
	someKey, err := crypto.GeneratePrivateKey(keyTag.Type())
	if err != nil {
		panic(err)
	}
	signatures[0].Signature, _, err = someKey.Sign([]byte("message"))
	if err != nil {
		panic(err)
	}

	proof, err := agg.Aggregate(valset, keyTag, signatures[0].MessageHash, signatures)
	if err != nil {
		panic(err)
	}

	success, err := agg.Verify(valset, keyTag, proof)
	if err == nil {
		t.Fatal(errors.New("verification passed"))
	}
	if success {
		t.Fatal(errors.New("verification passed"))
	}
}

func TestZkAggregator(t *testing.T) {
	t.Skipf("it works too long, so set skip here. For local debugging can remove this skip")
	prover := proof2.NewZkProver()
	agg := zk.NewAggregator(prover)
	valset, signatures, keyTag := genCorrectTest(10, []int{1, 2, 3})
	proof, err := agg.Aggregate(valset, keyTag, signatures[0].MessageHash, signatures)
	if err != nil {
		panic(err)
	}

	success, err := agg.Verify(valset, keyTag, proof)
	if err != nil {
		panic(err)
	}

	if !success {
		t.Fatal(errors.New("verification failed"))
	}
}

func genCorrectTest(numValidators int, nonSigners []int) (entity.ValidatorSet, []entity.SignatureExtended, entity.KeyTag) {
	valset := entity.ValidatorSet{}
	signatures := make([]entity.SignatureExtended, 0)
	pks := make([]crypto.PrivateKey, numValidators)
	msg := []byte("message")
	keyTag := entity.KeyTag(1)
	valset.Validators = make([]entity.Validator, numValidators)
	for i := 0; i < numValidators; i++ {
		var err error
		pks[i], err = crypto.GeneratePrivateKey(keyTag.Type())
		if err != nil {
			panic(err)
		}

		valset.Validators[i].Keys = []entity.ValidatorKey{
			{
				Tag:     keyTag,
				Payload: pks[i].PublicKey().OnChain(),
			},
		}

		pk2, err := crypto.GeneratePrivateKey(entity.KeyTypeEcdsaSecp256k1)
		if err != nil {
			panic(err)
		}

		pkEcdsa, err := crypto2.ToECDSA(big.NewInt(0).SetBytes(pk2.Bytes()).FillBytes(make([]byte, 32)))
		if err != nil {
			panic(err)
		}
		valset.Validators[i].Operator = crypto2.PubkeyToAddress(pkEcdsa.PublicKey)
		valset.Validators[i].IsActive = true
		valset.Validators[i].VotingPower = entity.ToVotingPower(big.NewInt(100))
	}

	valset.QuorumThreshold = entity.ToVotingPower(big.NewInt(int64(100 * (numValidators - len(nonSigners)))))
	valset.RequiredKeyTag = keyTag
	nonSignersMap := make(map[int]bool)
	for i := 0; i < len(nonSigners); i++ {
		nonSignersMap[nonSigners[i]] = true
	}

	for i := 0; i < numValidators; i++ {
		if _, ok := nonSignersMap[i]; ok {
			continue
		}
		sig, msgHash, err := pks[i].Sign(msg)
		if err != nil {
			panic(err)
		}
		signatures = append(signatures, entity.SignatureExtended{
			MessageHash: msgHash,
			Signature:   sig,
			PublicKey:   pks[i].PublicKey().Raw(),
		})
	}

	return valset, signatures, keyTag
}
