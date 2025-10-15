package badger

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

func TestValidatorProtoConversion(t *testing.T) {
	original := symbiotic.Validator{
		Operator:    common.HexToAddress("0x1234567890123456789012345678901234567890"),
		VotingPower: symbiotic.ToVotingPower(big.NewInt(1000000)),
		IsActive:    true,
		Keys: []symbiotic.ValidatorKey{
			{
				Tag:     symbiotic.KeyTag(1),
				Payload: []byte("test-key-payload-1"),
			},
			{
				Tag:     symbiotic.KeyTag(2),
				Payload: []byte("test-key-payload-2"),
			},
		},
		Vaults: []symbiotic.ValidatorVault{
			{
				ChainID:     1,
				Vault:       common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(500000)),
			},
			{
				ChainID:     2,
				Vault:       common.HexToAddress("0x1111111111111111111111111111111111111111"),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(500000)),
			},
		},
	}

	activeIndex := uint32(42)

	bytes, err := validatorToBytes(original, activeIndex)
	require.NoError(t, err)
	require.NotEmpty(t, bytes)

	decoded, decodedActiveIndex, err := bytesToValidator(bytes)
	require.NoError(t, err)

	assert.Equal(t, original.Operator, decoded.Operator)
	assert.Equal(t, original.VotingPower, decoded.VotingPower)
	assert.Equal(t, original.IsActive, decoded.IsActive)
	assert.Equal(t, activeIndex, decodedActiveIndex)
	assert.Len(t, decoded.Keys, len(original.Keys))
	assert.Len(t, decoded.Vaults, len(original.Vaults))

	for i, key := range original.Keys {
		assert.Equal(t, key.Tag, decoded.Keys[i].Tag)
		assert.Equal(t, key.Payload, decoded.Keys[i].Payload)
	}

	for i, vault := range original.Vaults {
		assert.Equal(t, vault.ChainID, decoded.Vaults[i].ChainID)
		assert.Equal(t, vault.Vault, decoded.Vaults[i].Vault)
		assert.Equal(t, vault.VotingPower, decoded.Vaults[i].VotingPower)
	}
}

func TestValidatorProtoConversion_EmptyArrays(t *testing.T) {
	original := symbiotic.Validator{
		Operator:    common.HexToAddress("0x1234567890123456789012345678901234567890"),
		VotingPower: symbiotic.ToVotingPower(big.NewInt(1000000)),
		IsActive:    false,
		Keys:        []symbiotic.ValidatorKey{},
		Vaults:      []symbiotic.ValidatorVault{},
	}

	bytes, err := validatorToBytes(original, 0)
	require.NoError(t, err)

	decoded, _, err := bytesToValidator(bytes)
	require.NoError(t, err)

	assert.Equal(t, original.Operator, decoded.Operator)
	assert.Equal(t, original.VotingPower, decoded.VotingPower)
	assert.Equal(t, original.IsActive, decoded.IsActive)
	assert.Empty(t, decoded.Keys)
	assert.Empty(t, decoded.Vaults)
}

