package bbolt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestRepository_Signature(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)

	priv1, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	priv2, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	sig1 := symbiotic.Signature{
		MessageHash: []byte("message1"),
		KeyTag:      15,
		Epoch:       100,
		Signature:   []byte("signature1"),
		PublicKey:   priv1.PublicKey(),
	}

	sig2 := symbiotic.Signature{
		MessageHash: []byte("message1"),
		KeyTag:      15,
		Epoch:       100,
		Signature:   []byte("signature2"),
		PublicKey:   priv2.PublicKey(),
	}

	t.Run("saveSignature and GetSignatureByIndex", func(t *testing.T) {
		err := repo.saveSignature(context.Background(), 5, sig1)
		require.NoError(t, err)

		retrievedSig, err := repo.GetSignatureByIndex(context.Background(), sig1.RequestID(), 5)
		require.NoError(t, err)
		require.Equal(t, sig1, retrievedSig)
	})

	t.Run("GetSignatureByIndex - not found", func(t *testing.T) {
		_, err := repo.GetSignatureByIndex(context.Background(), sig1.RequestID(), 999)
		require.Error(t, err)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
	})

	t.Run("GetAllSignatures - multiple signatures", func(t *testing.T) {
		err := repo.saveSignature(context.Background(), 10, sig1)
		require.NoError(t, err)
		err = repo.saveSignature(context.Background(), 20, sig2)
		require.NoError(t, err)

		signatures, err := repo.GetAllSignatures(context.Background(), sig1.RequestID())
		require.NoError(t, err)
		require.Len(t, signatures, 3) // sig1 at 5, sig1 at 10, sig2 at 20

		retrievedSig1, err := repo.GetSignatureByIndex(context.Background(), sig1.RequestID(), 10)
		require.NoError(t, err)
		require.Equal(t, sig1, retrievedSig1)

		retrievedSig2, err := repo.GetSignatureByIndex(context.Background(), sig1.RequestID(), 20)
		require.NoError(t, err)
		require.Equal(t, sig2, retrievedSig2)
	})
}

func TestRepository_SignatureOrdering(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)

	testIndices := []uint32{9, 11, 100, 1000, 2}
	expectedOrder := []uint32{2, 9, 11, 100, 1000}

	priv1, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	sig := symbiotic.Signature{
		MessageHash: []byte("message"),
		Signature:   []byte("signature"),
		KeyTag:      symbiotic.KeyTag(15),
		Epoch:       777,
		PublicKey:   priv1.PublicKey(),
	}

	pubs := []crypto.PublicKey{}
	for _, index := range testIndices {
		priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
		require.NoError(t, err)
		pubs = append(pubs, priv.PublicKey())
		sigCopy := sig
		sigCopy.PublicKey = priv.PublicKey()
		err = repo.saveSignature(context.Background(), index, sigCopy)
		require.NoError(t, err)
	}

	signatures, err := repo.GetAllSignatures(context.Background(), sig.RequestID())
	require.NoError(t, err)
	require.Len(t, signatures, len(testIndices))

	for i, expectedIndex := range expectedOrder {
		retrievedSig, err := repo.GetSignatureByIndex(context.Background(), sig.RequestID(), expectedIndex)
		require.NoError(t, err)

		originalIndex := -1
		for j, origIndex := range testIndices {
			if origIndex == expectedIndex {
				originalIndex = j
				break
			}
		}
		require.NotEqual(t, -1, originalIndex)
		require.Equal(t, pubs[originalIndex].Raw(), retrievedSig.PublicKey.Raw())
		require.Equal(t, retrievedSig, signatures[i])
	}
}

func TestRepository_GetSignaturesStartingFromEpoch(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)

	priv1, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)
	priv2, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)
	priv3, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	sig1 := symbiotic.Signature{
		MessageHash: []byte("message1"),
		KeyTag:      15,
		Epoch:       1,
		Signature:   []byte("signature1"),
		PublicKey:   priv1.PublicKey(),
	}

	sig2 := symbiotic.Signature{
		MessageHash: []byte("message2"),
		KeyTag:      15,
		Epoch:       2,
		Signature:   []byte("signature2"),
		PublicKey:   priv2.PublicKey(),
	}

	sig3 := symbiotic.Signature{
		MessageHash: []byte("message3"),
		KeyTag:      15,
		Epoch:       3,
		Signature:   []byte("signature3"),
		PublicKey:   priv3.PublicKey(),
	}

	require.NoError(t, repo.saveSignature(context.Background(), 1, sig1))
	require.NoError(t, repo.saveSignature(context.Background(), 1, sig2))
	require.NoError(t, repo.saveSignature(context.Background(), 1, sig3))

	t.Run("get signatures starting from epoch 2", func(t *testing.T) {
		signatures, err := repo.GetSignaturesStartingFromEpoch(context.Background(), 2)
		require.NoError(t, err)
		require.Len(t, signatures, 2)
		require.Equal(t, symbiotic.Epoch(2), signatures[0].Epoch)
		require.Equal(t, sig2, signatures[0])
		require.Equal(t, symbiotic.Epoch(3), signatures[1].Epoch)
		require.Equal(t, sig3, signatures[1])
	})

	t.Run("get signatures starting from epoch 1", func(t *testing.T) {
		signatures, err := repo.GetSignaturesStartingFromEpoch(context.Background(), 1)
		require.NoError(t, err)
		require.Len(t, signatures, 3)
	})

	t.Run("get signatures starting from non-existent epoch", func(t *testing.T) {
		signatures, err := repo.GetSignaturesStartingFromEpoch(context.Background(), 10)
		require.NoError(t, err)
		require.Empty(t, signatures)
	})
}

func TestRepository_GetSignaturesByEpoch(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)

	priv1, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)
	priv2, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	sig1 := symbiotic.Signature{
		MessageHash: []byte("message1"),
		KeyTag:      15,
		Epoch:       1,
		Signature:   []byte("signature1"),
		PublicKey:   priv1.PublicKey(),
	}

	sig2 := symbiotic.Signature{
		MessageHash: []byte("message2"),
		KeyTag:      15,
		Epoch:       2,
		Signature:   []byte("signature2"),
		PublicKey:   priv2.PublicKey(),
	}

	require.NoError(t, repo.saveSignature(context.Background(), 1, sig1))
	require.NoError(t, repo.saveSignature(context.Background(), 1, sig2))

	t.Run("get signatures for epoch 1", func(t *testing.T) {
		signatures, err := repo.GetSignaturesByEpoch(context.Background(), 1)
		require.NoError(t, err)
		require.Len(t, signatures, 1)
		require.Equal(t, sig1, signatures[0])
	})

	t.Run("get signatures for epoch 2", func(t *testing.T) {
		signatures, err := repo.GetSignaturesByEpoch(context.Background(), 2)
		require.NoError(t, err)
		require.Len(t, signatures, 1)
		require.Equal(t, sig2, signatures[0])
	})

	t.Run("get signatures for non-existent epoch", func(t *testing.T) {
		signatures, err := repo.GetSignaturesByEpoch(context.Background(), 10)
		require.NoError(t, err)
		require.Empty(t, signatures)
	})
}
