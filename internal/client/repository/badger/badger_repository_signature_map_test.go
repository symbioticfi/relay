package badger

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/core/entity"
)

// randomSignatureTargetID generates a random signature target ID for testing
func randomSignatureTargetID(t *testing.T) common.Hash {
	t.Helper()
	return common.BytesToHash(randomBytes(t, 32))
}

// randomSignatureMap creates a SignatureMap with test data
func randomSignatureMap(t *testing.T, signatureTargetId common.Hash) entity.SignatureMap {
	t.Helper()

	return entity.SignatureMap{
		SignatureTargetID:      signatureTargetId,
		Epoch:                  entity.Epoch(randomBigInt(t).Uint64()),
		SignedValidatorsBitmap: entity.NewBitmapOf(0, 1, 2),
		CurrentVotingPower:     entity.ToVotingPower(randomBigInt(t)),
	}
}

// assertSignatureMapsEqual performs deep equality check on SignatureMaps
func assertSignatureMapsEqual(t *testing.T, expected, actual entity.SignatureMap) {
	t.Helper()

	assert.Equal(t, expected.SignatureTargetID, actual.SignatureTargetID, "SignatureTargetID mismatch")
	assert.Equal(t, expected.Epoch, actual.Epoch, "Epoch mismatch")
	assert.True(t, expected.SignedValidatorsBitmap.Equals(actual.SignedValidatorsBitmap.Bitmap), "SignedValidatorsBitmap mismatch")
	assert.Equal(t, expected.CurrentVotingPower.String(), actual.CurrentVotingPower.String(), "CurrentVotingPower mismatch")
}

func TestBadgerRepository_SignatureMap(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	signatureTargetID1 := randomSignatureTargetID(t)
	signatureTargetID2 := randomSignatureTargetID(t)
	vm1 := randomSignatureMap(t, signatureTargetID1)
	vm2 := randomSignatureMap(t, signatureTargetID2)

	t.Run("UpdateSignatureMap - Success", func(t *testing.T) {
		err := repo.UpdateSignatureMap(context.Background(), vm1)
		require.NoError(t, err)

		// Verify data was saved correctly
		retrieved, err := repo.GetSignatureMap(context.Background(), signatureTargetID1)
		require.NoError(t, err)
		assertSignatureMapsEqual(t, vm1, retrieved)
	})

	t.Run("UpdateSignatureMap - Update Existing", func(t *testing.T) {
		// Save initial signature map
		err := repo.UpdateSignatureMap(context.Background(), vm1)
		require.NoError(t, err)

		// Update with modified data
		updatedVM := vm1
		updatedVM.Epoch = vm1.Epoch + 1
		updatedVM.CurrentVotingPower = entity.ToVotingPower(big.NewInt(999))

		err = repo.UpdateSignatureMap(context.Background(), updatedVM)
		require.NoError(t, err)

		// Verify updated data
		retrieved, err := repo.GetSignatureMap(context.Background(), signatureTargetID1)
		require.NoError(t, err)
		assertSignatureMapsEqual(t, updatedVM, retrieved)
	})

	t.Run("GetSignatureMap - Success", func(t *testing.T) {
		// Save two different signature maps
		err := repo.UpdateSignatureMap(context.Background(), vm1)
		require.NoError(t, err)
		err = repo.UpdateSignatureMap(context.Background(), vm2)
		require.NoError(t, err)

		// Retrieve first signature map
		retrieved1, err := repo.GetSignatureMap(context.Background(), signatureTargetID1)
		require.NoError(t, err)
		assertSignatureMapsEqual(t, vm1, retrieved1)

		// Retrieve second signature map
		retrieved2, err := repo.GetSignatureMap(context.Background(), signatureTargetID2)
		require.NoError(t, err)
		assertSignatureMapsEqual(t, vm2, retrieved2)
	})

	t.Run("GetSignatureMap - Not Found", func(t *testing.T) {
		nonExistentHash := randomSignatureTargetID(t)

		_, err := repo.GetSignatureMap(context.Background(), nonExistentHash)
		require.Error(t, err)
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound), "Expected ErrEntityNotFound, got: %v", err)
	})

	t.Run("Multiple SignatureMaps - Independence", func(t *testing.T) {
		// Create multiple signature maps
		hashes := make([]common.Hash, 5)
		vms := make([]entity.SignatureMap, 5)

		for i := 0; i < 5; i++ {
			hashes[i] = randomSignatureTargetID(t)
			vms[i] = randomSignatureMap(t, hashes[i])

			err := repo.UpdateSignatureMap(context.Background(), vms[i])
			require.NoError(t, err)
		}

		// Verify all can be retrieved correctly
		for i := 0; i < 5; i++ {
			retrieved, err := repo.GetSignatureMap(context.Background(), hashes[i])
			require.NoError(t, err)
			assertSignatureMapsEqual(t, vms[i], retrieved)
		}
	})
}