func TestNetworkConfigProtoConversion(t *testing.T) {
	original := symbiotic.NetworkConfig{
		VotingPowerProviders: []symbiotic.CrossChainAddress{
			{ChainId: 1, Address: common.HexToAddress("0x1111111111111111111111111111111111111111")},
			{ChainId: 2, Address: common.HexToAddress("0x2222222222222222222222222222222222222222")},
		},
		KeysProvider: symbiotic.CrossChainAddress{
			ChainId: 1,
			Address: common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		},
		Settlements: []symbiotic.CrossChainAddress{
			{ChainId: 1, Address: common.HexToAddress("0x3333333333333333333333333333333333333333")},
		},
		VerificationType:        symbiotic.VerificationType(1),
		MaxVotingPower:          symbiotic.ToVotingPower(big.NewInt(10000000)),
		MinInclusionVotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
		MaxValidatorsCount:      symbiotic.ToVotingPower(big.NewInt(100)),
		RequiredKeyTags:         []symbiotic.KeyTag{1, 2, 3},
		RequiredHeaderKeyTag:    symbiotic.KeyTag(1),
		QuorumThresholds: []symbiotic.QuorumThreshold{
			{
				KeyTag:          symbiotic.KeyTag(1),
				QuorumThreshold: symbiotic.ToQuorumThresholdPct(big.NewInt(6667)),
			},
			{
				KeyTag:          symbiotic.KeyTag(2),
				QuorumThreshold: symbiotic.ToQuorumThresholdPct(big.NewInt(5000)),
			},
		},
		NumCommitters:         10,
		NumAggregators:        5,
		CommitterSlotDuration: 12,
	}

	bytes, err := networkConfigToBytes(original)
	require.NoError(t, err)
	require.NotEmpty(t, bytes)

	decoded, err := bytesToNetworkConfig(bytes)
	require.NoError(t, err)

	assert.Equal(t, original.VerificationType, decoded.VerificationType)
	assert.Equal(t, original.MaxVotingPower, decoded.MaxVotingPower)
	assert.Equal(t, original.MinInclusionVotingPower, decoded.MinInclusionVotingPower)
	assert.Equal(t, original.MaxValidatorsCount, decoded.MaxValidatorsCount)
	assert.Equal(t, original.RequiredHeaderKeyTag, decoded.RequiredHeaderKeyTag)
	assert.Equal(t, original.NumCommitters, decoded.NumCommitters)
	assert.Equal(t, original.NumAggregators, decoded.NumAggregators)
	assert.Equal(t, original.CommitterSlotDuration, decoded.CommitterSlotDuration)

	assert.Equal(t, original.KeysProvider, decoded.KeysProvider)
	assert.Equal(t, original.VotingPowerProviders, decoded.VotingPowerProviders)
	assert.Equal(t, original.Settlements, decoded.Settlements)
	assert.Equal(t, original.RequiredKeyTags, decoded.RequiredKeyTags)
	assert.Equal(t, original.QuorumThresholds, decoded.QuorumThresholds)
}

func TestValidatorSetHeaderProtoConversion(t *testing.T) {
	valset := symbiotic.ValidatorSet{
		Version:          1,
		RequiredKeyTag:   symbiotic.KeyTag(1),
		Epoch:            symbiotic.Epoch(100),
		CaptureTimestamp: symbiotic.Timestamp(1234567890),
		QuorumThreshold:  symbiotic.ToVotingPower(big.NewInt(6667)),
		Validators: symbiotic.Validators{
			{
				Operator:    common.HexToAddress("0x1111111111111111111111111111111111111111"),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(5000000)),
				IsActive:    true,
				Keys: []symbiotic.ValidatorKey{
					{Tag: 1, Payload: []byte("key1")},
				},
				Vaults: []symbiotic.ValidatorVault{},
			},
			{
				Operator:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(5000000)),
				IsActive:    true,
				Keys: []symbiotic.ValidatorKey{
					{Tag: 1, Payload: []byte("key2")},
				},
				Vaults: []symbiotic.ValidatorVault{},
			},
		},
		AggregatorIndices: []uint32{1, 2, 3},
		CommitterIndices:  []uint32{4, 5, 6},
	}

	bytes, err := validatorSetHeaderToBytes(valset)
	require.NoError(t, err)
	require.NotEmpty(t, bytes)

	decoded, err := bytesToValidatorSetHeader(bytes)
	require.NoError(t, err)

	assert.Equal(t, valset.Version, decoded.Version)
	assert.Equal(t, valset.RequiredKeyTag, decoded.RequiredKeyTag)
	assert.Equal(t, valset.Epoch, decoded.Epoch)
	assert.Equal(t, valset.CaptureTimestamp, decoded.CaptureTimestamp)
	assert.Equal(t, valset.QuorumThreshold, decoded.QuorumThreshold)
	assert.NotNil(t, decoded.TotalVotingPower)
	assert.NotEqual(t, common.Hash{}, decoded.ValidatorsSszMRoot)
}

