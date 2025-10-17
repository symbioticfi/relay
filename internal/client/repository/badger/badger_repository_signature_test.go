package badger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/symbioticfi/relay/symbiotic/usecase/crypto"

	"github.com/symbioticfi/relay/internal/entity"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

func TestBadgerRepository_Signature(t *testing.T) {
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
		err := repo.doUpdateInTx(context.Background(), "test", func(ctx context.Context) error {
			return repo.saveSignature(ctx, 5, sig1)
		})
		require.NoError(t, err)

		retrievedSig, err := repo.GetSignatureByIndex(context.Background(), 100, sig1.RequestID(), 5)
		require.NoError(t, err)
		require.Equal(t, sig1, retrievedSig)
	})

	t.Run("GetSignatureByIndex - not found", func(t *testing.T) {
		_, err := repo.GetSignatureByIndex(context.Background(), 100, sig1.RequestID(), 999)
		require.Error(t, err)
		require.ErrorIs(t, err, entity.ErrEntityNotFound)
	})

	t.Run("GetAllSignatures - multiple signatures", func(t *testing.T) {
		err := repo.doUpdateInTx(context.Background(), "", func(ctx context.Context) error {
			if err := repo.saveSignature(ctx, 10, sig1); err != nil {
				return err
			}
			return repo.saveSignature(ctx, 20, sig2)
		})
		require.NoError(t, err)

		signatures, err := repo.GetAllSignatures(context.Background(), 100, sig1.RequestID())
		require.NoError(t, err)
		require.Len(t, signatures, 3) // sig1 from first test + sig1 and sig2 from this test

		// Verify we can retrieve each signature by index
		retrievedSig1, err := repo.GetSignatureByIndex(context.Background(), 100, sig1.RequestID(), 10)
		require.NoError(t, err)
		require.Equal(t, sig1, retrievedSig1)

		retrievedSig2, err := repo.GetSignatureByIndex(context.Background(), 100, sig1.RequestID(), 20)
		require.NoError(t, err)
		require.Equal(t, sig2, retrievedSig2)
	})
}

func TestBadgerRepository_SignatureOrdering(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)

	// Test numeric ordering with indices that would be wrong in lexicographic string order
	testIndices := []uint32{9, 11, 100, 1000, 2}
	expectedOrder := []uint32{2, 9, 11, 100, 1000} // Expected numeric order

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
	// Save signatures with test indices
	err = repo.doUpdateInTx(context.Background(), "test", func(ctx context.Context) error {
		for _, index := range testIndices {
			priv, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
			require.NoError(t, err)
			pubs = append(pubs, priv.PublicKey())
			sigCopy := sig
			sigCopy.PublicKey = priv.PublicKey() // Different public key for each
			if err := repo.saveSignature(ctx, index, sigCopy); err != nil {
				return err
			}
		}
		return nil
	})
	require.NoError(t, err)

	// Retrieve all signatures and verify they are returned in numeric order
	signatures, err := repo.GetAllSignatures(context.Background(), 777, sig.RequestID())
	require.NoError(t, err)
	require.Len(t, signatures, len(testIndices))

	// Verify each signature can be retrieved by its expected index
	for i, expectedIndex := range expectedOrder {
		retrievedSig, err := repo.GetSignatureByIndex(context.Background(), 777, sig.RequestID(), expectedIndex)
		require.NoError(t, err, "failed to retrieve signature for index %d", expectedIndex)

		// Verify this is the signature we expect (by checking the public key byte)
		originalIndex := -1
		for j, origIndex := range testIndices {
			if origIndex == expectedIndex {
				originalIndex = j
				break
			}
		}
		require.NotEqual(t, -1, originalIndex, "could not find original index")
		require.Equal(t, pubs[originalIndex].Raw(), retrievedSig.PublicKey.Raw())

		// The signature returned by GetAllSignatures should match
		require.Equal(t, retrievedSig, signatures[i], "signature order mismatch at position %d", i)
	}
}

func TestBadgerRepository_GetSignaturesByEpoch(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)

	priv1, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)
	priv2, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)
	priv3, err := crypto.GeneratePrivateKey(symbiotic.KeyTypeBlsBn254)
	require.NoError(t, err)

	// Create three signatures with epochs 1, 2, 3
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

	// Save all three signatures
	err = repo.doUpdateInTx(context.Background(), "test", func(ctx context.Context) error {
		if err := repo.saveSignature(ctx, 1, sig1); err != nil {
			return err
		}
		return nil
	})
	require.NoError(t, err)

	err = repo.doUpdateInTx(context.Background(), "test", func(ctx context.Context) error {
		if err := repo.saveSignature(ctx, 1, sig2); err != nil {
			return err
		}
		return nil
	})
	require.NoError(t, err)

	err = repo.doUpdateInTx(context.Background(), "test", func(ctx context.Context) error {
		if err := repo.saveSignature(ctx, 1, sig3); err != nil {
			return err
		}
		return nil
	})
	require.NoError(t, err)

	t.Run("get signatures starting from epoch 2", func(t *testing.T) {
		// Query starting from epoch 2
		signatures, err := repo.GetSignaturesByEpoch(context.Background(), 2)
		require.NoError(t, err)

		// Should return exactly 2 signatures (epochs 2 and 3)
		require.Len(t, signatures, 2, "Should return 2 signatures (epochs 2 and 3)")

		// Verify epochs are 2 and 3
		require.Equal(t, symbiotic.Epoch(2), signatures[0].Epoch)
		require.Equal(t, sig2, signatures[0])

		require.Equal(t, symbiotic.Epoch(3), signatures[1].Epoch)
		require.Equal(t, sig3, signatures[1])
	})

	t.Run("get signatures starting from epoch 1", func(t *testing.T) {
		// Query starting from epoch 1 - should return all 3
		signatures, err := repo.GetSignaturesByEpoch(context.Background(), 1)
		require.NoError(t, err)

		require.Len(t, signatures, 3, "Should return all 3 signatures")
		require.Equal(t, symbiotic.Epoch(1), signatures[0].Epoch)
		require.Equal(t, symbiotic.Epoch(2), signatures[1].Epoch)
		require.Equal(t, symbiotic.Epoch(3), signatures[2].Epoch)
	})

	t.Run("get signatures starting from epoch 3", func(t *testing.T) {
		// Query starting from epoch 3 - should return only epoch 3
		signatures, err := repo.GetSignaturesByEpoch(context.Background(), 3)
		require.NoError(t, err)

		require.Len(t, signatures, 1, "Should return only 1 signature (epoch 3)")
		require.Equal(t, symbiotic.Epoch(3), signatures[0].Epoch)
		require.Equal(t, sig3, signatures[0])
	})

	t.Run("get signatures starting from non-existent epoch", func(t *testing.T) {
		// Query starting from epoch 10 (doesn't exist) - should return empty
		signatures, err := repo.GetSignaturesByEpoch(context.Background(), 10)
		require.NoError(t, err)
		require.Empty(t, signatures, "Should return empty slice for non-existent epoch")
	})
}
