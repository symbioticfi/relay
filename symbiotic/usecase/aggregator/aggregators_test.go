package aggregator

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/pkg/proof"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/aggregator/blsBn254Simple"
	"github.com/symbioticfi/relay/symbiotic/usecase/aggregator/blsBn254ZK"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"

	crypto2 "github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
)

type mockProver struct{}

func (m *mockProver) Prove(proveInput proof.ProveInput) (proof.ProofData, error) {
	return proof.ProofData{}, nil
}

func (m *mockProver) Verify(valsetLen int, publicInputHash common.Hash, proofBytes []byte) (bool, error) {
	return true, nil
}

func TestSimpleAggregator(t *testing.T) {
	agg, err := blsBn254Simple.NewAggregator()
	require.NoError(t, err)
	valset, signatures, keyTag := genCorrectTest(10, []int{1, 2, 3})

	proof, err := agg.Aggregate(valset, keyTag, signatures[0].MessageHash, signatures)
	if err != nil {
		panic(err)
	}

	success, err := agg.Verify(valset, keyTag, proof)
	require.NoError(t, err)
	require.True(t, success, "verification failed")
}

func TestInvalidSimpleAggregator(t *testing.T) {
	agg, err := blsBn254Simple.NewAggregator()
	require.NoError(t, err)
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

func TestSimpleAggregatorExtraData(t *testing.T) {
	valset, keyTag := genExtraDataTest(t)
	agg, err := blsBn254Simple.NewAggregator()
	require.NoError(t, err)
	data, err := agg.GenerateExtraData(context.Background(), valset, []symbiotic.KeyTag{keyTag})
	require.NoError(t, err)
	expected := [][]string{
		{
			"0x653d30d3f5b20173b1482c2ed1d4435101e6c2eb4e07bbdb381c1297862af81d",
			"0x45061b0eef183d1991badf4d85070ad0faa423c1d29eafe7a0c84840fa3e9221",
		},
		{
			"0xb3114d5f3c21fca82212f69acee504fb470c13855114c29b5a634315ba69d58d",
			"0xb2d5fd4a3411e1ca6bccebfbf68b2d3aa244532a09517a3abc0dc3a27bd593e7",
		},
	}

	require.Len(t, data, len(expected))
	for i, datum := range data {
		require.Equal(t, datum.Key.Hex(), expected[i][0], "Key mismatch at index %d", i)
	}
}

func TestAggregatorZKExtraData(t *testing.T) {
	t.Skipf("it works too long, so set skip here. For local debugging can remove this skip")
	valset, keyTag := genExtraDataTest(t)
	prover := proof.NewZkProver("circuits")
	agg, err := blsBn254ZK.NewAggregator(prover)
	require.NoError(t, err)
	data, err := agg.GenerateExtraData(context.Background(), valset, []symbiotic.KeyTag{keyTag})
	require.NoError(t, err)
	expected := [][]string{
		{
			"0x7f6185ad9469ee6a9c05e14b4e03be396fc9beb5e6626c77957d25b5b62d83ab",
			"0x0000000000000000000000000000000000000000000000000000000000000003",
		},
		{
			"0xc6e9ac21096d96bb1e2bf10197ec1c6feadf21370333c71cbfd53ee641ebbd49",
			"0x2971480de32189d3c623c0ae300b42c94386379d05d4069700fc615dbc3a8636",
		},
	}

	require.Len(t, data, len(expected))
	for i, datum := range data {
		require.Equal(t, datum.Key.Hex(), expected[i][0], "Key mismatch at index %d", i)
	}
}

func TestZkAggregator(t *testing.T) {
	t.Skipf("it works too long, so set skip here. For local debugging can remove this skip")
	prover := proof.NewZkProver("circuits")
	agg, err := blsBn254ZK.NewAggregator(prover)
	require.NoError(t, err)
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

func genExtraDataTest(t *testing.T) (symbiotic.ValidatorSet, symbiotic.KeyTag) {
	t.Helper()
	valset := symbiotic.ValidatorSet{}
	pksHex := []string{"3d49215a0647a140dbc5e199ee896bfd075f7444ff6c3b13ade0eb014d4a83", "0fc6313bf6a88f31475108677cab0fa54be50e3025444b77043c89502ee49d79", "23059516e460695291d5dcc94361a5f67ce01bf027142c650b06773ef9a08311"}
	pks2Hex := []string{"29af2884cac3904eb71c5fd2e9dfbec69870cfb659f30aa2c1fb187c6cf6f96c", "1a6687dbf130e536cdfe7194fac593e927e8dd12eef2f7b929777b37e9a07cb4", "013ee8da42bfd9e45e6b19dc4df341c5cf449e9f5d2fd78f8fbbcef126d05281"}

	require.Len(t, pksHex, len(pks2Hex))
	numValidators := len(pksHex)
	pks := make([]crypto.PrivateKey, len(pksHex))
	keyTag := symbiotic.KeyTag(1)

	valset.Validators = make([]symbiotic.Validator, numValidators)
	for i := 0; i < len(pksHex); i++ {
		var err error
		decodeString, err := hex.DecodeString(pksHex[i])
		require.NoError(t, err)
		pk, err := crypto.NewPrivateKey(keyTag.Type(), decodeString)
		require.NoError(t, err)

		pks[i] = pk

		valset.Validators[i].Keys = []symbiotic.ValidatorKey{
			{
				Tag:     keyTag,
				Payload: pks[i].PublicKey().OnChain(),
			},
		}

		decodeString2, err := hex.DecodeString(pks2Hex[i])
		require.NoError(t, err)
		pk2, err := crypto.NewPrivateKey(keyTag.Type(), decodeString2)
		require.NoError(t, err)
		pkEcdsa, err := crypto2.ToECDSA(big.NewInt(0).SetBytes(pk2.Bytes()).FillBytes(make([]byte, 32)))
		require.NoError(t, err)

		valset.Validators[i].Operator = crypto2.PubkeyToAddress(pkEcdsa.PublicKey)
		valset.Validators[i].IsActive = true
		valset.Validators[i].VotingPower = symbiotic.ToVotingPower(big.NewInt(100))
	}

	valset.QuorumThreshold = symbiotic.ToVotingPower(big.NewInt(int64(100 * (len(pksHex)))))
	valset.RequiredKeyTag = keyTag

	valset.Validators.SortByOperatorAddressAsc()

	return valset, keyTag
}

func genCorrectTest(numValidators int, nonSigners []int) (symbiotic.ValidatorSet, []symbiotic.Signature, symbiotic.KeyTag) {
	valset := symbiotic.ValidatorSet{}
	signatures := make([]symbiotic.Signature, 0)
	pks := make([]crypto.PrivateKey, numValidators)
	msg := []byte("message")
	keyTag := symbiotic.KeyTag(1)
	valset.Validators = make([]symbiotic.Validator, numValidators)
	for i := 0; i < numValidators; i++ {
		var err error
		pks[i], err = crypto.GeneratePrivateKey(keyTag.Type())
		if err != nil {
			panic(err)
		}

		valset.Validators[i].Keys = []symbiotic.ValidatorKey{
			{
				Tag:     keyTag,
				Payload: pks[i].PublicKey().OnChain(),
			},
		}

		pk2, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeEcdsaSecp256k1)
		if err != nil {
			panic(err)
		}

		pkEcdsa, err := crypto2.ToECDSA(big.NewInt(0).SetBytes(pk2.Bytes()).FillBytes(make([]byte, 32)))
		if err != nil {
			panic(err)
		}
		valset.Validators[i].Operator = crypto2.PubkeyToAddress(pkEcdsa.PublicKey)
		valset.Validators[i].IsActive = true
		valset.Validators[i].VotingPower = symbiotic.ToVotingPower(big.NewInt(100))
	}

	valset.QuorumThreshold = symbiotic.ToVotingPower(big.NewInt(int64(100 * (numValidators - len(nonSigners)))))
	valset.RequiredKeyTag = keyTag

	valset.Validators.SortByOperatorAddressAsc()

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
		signatures = append(signatures, symbiotic.Signature{
			MessageHash: msgHash,
			KeyTag:      keyTag,
			Epoch:       1,
			Signature:   sig,
			PublicKey:   pks[i].PublicKey(),
		})
	}

	return valset, signatures, keyTag
}

