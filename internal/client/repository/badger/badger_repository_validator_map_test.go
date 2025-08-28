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

// randomRequestHash generates a random request hash for testing
func randomRequestHash(t *testing.T) common.Hash {
	t.Helper()
	return common.BytesToHash(randomBytes(t, 32))
}

// randomValidatorMap creates a ValidatorMap with test data
func randomValidatorMap(t *testing.T, requestHash common.Hash) entity.ValidatorMap {
	t.Helper()

	// Create addresses for validators
	addr1 := common.BytesToAddress(randomBytes(t, 20))
	addr2 := common.BytesToAddress(randomBytes(t, 20))
	addr3 := common.BytesToAddress(randomBytes(t, 20))

	activeValidators := map[common.Address]struct{}{
		addr1: {},
		addr2: {},
		addr3: {},
	}

	// Some validators are present (have provided signatures)
	isPresent := map[common.Address]struct{}{
		addr1: {},
		addr2: {},
	}

	return entity.ValidatorMap{
		RequestHash:         requestHash,
		Epoch:               randomBigInt(t).Uint64(),
		ActiveValidatorsMap: activeValidators,
		IsPresent:           isPresent,
		QuorumThreshold:     entity.ToVotingPower(randomBigInt(t)),
		TotalVotingPower:    entity.ToVotingPower(randomBigInt(t)),
		CurrentVotingPower:  entity.ToVotingPower(randomBigInt(t)),
	}
}

// assertValidatorMapsEqual performs deep equality check on ValidatorMaps
func assertValidatorMapsEqual(t *testing.T, expected, actual entity.ValidatorMap) {
	t.Helper()

	assert.Equal(t, expected.RequestHash, actual.RequestHash, "RequestHash mismatch")
	assert.Equal(t, expected.Epoch, actual.Epoch, "Epoch mismatch")
	assert.Equal(t, expected.ActiveValidatorsMap, actual.ActiveValidatorsMap, "ActiveValidatorsMap mismatch")
	assert.Equal(t, expected.IsPresent, actual.IsPresent, "IsPresent mismatch")
	assert.Equal(t, expected.QuorumThreshold.String(), actual.QuorumThreshold.String(), "QuorumThreshold mismatch")
	assert.Equal(t, expected.TotalVotingPower.String(), actual.TotalVotingPower.String(), "TotalVotingPower mismatch")
	assert.Equal(t, expected.CurrentVotingPower.String(), actual.CurrentVotingPower.String(), "CurrentVotingPower mismatch")
}

func TestBadgerRepository_ValidatorMap(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	requestHash1 := randomRequestHash(t)
	requestHash2 := randomRequestHash(t)
	vm1 := randomValidatorMap(t, requestHash1)
	vm2 := randomValidatorMap(t, requestHash2)

	t.Run("UpdateValidatorMap - Success", func(t *testing.T) {
		err := repo.UpdateValidatorMap(context.Background(), vm1)
		require.NoError(t, err)

		// Verify data was saved correctly
		retrieved, err := repo.GetValidatorMap(context.Background(), requestHash1)
		require.NoError(t, err)
		assertValidatorMapsEqual(t, vm1, retrieved)
	})

	t.Run("UpdateValidatorMap - Update Existing", func(t *testing.T) {
		// Save initial validator map
		err := repo.UpdateValidatorMap(context.Background(), vm1)
		require.NoError(t, err)

		// Update with modified data
		updatedVM := vm1
		updatedVM.Epoch = vm1.Epoch + 1
		updatedVM.CurrentVotingPower = entity.ToVotingPower(big.NewInt(999))

		err = repo.UpdateValidatorMap(context.Background(), updatedVM)
		require.NoError(t, err)

		// Verify updated data
		retrieved, err := repo.GetValidatorMap(context.Background(), requestHash1)
		require.NoError(t, err)
		assertValidatorMapsEqual(t, updatedVM, retrieved)
	})

	t.Run("GetValidatorMap - Success", func(t *testing.T) {
		// Save two different validator maps
		err := repo.UpdateValidatorMap(context.Background(), vm1)
		require.NoError(t, err)
		err = repo.UpdateValidatorMap(context.Background(), vm2)
		require.NoError(t, err)

		// Retrieve first validator map
		retrieved1, err := repo.GetValidatorMap(context.Background(), requestHash1)
		require.NoError(t, err)
		assertValidatorMapsEqual(t, vm1, retrieved1)

		// Retrieve second validator map
		retrieved2, err := repo.GetValidatorMap(context.Background(), requestHash2)
		require.NoError(t, err)
		assertValidatorMapsEqual(t, vm2, retrieved2)
	})

	t.Run("GetValidatorMap - Not Found", func(t *testing.T) {
		nonExistentHash := randomRequestHash(t)

		_, err := repo.GetValidatorMap(context.Background(), nonExistentHash)
		require.Error(t, err)
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound), "Expected ErrEntityNotFound, got: %v", err)
	})

	t.Run("Multiple ValidatorMaps - Independence", func(t *testing.T) {
		// Create multiple validator maps
		hashes := make([]common.Hash, 5)
		vms := make([]entity.ValidatorMap, 5)

		for i := 0; i < 5; i++ {
			hashes[i] = randomRequestHash(t)
			vms[i] = randomValidatorMap(t, hashes[i])

			err := repo.UpdateValidatorMap(context.Background(), vms[i])
			require.NoError(t, err)
		}

		// Verify all can be retrieved correctly
		for i := 0; i < 5; i++ {
			retrieved, err := repo.GetValidatorMap(context.Background(), hashes[i])
			require.NoError(t, err)
			assertValidatorMapsEqual(t, vms[i], retrieved)
		}
	})
}

