package blsBn254Simple

import (
	"context"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestNewAggregator_Success(t *testing.T) {
	agg, err := NewAggregator()

	require.NoError(t, err)
	assert.NotNil(t, agg)
	assert.NotNil(t, agg.abiTypes.g1Type)
	assert.NotNil(t, agg.abiTypes.g2Type)
	assert.NotNil(t, agg.abiTypes.validatorsDataType)
}

func TestCompress_WithValidG1Point_ReturnsCompressedHash(t *testing.T) {
	g1 := new(bn254.G1Affine)
	g1.X.SetOne()
	g1.Y.SetUint64(2)

	compressed, err := compress(g1)

	require.NoError(t, err)
	assert.NotEqual(t, common.Hash{}, compressed)
}

func TestCompress_WithZeroPoint_ReturnsHash(t *testing.T) {
	g1 := new(bn254.G1Affine)

	compressed, err := compress(g1)

	require.NoError(t, err)
	assert.NotEqual(t, common.Hash{}, compressed)
}

func TestCompress_WithDifferentPoints_ReturnsDifferentHashes(t *testing.T) {
	g1Point1 := new(bn254.G1Affine)
	g1Point1.X.SetOne()
	g1Point1.Y.SetUint64(2)

	g1Point2 := new(bn254.G1Affine)
	g1Point2.X.SetUint64(3)
	g1Point2.Y.SetUint64(4)

	compressed1, err := compress(g1Point1)
	require.NoError(t, err)

	compressed2, err := compress(g1Point2)
	require.NoError(t, err)

	assert.NotEqual(t, compressed1, compressed2)
}

func TestDecompress_WithValidCompressedPoint_ReturnsG1Point(t *testing.T) {
	originalG1 := new(bn254.G1Affine)
	originalG1.X.SetOne()
	originalG1.Y.SetUint64(2)

	compressed, err := compress(originalG1)
	require.NoError(t, err)

	decompressed, err := decompress(compressed)

	require.NoError(t, err)
	assert.NotNil(t, decompressed)
}

func TestCompress_Decompress_RoundTrip_PreservesPoint(t *testing.T) {
	_, _, g1Gen, _ := bn254.Generators()

	tests := []struct {
		name string
		g1   *bn254.G1Affine
	}{
		{
			name: "generator point",
			g1:   &g1Gen,
		},
		{
			name: "scaled generator",
			g1:   new(bn254.G1Affine).ScalarMultiplication(&g1Gen, big.NewInt(42)),
		},
		{
			name: "another scaled point",
			g1:   new(bn254.G1Affine).ScalarMultiplication(&g1Gen, big.NewInt(12345)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := compress(tt.g1)
			require.NoError(t, err)

			decompressed, err := decompress(compressed)
			require.NoError(t, err)

			assert.Equal(t, tt.g1.X.String(), decompressed.X.String())
		})
	}
}

func TestFindYFromX_WithValidX_ReturnsY(t *testing.T) {
	x := big.NewInt(1)

	y, err := findYFromX(x)

	require.NoError(t, err)
	assert.NotNil(t, y)
	assert.Greater(t, y.Cmp(big.NewInt(0)), 0) //nolint:testifylint // assert.Positive doesn't work with *big.Int
}

func TestFindYFromX_WithZeroX_ReturnsY(t *testing.T) {
	x := big.NewInt(0)

	y, err := findYFromX(x)

	require.NoError(t, err)
	assert.NotNil(t, y)
}

func TestFindYFromX_WithLargeX_ReturnsY(t *testing.T) {
	x := new(big.Int).SetUint64(999999999)

	y, err := findYFromX(x)

	require.NoError(t, err)
	assert.NotNil(t, y)
}

func TestProcessValidators_WithNoValidators_ReturnsEmpty(t *testing.T) {
	validators := []symbiotic.Validator{}
	keyTag := symbiotic.KeyTag(1)

	result, err := processValidators(validators, keyTag)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestProcessValidators_WithInactiveValidators_ReturnsEmpty(t *testing.T) {
	g1 := new(bn254.G1Affine)
	g1.X.SetOne()
	g1.Y.SetUint64(2)
	g1Bytes := g1.Bytes()

	validators := []symbiotic.Validator{
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

	result, err := processValidators(validators, keyTag)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestProcessValidators_WithActiveValidators_ReturnsValidatorData(t *testing.T) {
	_, _, g1Gen, _ := bn254.Generators()
	g1Bytes := g1Gen.Bytes()

	validators := []symbiotic.Validator{
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

	result, err := processValidators(validators, keyTag)

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.NotEqual(t, common.Hash{}, result[0].KeySerialized)
	assert.Equal(t, big.NewInt(100), result[0].VotingPower)
}

func TestProcessValidators_WithMultipleValidators_ReturnsSortedData(t *testing.T) {
	_, _, g1Gen, _ := bn254.Generators()

	g1Point1 := new(bn254.G1Affine).ScalarMultiplication(&g1Gen, big.NewInt(1))
	g1Point2 := new(bn254.G1Affine).ScalarMultiplication(&g1Gen, big.NewInt(2))
	g1Point3 := new(bn254.G1Affine).ScalarMultiplication(&g1Gen, big.NewInt(3))

	g1Bytes1 := g1Point1.Bytes()
	g1Bytes2 := g1Point2.Bytes()
	g1Bytes3 := g1Point3.Bytes()

	validators := []symbiotic.Validator{
		{
			IsActive: true,
			Operator: common.HexToAddress("0x3333"),
			Keys: []symbiotic.ValidatorKey{
				{
					Tag:     symbiotic.KeyTag(1),
					Payload: g1Bytes3[:],
				},
			},
			VotingPower: symbiotic.ToVotingPower(big.NewInt(300)),
		},
		{
			IsActive: true,
			Operator: common.HexToAddress("0x1111"),
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
			Operator: common.HexToAddress("0x2222"),
			Keys: []symbiotic.ValidatorKey{
				{
					Tag:     symbiotic.KeyTag(1),
					Payload: g1Bytes2[:],
				},
			},
			VotingPower: symbiotic.ToVotingPower(big.NewInt(200)),
		},
	}
	keyTag := symbiotic.KeyTag(1)

	result, err := processValidators(validators, keyTag)

	require.NoError(t, err)
	require.Len(t, result, 3)

	for i := 0; i < len(result)-1; i++ {
		assert.Negative(t, result[i].KeySerialized.Cmp(result[i+1].KeySerialized), "validators should be sorted by KeySerialized")
	}
}

func TestProcessValidators_WithMissingKeyTag_ReturnsError(t *testing.T) {
	_, _, g1Gen, _ := bn254.Generators()
	g1Bytes := g1Gen.Bytes()

	validators := []symbiotic.Validator{
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
	keyTag := symbiotic.KeyTag(2)

	result, err := processValidators(validators, keyTag)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find key by keyTag")
}

func TestCalcAlpha_WithValidInputs_ReturnsAlpha(t *testing.T) {
	_, _, g1Gen, g2Gen := bn254.Generators()
	messageHash := make([]byte, 32)
	messageHash[0] = 0x01

	alpha := calcAlpha(&g1Gen, &g2Gen, &g1Gen, messageHash)

	require.NotNil(t, alpha)
	assert.Greater(t, alpha.Cmp(big.NewInt(0)), 0) //nolint:testifylint // assert.Positive doesn't work with *big.Int
}

func TestCalcAlpha_WithDifferentInputs_ReturnsDifferentAlphas(t *testing.T) {
	_, _, g1Gen, g2Gen := bn254.Generators()

	messageHash1 := make([]byte, 32)
	messageHash1[0] = 0x01

	messageHash2 := make([]byte, 32)
	messageHash2[0] = 0x02

	alpha1 := calcAlpha(&g1Gen, &g2Gen, &g1Gen, messageHash1)
	alpha2 := calcAlpha(&g1Gen, &g2Gen, &g1Gen, messageHash2)

	assert.NotEqual(t, alpha1, alpha2)
}

func TestCalcAlpha_WithSameInputs_ReturnsSameAlpha(t *testing.T) {
	_, _, g1Gen, g2Gen := bn254.Generators()
	messageHash := make([]byte, 32)
	messageHash[0] = 0x01

	alpha1 := calcAlpha(&g1Gen, &g2Gen, &g1Gen, messageHash)
	alpha2 := calcAlpha(&g1Gen, &g2Gen, &g1Gen, messageHash)

	assert.Equal(t, alpha1, alpha2)
}

func TestAggregator_Aggregate_WithEmptySignatures_Fail(t *testing.T) {
	ctx := context.Background()
	agg, err := NewAggregator()
	require.NoError(t, err)

	valset := symbiotic.ValidatorSet{
		Validators: []symbiotic.Validator{},
	}
	signatures := []symbiotic.Signature{}

	_, err = agg.Aggregate(ctx, valset, signatures)
	require.EqualError(t, err, "invalid signatures: empty signatures slice")
}

func TestAggregator_Aggregate_WithMismatchedMessageHashes_ReturnsError(t *testing.T) {
	ctx := context.Background()
	agg, err := NewAggregator()
	require.NoError(t, err)

	valset := symbiotic.ValidatorSet{}
	messageHash := []byte("test-message")
	signatures := []symbiotic.Signature{
		{MessageHash: []byte("different-message")},
		{MessageHash: messageHash},
	}

	_, err = agg.Aggregate(ctx, valset, signatures)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "signatures have different message hashes")
}

func TestAggregator_Verify_WithInvalidMessageHashLength_ReturnsError(t *testing.T) {
	ctx := context.Background()
	agg, err := NewAggregator()
	require.NoError(t, err)

	valset := symbiotic.ValidatorSet{}
	keyTag := symbiotic.KeyTag(1)
	proof := symbiotic.AggregationProof{
		MessageHash: make([]byte, 16),
		Proof:       make([]byte, 224),
	}

	success, err := agg.Verify(ctx, valset, keyTag, proof)

	require.Error(t, err)
	assert.False(t, success)
	assert.Contains(t, err.Error(), "invalid length")
}

func TestAggregator_Verify_WithShortProof_ReturnsError(t *testing.T) {
	ctx := context.Background()
	agg, err := NewAggregator()
	require.NoError(t, err)

	valset := symbiotic.ValidatorSet{}
	keyTag := symbiotic.KeyTag(1)
	proof := symbiotic.AggregationProof{
		MessageHash: make([]byte, 32),
		Proof:       make([]byte, 100),
	}

	success, err := agg.Verify(ctx, valset, keyTag, proof)

	require.Error(t, err)
	assert.False(t, success)
	assert.Contains(t, err.Error(), "too short")
}

func TestAggregator_PackValidatorsData_WithValidData_ReturnsBytes(t *testing.T) {
	agg, err := NewAggregator()
	require.NoError(t, err)

	validatorsData := []ValidatorData{
		{
			KeySerialized: common.HexToHash("0x1234"),
			VotingPower:   big.NewInt(100),
		},
	}

	result, err := agg.packValidatorsData(validatorsData)

	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestAggregator_PackValidatorsData_WithEmptyData_ReturnsBytes(t *testing.T) {
	agg, err := NewAggregator()
	require.NoError(t, err)

	validatorsData := []ValidatorData{}

	result, err := agg.packValidatorsData(validatorsData)

	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestAggregator_CalculateValidatorsKeccak_WithValidData_ReturnsHash(t *testing.T) {
	agg, err := NewAggregator()
	require.NoError(t, err)

	validatorsData := []ValidatorData{
		{
			KeySerialized: common.HexToHash("0x1234"),
			VotingPower:   big.NewInt(100),
		},
	}

	result, err := agg.calculateValidatorsKeccak(validatorsData)

	require.NoError(t, err)
	assert.NotEqual(t, common.Hash{}, result)
}

func TestAggregator_CalculateValidatorsKeccak_WithSameData_ReturnsSameHash(t *testing.T) {
	agg, err := NewAggregator()
	require.NoError(t, err)

	validatorsData := []ValidatorData{
		{
			KeySerialized: common.HexToHash("0x1234"),
			VotingPower:   big.NewInt(100),
		},
	}

	hash1, err := agg.calculateValidatorsKeccak(validatorsData)
	require.NoError(t, err)

	hash2, err := agg.calculateValidatorsKeccak(validatorsData)
	require.NoError(t, err)

	assert.Equal(t, hash1, hash2)
}

func TestAggregator_CalculateValidatorsKeccak_WithDifferentData_ReturnsDifferentHashes(t *testing.T) {
	agg, err := NewAggregator()
	require.NoError(t, err)

	validatorsData1 := []ValidatorData{
		{
			KeySerialized: common.HexToHash("0x1234"),
			VotingPower:   big.NewInt(100),
		},
	}

	validatorsData2 := []ValidatorData{
		{
			KeySerialized: common.HexToHash("0x5678"),
			VotingPower:   big.NewInt(200),
		},
	}

	hash1, err := agg.calculateValidatorsKeccak(validatorsData1)
	require.NoError(t, err)

	hash2, err := agg.calculateValidatorsKeccak(validatorsData2)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
}

func TestAggregator_GenerateExtraData_WithValidValidatorSet_ReturnsExtraData(t *testing.T) {
	ctx := context.Background()
	agg, err := NewAggregator()
	require.NoError(t, err)

	_, _, g1Gen, _ := bn254.Generators()
	g1Bytes := g1Gen.Bytes()

	valset := symbiotic.ValidatorSet{
		Validators: []symbiotic.Validator{
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
		},
	}

	keyTags := []symbiotic.KeyTag{symbiotic.KeyTag(1)}

	result, err := agg.GenerateExtraData(ctx, valset, keyTags)

	require.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Len(t, result, 2)
}

func TestAggregator_GenerateExtraData_WithMultipleKeyTags_ReturnsMultipleExtraData(t *testing.T) {
	ctx := context.Background()
	agg, err := NewAggregator()
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
				Operator: common.HexToAddress("0x1234"),
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

	result, err := agg.GenerateExtraData(ctx, valset, keyTags)

	require.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.GreaterOrEqual(t, len(result), 4)
}

func TestAggregator_GenerateExtraData_ReturnsSortedData(t *testing.T) {
	ctx := context.Background()
	agg, err := NewAggregator()
	require.NoError(t, err)

	_, _, g1Gen, _ := bn254.Generators()
	g1Bytes := g1Gen.Bytes()

	valset := symbiotic.ValidatorSet{
		Validators: []symbiotic.Validator{
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
		},
	}

	keyTags := []symbiotic.KeyTag{symbiotic.KeyTag(1)}

	result, err := agg.GenerateExtraData(ctx, valset, keyTags)

	require.NoError(t, err)
	require.NotEmpty(t, result)

	for i := 0; i < len(result)-1; i++ {
		assert.Negative(t, result[i].Key.Cmp(result[i+1].Key), "extra data should be sorted by key")
	}
}
