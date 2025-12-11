package badger

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestCachedRepository_NetworkConfig(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create base repository
	baseRepo, err := New(Config{Dir: tempDir, Metrics: DoNothingMetrics{}})
	require.NoError(t, err)
	defer func() {
		err := baseRepo.Close()
		require.NoError(t, err)
	}()

	// Create cached repository
	cachedRepo, err := NewCached(baseRepo, CachedConfig{
		NetworkConfigCacheSize: 10,
		ValidatorSetCacheSize:  10,
	})
	require.NoError(t, err)

	ctx := context.Background()
	epoch := symbiotic.Epoch(123)

	// Create test network config
	testConfig := symbiotic.NetworkConfig{
		VerificationType: symbiotic.VerificationTypeBlsBn254Simple,
	}

	// Test cache miss - config doesn't exist
	_, err = cachedRepo.GetConfigByEpoch(ctx, epoch)
	require.Error(t, err)

	// Save config
	err = cachedRepo.SaveConfig(ctx, testConfig, epoch)
	require.NoError(t, err)

	// Test cache hit - should retrieve from cache
	retrievedConfig, err := cachedRepo.GetConfigByEpoch(ctx, epoch)
	require.NoError(t, err)
	require.Equal(t, testConfig.VerificationType, retrievedConfig.VerificationType)

	// Test cache hit again - should still work
	retrievedConfig2, err := cachedRepo.GetConfigByEpoch(ctx, epoch)
	require.NoError(t, err)
	require.Equal(t, testConfig.VerificationType, retrievedConfig2.VerificationType)
}

func TestCachedRepository_InheritedMethods(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create base repository
	baseRepo, err := New(Config{Dir: tempDir, Metrics: DoNothingMetrics{}})
	require.NoError(t, err)
	defer func() {
		err := baseRepo.Close()
		require.NoError(t, err)
	}()

	// Create cached repository
	cachedRepo, err := NewCached(baseRepo, CachedConfig{
		NetworkConfigCacheSize: 10,
		ValidatorSetCacheSize:  10,
	})
	require.NoError(t, err)

	// Test that inherited methods work (non-cached methods)
	// This tests that embedding is working correctly
	err = cachedRepo.Close()
	require.NoError(t, err)
}

func TestCachedRepository_ValidatorSet(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create base repository
	baseRepo, err := New(Config{Dir: tempDir, Metrics: DoNothingMetrics{}})
	require.NoError(t, err)
	defer func() {
		err := baseRepo.Close()
		require.NoError(t, err)
	}()

	// Create cached repository
	cachedRepo, err := NewCached(baseRepo, CachedConfig{
		NetworkConfigCacheSize: 10,
		ValidatorSetCacheSize:  10,
	})
	require.NoError(t, err)

	ctx := context.Background()
	epoch := symbiotic.Epoch(456)

	// Create test validator set
	testValidatorSet := symbiotic.ValidatorSet{
		Version:          1,
		RequiredKeyTag:   symbiotic.KeyTag(15),
		Epoch:            epoch,
		CaptureTimestamp: 1234567890,
		QuorumThreshold:  symbiotic.ToVotingPower(big.NewInt(100)),
		Validators: symbiotic.Validators{
			{
				Operator:    common.HexToAddress("0x1234567890123456789012345678901234567890"),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(50)),
				IsActive:    true,
			},
		},
	}

	// Test cache miss - validator set doesn't exist
	_, err = cachedRepo.GetValidatorSetByEpoch(ctx, epoch)
	require.Error(t, err)

	// Save validator set
	err = cachedRepo.saveValidatorSet(ctx, testValidatorSet)
	require.NoError(t, err)

	// Test cache hit - should retrieve from cache
	retrievedValidatorSet, err := cachedRepo.GetValidatorSetByEpoch(ctx, epoch)
	require.NoError(t, err)
	require.Equal(t, testValidatorSet.Epoch, retrievedValidatorSet.Epoch)
	require.Equal(t, testValidatorSet.RequiredKeyTag, retrievedValidatorSet.RequiredKeyTag)
	require.Len(t, retrievedValidatorSet.Validators, len(testValidatorSet.Validators))

	// Test cache hit again - should still work
	retrievedValidatorSet2, err := cachedRepo.GetValidatorSetByEpoch(ctx, epoch)
	require.NoError(t, err)
	require.Equal(t, testValidatorSet.Epoch, retrievedValidatorSet2.Epoch)
}
