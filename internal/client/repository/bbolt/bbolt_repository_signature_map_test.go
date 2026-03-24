package bbolt

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"
)

func randomRequestID(t *testing.T) common.Hash {
	t.Helper()
	req := symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(15),
		RequiredEpoch: symbiotic.Epoch(randomBigInt(t).Uint64()),
		Message:       randomBytes(t, 32),
	}
	priv, err := crypto.GeneratePrivateKey(req.KeyTag.Type())
	require.NoError(t, err)
	_, messageHash, err := priv.Sign(req.Message)
	require.NoError(t, err)

	sig := symbiotic.Signature{
		KeyTag:      req.KeyTag,
		Epoch:       req.RequiredEpoch,
		MessageHash: messageHash,
		PublicKey:   priv.PublicKey(),
	}
	return sig.RequestID()
}

func randomSignatureMap(t *testing.T, requestID common.Hash) entity.SignatureMap {
	t.Helper()
	return entity.SignatureMap{
		RequestID:              requestID,
		Epoch:                  symbiotic.Epoch(randomBigInt(t).Uint64()),
		SignedValidatorsBitmap: entity.NewBitmapOf(0, 1, 2),
		CurrentVotingPower:     symbiotic.ToVotingPower(randomBigInt(t)),
	}
}

func assertSignatureMapsEqual(t *testing.T, expected, actual entity.SignatureMap) {
	t.Helper()
	assert.Equal(t, expected.RequestID, actual.RequestID)
	assert.Equal(t, expected.Epoch, actual.Epoch)
	assert.True(t, expected.SignedValidatorsBitmap.Equals(actual.SignedValidatorsBitmap.Bitmap))
	assert.Equal(t, expected.CurrentVotingPower.String(), actual.CurrentVotingPower.String())
}

func TestRepository_SignatureMap(t *testing.T) {
	t.Parallel()
	repo := setupTestRepository(t)

	requestID1 := randomRequestID(t)
	requestID2 := randomRequestID(t)
	vm1 := randomSignatureMap(t, requestID1)
	vm2 := randomSignatureMap(t, requestID2)

	t.Run("UpdateSignatureMap - Success", func(t *testing.T) {
		err := repo.UpdateSignatureMap(context.Background(), vm1)
		require.NoError(t, err)

		retrieved, err := repo.GetSignatureMap(context.Background(), requestID1)
		require.NoError(t, err)
		assertSignatureMapsEqual(t, vm1, retrieved)
	})

	t.Run("UpdateSignatureMap - Update Existing", func(t *testing.T) {
		updatedVM := vm1
		updatedVM.Epoch = vm1.Epoch + 1
		updatedVM.CurrentVotingPower = symbiotic.ToVotingPower(big.NewInt(999))

		err := repo.UpdateSignatureMap(context.Background(), updatedVM)
		require.NoError(t, err)

		retrieved, err := repo.GetSignatureMap(context.Background(), requestID1)
		require.NoError(t, err)
		assertSignatureMapsEqual(t, updatedVM, retrieved)
	})

	t.Run("GetSignatureMap - Success", func(t *testing.T) {
		err := repo.UpdateSignatureMap(context.Background(), vm1)
		require.NoError(t, err)
		err = repo.UpdateSignatureMap(context.Background(), vm2)
		require.NoError(t, err)

		retrieved1, err := repo.GetSignatureMap(context.Background(), requestID1)
		require.NoError(t, err)
		assertSignatureMapsEqual(t, vm1, retrieved1)

		retrieved2, err := repo.GetSignatureMap(context.Background(), requestID2)
		require.NoError(t, err)
		assertSignatureMapsEqual(t, vm2, retrieved2)
	})

	t.Run("GetSignatureMap - Not Found", func(t *testing.T) {
		nonExistentHash := randomRequestID(t)
		_, err := repo.GetSignatureMap(context.Background(), nonExistentHash)
		require.Error(t, err)
		assert.True(t, errors.Is(err, entity.ErrEntityNotFound))
	})
}