func TestValidatorMapSerialization(t *testing.T) {
	t.Parallel()

	t.Run("Serialization Round-Trip - Complete Data", func(t *testing.T) {
		original := randomValidatorMap(t, randomRequestHash(t))

		// Serialize
		data, err := validatorMapToBytes(original)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		// Deserialize
		deserialized, err := bytesToValidatorMap(data)
		require.NoError(t, err)

		// Verify round-trip preservation
		assertValidatorMapsEqual(t, original, deserialized)
	})

	t.Run("Serialization - Empty Maps", func(t *testing.T) {
		vm := entity.ValidatorMap{
			RequestHash:         randomRequestHash(t),
			Epoch:               123,
			ActiveValidatorsMap: make(map[common.Address]struct{}),
			IsPresent:           make(map[common.Address]struct{}),
			QuorumThreshold:     entity.ToVotingPower(big.NewInt(1000)),
			TotalVotingPower:    entity.ToVotingPower(big.NewInt(5000)),
			CurrentVotingPower:  entity.ToVotingPower(big.NewInt(0)),
		}

		// Serialize
		data, err := validatorMapToBytes(vm)
		require.NoError(t, err)

		// Deserialize
		deserialized, err := bytesToValidatorMap(data)
		require.NoError(t, err)

		assertValidatorMapsEqual(t, vm, deserialized)
	})

	t.Run("Serialization - Large Numbers", func(t *testing.T) {
		// Create ValidatorMap with large big.Int values
		largeBigInt := new(big.Int)
		largeBigInt.SetString("123456789012345678901234567890", 10)

		vm := entity.ValidatorMap{
			RequestHash:         randomRequestHash(t),
			Epoch:               18446744073709551615, // Max uint64
			ActiveValidatorsMap: make(map[common.Address]struct{}),
			IsPresent:           make(map[common.Address]struct{}),
			QuorumThreshold:     entity.ToVotingPower(largeBigInt),
			TotalVotingPower:    entity.ToVotingPower(new(big.Int).Mul(largeBigInt, big.NewInt(2))),
			CurrentVotingPower:  entity.ToVotingPower(new(big.Int).Mul(largeBigInt, big.NewInt(3))),
		}

		// Serialize and deserialize
		data, err := validatorMapToBytes(vm)
		require.NoError(t, err)

		deserialized, err := bytesToValidatorMap(data)
		require.NoError(t, err)

		assertValidatorMapsEqual(t, vm, deserialized)
	})

	t.Run("Serialization - Address Conversion", func(t *testing.T) {
		// Test specific addresses to ensure proper hex conversion
		addr1 := common.HexToAddress("0x1234567890123456789012345678901234567890")
		addr2 := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")

		vm := entity.ValidatorMap{
			RequestHash: randomRequestHash(t),
			Epoch:       42,
			ActiveValidatorsMap: map[common.Address]struct{}{
				addr1: {},
				addr2: {},
			},
			IsPresent: map[common.Address]struct{}{
				addr1: {},
			},
			QuorumThreshold:    entity.ToVotingPower(big.NewInt(100)),
			TotalVotingPower:   entity.ToVotingPower(big.NewInt(200)),
			CurrentVotingPower: entity.ToVotingPower(big.NewInt(150)),
		}

		// Serialize and deserialize
		data, err := validatorMapToBytes(vm)
		require.NoError(t, err)

		deserialized, err := bytesToValidatorMap(data)
		require.NoError(t, err)

		assertValidatorMapsEqual(t, vm, deserialized)

		// Verify specific addresses are preserved
		_, hasAddr1 := deserialized.ActiveValidatorsMap[addr1]
		_, hasAddr2 := deserialized.ActiveValidatorsMap[addr2]
		assert.True(t, hasAddr1, "Address 1 should be present in ActiveValidatorsMap")
		assert.True(t, hasAddr2, "Address 2 should be present in ActiveValidatorsMap")

		_, isPresentAddr1 := deserialized.IsPresent[addr1]
		_, isPresentAddr2 := deserialized.IsPresent[addr2]
		assert.True(t, isPresentAddr1, "Address 1 should be present in IsPresent")
		assert.False(t, isPresentAddr2, "Address 2 should NOT be present in IsPresent")
	})

	t.Run("Deserialization - Invalid JSON", func(t *testing.T) {
		invalidJSON := []byte(`{"invalid": "json", "missing_fields"}`)

		_, err := bytesToValidatorMap(invalidJSON)
		assert.Error(t, err, "Should fail with invalid JSON")
	})
}

