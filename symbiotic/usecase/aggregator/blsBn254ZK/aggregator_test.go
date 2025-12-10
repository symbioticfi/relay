package blsBn254ZK

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/pkg/proof"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type mockProver struct{}

func (m *mockProver) Prove(proveInput proof.ProveInput) (proof.ProofData, error) {
	return proof.ProofData{}, nil
}

func (m *mockProver) Verify(valsetLen int, publicInputHash common.Hash, proofBytes []byte) (bool, error) {
	return true, nil
}

func TestNewAggregator_WithProver_Success(t *testing.T) {
	prover := &mockProver{}

	agg, err := NewAggregator(prover)

	require.NoError(t, err)
	assert.NotNil(t, agg)
	assert.NotNil(t, agg.prover)
}

func TestNewAggregator_WithNilProver_Success(t *testing.T) {
	agg, err := NewAggregator(nil)

	require.NoError(t, err)
	assert.NotNil(t, agg)
}

func TestAggregator_Aggregate_WithMismatchedMessageHashes_ReturnsError(t *testing.T) {
	prover := &mockProver{}
	agg, err := NewAggregator(prover)
	require.NoError(t, err)

	valset := symbiotic.ValidatorSet{}
	signatures := []symbiotic.Signature{
		{MessageHash: []byte("different-message")},
		{MessageHash: []byte("first-message")},
	}

	_, err = agg.Aggregate(valset, signatures)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "signatures have different message hashes")
}

func TestValidatorSetMimcAccumulator_WithNoValidators_ReturnsZeroHash(t *testing.T) {
	validators := []symbiotic.Validator{}
	keyTag := symbiotic.KeyTag(1)

	result, err := validatorSetMimcAccumulator(validators, keyTag)

	require.NoError(t, err)
	assert.Equal(t, common.Hash{}, result)
}

func TestValidatorSetMimcAccumulator_WithActiveValidators_ReturnsHash(t *testing.T) {
	_, _, g1Gen, _ := bn254.Generators()
	g1Bytes := g1Gen.Bytes()

	validators := []symbiotic.Validator{
		{
			IsActive: true,
			Keys: []symbiotic.ValidatorKey{
				{
					Tag:     symbiotic.KeyTag(1),
					Payload: g1Bytes[:],
				},
			},
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		},
	}
	keyTag := symbiotic.KeyTag(1)

	result, err := validatorSetMimcAccumulator(validators, keyTag)

	require.NoError(t, err)
	assert.NotEqual(t, common.Hash{}, result)
}

func TestValidatorSetMimcAccumulator_WithSameValidators_ReturnsSameHash(t *testing.T) {
	_, _, g1Gen, _ := bn254.Generators()
	g1Bytes := g1Gen.Bytes()

	validators := []symbiotic.Validator{
		{
			IsActive: true,
			Keys: []symbiotic.ValidatorKey{
				{
					Tag:     symbiotic.KeyTag(1),
					Payload: g1Bytes[:],
				},
			},
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		},
	}
	keyTag := symbiotic.KeyTag(1)

	hash1, err := validatorSetMimcAccumulator(validators, keyTag)
	require.NoError(t, err)

	hash2, err := validatorSetMimcAccumulator(validators, keyTag)
	require.NoError(t, err)

	assert.Equal(t, hash1, hash2)
}

func TestValidatorSetMimcAccumulator_WithDifferentValidators_ReturnsDifferentHashes(t *testing.T) {
	_, _, g1Gen, _ := bn254.Generators()

	g1Point1 := new(bn254.G1Affine).ScalarMultiplication(&g1Gen, big.NewInt(1))
	g1Point2 := new(bn254.G1Affine).ScalarMultiplication(&g1Gen, big.NewInt(2))

	g1Bytes1 := g1Point1.Bytes()
	g1Bytes2 := g1Point2.Bytes()

	validators1 := []symbiotic.Validator{
		{
			IsActive: true,
			Keys: []symbiotic.ValidatorKey{
				{
					Tag:     symbiotic.KeyTag(1),
					Payload: g1Bytes1[:],
				},
			},
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		},
	}

	validators2 := []symbiotic.Validator{
		{
			IsActive: true,
			Keys: []symbiotic.ValidatorKey{
				{
					Tag:     symbiotic.KeyTag(1),
					Payload: g1Bytes2[:],
				},
			},
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		},
	}

	keyTag := symbiotic.KeyTag(1)

	hash1, err := validatorSetMimcAccumulator(validators1, keyTag)
	require.NoError(t, err)

	hash2, err := validatorSetMimcAccumulator(validators2, keyTag)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
}

