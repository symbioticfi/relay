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
		err := repo.doUpdateInTx(context.Background(), "", func(ctx context.Context) error {
			if err := repo.saveSignature(ctx, 10, sig1); err != nil {
				return err
			}
			return repo.saveSignature(ctx, 20, sig2)
		})
		require.NoError(t, err)

		signatures, err := repo.GetAllSignatures(context.Background(), sig1.RequestID())
		require.NoError(t, err)
		require.Len(t, signatures, 3) // sig1 from first test + sig1 and sig2 from this test

		// Verify we can retrieve each signature by index
		retrievedSig1, err := repo.GetSignatureByIndex(context.Background(), sig1.RequestID(), 10)
		require.NoError(t, err)
		require.Equal(t, sig1, retrievedSig1)

		retrievedSig2, err := repo.GetSignatureByIndex(context.Background(), sig1.RequestID(), 20)
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
	signatures, err := repo.GetAllSignatures(context.Background(), sig.RequestID())
	require.NoError(t, err)
	require.Len(t, signatures, len(testIndices))

	// Verify each signature can be retrieved by its expected index
	for i, expectedIndex := range expectedOrder {
		retrievedSig, err := repo.GetSignatureByIndex(context.Background(), sig.RequestID(), expectedIndex)
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
