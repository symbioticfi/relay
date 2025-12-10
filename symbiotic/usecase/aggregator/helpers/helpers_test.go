package helpers

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestCompareMessageHasher_WithMatchingHashes_ReturnsTrue(t *testing.T) {
	msgHash := []byte("test-message-hash")
	signatures := []symbiotic.Signature{
		{MessageHash: msgHash},
		{MessageHash: msgHash},
		{MessageHash: msgHash},
	}

	assert.NoError(t, CheckSignaturesHaveSameTagAndMessageHash(signatures))
}

func TestCompareMessageHasher_WithDifferentHash_ReturnsFalse(t *testing.T) {
	msgHash := []byte("test-message-hash")
	differentHash := []byte("different-hash")
	signatures := []symbiotic.Signature{
		{MessageHash: msgHash},
		{MessageHash: differentHash},
		{MessageHash: msgHash},
	}

	assert.Error(t, CheckSignaturesHaveSameTagAndMessageHash(signatures), "")
}

func TestCompareMessageHasher_WithEmptySignatures_ReturnsTrue(t *testing.T) {
	msgHash := []byte("test-message-hash")
	signatures := []symbiotic.Signature{
		{MessageHash: msgHash},
	}

	err := CheckSignaturesHaveSameTagAndMessageHash(signatures)
	assert.NoError(t, err)
}

