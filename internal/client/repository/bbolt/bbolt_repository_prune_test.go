package bbolt

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
		Keys:        []symbiotic.ValidatorKey{{Tag: symbiotic.KeyTag(15), Payload: randomBytes(t, 96)}},
		Vaults:      []symbiotic.ValidatorVault{{ChainID: 1, Vault: common.BytesToAddress(randomBytes(t, 20)), VotingPower: symbiotic.ToVotingPower(big.NewInt(1000))}},
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

	networkConfig := randomNetworkConfig(t)
	err = repo.SaveConfig(ctx, networkConfig, epoch)
	require.NoError(t, err)

	_, messageHash, err := priv.Sign(randomBytes(t, 32))
	require.NoError(t, err)
	signature := symbiotic.Signature{
		KeyTag:      symbiotic.KeyTag(15),
		Epoch:       epoch,
		MessageHash: messageHash,
		Signature:   randomBytes(t, 96),
		PublicKey:   priv.PublicKey(),
	}
	requestID := signature.RequestID()

	err = repo.saveProofCommitPending(ctx, epoch, requestID)
	require.NoError(t, err)

	sigRequest := symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(15),
		RequiredEpoch: epoch,
		Message:       randomBytes(t, 32),
	}
	err = repo.SaveSignatureRequest(ctx, requestID, sigRequest)
	require.NoError(t, err)

	err = repo.saveSignature(ctx, 0, signature)
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
		err := repo.PruneSignatureEntitiesForEpoch(ctx, epoch)
		require.NoError(t, err)

		err = repo.PruneProofEntities(ctx, epoch)
		require.NoError(t, err)

		err = repo.PruneValsetEntities(ctx, epoch)
		require.NoError(t, err)
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
	require.NoError(t, repo.saveValidatorSet(ctx, valset))

	networkConfig := randomNetworkConfig(t)
	require.NoError(t, repo.SaveConfig(ctx, networkConfig, epoch))

	_, messageHash, err := priv.Sign(randomBytes(t, 32))
	require.NoError(t, err)
	signature := symbiotic.Signature{
		KeyTag:      symbiotic.KeyTag(15),
		Epoch:       epoch,
		MessageHash: messageHash,
		Signature:   randomBytes(t, 96),
		PublicKey:   priv.PublicKey(),
	}
	requestID := signature.RequestID()

	require.NoError(t, repo.saveProofCommitPending(ctx, epoch, requestID))
	require.NoError(t, repo.SaveSignatureRequest(ctx, requestID, symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(15),
		RequiredEpoch: epoch,
		Message:       randomBytes(t, 32),
	}))
	require.NoError(t, repo.saveSignature(ctx, 0, signature))
	require.NoError(t, repo.UpdateSignatureMap(ctx, entity.SignatureMap{
		RequestID:              requestID,
		Epoch:                  epoch,
		SignedValidatorsBitmap: entity.NewBitmapOf(0),
		CurrentVotingPower:     symbiotic.ToVotingPower(big.NewInt(1000)),
	}))
	require.NoError(t, repo.saveAggregationProof(ctx, requestID, symbiotic.AggregationProof{
		MessageHash: messageHash,
		KeyTag:      symbiotic.KeyTag(15),
		Epoch:       epoch,
		Proof:       randomBytes(t, 96),
	}))

	t.Run("prune signatures only", func(t *testing.T) {
		require.NoError(t, repo.PruneSignatureEntitiesForEpoch(ctx, epoch))

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
	})

	t.Run("prune proofs after signatures", func(t *testing.T) {
		require.NoError(t, repo.PruneProofEntities(ctx, epoch))

		_, err = repo.GetAggregationProof(ctx, requestID)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
		_, err = repo.GetValidatorSetByEpoch(ctx, epoch)
		require.NoError(t, err)
	})

	t.Run("prune valsets after proofs", func(t *testing.T) {
		require.NoError(t, repo.PruneValsetEntities(ctx, epoch))

		_, err = repo.GetValidatorSetByEpoch(ctx, epoch)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
		_, err = repo.GetConfigByEpoch(ctx, epoch)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
	})
}

func TestRepository_PruneRequestIDEpochIndices(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)
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
	require.NoError(t, repo.saveSignature(context.Background(), 0, signature))

	requestIDs := repo.getRequestIDsByEpoch(ctx, epoch)
	require.Len(t, requestIDs, 1)

	t.Run("index remains when only proof is deleted", func(t *testing.T) {
		require.NoError(t, repo.PruneProofEntities(ctx, epoch))
		require.NoError(t, repo.PruneRequestIDEpochIndices(ctx, epoch))

		remainingIDs := repo.getRequestIDsByEpoch(ctx, epoch)
		require.Len(t, remainingIDs, 1)
	})

	t.Run("index is deleted when both proof and signatures are gone", func(t *testing.T) {
		require.NoError(t, repo.PruneSignatureEntitiesForEpoch(ctx, epoch))
		require.NoError(t, repo.PruneRequestIDEpochIndices(ctx, epoch))

		finalIDs := repo.getRequestIDsByEpoch(ctx, epoch)
		require.Empty(t, finalIDs)
	})
}