func TestToValidatorsData_WithNoValidators_ReturnsNormalizedData(t *testing.T) {
	signers := []symbiotic.Validator{}
	allValidators := []symbiotic.Validator{}
	keyTag := symbiotic.KeyTag(1)

	result, err := toValidatorsData(signers, allValidators, keyTag)

	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestToValidatorsData_WithInactiveValidators_ReturnsNormalizedData(t *testing.T) {
	_, _, g1Gen, _ := bn254.Generators()
	g1Bytes := g1Gen.Bytes()

	signers := []symbiotic.Validator{}
	allValidators := []symbiotic.Validator{
		{
			IsActive: false,
			Keys: []symbiotic.ValidatorKey{
				{
					Tag:     symbiotic.KeyTag(1),
					Payload: g1Bytes[:],
				},
			},
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		},
	}
	keyTag := symbiotic.KeyTag(1)

	result, err := toValidatorsData(signers, allValidators, keyTag)

	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestToValidatorsData_WithActiveValidators_ReturnsData(t *testing.T) {
	_, _, g1Gen, _ := bn254.Generators()
	g1Bytes := g1Gen.Bytes()

	signers := []symbiotic.Validator{}
	allValidators := []symbiotic.Validator{
		{
			IsActive: true,
			Operator: common.HexToAddress("0x1234"),
			Keys: []symbiotic.ValidatorKey{
				{
					Tag:     symbiotic.KeyTag(1),
					Payload: g1Bytes[:],
				},
			},
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		},
	}
	keyTag := symbiotic.KeyTag(1)

	result, err := toValidatorsData(signers, allValidators, keyTag)

	require.NoError(t, err)
	require.NotEmpty(t, result)
	assert.True(t, result[0].IsNonSigner)
}

func TestToValidatorsData_WithSigners_MarksSignersCorrectly(t *testing.T) {
	_, _, g1Gen, _ := bn254.Generators()
	g1Bytes := g1Gen.Bytes()

	operatorAddr := common.HexToAddress("0x1234")

	signers := []symbiotic.Validator{
		{
			Operator: operatorAddr,
		},
	}

	allValidators := []symbiotic.Validator{
		{
			IsActive: true,
			Operator: operatorAddr,
			Keys: []symbiotic.ValidatorKey{
				{
					Tag:     symbiotic.KeyTag(1),
					Payload: g1Bytes[:],
				},
			},
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		},
	}
	keyTag := symbiotic.KeyTag(1)

	result, err := toValidatorsData(signers, allValidators, keyTag)

	require.NoError(t, err)
	require.NotEmpty(t, result)
	assert.False(t, result[0].IsNonSigner)
}

func TestToValidatorsData_WithMixedSignersAndNonSigners_MarksCorrectly(t *testing.T) {
	_, _, g1Gen, _ := bn254.Generators()

	g1Point1 := new(bn254.G1Affine).ScalarMultiplication(&g1Gen, big.NewInt(1))
	g1Point2 := new(bn254.G1Affine).ScalarMultiplication(&g1Gen, big.NewInt(2))

	g1Bytes1 := g1Point1.Bytes()
	g1Bytes2 := g1Point2.Bytes()

	signer1Addr := common.HexToAddress("0x1111")
	nonSigner1Addr := common.HexToAddress("0x2222")

	signers := []symbiotic.Validator{
		{
			Operator: signer1Addr,
		},
	}

	allValidators := []symbiotic.Validator{
		{
			IsActive: true,
			Operator: signer1Addr,
			Keys: []symbiotic.ValidatorKey{
				{
					Tag:     symbiotic.KeyTag(1),
					Payload: g1Bytes1[:],
				},
			},
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		},
		{
			IsActive: true,
			Operator: nonSigner1Addr,
			Keys: []symbiotic.ValidatorKey{
				{
					Tag:     symbiotic.KeyTag(1),
					Payload: g1Bytes2[:],
				},
			},
			VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		},
	}
	keyTag := symbiotic.KeyTag(1)

	result, err := toValidatorsData(signers, allValidators, keyTag)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 2)
	assert.False(t, result[0].IsNonSigner)
	assert.True(t, result[1].IsNonSigner)
}

func TestAggregator_GenerateExtraData_WithValidValidators_ReturnsExtraData(t *testing.T) {
	prover := &mockProver{}
	agg, err := NewAggregator(prover)
	require.NoError(t, err)

	_, _, g1Gen, _ := bn254.Generators()
	g1Bytes := g1Gen.Bytes()

	valset := symbiotic.ValidatorSet{
		Validators: []symbiotic.Validator{
			{
				IsActive: true,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: g1Bytes[:],
					},
				},
				VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
			},
		},
	}

	keyTags := []symbiotic.KeyTag{symbiotic.KeyTag(1)}

	result, err := agg.GenerateExtraData(valset, keyTags)

	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestAggregator_GenerateExtraData_WithMultipleKeyTags_ReturnsMultipleExtraData(t *testing.T) {
	prover := &mockProver{}
	agg, err := NewAggregator(prover)
	require.NoError(t, err)

	_, _, g1Gen, _ := bn254.Generators()

	g1Point1 := new(bn254.G1Affine).ScalarMultiplication(&g1Gen, big.NewInt(1))
	g1Point2 := new(bn254.G1Affine).ScalarMultiplication(&g1Gen, big.NewInt(2))

	g1Bytes1 := g1Point1.Bytes()
	g1Bytes2 := g1Point2.Bytes()

	valset := symbiotic.ValidatorSet{
		Validators: []symbiotic.Validator{
			{
				IsActive: true,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: g1Bytes1[:],
					},
					{
						Tag:     symbiotic.KeyTag(2),
						Payload: g1Bytes2[:],
					},
				},
				VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
			},
		},
	}

	keyTags := []symbiotic.KeyTag{symbiotic.KeyTag(1), symbiotic.KeyTag(2)}

	result, err := agg.GenerateExtraData(valset, keyTags)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 3)
}

