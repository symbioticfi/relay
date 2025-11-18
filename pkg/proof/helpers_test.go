package proof

import (
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetMaxValidators tests the GetMaxValidators function
func TestGetMaxValidators(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected []int
	}{
		{
			name:     "default values when env var is unset",
			envValue: "",
			expected: []int{10, 100, 1000},
		},
		{
			name:     "custom comma-separated values",
			envValue: "5,20,50",
			expected: []int{5, 20, 50},
		},
		{
			name:     "values with whitespace",
			envValue: "10, 100 , 1000",
			expected: []int{10, 100, 1000},
		},
		{
			name:     "ignores invalid non-integer values",
			envValue: "10,invalid,100",
			expected: []int{10, 100},
		},
		{
			name:     "ignores zero and negative values",
			envValue: "10,0,-5,100",
			expected: []int{10, 100},
		},
		{
			name:     "falls back to default if all values invalid",
			envValue: "invalid,zero,negative",
			expected: []int{10, 100, 1000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				t.Setenv("MAX_VALIDATORS", tt.envValue)
			} else {
				os.Unsetenv("MAX_VALIDATORS")
			}

			result := GetMaxValidators()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetOptimalN tests the getOptimalN function
func TestGetOptimalN(t *testing.T) {
	tests := []struct {
		name          string
		valsetLength  int
		maxValidators []int
		expected      int
	}{
		{
			name:          "exact match",
			valsetLength:  10,
			maxValidators: []int{10, 100, 1000},
			expected:      10,
		},
		{
			name:          "between sizes - returns next larger",
			valsetLength:  50,
			maxValidators: []int{10, 100, 1000},
			expected:      100,
		},
		{
			name:          "exceeds all sizes - returns 0",
			valsetLength:  2000,
			maxValidators: []int{10, 100, 1000},
			expected:      0,
		},
		{
			name:          "single validator",
			valsetLength:  1,
			maxValidators: []int{10, 100, 1000},
			expected:      10,
		},
		{
			name:          "at boundary",
			valsetLength:  1000,
			maxValidators: []int{10, 100, 1000},
			expected:      1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Temporarily override max validators
			envParts := make([]string, 0, len(tt.maxValidators))
			for _, v := range tt.maxValidators {
				envParts = append(envParts, big.NewInt(int64(v)).String())
			}
			envValue := strings.Join(envParts, ",")
			t.Setenv("MAX_VALIDATORS", envValue)

			result := getOptimalN(tt.valsetLength)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNormalizeValset tests the NormalizeValset function
func TestNormalizeValset(t *testing.T) {
	t.Run("sorts validators by key", func(t *testing.T) {
		valset := []ValidatorData{
			{
				PrivateKey:  big.NewInt(30),
				Key:         getPubkeyG1(big.NewInt(30)),
				KeyG2:       getPubkeyG2(big.NewInt(30)),
				VotingPower: big.NewInt(100),
			},
			{
				PrivateKey:  big.NewInt(10),
				Key:         getPubkeyG1(big.NewInt(10)),
				KeyG2:       getPubkeyG2(big.NewInt(10)),
				VotingPower: big.NewInt(200),
			},
			{
				PrivateKey:  big.NewInt(20),
				Key:         getPubkeyG1(big.NewInt(20)),
				KeyG2:       getPubkeyG2(big.NewInt(20)),
				VotingPower: big.NewInt(150),
			},
		}

		// Set MAX_VALIDATORS to ensure we get a specific size
		t.Setenv("MAX_VALIDATORS", "10")

		result := NormalizeValset(valset)

		// Should be sorted and padded to 10
		require.Len(t, result, 10)

		// Check that first 3 are sorted (we need to verify sorting logic)
		// The sorting is done by key.X or key.Y comparison
		for i := 0; i < 2; i++ {
			cmpX := result[i].Key.X.Cmp(&result[i+1].Key.X)
			cmpY := result[i].Key.Y.Cmp(&result[i+1].Key.Y)
			assert.True(t, cmpX < 0 || cmpY < 0, "validators should be sorted")
		}

		// Check that remaining slots are padded with zero points
		zeroPoint := new(bn254.G1Affine)
		zeroPoint.SetInfinity()
		for i := 3; i < 10; i++ {
			assert.Equal(t, zeroPoint.X, result[i].Key.X, "padding should have zero X")
			assert.Equal(t, zeroPoint.Y, result[i].Key.Y, "padding should have zero Y")
			assert.Equal(t, big.NewInt(0), result[i].VotingPower, "padding should have zero voting power")
		}
	})

	t.Run("handles empty validator set", func(t *testing.T) {
		t.Setenv("MAX_VALIDATORS", "10")

		result := NormalizeValset([]ValidatorData{})
		require.Len(t, result, 10)

		// All should be zero points
		zeroPoint := new(bn254.G1Affine)
		zeroPoint.SetInfinity()
		for i := 0; i < 10; i++ {
			assert.Equal(t, zeroPoint.X, result[i].Key.X)
			assert.Equal(t, zeroPoint.Y, result[i].Key.Y)
		}
	})
}

// TestHashValset tests the HashValset function
func TestHashValset(t *testing.T) {
	t.Run("produces consistent hash for same input", func(t *testing.T) {
		valset := genValset(3, []int{})
		hash1 := HashValset(valset)
		hash2 := HashValset(valset)

		assert.Equal(t, hash1, hash2, "hash should be deterministic")
		assert.Len(t, hash1, 32, "hash should be 32 bytes")
	})

	t.Run("produces different hash for different validators", func(t *testing.T) {
		valset1 := genValset(3, []int{})
		valset2 := genValset(5, []int{})

		hash1 := HashValset(valset1)
		hash2 := HashValset(valset2)

		assert.NotEqual(t, hash1, hash2, "different valsets should have different hashes")
	})

	t.Run("stops at zero point", func(t *testing.T) {
		valset := genValset(3, []int{})
		// Add zero point validator
		zeroPoint := new(bn254.G1Affine)
		zeroPoint.SetInfinity()
		valset = append(valset, ValidatorData{
			Key:         *zeroPoint,
			VotingPower: big.NewInt(100),
		})

		// Hash should ignore the zero point validator
		hash1 := HashValset(valset[:3])
		hash2 := HashValset(valset)

		assert.Equal(t, hash1, hash2, "hash should stop at zero point")
	})

	t.Run("handles empty valset", func(t *testing.T) {
		hash := HashValset([]ValidatorData{})
		assert.NotNil(t, hash)
		assert.Len(t, hash, 32)
	})
}

// TestGetNonSignersData tests the getNonSignersData function
func TestGetNonSignersData(t *testing.T) {
	t.Run("aggregates non-signers correctly", func(t *testing.T) {
		valset := genValset(5, []int{1, 3}) // Validators at index 1 and 3 are non-signers

		aggKey, aggVotingPower, totalVotingPower := getNonSignersData(valset)

		// Total voting power should be sum of all validators
		assert.Equal(t, big.NewInt(500), totalVotingPower, "total should be 5 * 100")

		// Non-signers voting power should be sum of 2 validators
		assert.Equal(t, big.NewInt(200), aggVotingPower, "non-signers should be 2 * 100")

		// Aggregated key should not be infinity (we have non-signers)
		zeroPoint := new(bn254.G1Affine)
		zeroPoint.SetInfinity()
		assert.False(t, aggKey.X.Cmp(&zeroPoint.X) == 0 && aggKey.Y.Cmp(&zeroPoint.Y) == 0,
			"aggKey should not be infinity when there are non-signers")
	})

	t.Run("handles no non-signers", func(t *testing.T) {
		valset := genValset(3, []int{}) // All are signers

		aggKey, aggVotingPower, totalVotingPower := getNonSignersData(valset)

		// Total voting power
		assert.Equal(t, big.NewInt(300), totalVotingPower)

		// No non-signers
		assert.Equal(t, big.NewInt(0), aggVotingPower)

		// Aggregated key should be infinity
		zeroPoint := new(bn254.G1Affine)
		zeroPoint.SetInfinity()
		assert.Equal(t, zeroPoint.X, aggKey.X)
		assert.Equal(t, zeroPoint.Y, aggKey.Y)
	})

	t.Run("handles all non-signers", func(t *testing.T) {
		valset := genValset(3, []int{0, 1, 2}) // All are non-signers

		_, aggVotingPower, totalVotingPower := getNonSignersData(valset)

		// All voting power goes to non-signers
		assert.Equal(t, totalVotingPower, aggVotingPower)
		assert.Equal(t, big.NewInt(300), aggVotingPower)
	})
}

// TestGetAggSignature tests the getAggSignature function
func TestGetAggSignature(t *testing.T) {
	t.Run("aggregates signatures from signers only", func(t *testing.T) {
		valset := genValset(5, []int{1, 3}) // Validators 1 and 3 are non-signers

		// Use a simple message point for testing
		messageG1 := getPubkeyG1(big.NewInt(999))

		signature, aggKeyG2, aggKeyG1 := getAggSignature(messageG1, &valset)

		// Should not be infinity points (we have signers)
		zeroPointG1 := new(bn254.G1Affine)
		zeroPointG1.SetInfinity()
		zeroPointG2 := new(bn254.G2Affine)
		zeroPointG2.SetInfinity()

		assert.False(t, signature.X.Cmp(&zeroPointG1.X) == 0 && signature.Y.Cmp(&zeroPointG1.Y) == 0,
			"signature should not be infinity")
		assert.False(t, aggKeyG1.X.Cmp(&zeroPointG1.X) == 0 && aggKeyG1.Y.Cmp(&zeroPointG1.Y) == 0,
			"aggKeyG1 should not be infinity")
		assert.NotNil(t, aggKeyG2)
	})

	t.Run("handles all signers", func(t *testing.T) {
		valset := genValset(3, []int{}) // All are signers

		// Use a simple message point for testing
		messageG1 := getPubkeyG1(big.NewInt(999))

		signature, aggKeyG2, aggKeyG1 := getAggSignature(messageG1, &valset)

		assert.NotNil(t, signature)
		assert.NotNil(t, aggKeyG2)
		assert.NotNil(t, aggKeyG1)
	})

	t.Run("handles all non-signers", func(t *testing.T) {
		valset := genValset(3, []int{0, 1, 2}) // All are non-signers

		// Use a simple message point for testing
		messageG1 := getPubkeyG1(big.NewInt(999))

		signature, _, aggKeyG1 := getAggSignature(messageG1, &valset)

		// Should be infinity points (no signers)
		zeroPointG1 := new(bn254.G1Affine)
		zeroPointG1.SetInfinity()

		assert.Equal(t, zeroPointG1.X, signature.X)
		assert.Equal(t, zeroPointG1.Y, signature.Y)
		assert.Equal(t, zeroPointG1.X, aggKeyG1.X)
		assert.Equal(t, zeroPointG1.Y, aggKeyG1.Y)
	})
}

// TestGetPubkeyG1 tests the getPubkeyG1 function
func TestGetPubkeyG1(t *testing.T) {
	t.Run("generates consistent key for same private key", func(t *testing.T) {
		pk := big.NewInt(12345)
		key1 := getPubkeyG1(pk)
		key2 := getPubkeyG1(pk)

		assert.Equal(t, key1.X, key2.X)
		assert.Equal(t, key1.Y, key2.Y)
	})

	t.Run("generates different keys for different private keys", func(t *testing.T) {
		pk1 := big.NewInt(12345)
		pk2 := big.NewInt(67890)

		key1 := getPubkeyG1(pk1)
		key2 := getPubkeyG1(pk2)

		assert.False(t, key1.X.Cmp(&key2.X) == 0 && key1.Y.Cmp(&key2.Y) == 0,
			"different private keys should generate different public keys")
	})

	t.Run("handles edge cases", func(t *testing.T) {
		// Zero should give infinity point
		pk0 := big.NewInt(0)
		key0 := getPubkeyG1(pk0)
		zeroPoint := new(bn254.G1Affine)
		zeroPoint.SetInfinity()
		assert.Equal(t, zeroPoint.X, key0.X)
		assert.Equal(t, zeroPoint.Y, key0.Y)

		// One should give generator
		pk1 := big.NewInt(1)
		key1 := getPubkeyG1(pk1)
		_, _, g1Gen, _ := bn254.Generators()
		assert.Equal(t, g1Gen.X, key1.X)
		assert.Equal(t, g1Gen.Y, key1.Y)
	})
}

// TestGetPubkeyG2 tests the getPubkeyG2 function
func TestGetPubkeyG2(t *testing.T) {
	t.Run("generates consistent key for same private key", func(t *testing.T) {
		pk := big.NewInt(12345)
		key1 := getPubkeyG2(pk)
		key2 := getPubkeyG2(pk)

		assert.Equal(t, key1.X.A0, key2.X.A0)
		assert.Equal(t, key1.X.A1, key2.X.A1)
		assert.Equal(t, key1.Y.A0, key2.Y.A0)
		assert.Equal(t, key1.Y.A1, key2.Y.A1)
	})

	t.Run("generates different keys for different private keys", func(t *testing.T) {
		pk1 := big.NewInt(12345)
		pk2 := big.NewInt(67890)

		key1 := getPubkeyG2(pk1)
		key2 := getPubkeyG2(pk2)

		assert.False(t,
			key1.X.A0.Cmp(&key2.X.A0) == 0 &&
				key1.X.A1.Cmp(&key2.X.A1) == 0 &&
				key1.Y.A0.Cmp(&key2.Y.A0) == 0 &&
				key1.Y.A1.Cmp(&key2.Y.A1) == 0,
			"different private keys should generate different public keys")
	})

	t.Run("handles edge cases", func(t *testing.T) {
		// Zero should give infinity point
		pk0 := big.NewInt(0)
		key0 := getPubkeyG2(pk0)
		zeroPoint := new(bn254.G2Affine)
		zeroPoint.SetInfinity()
		assert.Equal(t, zeroPoint.X.A0, key0.X.A0)
		assert.Equal(t, zeroPoint.X.A1, key0.X.A1)

		// One should give generator
		pk1 := big.NewInt(1)
		key1 := getPubkeyG2(pk1)
		_, _, _, g2Gen := bn254.Generators()
		assert.Equal(t, g2Gen.X.A0, key1.X.A0)
		assert.Equal(t, g2Gen.X.A1, key1.X.A1)
	})
}

// TestProofDataMarshal tests the ProofData.Marshal function
func TestProofDataMarshal(t *testing.T) {
	t.Run("marshals proof data correctly", func(t *testing.T) {
		proofData := ProofData{
			Proof:                 make([]byte, 256),
			Commitments:           make([]byte, 64),
			CommitmentPok:         make([]byte, 64),
			SignersAggVotingPower: big.NewInt(123456789),
		}

		// Fill with test data
		for i := range proofData.Proof {
			proofData.Proof[i] = byte(i % 256)
		}
		for i := range proofData.Commitments {
			proofData.Commitments[i] = byte((i + 1) % 256)
		}
		for i := range proofData.CommitmentPok {
			proofData.CommitmentPok[i] = byte((i + 2) % 256)
		}

		marshaled := proofData.Marshal()

		// Should be 256 + 64 + 64 + 32 = 416 bytes
		assert.Len(t, marshaled, 416)

		// Check that proof is at the beginning
		assert.Equal(t, proofData.Proof, marshaled[:256])

		// Check commitments
		assert.Equal(t, proofData.Commitments, marshaled[256:320])

		// Check commitment PoK
		assert.Equal(t, proofData.CommitmentPok, marshaled[320:384])

		// Check voting power (last 32 bytes)
		votingPowerBytes := make([]byte, 32)
		proofData.SignersAggVotingPower.FillBytes(votingPowerBytes)
		assert.Equal(t, votingPowerBytes, marshaled[384:416])
	})

	t.Run("handles zero voting power", func(t *testing.T) {
		proofData := ProofData{
			Proof:                 make([]byte, 256),
			Commitments:           make([]byte, 64),
			CommitmentPok:         make([]byte, 64),
			SignersAggVotingPower: big.NewInt(0),
		}

		marshaled := proofData.Marshal()
		assert.Len(t, marshaled, 416)

		// Last 32 bytes should be zero
		zeroBytes := make([]byte, 32)
		assert.Equal(t, zeroBytes, marshaled[384:416])
	})
}

// TestExists tests the exists function
func TestExists(t *testing.T) {
	t.Run("returns false for non-existent file", func(t *testing.T) {
		assert.False(t, exists("/nonexistent/file/path.txt"))
	})

	t.Run("returns false for empty string", func(t *testing.T) {
		assert.False(t, exists(""))
	})

	t.Run("returns true for current directory", func(t *testing.T) {
		assert.True(t, exists("."))
	})

	t.Run("returns true for existing file", func(t *testing.T) {
		// Create temp file
		tmpFile := t.TempDir() + "/test.txt"
		err := os.WriteFile(tmpFile, []byte("test"), 0600)
		require.NoError(t, err)

		assert.True(t, exists(tmpFile))
	})

	t.Run("returns true for existing directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		assert.True(t, exists(tmpDir))
	})

	t.Run("returns false for file in non-existent directory", func(t *testing.T) {
		assert.False(t, exists("/nonexistent/dir/file.txt"))
	})
}
