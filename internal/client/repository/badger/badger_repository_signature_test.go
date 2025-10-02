package badger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/symbioticfi/relay/core/entity"
)

func TestBadgerRepository_Signature(t *testing.T) {
	t.Parallel()

	repo := setupTestRepository(t)

	sig1 := entity.SignatureExtended{
		MessageHash: []byte("message1"),
		KeyTag:      15,
		Epoch:       100,
		Signature:   []byte("signature1"),
		PublicKey:   []byte("publickey1"),
	}

	sig2 := entity.SignatureExtended{
		MessageHash: []byte("message1"),
		KeyTag:      15,
		Epoch:       100,
		Signature:   []byte("signature2"),
		PublicKey:   []byte("publickey2"),
	}

	t.Run("SaveSignature and GetSignatureByIndex", func(t *testing.T) {
		err := repo.doUpdateInTx(context.Background(), "test", func(ctx context.Context) error {
			return repo.SaveSignature(ctx, 5, sig1)
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
			if err := repo.SaveSignature(ctx, 10, sig1); err != nil {
				return err
			}
			return repo.SaveSignature(ctx, 20, sig2)
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

	sig := entity.SignatureExtended{
		MessageHash: []byte("message"),
		Signature:   []byte("signature"),
		KeyTag:      entity.KeyTag(15),
		Epoch:       777,
		PublicKey:   randomBytes(t, 10),
	}

	// Save signatures with test indices
	err := repo.doUpdateInTx(context.Background(), "test", func(ctx context.Context) error {
		for i, index := range testIndices {
			sigCopy := sig
			sigCopy.PublicKey = []byte{byte(i)} // Different public key for each
			if err := repo.SaveSignature(ctx, index, sigCopy); err != nil {
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
		require.Equal(t, byte(originalIndex), retrievedSig.PublicKey[0])

		// The signature returned by GetAllSignatures should match
		require.Equal(t, retrievedSig, signatures[i], "signature order mismatch at position %d", i)
	}
}