func TestAggregator_GenerateExtraData_ReturnsSortedData(t *testing.T) {
	prover := &mockProver{}
	agg, err := NewAggregator(prover)
	require.NoError(t, err)

	_, _, g1Gen, _ := bn254.Generators()
	g1Bytes := g1Gen.Bytes()

	valset := symbiotic.ValidatorSet{
		Validators: []symbiotic.Validator{
			{
				IsActive: true,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: g1Bytes[:],
					},
				},
				VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
			},
		},
	}

	keyTags := []symbiotic.KeyTag{symbiotic.KeyTag(1)}

	result, err := agg.GenerateExtraData(valset, keyTags)

	require.NoError(t, err)
	require.NotEmpty(t, result)

	for i := 0; i < len(result)-1; i++ {
		assert.Negative(t, result[i].Key.Cmp(result[i+1].Key), "extra data should be sorted by key")
	}
}

func TestAggregator_Verify_WithInvalidMessageHash_ReturnsError(t *testing.T) {
	prover := &mockProver{}
	agg, err := NewAggregator(prover)
	require.NoError(t, err)

	_, _, g1Gen, _ := bn254.Generators()
	g1Bytes := g1Gen.Bytes()

	valset := symbiotic.ValidatorSet{
		Validators: []symbiotic.Validator{
			{
				IsActive: true,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: g1Bytes[:],
					},
				},
				VotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
			},
		},
		QuorumThreshold: symbiotic.ToVotingPower(big.NewInt(50)),
	}
	keyTag := symbiotic.KeyTag(1)

	proofBytes := make([]byte, 64)
	aggregationProof := symbiotic.AggregationProof{
		MessageHash: []byte("invalid"),
		Proof:       proofBytes,
	}

	success, err := agg.Verify(valset, keyTag, aggregationProof)

	require.Error(t, err)
	assert.False(t, success)
}

func TestAggregator_Verify_WithInsufficientVotingPower_ReturnsError(t *testing.T) {
	prover := &mockProver{}
	agg, err := NewAggregator(prover)
	require.NoError(t, err)

	_, _, g1Gen, _ := bn254.Generators()
	g1Bytes := g1Gen.Bytes()

	valset := symbiotic.ValidatorSet{
		Validators: []symbiotic.Validator{
			{
				IsActive: true,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: g1Bytes[:],
					},
				},
				VotingPower: symbiotic.ToVotingPower(big.NewInt(50)),
			},
		},
		QuorumThreshold: symbiotic.ToVotingPower(big.NewInt(100)),
	}
	keyTag := symbiotic.KeyTag(1)

	proofBytes := make([]byte, 64)
	proof := symbiotic.AggregationProof{
		MessageHash: make([]byte, 32),
		Proof:       proofBytes,
	}

	success, err := agg.Verify(valset, keyTag, proof)

	require.Error(t, err)
	assert.False(t, success)
	assert.Contains(t, err.Error(), "less than quorum threshold")
}