func TestSignatureMapSerialization(t *testing.T) {
	t.Parallel()

	t.Run("Serialization Round-Trip - Complete Data", func(t *testing.T) {
		original := randomSignatureMap(t, randomSignatureTargetID(t))

		// Serialize
		data, err := signatureMapToBytes(original)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		// Deserialize
		deserialized, err := bytesToSignatureMap(data)
		require.NoError(t, err)

		// Verify round-trip preservation
		assertSignatureMapsEqual(t, original, deserialized)
	})

	t.Run("Serialization - Empty Maps", func(t *testing.T) {
		vm := entity.SignatureMap{
			SignatureTargetID:      randomSignatureTargetID(t),
			Epoch:                  123,
			SignedValidatorsBitmap: entity.NewBitmap(),
			CurrentVotingPower:     entity.ToVotingPower(big.NewInt(0)),
		}

		// Serialize
		data, err := signatureMapToBytes(vm)
		require.NoError(t, err)

		// Deserialize
		deserialized, err := bytesToSignatureMap(data)
		require.NoError(t, err)

		assertSignatureMapsEqual(t, vm, deserialized)
	})

	t.Run("Serialization - Large Numbers", func(t *testing.T) {
		// Create SignatureMap with large big.Int values
		largeBigInt := new(big.Int)
		largeBigInt.SetString("123456789012345678901234567890", 10)

		vm := entity.SignatureMap{
			SignatureTargetID:      randomSignatureTargetID(t),
			Epoch:                  18446744073709551615, // Max uint64
			SignedValidatorsBitmap: entity.NewBitmap(),
			CurrentVotingPower:     entity.ToVotingPower(new(big.Int).Mul(largeBigInt, big.NewInt(3))),
		}

		// Serialize and deserialize
		data, err := signatureMapToBytes(vm)
		require.NoError(t, err)

		deserialized, err := bytesToSignatureMap(data)
		require.NoError(t, err)

		assertSignatureMapsEqual(t, vm, deserialized)
	})

	t.Run("Serialization - Address Conversion", func(t *testing.T) {
		// Test roaring bitmap with specific indexes
		bitmap := entity.NewBitmapOf(0) // Only validator at index 0 is present
		vm := entity.SignatureMap{
			SignatureTargetID:      randomSignatureTargetID(t),
			Epoch:                  42,
			SignedValidatorsBitmap: bitmap,
			CurrentVotingPower:     entity.ToVotingPower(big.NewInt(150)),
		}

		// Serialize and deserialize
		data, err := signatureMapToBytes(vm)
		require.NoError(t, err)

		deserialized, err := bytesToSignatureMap(data)
		require.NoError(t, err)

		assertSignatureMapsEqual(t, vm, deserialized)

		// Verify bitmap contains expected validator index
		assert.True(t, deserialized.SignedValidatorsBitmap.Contains(0), "Validator index 0 should be present in bitmap")
	})

	t.Run("Deserialization - Invalid JSON", func(t *testing.T) {
		invalidJSON := []byte(`{"invalid": "json", "missing_fields"}`)

		_, err := bytesToSignatureMap(invalidJSON)
		assert.Error(t, err, "Should fail with invalid JSON")
	})
}

