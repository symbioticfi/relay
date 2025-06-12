package badger

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"middleware-offchain/core/entity"
)

func TestBadgerRepository_NetworkConfig(t *testing.T) {
	t.Parallel()
	repo, err := New(Config{Dir: t.TempDir()})
	require.NoError(t, err)

	config := randomNetworkConfig(t)

	err = repo.SaveConfig(t.Context(), config, 1)
	require.NoError(t, err)

	loadedConfig, err := repo.GetConfigByEpoch(t.Context(), 1)
	require.NoError(t, err)
	require.Equal(t, config, loadedConfig)

	err = repo.SaveConfig(t.Context(), config, 1)
	require.ErrorIs(t, err, entity.ErrEntityAlreadyExist)
}

func TestBadgerRepository_SignatureRequest(t *testing.T) {
	t.Parallel()
	repo, err := New(Config{Dir: t.TempDir()})
	require.NoError(t, err)

	req := randomSignatureRequest(t)

	err = repo.SaveSignatureRequest(t.Context(), req)
	require.NoError(t, err)

	loadedConfig, err := repo.GetSignatureRequest(t.Context(), req.Hash())
	require.NoError(t, err)
	require.Equal(t, req, loadedConfig)
}

func TestBadgerRepository_AggregationProof(t *testing.T) {
	t.Parallel()
	repo, err := New(Config{Dir: t.TempDir()})
	require.NoError(t, err)

	ap := randomAggregationProof(t)

	hash := common.BytesToHash(randomBytes(t, 32))

	err = repo.SaveAggregationProof(t.Context(), hash, ap)
	require.NoError(t, err)

	loadedConfig, err := repo.GetAggregationProof(t.Context(), hash)
	require.NoError(t, err)
	require.Equal(t, ap, loadedConfig)
}

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

func randomAggregationProof(t *testing.T) entity.AggregationProof {
	t.Helper()

	return entity.AggregationProof{
		VerificationType: entity.VerificationTypeSimple,
		MessageHash:      randomBytes(t, 32),
		Proof:            randomBytes(t, 32),
	}
}

func randomSignatureRequest(t *testing.T) entity.SignatureRequest {
	t.Helper()
	return entity.SignatureRequest{
		KeyTag:        entity.KeyTag(15),
		RequiredEpoch: randomBigInt(t).Uint64(),
		Message:       randomBytes(t, 32),
	}
}

func randomNetworkConfig(t *testing.T) entity.NetworkConfig {
	t.Helper()
	return entity.NetworkConfig{
		VotingPowerProviders:    []entity.CrossChainAddress{randomAddr(t)},
		KeysProvider:            randomAddr(t),
		Replicas:                []entity.CrossChainAddress{randomAddr(t)},
		VerificationType:        entity.VerificationTypeSimple,
		MaxVotingPower:          randomBigInt(t),
		MinInclusionVotingPower: randomBigInt(t),
		MaxValidatorsCount:      randomBigInt(t),
		RequiredKeyTags:         []entity.KeyTag{15},
	}
}

func randomBytes(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return b
}

func randomAddr(t *testing.T) entity.CrossChainAddress {
	t.Helper()

	chainID, err := rand.Int(rand.Reader, big.NewInt(10000))
	require.NoError(t, err)

	return entity.CrossChainAddress{
		Address: common.BytesToAddress(randomBytes(t, 20)),
		ChainId: chainID.Uint64(),
	}
}

func randomBigInt(t *testing.T) *big.Int {
	t.Helper()
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	require.NoError(t, err)
	return n
}