func TestCompareMessageHasher_WithSingleSignature_ReturnsCorrectResult(t *testing.T) {
	tests := []struct {
		name           string
		msgHash        []byte
		signatureHash  []byte
		expectedResult string
	}{
		{
			name:           "matching single signature",
			msgHash:        []byte("test-hash"),
			signatureHash:  []byte("test-hash"),
			expectedResult: "",
		},
		{
			name:           "non-matching single signature",
			msgHash:        []byte("test-hash"),
			signatureHash:  []byte("different-hash"),
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signatures := []symbiotic.Signature{{MessageHash: tt.signatureHash}}

			err := CheckSignaturesHaveSameTagAndMessageHash(signatures)
			if tt.expectedResult != "" {
				require.Error(t, err)
				require.EqualError(t, err, tt.expectedResult)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestGetExtraDataKey_WithValidInputs_ReturnsHash(t *testing.T) {
	verificationType := symbiotic.VerificationTypeBlsBn254Simple
	nameHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	result, err := GetExtraDataKey(verificationType, nameHash)

	require.NoError(t, err)
	assert.NotEqual(t, common.Hash{}, result)
}

func TestGetExtraDataKey_WithDifferentVerificationTypes_ReturnsDifferentHashes(t *testing.T) {
	nameHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	hash1, err := GetExtraDataKey(symbiotic.VerificationTypeBlsBn254Simple, nameHash)
	require.NoError(t, err)

	hash2, err := GetExtraDataKey(symbiotic.VerificationTypeBlsBn254ZK, nameHash)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
}

func TestGetExtraDataKey_WithDifferentNameHashes_ReturnsDifferentHashes(t *testing.T) {
	verificationType := symbiotic.VerificationTypeBlsBn254Simple
	nameHash1 := common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111")
	nameHash2 := common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222")

	hash1, err := GetExtraDataKey(verificationType, nameHash1)
	require.NoError(t, err)

	hash2, err := GetExtraDataKey(verificationType, nameHash2)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
}

func TestGetExtraDataKey_WithSameInputs_ReturnsSameHash(t *testing.T) {
	verificationType := symbiotic.VerificationTypeBlsBn254Simple
	nameHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	hash1, err := GetExtraDataKey(verificationType, nameHash)
	require.NoError(t, err)

	hash2, err := GetExtraDataKey(verificationType, nameHash)
	require.NoError(t, err)

	assert.Equal(t, hash1, hash2)
}

func TestGetExtraDataKeyTagged_WithValidInputs_ReturnsHash(t *testing.T) {
	verificationType := symbiotic.VerificationTypeBlsBn254Simple
	keyTag := symbiotic.KeyTag(1)
	nameHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	result, err := GetExtraDataKeyTagged(verificationType, keyTag, nameHash)

	require.NoError(t, err)
	assert.NotEqual(t, common.Hash{}, result)
}

func TestGetExtraDataKeyTagged_WithDifferentKeyTags_ReturnsDifferentHashes(t *testing.T) {
	verificationType := symbiotic.VerificationTypeBlsBn254Simple
	nameHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	hash1, err := GetExtraDataKeyTagged(verificationType, symbiotic.KeyTag(1), nameHash)
	require.NoError(t, err)

	hash2, err := GetExtraDataKeyTagged(verificationType, symbiotic.KeyTag(2), nameHash)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
}

func TestGetExtraDataKeyTagged_WithDifferentVerificationTypes_ReturnsDifferentHashes(t *testing.T) {
	keyTag := symbiotic.KeyTag(1)
	nameHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	hash1, err := GetExtraDataKeyTagged(symbiotic.VerificationTypeBlsBn254Simple, keyTag, nameHash)
	require.NoError(t, err)

	hash2, err := GetExtraDataKeyTagged(symbiotic.VerificationTypeBlsBn254ZK, keyTag, nameHash)
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
}

func TestGetExtraDataKeyTagged_WithSameInputs_ReturnsSameHash(t *testing.T) {
	verificationType := symbiotic.VerificationTypeBlsBn254Simple
	keyTag := symbiotic.KeyTag(1)
	nameHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	hash1, err := GetExtraDataKeyTagged(verificationType, keyTag, nameHash)
	require.NoError(t, err)

	hash2, err := GetExtraDataKeyTagged(verificationType, keyTag, nameHash)
	require.NoError(t, err)

	assert.Equal(t, hash1, hash2)
}

func TestGetAggregatedPubKeys_WithNoValidators_ReturnsZeroPoint(t *testing.T) {
	valset := symbiotic.ValidatorSet{
		Validators: []symbiotic.Validator{},
	}
	keyTags := []symbiotic.KeyTag{symbiotic.KeyTag(1)}

	result := GetAggregatedPubKeys(valset, keyTags)

	require.Len(t, result, 1)
	assert.Equal(t, symbiotic.KeyTag(1), result[0].Tag)
	assert.NotEmpty(t, result[0].Payload)
}

func TestGetAggregatedPubKeys_WithInactiveValidators_ReturnsZeroPoint(t *testing.T) {
	g1Point := bn254.G1Affine{}
	g1Point.X.SetOne()
	g1Point.Y.SetOne()
	g1Bytes := g1Point.Bytes()

	valset := symbiotic.ValidatorSet{
		Validators: []symbiotic.Validator{
			{
				IsActive: false,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: g1Bytes[:],
					},
				},
			},
		},
	}
	keyTags := []symbiotic.KeyTag{symbiotic.KeyTag(1)}

	result := GetAggregatedPubKeys(valset, keyTags)

	require.Len(t, result, 1)
	assert.Equal(t, symbiotic.KeyTag(1), result[0].Tag)
	assert.NotEmpty(t, result[0].Payload)
}

func TestGetAggregatedPubKeys_WithSingleActiveValidator_ReturnsAggregatedKey(t *testing.T) {
	g1Point := bn254.G1Affine{}
	g1Point.X.SetOne()
	g1Point.Y.SetUint64(2)
	g1Bytes := g1Point.Bytes()

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
			},
		},
	}
	keyTags := []symbiotic.KeyTag{symbiotic.KeyTag(1)}

	result := GetAggregatedPubKeys(valset, keyTags)

	require.Len(t, result, 1)
	assert.Equal(t, symbiotic.KeyTag(1), result[0].Tag)
	assert.NotEmpty(t, result[0].Payload)
}

func TestGetAggregatedPubKeys_WithMultipleActiveValidators_ReturnsAggregatedKey(t *testing.T) {
	g1Point1 := bn254.G1Affine{}
	g1Point1.X.SetOne()
	g1Point1.Y.SetUint64(2)
	g1Bytes1 := g1Point1.Bytes()

	g1Point2 := bn254.G1Affine{}
	g1Point2.X.SetUint64(3)
	g1Point2.Y.SetUint64(4)
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
				},
			},
			{
				IsActive: true,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: g1Bytes2[:],
					},
				},
			},
		},
	}
	keyTags := []symbiotic.KeyTag{symbiotic.KeyTag(1)}

	result := GetAggregatedPubKeys(valset, keyTags)

	require.Len(t, result, 1)
	assert.Equal(t, symbiotic.KeyTag(1), result[0].Tag)
	assert.NotEmpty(t, result[0].Payload)
}