func TestValidatorSetMetadataProtoConversion(t *testing.T) {
	original := symbiotic.ValidatorSetMetadata{
		RequestID: common.HexToHash("0x1234567890123456789012345678901234567890123456789012345678901234"),
		Epoch:     symbiotic.Epoch(100),
		ExtraData: []symbiotic.ExtraData{
			{
				Key:   common.HexToHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
				Value: common.HexToHash("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
			},
		},
		CommitmentData: []byte("test-commitment-data"),
	}

	bytes, err := validatorSetMetadataToBytes(original)
	require.NoError(t, err)
	require.NotEmpty(t, bytes)

	decoded, err := bytesToValidatorSetMetadata(bytes)
	require.NoError(t, err)

	assert.Equal(t, original.RequestID, decoded.RequestID)
	assert.Equal(t, original.Epoch, decoded.Epoch)
	assert.Equal(t, original.CommitmentData, decoded.CommitmentData)
	assert.Equal(t, original.ExtraData, decoded.ExtraData)
}

func TestSignatureMapProtoConversion(t *testing.T) {
	bitmap := entity.NewBitmapOf(1, 5, 10, 15)

	original := entity.SignatureMap{
		RequestID:              common.HexToHash("0x1234567890123456789012345678901234567890123456789012345678901234"),
		Epoch:                  symbiotic.Epoch(100),
		SignedValidatorsBitmap: bitmap,
		CurrentVotingPower:     symbiotic.ToVotingPower(big.NewInt(5000000)),
		TotalValidators:        20,
	}

	bytes, err := signatureMapToBytes(original)
	require.NoError(t, err)
	require.NotEmpty(t, bytes)

	decoded, err := bytesToSignatureMap(bytes)
	require.NoError(t, err)

	assert.Equal(t, original.RequestID, decoded.RequestID)
	assert.Equal(t, original.Epoch, decoded.Epoch)
	assert.Equal(t, original.CurrentVotingPower, decoded.CurrentVotingPower)
	assert.Equal(t, original.TotalValidators, decoded.TotalValidators)
	assert.Equal(t, bitmap.ToArray(), decoded.SignedValidatorsBitmap.ToArray())
}

func TestSignatureProtoConversion(t *testing.T) {
	t.Run("BLS signature conversion", func(t *testing.T) {
		privateKey, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
		require.NoError(t, err)

		original := symbiotic.Signature{
			MessageHash: []byte("test-message-hash-32-bytes-long!"),
			KeyTag:      symbiotic.KeyTag(4), // 4 is BLS (upper nibble = 0)
			Epoch:       symbiotic.Epoch(100),
			Signature:   []byte("test-signature-data"),
			PublicKey:   privateKey.PublicKey(),
		}

		bytes, err := signatureToBytes(original)
		require.NoError(t, err)
		require.NotEmpty(t, bytes)

		decoded, err := bytesToSignature(bytes)
		require.NoError(t, err)

		assert.Equal(t, original.MessageHash, decoded.MessageHash)
		assert.Equal(t, original.KeyTag, decoded.KeyTag)
		assert.Equal(t, original.Epoch, decoded.Epoch)
		assert.Equal(t, original.Signature, decoded.Signature)
		assert.Equal(t, original.PublicKey.Raw(), decoded.PublicKey.Raw())
	})

	t.Run("ECDSA signature conversion", func(t *testing.T) {
		privateKey, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeEcdsaSecp256k1)
		require.NoError(t, err)

		original := symbiotic.Signature{
			MessageHash: []byte("test-message-hash-32-bytes-long!"),
			KeyTag:      symbiotic.KeyTag(16), // 16 is ECDSA (upper nibble = 1)
			Epoch:       symbiotic.Epoch(100),
			Signature:   []byte("test-signature-data"),
			PublicKey:   privateKey.PublicKey(),
		}

		bytes, err := signatureToBytes(original)
		require.NoError(t, err)
		require.NotEmpty(t, bytes)

		decoded, err := bytesToSignature(bytes)
		require.NoError(t, err)

		assert.Equal(t, original.MessageHash, decoded.MessageHash)
		assert.Equal(t, original.KeyTag, decoded.KeyTag)
		assert.Equal(t, original.Epoch, decoded.Epoch)
		assert.Equal(t, original.Signature, decoded.Signature)
		assert.Equal(t, original.PublicKey.Raw(), decoded.PublicKey.Raw())
	})
}

