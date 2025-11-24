package badger

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

func TestRepository_PruneAllEntityTypes(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)
	ctx := t.Context()

	epoch := symbiotic.Epoch(100)

	// Setup: Create all entities for the epoch

	// 1. Create and save validator set
	priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	validator := symbiotic.Validator{
		Operator:    common.BytesToAddress(randomBytes(t, 20)),
		VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
		IsActive:    true,
		Keys: []symbiotic.ValidatorKey{
			{
				Tag:     symbiotic.KeyTag(15),
				Payload: randomBytes(t, 96), // CompactPublicKey as bytes
			},
		},
		Vaults: []symbiotic.ValidatorVault{
			{
				ChainID:     1,
				Vault:       common.BytesToAddress(randomBytes(t, 20)),
				VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
			},
		},
	}

	valset := symbiotic.ValidatorSet{
		Version:          1,
		RequiredKeyTag:   symbiotic.KeyTag(15),
		Epoch:            epoch,
		CaptureTimestamp: 1234567890,
		QuorumThreshold:  symbiotic.ToVotingPower(big.NewInt(500)),
		Validators:       []symbiotic.Validator{validator},
		Status:           symbiotic.HeaderDerived,
	}

	err = repo.SaveValidatorSet(ctx, valset)
	require.NoError(t, err)

	// 2. Save network config
	networkConfig := symbiotic.NetworkConfig{
		VotingPowerProviders:    []symbiotic.CrossChainAddress{randomAddr(t)},
		KeysProvider:            randomAddr(t),
		Settlements:             []symbiotic.CrossChainAddress{randomAddr(t)},
		VerificationType:        symbiotic.VerificationTypeBlsBn254Simple,
		MaxVotingPower:          symbiotic.ToVotingPower(big.NewInt(10000)),
		MinInclusionVotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		MaxValidatorsCount:      symbiotic.ToVotingPower(big.NewInt(100)),
		RequiredKeyTags:         []symbiotic.KeyTag{15},
		RequiredHeaderKeyTag:    15,
		QuorumThresholds:        []symbiotic.QuorumThreshold{{KeyTag: 15, QuorumThreshold: symbiotic.ToQuorumThresholdPct(big.NewInt(67))}},
		NumCommitters:           3,
		NumAggregators:          5,
	}
	err = repo.SaveConfig(ctx, networkConfig, epoch)
	require.NoError(t, err)

	// 3. Save signature request and compute requestID
	sigRequest := symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(15),
		RequiredEpoch: epoch,
		Message:       randomBytes(t, 32),
	}

	// Compute requestID for signatures
	_, messageHash, err := priv.Sign(sigRequest.Message)
	require.NoError(t, err)
	signature := symbiotic.Signature{
		KeyTag:      sigRequest.KeyTag,
		Epoch:       epoch,
		MessageHash: messageHash,
		Signature:   randomBytes(t, 96),
		PublicKey:   priv.PublicKey(),
	}
	requestID := signature.RequestID()

	// 4. Save proof commit pending
	err = repo.SaveProofCommitPending(ctx, epoch, requestID)
	require.NoError(t, err)

	// 5. Save signature request
	err = repo.SaveSignatureRequest(ctx, requestID, sigRequest)
	require.NoError(t, err)

	// 6. Save signature
	err = repo.doUpdateInTx(ctx, "saveSignature", func(ctx context.Context) error {
		return repo.saveSignature(ctx, 0, signature)
	})
	require.NoError(t, err)

	// 7. Save signature map
	sigMap := entity.SignatureMap{
		RequestID:              requestID,
		Epoch:                  epoch,
		SignedValidatorsBitmap: entity.NewBitmapOf(0),
		CurrentVotingPower:     symbiotic.ToVotingPower(big.NewInt(1000)),
	}
	err = repo.UpdateSignatureMap(ctx, sigMap)
	require.NoError(t, err)

	// 8. Save aggregation proof
	aggProof := symbiotic.AggregationProof{
		MessageHash: messageHash,
		KeyTag:      symbiotic.KeyTag(15),
		Epoch:       epoch,
		Proof:       randomBytes(t, 96),
	}
	err = repo.saveAggregationProof(ctx, requestID, aggProof)
	require.NoError(t, err)

	// Verify entities exist before pruning
	t.Run("verify entities exist before pruning", func(t *testing.T) {
		// Check validator set
		_, err := repo.GetValidatorSetByEpoch(ctx, epoch)
		require.NoError(t, err)

		// Check network config
		_, err = repo.GetConfigByEpoch(ctx, epoch)
		require.NoError(t, err)

		// Check signature request
		_, err = repo.GetSignatureRequest(ctx, requestID)
		require.NoError(t, err)

		// Check signature
		_, err = repo.GetSignatureByIndex(ctx, requestID, 0)
		require.NoError(t, err)

		// Check signature map
		_, err = repo.GetSignatureMap(ctx, requestID)
		require.NoError(t, err)

		// Check aggregation proof
		_, err = repo.GetAggregationProof(ctx, requestID)
		require.NoError(t, err)
	})

	// Execute: Prune all entity types for the epoch
	t.Run("prune all entity types", func(t *testing.T) {
		// Prune in reverse dependency order
		err := repo.PruneSignatureEntitiesForEpoch(ctx, epoch)
		require.NoError(t, err)

		err = repo.PruneProofEntities(ctx, epoch)
		require.NoError(t, err)

		err = repo.PruneValsetEntities(ctx, epoch)
		require.NoError(t, err)
	})

	// Verify: All entities should be deleted
	t.Run("verify all entities deleted after pruning", func(t *testing.T) {
		// Check validator set deleted
		_, err := repo.GetValidatorSetByEpoch(ctx, epoch)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)

		// Check network config deleted
		_, err = repo.GetConfigByEpoch(ctx, epoch)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)

		// Check signature request deleted
		_, err = repo.GetSignatureRequest(ctx, requestID)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)

		// Check signature deleted
		_, err = repo.GetSignatureByIndex(ctx, requestID, 0)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)

		// Check signature map deleted
		_, err = repo.GetSignatureMap(ctx, requestID)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)

		// Check aggregation proof deleted
		_, err = repo.GetAggregationProof(ctx, requestID)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
	})
}