func TestSignatureMapTransactions(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	t.Run("Operations Within Transaction", func(t *testing.T) {
		signatureTargetID := randomSignatureTargetID(t)
		vm := randomSignatureMap(t, signatureTargetID)

		err := repo.DoUpdateInTx(context.Background(), func(ctx context.Context) error {
			// Update within transaction
			err := repo.UpdateSignatureMap(ctx, vm)
			require.NoError(t, err)

			// Get within same transaction - should work
			retrieved, err := repo.GetSignatureMap(ctx, signatureTargetID)
			require.NoError(t, err)
			assertSignatureMapsEqual(t, vm, retrieved)

			return nil
		})
		require.NoError(t, err)

		// Verify data persisted after transaction
		retrieved, err := repo.GetSignatureMap(context.Background(), signatureTargetID)
		require.NoError(t, err)
		assertSignatureMapsEqual(t, vm, retrieved)
	})

	t.Run("Transaction Rollback", func(t *testing.T) {
		signatureTargetID := randomSignatureTargetID(t)
		vm := randomSignatureMap(t, signatureTargetID)

		// Transaction that will rollback due to error
		err := repo.DoUpdateInTx(context.Background(), func(ctx context.Context) error {
			err := repo.UpdateSignatureMap(ctx, vm)
			require.NoError(t, err)

			// Verify data exists within transaction
			_, err = repo.GetSignatureMap(ctx, signatureTargetID)
			require.NoError(t, err)

			// Return error to trigger rollback
			return errors.New("intentional error for rollback")
		})
		require.Error(t, err)

		// Verify data was not persisted due to rollback
		_, err = repo.GetSignatureMap(context.Background(), signatureTargetID)
		require.Error(t, err)
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
	})

	t.Run("Mixed Read and Write in Transaction", func(t *testing.T) {
		// Setup existing data
		existingHash := randomSignatureTargetID(t)
		existingVM := randomSignatureMap(t, existingHash)
		err := repo.UpdateSignatureMap(context.Background(), existingVM)
		require.NoError(t, err)

		newHash := randomSignatureTargetID(t)
		newVM := randomSignatureMap(t, newHash)

		err = repo.DoUpdateInTx(context.Background(), func(ctx context.Context) error {
			// Read existing data
			retrieved, err := repo.GetSignatureMap(ctx, existingHash)
			require.NoError(t, err)
			assertSignatureMapsEqual(t, existingVM, retrieved)

			// Write new data
			err = repo.UpdateSignatureMap(ctx, newVM)
			require.NoError(t, err)

			// Read newly written data within same transaction
			newRetrieved, err := repo.GetSignatureMap(ctx, newHash)
			require.NoError(t, err)
			assertSignatureMapsEqual(t, newVM, newRetrieved)

			return nil
		})
		require.NoError(t, err)

		// Verify both datasets exist after transaction
		retrieved1, err := repo.GetSignatureMap(context.Background(), existingHash)
		require.NoError(t, err)
		assertSignatureMapsEqual(t, existingVM, retrieved1)

		retrieved2, err := repo.GetSignatureMap(context.Background(), newHash)
		require.NoError(t, err)
		assertSignatureMapsEqual(t, newVM, retrieved2)
	})
}

func TestSignatureMapKeyGeneration(t *testing.T) {
	t.Parallel()

	t.Run("Key Format", func(t *testing.T) {
		hash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		key := keySignatureMap(hash)

		expectedKey := "signature_map:0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		assert.Equal(t, expectedKey, string(key))
	})

	t.Run("Key Consistency", func(t *testing.T) {
		hash := randomSignatureTargetID(t)

		key1 := keySignatureMap(hash)
		key2 := keySignatureMap(hash)

		assert.Equal(t, key1, key2, "Same hash should produce same key")
	})

	t.Run("Key Uniqueness", func(t *testing.T) {
		hash1 := randomSignatureTargetID(t)
		hash2 := randomSignatureTargetID(t)

		key1 := keySignatureMap(hash1)
		key2 := keySignatureMap(hash2)

		assert.NotEqual(t, key1, key2, "Different hashes should produce different keys")
	})
}

func TestSignatureMapEdgeCases(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	t.Run("Zero Values", func(t *testing.T) {
		vm := entity.SignatureMap{
			SignatureTargetID:      common.Hash{}, // Zero hash
			Epoch:                  0,
			SignedValidatorsBitmap: entity.NewBitmap(),
			CurrentVotingPower:     entity.ToVotingPower(big.NewInt(0)),
		}

		err := repo.UpdateSignatureMap(context.Background(), vm)
		require.NoError(t, err)

		retrieved, err := repo.GetSignatureMap(context.Background(), common.Hash{})
		require.NoError(t, err)
		assertSignatureMapsEqual(t, vm, retrieved)
	})

	t.Run("Single Validator", func(t *testing.T) {
		// Test single validator scenario
		vm := entity.SignatureMap{
			SignatureTargetID:      randomSignatureTargetID(t),
			Epoch:                  1,
			SignedValidatorsBitmap: entity.NewBitmapOf(0), // Single validator at index 0
			CurrentVotingPower:     entity.ToVotingPower(big.NewInt(1)),
		}

		err := repo.UpdateSignatureMap(context.Background(), vm)
		require.NoError(t, err)

		retrieved, err := repo.GetSignatureMap(context.Background(), vm.SignatureTargetID)
		require.NoError(t, err)
		assertSignatureMapsEqual(t, vm, retrieved)
	})

	t.Run("Many Validators", func(t *testing.T) {
		// Create signature map with many validators
		bitmap := entity.NewBitmap()
		// Add even indexes (50 validators present out of 100)
		for i := uint32(0); i < 100; i += 2 {
			bitmap.Add(i)
		}

		vm := entity.SignatureMap{
			SignatureTargetID:      randomSignatureTargetID(t),
			Epoch:                  100,
			SignedValidatorsBitmap: bitmap,
			CurrentVotingPower:     entity.ToVotingPower(big.NewInt(5000)),
		}

		err := repo.UpdateSignatureMap(context.Background(), vm)
		require.NoError(t, err)

		retrieved, err := repo.GetSignatureMap(context.Background(), vm.SignatureTargetID)
		require.NoError(t, err)
		assertSignatureMapsEqual(t, vm, retrieved)
	})
}
