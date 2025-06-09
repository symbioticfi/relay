package signer

import (
	"testing"

	"github.com/stretchr/testify/require"

	"middleware-offchain/internal/entity"
	keyprovider "middleware-offchain/internal/usecase/key-provider"
	"middleware-offchain/pkg/bls"
)

func TestBLSBn254(t *testing.T) {
	kp, err := keyprovider.NewSimpleKeystoreProvider()
	require.NoError(t, err)
	key := []byte("testrandomkey")
	require.NoError(t, kp.AddKey(entity.ValsetHeaderKeyTag, key))

	signer := NewSigner(kp)

	msg := []byte("message")

	hash, err := signer.Hash(entity.ValsetHeaderKeyTag, msg)
	require.NoError(t, err)

	signature, err := signer.Sign(entity.ValsetHeaderKeyTag, msg)
	require.NoError(t, err)
	require.Equal(t, hash, signature.MessageHash)

	public := bls.ComputeKeyPair(key).PackPublicG1G2()
	require.NoError(t, err)
	require.Len(t, public, 96)

	_, g2, err := bls.UnpackPublicG1G2(public)
	require.NoError(t, err)

	g1Sig, err := bls.DeserializeG1(signature.Signature)
	require.NoError(t, err)

	result, err := g2.Verify(g1Sig, hash)
	require.NoError(t, err)
	require.True(t, result)
}