func TestValidatorMapTransactions(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	t.Run("Operations Within Transaction", func(t *testing.T) {
		requestHash := randomRequestHash(t)
		vm := randomValidatorMap(t, requestHash)

		err := repo.DoUpdateInTx(context.Background(), func(ctx context.Context) error {
			// Update within transaction
			err := repo.UpdateValidatorMap(ctx, vm)
			require.NoError(t, err)

			// Get within same transaction - should work
			retrieved, err := repo.GetValidatorMap(ctx, requestHash)
			require.NoError(t, err)
			assertValidatorMapsEqual(t, vm, retrieved)

			return nil
		})
		require.NoError(t, err)

		// Verify data persisted after transaction
		retrieved, err := repo.GetValidatorMap(context.Background(), requestHash)
		require.NoError(t, err)
		assertValidatorMapsEqual(t, vm, retrieved)
	})

	t.Run("Transaction Rollback", func(t *testing.T) {
		requestHash := randomRequestHash(t)
		vm := randomValidatorMap(t, requestHash)

		// Transaction that will rollback due to error
		err := repo.DoUpdateInTx(context.Background(), func(ctx context.Context) error {
			err := repo.UpdateValidatorMap(ctx, vm)
			require.NoError(t, err)

			// Verify data exists within transaction
			_, err = repo.GetValidatorMap(ctx, requestHash)
			require.NoError(t, err)

			// Return error to trigger rollback
			return errors.New("intentional error for rollback")
		})
		require.Error(t, err)

		// Verify data was not persisted due to rollback
		_, err = repo.GetValidatorMap(context.Background(), requestHash)
		require.Error(t, err)
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
	})

	t.Run("Mixed Read and Write in Transaction", func(t *testing.T) {
		// Setup existing data
		existingHash := randomRequestHash(t)
		existingVM := randomValidatorMap(t, existingHash)
		err := repo.UpdateValidatorMap(context.Background(), existingVM)
		require.NoError(t, err)

		newHash := randomRequestHash(t)
		newVM := randomValidatorMap(t, newHash)

		err = repo.DoUpdateInTx(context.Background(), func(ctx context.Context) error {
			// Read existing data
			retrieved, err := repo.GetValidatorMap(ctx, existingHash)
			require.NoError(t, err)
			assertValidatorMapsEqual(t, existingVM, retrieved)

			// Write new data
			err = repo.UpdateValidatorMap(ctx, newVM)
			require.NoError(t, err)

			// Read newly written data within same transaction
			newRetrieved, err := repo.GetValidatorMap(ctx, newHash)
			require.NoError(t, err)
			assertValidatorMapsEqual(t, newVM, newRetrieved)

			return nil
		})
		require.NoError(t, err)

		// Verify both datasets exist after transaction
		retrieved1, err := repo.GetValidatorMap(context.Background(), existingHash)
		require.NoError(t, err)
		assertValidatorMapsEqual(t, existingVM, retrieved1)

		retrieved2, err := repo.GetValidatorMap(context.Background(), newHash)
		require.NoError(t, err)
		assertValidatorMapsEqual(t, newVM, retrieved2)
	})
}