func TestRepository_PruneEntityTypes_Separately(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)
	ctx := t.Context()

	epoch := symbiotic.Epoch(100)

	// Setup: Create minimal entities for testing
	priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	// 1. Create validator set
	valset := symbiotic.ValidatorSet{
		Version:          1,
		RequiredKeyTag:   symbiotic.KeyTag(15),
		Epoch:            epoch,
		CaptureTimestamp: 1234567890,
		QuorumThreshold:  symbiotic.ToVotingPower(big.NewInt(500)),
		Validators: []symbiotic.Validator{{
			Operator:    common.BytesToAddress(randomBytes(t, 20)),
			VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
			IsActive:    true,
			Keys:        []symbiotic.ValidatorKey{{Tag: symbiotic.KeyTag(15), Payload: randomBytes(t, 96)}},
			Vaults:      []symbiotic.ValidatorVault{{ChainID: 1, Vault: common.BytesToAddress(randomBytes(t, 20)), VotingPower: symbiotic.ToVotingPower(big.NewInt(1000))}},
		}},
		Status: symbiotic.HeaderDerived,
	}
	err = repo.SaveValidatorSet(ctx, valset)
	require.NoError(t, err)

	// 2. Save network config
	networkConfig := symbiotic.NetworkConfig{
		VotingPowerProviders:    []symbiotic.CrossChainAddress{randomAddr(t)},
		KeysProvider:            randomAddr(t),
		Settlements:             []symbiotic.CrossChainAddress{randomAddr(t)},
		VerificationType:        symbiotic.VerificationTypeBlsBn254Simple,
		MaxVotingPower:          symbiotic.ToVotingPower(big.NewInt(10000)),
		MinInclusionVotingPower: symbiotic.ToVotingPower(big.NewInt(100)),
		MaxValidatorsCount:      symbiotic.ToVotingPower(big.NewInt(100)),
		RequiredKeyTags:         []symbiotic.KeyTag{15},
		RequiredHeaderKeyTag:    15,
		QuorumThresholds:        []symbiotic.QuorumThreshold{{KeyTag: 15, QuorumThreshold: symbiotic.ToQuorumThresholdPct(big.NewInt(67))}},
		NumCommitters:           3,
		NumAggregators:          5,
	}
	err = repo.SaveConfig(ctx, networkConfig, epoch)
	require.NoError(t, err)

	// 3. Create signature and proof entities
	sigRequest := symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(15),
		RequiredEpoch: epoch,
		Message:       randomBytes(t, 32),
	}
	_, messageHash, err := priv.Sign(sigRequest.Message)
	require.NoError(t, err)
	signature := symbiotic.Signature{
		KeyTag:      sigRequest.KeyTag,
		Epoch:       epoch,
		MessageHash: messageHash,
		Signature:   randomBytes(t, 96),
		PublicKey:   priv.PublicKey(),
	}
	requestID := signature.RequestID()

	err = repo.SaveProofCommitPending(ctx, epoch, requestID)
	require.NoError(t, err)
	err = repo.SaveSignatureRequest(ctx, requestID, sigRequest)
	require.NoError(t, err)
	err = repo.doUpdateInTx(ctx, "saveSignature", func(ctx context.Context) error {
		return repo.saveSignature(ctx, 0, signature)
	})
	require.NoError(t, err)
	sigMap := entity.SignatureMap{
		RequestID:              requestID,
		Epoch:                  epoch,
		SignedValidatorsBitmap: entity.NewBitmapOf(0),
		CurrentVotingPower:     symbiotic.ToVotingPower(big.NewInt(1000)),
	}
	err = repo.UpdateSignatureMap(ctx, sigMap)
	require.NoError(t, err)
	aggProof := symbiotic.AggregationProof{
		MessageHash: messageHash,
		KeyTag:      symbiotic.KeyTag(15),
		Epoch:       epoch,
		Proof:       randomBytes(t, 96),
	}
	err = repo.saveAggregationProof(ctx, requestID, aggProof)
	require.NoError(t, err)

	// Test: Prune only signatures, verify proofs and valsets remain
	t.Run("prune signatures only", func(t *testing.T) {
		err := repo.PruneSignatureEntitiesForEpoch(ctx, epoch)
		require.NoError(t, err)

		// Signatures should be deleted
		_, err = repo.GetSignatureByIndex(ctx, requestID, 0)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
		_, err = repo.GetSignatureMap(ctx, requestID)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
		_, err = repo.GetSignatureRequest(ctx, requestID)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)

		// Proofs should still exist
		_, err = repo.GetAggregationProof(ctx, requestID)
		require.NoError(t, err)

		// Validator sets should still exist
		_, err = repo.GetValidatorSetByEpoch(ctx, epoch)
		require.NoError(t, err)
		_, err = repo.GetConfigByEpoch(ctx, epoch)
		require.NoError(t, err)
	})

	// Test: Prune proofs, verify valsets remain
	t.Run("prune proofs after signatures", func(t *testing.T) {
		err := repo.PruneProofEntities(ctx, epoch)
		require.NoError(t, err)

		// Proofs should be deleted
		_, err = repo.GetAggregationProof(ctx, requestID)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)

		// Validator sets should still exist
		_, err = repo.GetValidatorSetByEpoch(ctx, epoch)
		require.NoError(t, err)
		_, err = repo.GetConfigByEpoch(ctx, epoch)
		require.NoError(t, err)
	})

	// Test: Prune valsets last
	t.Run("prune valsets after proofs", func(t *testing.T) {
		err := repo.PruneValsetEntities(ctx, epoch)
		require.NoError(t, err)

		// Validator sets should be deleted
		_, err = repo.GetValidatorSetByEpoch(ctx, epoch)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
		_, err = repo.GetConfigByEpoch(ctx, epoch)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
	})
}