func TestGetAggregatedPubKeys_WithMixedActiveInactive_ReturnsOnlyActiveAggregation(t *testing.T) {
	g1Point1 := bn254.G1Affine{}
	g1Point1.X.SetOne()
	g1Point1.Y.SetUint64(2)
	g1Bytes1 := g1Point1.Bytes()

	g1Point2 := bn254.G1Affine{}
	g1Point2.X.SetUint64(3)
	g1Point2.Y.SetUint64(4)
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
				},
			},
			{
				IsActive: false,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: g1Bytes2[:],
					},
				},
			},
		},
	}
	keyTags := []symbiotic.KeyTag{symbiotic.KeyTag(1)}

	result := GetAggregatedPubKeys(valset, keyTags)

	require.Len(t, result, 1)
	assert.Equal(t, symbiotic.KeyTag(1), result[0].Tag)
}

func TestGetAggregatedPubKeys_WithMultipleKeyTags_ReturnsMultipleAggregations(t *testing.T) {
	g1Point1 := bn254.G1Affine{}
	g1Point1.X.SetOne()
	g1Point1.Y.SetUint64(2)
	g1Bytes1 := g1Point1.Bytes()

	g1Point2 := bn254.G1Affine{}
	g1Point2.X.SetUint64(3)
	g1Point2.Y.SetUint64(4)
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
			},
		},
	}
	keyTags := []symbiotic.KeyTag{symbiotic.KeyTag(1), symbiotic.KeyTag(2)}

	result := GetAggregatedPubKeys(valset, keyTags)

	require.Len(t, result, 2)
}

func TestGetExtraDataKeyIndexed_WithValidInputs_ReturnsHash(t *testing.T) {
	verificationType := symbiotic.VerificationTypeBlsBn254Simple
	keyTag := symbiotic.KeyTag(1)
	nameHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	index := big.NewInt(0)

	result, err := GetExtraDataKeyIndexed(verificationType, keyTag, nameHash, index)

	require.NoError(t, err)
	assert.NotEqual(t, common.Hash{}, result)
}

func TestGetExtraDataKeyIndexed_WithZeroIndex_EqualsBaseHash(t *testing.T) {
	verificationType := symbiotic.VerificationTypeBlsBn254Simple
	keyTag := symbiotic.KeyTag(1)
	nameHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	baseHash, err := GetExtraDataKeyTagged(verificationType, keyTag, nameHash)
	require.NoError(t, err)

	indexedHash, err := GetExtraDataKeyIndexed(verificationType, keyTag, nameHash, big.NewInt(0))
	require.NoError(t, err)

	assert.Equal(t, baseHash, indexedHash)
}

func TestGetExtraDataKeyIndexed_WithDifferentIndexes_ReturnsDifferentHashes(t *testing.T) {
	verificationType := symbiotic.VerificationTypeBlsBn254Simple
	keyTag := symbiotic.KeyTag(1)
	nameHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	hash1, err := GetExtraDataKeyIndexed(verificationType, keyTag, nameHash, big.NewInt(0))
	require.NoError(t, err)

	hash2, err := GetExtraDataKeyIndexed(verificationType, keyTag, nameHash, big.NewInt(1))
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
}

func TestGetExtraDataKeyIndexed_WithLargeIndex_ReturnsHash(t *testing.T) {
	verificationType := symbiotic.VerificationTypeBlsBn254Simple
	keyTag := symbiotic.KeyTag(1)
	nameHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	largeIndex := big.NewInt(999999999)

	result, err := GetExtraDataKeyIndexed(verificationType, keyTag, nameHash, largeIndex)

	require.NoError(t, err)
	assert.NotEqual(t, common.Hash{}, result)
}

func TestGetValidatorsIndexesMapByKey_WithNoValidators_ReturnsEmptyMap(t *testing.T) {
	valset := symbiotic.ValidatorSet{
		Validators: []symbiotic.Validator{},
	}
	keyTag := symbiotic.KeyTag(1)

	result := GetValidatorsIndexesMapByKey(valset, keyTag)

	assert.Empty(t, result)
}

