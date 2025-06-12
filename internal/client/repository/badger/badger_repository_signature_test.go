package badger

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"middleware-offchain/core/entity"
)

func TestBadgerRepository_Signature(t *testing.T) {
	t.Parallel()
	repo, err := New(Config{Dir: t.TempDir()})
	require.NoError(t, err)

	// Create two signatures for the same request hash
	reqHash1 := common.BytesToHash(randomBytes(t, 32))
	sig1 := randomSignature(t)
	sig2 := randomSignature(t)

	// Create a signature for a different request hash
	reqHash2 := common.BytesToHash(randomBytes(t, 32))
	sig3 := randomSignature(t)

	// Save all signatures
	err = repo.SaveSignature(t.Context(), reqHash1, sig1.PublicKey, sig1)
	require.NoError(t, err)
	err = repo.SaveSignature(t.Context(), reqHash1, sig2.PublicKey, sig2)
	require.NoError(t, err)
	err = repo.SaveSignature(t.Context(), reqHash2, sig3.PublicKey, sig3)
	require.NoError(t, err)

	// Get signatures for reqHash1
	signatures, err := repo.GetAllSignatures(t.Context(), reqHash1)
	require.NoError(t, err)
	require.Len(t, signatures, 2)

	// Verify that we got the correct signatures
	found := make(map[string]bool)
	for _, sig := range signatures {
		if bytes.Equal(sig.MessageHash, sig1.MessageHash) &&
			bytes.Equal(sig.Signature, sig1.Signature) &&
			bytes.Equal(sig.PublicKey, sig1.PublicKey) {
			found["sig1"] = true
		}
		if bytes.Equal(sig.MessageHash, sig2.MessageHash) &&
			bytes.Equal(sig.Signature, sig2.Signature) &&
			bytes.Equal(sig.PublicKey, sig2.PublicKey) {
			found["sig2"] = true
		}
	}
	require.True(t, found["sig1"], "sig1 not found in results")
	require.True(t, found["sig2"], "sig2 not found in results")

	// Get signatures for reqHash2
	signatures, err = repo.GetAllSignatures(t.Context(), reqHash2)
	require.NoError(t, err)
	require.Len(t, signatures, 1)
	require.True(t, bytes.Equal(signatures[0].MessageHash, sig3.MessageHash))
	require.True(t, bytes.Equal(signatures[0].Signature, sig3.Signature))
	require.True(t, bytes.Equal(signatures[0].PublicKey, sig3.PublicKey))
}

func randomSignature(t *testing.T) entity.Signature {
	t.Helper()
	return entity.Signature{
		MessageHash: randomBytes(t, 32),
		Signature:   randomBytes(t, 65), // Typical ECDSA signature length
		PublicKey:   randomBytes(t, 33), // Compressed public key length
	}
}