func TestRepository_PruneAggregationProof_IndexCleanup(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)
	ctx := t.Context()

	priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	// Create aggregation proofs for three epochs
	epochs := []symbiotic.Epoch{100, 101, 102}
	requestIDs := make([]common.Hash, len(epochs))

	for i, epoch := range epochs {
		// Create signature to get requestID
		sigRequest := symbiotic.SignatureRequest{
			KeyTag:        symbiotic.KeyTag(15),
			RequiredEpoch: epoch,
			Message:       randomBytes(t, 32),
		}
		_, messageHash, err := priv.Sign(sigRequest.Message)
		require.NoError(t, err)
		signature := symbiotic.Signature{
			KeyTag:      sigRequest.KeyTag,
			Epoch:       epoch,
			MessageHash: messageHash,
			Signature:   randomBytes(t, 96),
			PublicKey:   priv.PublicKey(),
		}
		requestID := signature.RequestID()
		requestIDs[i] = requestID

		// Save aggregation proof
		aggProof := symbiotic.AggregationProof{
			MessageHash: messageHash,
			KeyTag:      symbiotic.KeyTag(15),
			Epoch:       epoch,
			Proof:       randomBytes(t, 96),
		}
		err = repo.saveAggregationProof(ctx, requestID, aggProof)
		require.NoError(t, err)
	}

	// Verify all proofs exist before pruning
	t.Run("verify all proofs exist before pruning", func(t *testing.T) {
		for i, epoch := range epochs {
			proofs, err := repo.GetAggregationProofsByEpoch(ctx, epoch)
			require.NoError(t, err)
			require.Len(t, proofs, 1)
			require.Equal(t, requestIDs[i], proofs[0].RequestID())
		}

		// Test GetAggregationProofsStartingFromEpoch
		proofs, err := repo.GetAggregationProofsStartingFromEpoch(ctx, epochs[0])
		require.NoError(t, err)
		require.Len(t, proofs, 3)
	})

	// Prune the middle epoch (101)
	t.Run("prune middle epoch", func(t *testing.T) {
		err := repo.PruneProofEntities(ctx, epochs[1])
		require.NoError(t, err)

		// Direct get should fail
		_, err = repo.GetAggregationProof(ctx, requestIDs[1])
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
	})

	// Verify GetAggregationProofsByEpoch returns empty for pruned epoch
	t.Run("GetAggregationProofsByEpoch returns empty for pruned epoch", func(t *testing.T) {
		proofs, err := repo.GetAggregationProofsByEpoch(ctx, epochs[1])
		require.NoError(t, err, "GetAggregationProofsByEpoch should not error on pruned epoch")
		require.Len(t, proofs, 0, "GetAggregationProofsByEpoch should return empty slice for pruned epoch")
	})

	// Verify GetAggregationProofsByEpoch still works for non-pruned epochs
	t.Run("GetAggregationProofsByEpoch works for non-pruned epochs", func(t *testing.T) {
		// First epoch should still have its proof
		proofs, err := repo.GetAggregationProofsByEpoch(ctx, epochs[0])
		require.NoError(t, err)
		require.Len(t, proofs, 1)
		require.Equal(t, requestIDs[0], proofs[0].RequestID())

		// Last epoch should still have its proof
		proofs, err = repo.GetAggregationProofsByEpoch(ctx, epochs[2])
		require.NoError(t, err)
		require.Len(t, proofs, 1)
		require.Equal(t, requestIDs[2], proofs[0].RequestID())
	})

	// Verify GetAggregationProofsStartingFromEpoch skips pruned epoch
	t.Run("GetAggregationProofsStartingFromEpoch skips pruned epoch", func(t *testing.T) {
		// Starting from epoch 100 should return proofs for epochs 100 and 102 only
		proofs, err := repo.GetAggregationProofsStartingFromEpoch(ctx, epochs[0])
		require.NoError(t, err, "GetAggregationProofsStartingFromEpoch should not error when iterating past pruned epochs")
		require.Len(t, proofs, 2, "GetAggregationProofsStartingFromEpoch should return 2 proofs (skipping pruned epoch 101)")

		// Verify the proofs are from epochs 100 and 102
		require.Equal(t, epochs[0], proofs[0].Epoch)
		require.Equal(t, requestIDs[0], proofs[0].RequestID())
		require.Equal(t, epochs[2], proofs[1].Epoch)
		require.Equal(t, requestIDs[2], proofs[1].RequestID())
	})

	// Verify GetAggregationProofsStartingFromEpoch works when starting from pruned epoch
	t.Run("GetAggregationProofsStartingFromEpoch works when starting from pruned epoch", func(t *testing.T) {
		// Starting from pruned epoch 101 should return only epoch 102
		proofs, err := repo.GetAggregationProofsStartingFromEpoch(ctx, epochs[1])
		require.NoError(t, err)
		require.Len(t, proofs, 1)
		require.Equal(t, epochs[2], proofs[0].Epoch)
		require.Equal(t, requestIDs[2], proofs[0].RequestID())
	})
}
