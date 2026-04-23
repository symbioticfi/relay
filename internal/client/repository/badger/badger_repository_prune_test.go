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

	priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	validator := symbiotic.Validator{
		Operator:    common.BytesToAddress(randomBytes(t, 20)),
		VotingPower: symbiotic.ToVotingPower(big.NewInt(1000)),
		IsActive:    true,
		Keys: []symbiotic.ValidatorKey{
			{Tag: symbiotic.KeyTag(15), Payload: randomBytes(t, 96)},
		},
		Vaults: []symbiotic.ValidatorVault{
			{ChainID: 1, Vault: common.BytesToAddress(randomBytes(t, 20)), VotingPower: symbiotic.ToVotingPower(big.NewInt(1000))},
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

	err = repo.saveValidatorSet(ctx, valset)
	require.NoError(t, err)

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

	err = repo.saveProofCommitPending(ctx, epoch, requestID)
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

	t.Run("verify entities exist before pruning", func(t *testing.T) {
		_, err := repo.GetValidatorSetByEpoch(ctx, epoch)
		require.NoError(t, err)
		_, err = repo.GetConfigByEpoch(ctx, epoch)
		require.NoError(t, err)
		_, err = repo.GetSignatureRequest(ctx, requestID)
		require.NoError(t, err)
		_, err = repo.GetSignatureByIndex(ctx, requestID, 0)
		require.NoError(t, err)
		_, err = repo.GetSignatureMap(ctx, requestID)
		require.NoError(t, err)
		_, err = repo.GetAggregationProof(ctx, requestID)
		require.NoError(t, err)
	})

	t.Run("prune all entity types", func(t *testing.T) {
		requestIDs, err := repo.GetRequestIDsByEpoch(ctx, epoch, 0)
		require.NoError(t, err)

		require.NoError(t, repo.PruneProofCommits(ctx, epoch))
		require.NoError(t, repo.PruneSignaturesByRequestIDs(ctx, epoch, requestIDs, 0))
		require.NoError(t, repo.PruneProofsByRequestIDs(ctx, epoch, requestIDs, 0))
		require.NoError(t, repo.PruneEpochIndicesByRequestIDs(ctx, epoch, requestIDs, 0))
		require.NoError(t, repo.PruneValsetEntities(ctx, epoch, 0))
	})

	t.Run("verify all entities deleted after pruning", func(t *testing.T) {
		_, err := repo.GetValidatorSetByEpoch(ctx, epoch)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
		_, err = repo.GetConfigByEpoch(ctx, epoch)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
		_, err = repo.GetSignatureRequest(ctx, requestID)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
		_, err = repo.GetSignatureByIndex(ctx, requestID, 0)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
		_, err = repo.GetSignatureMap(ctx, requestID)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
		_, err = repo.GetAggregationProof(ctx, requestID)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
	})
}

func TestRepository_PruneEntityTypes_Separately(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)
	ctx := t.Context()

	epoch := symbiotic.Epoch(100)

	priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

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
	err = repo.saveValidatorSet(ctx, valset)
	require.NoError(t, err)

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

	err = repo.saveProofCommitPending(ctx, epoch, requestID)
	require.NoError(t, err)
	err = repo.SaveSignatureRequest(ctx, requestID, sigRequest)
	require.NoError(t, err)
	err = repo.doUpdateInTx(ctx, "saveSignature", func(ctx context.Context) error {
		return repo.saveSignature(ctx, 0, signature)
	})
	require.NoError(t, err)
	err = repo.UpdateSignatureMap(ctx, entity.SignatureMap{
		RequestID:              requestID,
		Epoch:                  epoch,
		SignedValidatorsBitmap: entity.NewBitmapOf(0),
		CurrentVotingPower:     symbiotic.ToVotingPower(big.NewInt(1000)),
	})
	require.NoError(t, err)
	err = repo.saveAggregationProof(ctx, requestID, symbiotic.AggregationProof{
		MessageHash: messageHash,
		KeyTag:      symbiotic.KeyTag(15),
		Epoch:       epoch,
		Proof:       randomBytes(t, 96),
	})
	require.NoError(t, err)

	requestIDs, err := repo.GetRequestIDsByEpoch(ctx, epoch, 0)
	require.NoError(t, err)

	t.Run("prune signatures only", func(t *testing.T) {
		require.NoError(t, repo.PruneSignaturesByRequestIDs(ctx, epoch, requestIDs, 0))

		_, err = repo.GetSignatureByIndex(ctx, requestID, 0)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
		_, err = repo.GetSignatureMap(ctx, requestID)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
		_, err = repo.GetSignatureRequest(ctx, requestID)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)

		_, err = repo.GetAggregationProof(ctx, requestID)
		require.NoError(t, err)
		_, err = repo.GetValidatorSetByEpoch(ctx, epoch)
		require.NoError(t, err)
		_, err = repo.GetConfigByEpoch(ctx, epoch)
		require.NoError(t, err)
	})

	t.Run("prune proofs after signatures", func(t *testing.T) {
		require.NoError(t, repo.PruneProofCommits(ctx, epoch))
		require.NoError(t, repo.PruneProofsByRequestIDs(ctx, epoch, requestIDs, 0))

		_, err = repo.GetAggregationProof(ctx, requestID)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
		_, err = repo.GetValidatorSetByEpoch(ctx, epoch)
		require.NoError(t, err)
		_, err = repo.GetConfigByEpoch(ctx, epoch)
		require.NoError(t, err)
	})

	t.Run("prune valsets after proofs", func(t *testing.T) {
		require.NoError(t, repo.PruneValsetEntities(ctx, epoch, 0))

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

	epochs := []symbiotic.Epoch{100, 101, 102}
	requestIDs := make([]common.Hash, len(epochs))

	for i, epoch := range epochs {
		message := randomBytes(t, 32)
		_, messageHash, err := priv.Sign(message)
		require.NoError(t, err)
		signature := symbiotic.Signature{
			KeyTag:      symbiotic.KeyTag(15),
			Epoch:       epoch,
			MessageHash: messageHash,
			Signature:   randomBytes(t, 96),
			PublicKey:   priv.PublicKey(),
		}
		requestID := signature.RequestID()
		requestIDs[i] = requestID

		aggProof := symbiotic.AggregationProof{
			MessageHash: messageHash,
			KeyTag:      symbiotic.KeyTag(15),
			Epoch:       epoch,
			Proof:       randomBytes(t, 96),
		}
		err = repo.saveAggregationProof(ctx, requestID, aggProof)
		require.NoError(t, err)
	}

	t.Run("verify all proofs exist before pruning", func(t *testing.T) {
		for i, epoch := range epochs {
			proofs, err := repo.GetAggregationProofsByEpoch(ctx, epoch)
			require.NoError(t, err)
			require.Len(t, proofs, 1)
			require.Equal(t, requestIDs[i], proofs[0].RequestID())
		}

		proofs, err := repo.GetAggregationProofsStartingFromEpoch(ctx, epochs[0])
		require.NoError(t, err)
		require.Len(t, proofs, 3)
	})

	t.Run("prune middle epoch", func(t *testing.T) {
		epochRequestIDs, err := repo.GetRequestIDsByEpoch(ctx, epochs[1], 0)
		require.NoError(t, err)

		require.NoError(t, repo.PruneProofsByRequestIDs(ctx, epochs[1], epochRequestIDs, 0))

		_, err = repo.GetAggregationProof(ctx, requestIDs[1])
		require.ErrorIs(t, err, entity.ErrEntityNotFound)

		require.NoError(t, repo.PruneEpochIndicesByRequestIDs(ctx, epochs[1], epochRequestIDs, 0))
	})

	t.Run("GetAggregationProofsByEpoch returns empty for pruned epoch", func(t *testing.T) {
		proofs, err := repo.GetAggregationProofsByEpoch(ctx, epochs[1])
		require.NoError(t, err)
		require.Empty(t, proofs)
	})

	t.Run("GetAggregationProofsByEpoch works for non-pruned epochs", func(t *testing.T) {
		proofs, err := repo.GetAggregationProofsByEpoch(ctx, epochs[0])
		require.NoError(t, err)
		require.Len(t, proofs, 1)
		require.Equal(t, requestIDs[0], proofs[0].RequestID())

		proofs, err = repo.GetAggregationProofsByEpoch(ctx, epochs[2])
		require.NoError(t, err)
		require.Len(t, proofs, 1)
		require.Equal(t, requestIDs[2], proofs[0].RequestID())
	})

	t.Run("GetAggregationProofsStartingFromEpoch skips pruned epoch", func(t *testing.T) {
		proofs, err := repo.GetAggregationProofsStartingFromEpoch(ctx, epochs[0])
		require.NoError(t, err)
		require.Len(t, proofs, 2)
		require.Equal(t, epochs[0], proofs[0].Epoch)
		require.Equal(t, requestIDs[0], proofs[0].RequestID())
		require.Equal(t, epochs[2], proofs[1].Epoch)
		require.Equal(t, requestIDs[2], proofs[1].RequestID())
	})

	t.Run("GetAggregationProofsStartingFromEpoch works when starting from pruned epoch", func(t *testing.T) {
		proofs, err := repo.GetAggregationProofsStartingFromEpoch(ctx, epochs[1])
		require.NoError(t, err)
		require.Len(t, proofs, 1)
		require.Equal(t, epochs[2], proofs[0].Epoch)
		require.Equal(t, requestIDs[2], proofs[0].RequestID())
	})
}

func TestRepository_PruneEpochIndices_DeletionOrder(t *testing.T) {
	t.Parallel()

	setupTestData := func(t *testing.T, repo *Repository) symbiotic.Epoch {
		t.Helper()
		ctx := t.Context()

		priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
		require.NoError(t, err)

		epoch := symbiotic.Epoch(100)
		message := randomBytes(t, 32)
		_, messageHash, err := priv.Sign(message)
		require.NoError(t, err)

		signature := symbiotic.Signature{
			KeyTag:      symbiotic.KeyTag(15),
			Epoch:       epoch,
			MessageHash: messageHash,
			Signature:   randomBytes(t, 96),
			PublicKey:   priv.PublicKey(),
		}
		requestID := signature.RequestID()

		require.NoError(t, repo.saveAggregationProof(ctx, requestID, symbiotic.AggregationProof{
			MessageHash: messageHash,
			KeyTag:      symbiotic.KeyTag(15),
			Epoch:       epoch,
			Proof:       randomBytes(t, 96),
		}))
		require.NoError(t, repo.SaveSignatureRequest(ctx, requestID, symbiotic.SignatureRequest{
			KeyTag:  symbiotic.KeyTag(15),
			Message: message,
		}))
		require.NoError(t, repo.doUpdateInTx(ctx, "saveSignature", func(ctx context.Context) error {
			return repo.saveSignature(ctx, 0, signature)
		}))

		return epoch
	}

	t.Run("proof deleted first then signatures", func(t *testing.T) {
		repo := setupTestRepository(t)
		ctx := t.Context()

		epoch := setupTestData(t, repo)

		requestIDs, err := repo.GetRequestIDsByEpoch(ctx, epoch, 0)
		require.NoError(t, err)
		require.Len(t, requestIDs, 1)

		// Delete proofs, index should remain (signatures still exist)
		require.NoError(t, repo.PruneProofsByRequestIDs(ctx, epoch, requestIDs, 0))
		require.NoError(t, repo.PruneEpochIndicesByRequestIDs(ctx, epoch, requestIDs, 0))

		remainingIDs, err := repo.GetRequestIDsByEpoch(ctx, epoch, 0)
		require.NoError(t, err)
		require.Len(t, remainingIDs, 1, "index should remain when signatures still exist")

		// Delete signatures, now index should be cleaned up
		require.NoError(t, repo.PruneSignaturesByRequestIDs(ctx, epoch, requestIDs, 0))
		require.NoError(t, repo.PruneEpochIndicesByRequestIDs(ctx, epoch, requestIDs, 0))

		finalIDs, err := repo.GetRequestIDsByEpoch(ctx, epoch, 0)
		require.NoError(t, err)
		require.Empty(t, finalIDs, "index should be deleted when both proof and signatures are gone")
	})

	t.Run("signatures deleted first then proof", func(t *testing.T) {
		repo := setupTestRepository(t)
		ctx := t.Context()

		epoch := setupTestData(t, repo)

		requestIDs, err := repo.GetRequestIDsByEpoch(ctx, epoch, 0)
		require.NoError(t, err)
		require.Len(t, requestIDs, 1)

		// Delete signatures, index should remain (proof still exists)
		require.NoError(t, repo.PruneSignaturesByRequestIDs(ctx, epoch, requestIDs, 0))
		require.NoError(t, repo.PruneEpochIndicesByRequestIDs(ctx, epoch, requestIDs, 0))

		remainingIDs, err := repo.GetRequestIDsByEpoch(ctx, epoch, 0)
		require.NoError(t, err)
		require.Len(t, remainingIDs, 1, "index should remain when proof still exists")

		// Delete proofs, now index should be cleaned up
		require.NoError(t, repo.PruneProofsByRequestIDs(ctx, epoch, requestIDs, 0))
		require.NoError(t, repo.PruneEpochIndicesByRequestIDs(ctx, epoch, requestIDs, 0))

		finalIDs, err := repo.GetRequestIDsByEpoch(ctx, epoch, 0)
		require.NoError(t, err)
		require.Empty(t, finalIDs, "index should be deleted when both signatures and proof are gone")
	})
}