func TestGetValidatorsIndexesMapByKey_WithInactiveValidators_ReturnsEmptyMap(t *testing.T) {
	valset := symbiotic.ValidatorSet{
		Validators: []symbiotic.Validator{
			{
				IsActive: false,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: []byte("key-payload"),
					},
				},
			},
		},
	}
	keyTag := symbiotic.KeyTag(1)

	result := GetValidatorsIndexesMapByKey(valset, keyTag)

	assert.Empty(t, result)
}

func TestGetValidatorsIndexesMapByKey_WithSingleActiveValidator_ReturnsMapWithOneEntry(t *testing.T) {
	keyPayload := []byte("key-payload")
	valset := symbiotic.ValidatorSet{
		Validators: []symbiotic.Validator{
			{
				IsActive: true,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: keyPayload,
					},
				},
			},
		},
	}
	keyTag := symbiotic.KeyTag(1)

	result := GetValidatorsIndexesMapByKey(valset, keyTag)

	require.Len(t, result, 1)
	assert.Equal(t, 0, result[string(keyPayload)])
}

func TestGetValidatorsIndexesMapByKey_WithMultipleActiveValidators_ReturnsMapWithAllEntries(t *testing.T) {
	keyPayload1 := []byte("key-payload-1")
	keyPayload2 := []byte("key-payload-2")
	keyPayload3 := []byte("key-payload-3")

	valset := symbiotic.ValidatorSet{
		Validators: []symbiotic.Validator{
			{
				IsActive: true,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: keyPayload1,
					},
				},
			},
			{
				IsActive: true,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: keyPayload2,
					},
				},
			},
			{
				IsActive: true,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: keyPayload3,
					},
				},
			},
		},
	}
	keyTag := symbiotic.KeyTag(1)

	result := GetValidatorsIndexesMapByKey(valset, keyTag)

	require.Len(t, result, 3)
	assert.Equal(t, 0, result[string(keyPayload1)])
	assert.Equal(t, 1, result[string(keyPayload2)])
	assert.Equal(t, 2, result[string(keyPayload3)])
}

func TestGetValidatorsIndexesMapByKey_WithMixedActiveInactive_ReturnsOnlyActiveValidators(t *testing.T) {
	keyPayload1 := []byte("key-payload-1")
	keyPayload2 := []byte("key-payload-2")
	keyPayload3 := []byte("key-payload-3")

	valset := symbiotic.ValidatorSet{
		Validators: []symbiotic.Validator{
			{
				IsActive: true,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: keyPayload1,
					},
				},
			},
			{
				IsActive: false,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: keyPayload2,
					},
				},
			},
			{
				IsActive: true,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: keyPayload3,
					},
				},
			},
		},
	}
	keyTag := symbiotic.KeyTag(1)

	result := GetValidatorsIndexesMapByKey(valset, keyTag)

	require.Len(t, result, 2)
	assert.Equal(t, 0, result[string(keyPayload1)])
	assert.Equal(t, 2, result[string(keyPayload3)])
	_, exists := result[string(keyPayload2)]
	assert.False(t, exists)
}

func TestGetValidatorsIndexesMapByKey_WithNonMatchingKeyTag_ReturnsEmptyMap(t *testing.T) {
	valset := symbiotic.ValidatorSet{
		Validators: []symbiotic.Validator{
			{
				IsActive: true,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: []byte("key-payload"),
					},
				},
			},
		},
	}
	keyTag := symbiotic.KeyTag(2)

	result := GetValidatorsIndexesMapByKey(valset, keyTag)

	assert.Empty(t, result)
}

func TestGetValidatorsIndexesMapByKey_WithValidatorHavingMultipleKeys_OnlyMapsMatchingKeyTag(t *testing.T) {
	keyPayload1 := []byte("key-payload-1")
	keyPayload2 := []byte("key-payload-2")

	valset := symbiotic.ValidatorSet{
		Validators: []symbiotic.Validator{
			{
				IsActive: true,
				Keys: []symbiotic.ValidatorKey{
					{
						Tag:     symbiotic.KeyTag(1),
						Payload: keyPayload1,
					},
					{
						Tag:     symbiotic.KeyTag(2),
						Payload: keyPayload2,
					},
				},
			},
		},
	}
	keyTag := symbiotic.KeyTag(1)

	result := GetValidatorsIndexesMapByKey(valset, keyTag)

	require.Len(t, result, 1)
	assert.Equal(t, 0, result[string(keyPayload1)])
	_, exists := result[string(keyPayload2)]
	assert.False(t, exists)
}