func TestValidatorMapKeyGeneration(t *testing.T) {
	t.Parallel()

	t.Run("Key Format", func(t *testing.T) {
		hash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		key := keyValidatorMap(hash)

		expectedKey := "validator_map:0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		assert.Equal(t, expectedKey, string(key))
	})

	t.Run("Key Consistency", func(t *testing.T) {
		hash := randomRequestHash(t)

		key1 := keyValidatorMap(hash)
		key2 := keyValidatorMap(hash)

		assert.Equal(t, key1, key2, "Same hash should produce same key")
	})

	t.Run("Key Uniqueness", func(t *testing.T) {
		hash1 := randomRequestHash(t)
		hash2 := randomRequestHash(t)

		key1 := keyValidatorMap(hash1)
		key2 := keyValidatorMap(hash2)

		assert.NotEqual(t, key1, key2, "Different hashes should produce different keys")
	})
}

func TestValidatorMapEdgeCases(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	t.Run("Zero Values", func(t *testing.T) {
		vm := entity.ValidatorMap{
			RequestHash:         common.Hash{}, // Zero hash
			Epoch:               0,
			ActiveValidatorsMap: make(map[common.Address]struct{}),
			IsPresent:           make(map[common.Address]struct{}),
			QuorumThreshold:     entity.ToVotingPower(big.NewInt(0)),
			TotalVotingPower:    entity.ToVotingPower(big.NewInt(0)),
			CurrentVotingPower:  entity.ToVotingPower(big.NewInt(0)),
		}

		err := repo.UpdateValidatorMap(context.Background(), vm)
		require.NoError(t, err)

		retrieved, err := repo.GetValidatorMap(context.Background(), common.Hash{})
		require.NoError(t, err)
		assertValidatorMapsEqual(t, vm, retrieved)
	})

	t.Run("Single Validator", func(t *testing.T) {
		singleAddr := common.BytesToAddress(randomBytes(t, 20))
		vm := entity.ValidatorMap{
			RequestHash: randomRequestHash(t),
			Epoch:       1,
			ActiveValidatorsMap: map[common.Address]struct{}{
				singleAddr: {},
			},
			IsPresent: map[common.Address]struct{}{
				singleAddr: {},
			},
			QuorumThreshold:    entity.ToVotingPower(big.NewInt(1)),
			TotalVotingPower:   entity.ToVotingPower(big.NewInt(1)),
			CurrentVotingPower: entity.ToVotingPower(big.NewInt(1)),
		}

		err := repo.UpdateValidatorMap(context.Background(), vm)
		require.NoError(t, err)

		retrieved, err := repo.GetValidatorMap(context.Background(), vm.RequestHash)
		require.NoError(t, err)
		assertValidatorMapsEqual(t, vm, retrieved)
	})

	t.Run("Many Validators", func(t *testing.T) {
		// Create validator map with many validators
		activeValidators := make(map[common.Address]struct{})
		isPresent := make(map[common.Address]struct{})

		for i := 0; i < 100; i++ {
			addr := common.BytesToAddress(randomBytes(t, 20))
			activeValidators[addr] = struct{}{}

			// Half of them are present
			if i%2 == 0 {
				isPresent[addr] = struct{}{}
			}
		}

		vm := entity.ValidatorMap{
			RequestHash:         randomRequestHash(t),
			Epoch:               100,
			ActiveValidatorsMap: activeValidators,
			IsPresent:           isPresent,
			QuorumThreshold:     entity.ToVotingPower(big.NewInt(5000)),
			TotalVotingPower:    entity.ToVotingPower(big.NewInt(10000)),
			CurrentVotingPower:  entity.ToVotingPower(big.NewInt(5000)),
		}

		err := repo.UpdateValidatorMap(context.Background(), vm)
		require.NoError(t, err)

		retrieved, err := repo.GetValidatorMap(context.Background(), vm.RequestHash)
		require.NoError(t, err)
		assertValidatorMapsEqual(t, vm, retrieved)
	})
}