func TestSignatureRequestProtoConversion(t *testing.T) {
	original := symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(1),
		RequiredEpoch: symbiotic.Epoch(100),
		Message:       []byte("test-message-to-sign"),
	}

	bytes, err := signatureRequestToBytes(original)
	require.NoError(t, err)
	require.NotEmpty(t, bytes)

	decoded, err := bytesToSignatureRequest(bytes)
	require.NoError(t, err)

	assert.Equal(t, original.KeyTag, decoded.KeyTag)
	assert.Equal(t, original.RequiredEpoch, decoded.RequiredEpoch)
	assert.Equal(t, original.Message, decoded.Message)
}

func TestAggregationProofProtoConversion(t *testing.T) {
	original := symbiotic.AggregationProof{
		MessageHash: []byte("test-message-hash-32-bytes-long!"),
		KeyTag:      symbiotic.KeyTag(1),
		Epoch:       symbiotic.Epoch(100),
		Proof:       []byte("test-aggregation-proof-data"),
	}

	bytes, err := aggregationProofToBytes(original)
	require.NoError(t, err)
	require.NotEmpty(t, bytes)

	decoded, err := bytesToAggregationProof(bytes)
	require.NoError(t, err)

	assert.Equal(t, original.MessageHash, decoded.MessageHash)
	assert.Equal(t, original.KeyTag, decoded.KeyTag)
	assert.Equal(t, original.Epoch, decoded.Epoch)
	assert.Equal(t, original.Proof, decoded.Proof)
}

func TestBigIntEdgeCases(t *testing.T) {
	testCases := []struct {
		name  string
		value *big.Int
	}{
		{"zero", big.NewInt(0)},
		{"small", big.NewInt(1)},
		{"large", new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil)},
		{"max_uint256", new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			original := symbiotic.NetworkConfig{
				VotingPowerProviders:    []symbiotic.CrossChainAddress{},
				KeysProvider:            symbiotic.CrossChainAddress{},
				Settlements:             []symbiotic.CrossChainAddress{},
				VerificationType:        0,
				MaxVotingPower:          symbiotic.ToVotingPower(tc.value),
				MinInclusionVotingPower: symbiotic.ToVotingPower(tc.value),
				MaxValidatorsCount:      symbiotic.ToVotingPower(tc.value),
				RequiredKeyTags:         []symbiotic.KeyTag{},
				RequiredHeaderKeyTag:    0,
				QuorumThresholds:        []symbiotic.QuorumThreshold{},
			}

			bytes, err := networkConfigToBytes(original)
			require.NoError(t, err)

			decoded, err := bytesToNetworkConfig(bytes)
			require.NoError(t, err)

			assert.Equal(t, original.MaxVotingPower, decoded.MaxVotingPower, "MaxVotingPower mismatch")
			assert.Equal(t, original.MinInclusionVotingPower, decoded.MinInclusionVotingPower, "MinInclusionVotingPower mismatch")
			assert.Equal(t, original.MaxValidatorsCount, decoded.MaxValidatorsCount, "MaxValidatorsCount mismatch")
		})
	}
}