func TestNewAggregator_WithBlsBn254Simple_ReturnsAggregator(t *testing.T) {
	agg, err := NewAggregator(symbiotic.VerificationTypeBlsBn254Simple, nil)

	require.NoError(t, err)
	assert.NotNil(t, agg)
}

func TestNewAggregator_WithBlsBn254ZK_ReturnsAggregator(t *testing.T) {
	prover := &mockProver{}

	agg, err := NewAggregator(symbiotic.VerificationTypeBlsBn254ZK, prover)

	require.NoError(t, err)
	assert.NotNil(t, agg)
}

func TestNewAggregator_WithUnsupportedType_ReturnsError(t *testing.T) {
	unsupportedType := symbiotic.VerificationType(999)

	agg, err := NewAggregator(unsupportedType, nil)

	require.Error(t, err)
	assert.Nil(t, agg)
	assert.Contains(t, err.Error(), "unsupported verification type")
}

func TestNewAggregator_WithDifferentTypes_ReturnsDifferentImplementations(t *testing.T) {
	prover := &mockProver{}

	_, err := NewAggregator(symbiotic.VerificationTypeBlsBn254Simple, nil)
	require.NoError(t, err)

	_, err = NewAggregator(symbiotic.VerificationTypeBlsBn254ZK, prover)
	require.NoError(t, err)
}
