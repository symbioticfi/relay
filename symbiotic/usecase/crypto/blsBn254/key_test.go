package blsBn254

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBLSKeys(t *testing.T) {
	private, err := GenerateKey()
	require.NoError(t, err)
	public := private.PublicKey()

	someRandData := randData(t)

	sign, hash, err := private.Sign(someRandData)
	require.NoError(t, err)

	t.Run("Verify", func(t *testing.T) {
		err = public.VerifyWithHash(hash, sign)
		require.NoError(t, err)

		err = public.Verify(someRandData, sign)
		require.NoError(t, err)
	})

	t.Run("VerifyFromRaw", func(t *testing.T) {
		rawPublic := public.Raw()
		publicFromRaw, err := FromRaw(rawPublic)
		require.NoError(t, err)

		err = publicFromRaw.VerifyWithHash(hash, sign)
		require.NoError(t, err)

		err = publicFromRaw.Verify(someRandData, sign)
		require.NoError(t, err)
	})

	t.Run("VerifyFromPrivate", func(t *testing.T) {
		publicFromPrivate := FromPrivateKey(private)

		err = publicFromPrivate.VerifyWithHash(hash, sign)
		require.NoError(t, err)

		err = publicFromPrivate.Verify(someRandData, sign)
		require.NoError(t, err)
	})

	t.Run("VerifyFromNewPublicKey", func(t *testing.T) {
		key, ok := public.(*PublicKey)
		require.True(t, ok)
		newPublic := NewPublicKey(key.g1PubKey, key.g2PubKey)

		err = newPublic.VerifyWithHash(hash, sign)
		require.NoError(t, err)

		err = newPublic.Verify(someRandData, sign)
		require.NoError(t, err)
	})
}

func TestCheckPrivateKeyBytes(t *testing.T) {
	private, err := GenerateKey()
	require.NoError(t, err)
	someRandData := randData(t)

	sign, hash, err := private.Sign(someRandData)
	require.NoError(t, err)

	bytes := private.Bytes()
	newPrivateKey, err := NewPrivateKey(bytes)
	require.NoError(t, err)

	newSign, newHash, err := newPrivateKey.Sign(someRandData)
	require.NoError(t, err)

	require.Equal(t, sign, newSign)
	require.Equal(t, hash, newHash)
}

func TestMarshallText(t *testing.T) {
	key, err := NewPrivateKey(big.NewInt(123).FillBytes(make([]byte, 32)))
	require.NoError(t, err)

	text, err := key.PublicKey().MarshalText()
	require.NoError(t, err)

	require.Equal(t, "G1/0xdaa125a22bd902874034e67868aed40267e5575d5919677987e3bc6dd42a32fe;G2/0x9f1954b33144db2b5c90da089e8bde287ec7089d5d6433f3b6becaefdb678b1b2a9de38d14bef2cf9afc3c698a4211fa7ada7b4f036a2dfef0dc122b423259d0", string(text))
}

func TestInvalidVerification(t *testing.T) {
	private, err := GenerateKey()
	require.NoError(t, err)
	public := private.PublicKey()

	err = public.VerifyWithHash([]byte{1}, nil)
	require.Error(t, err)
	require.EqualError(t, err, "blsBn254: invalid message hash length")
	err = public.VerifyWithHash(make([]byte, 32), nil)
	require.Error(t, err)
	require.EqualError(t, err, "blsBn254: failed to set big into G1: short buffer")
	err = public.VerifyWithHash(make([]byte, 32), make([]byte, 65))
	require.Error(t, err)
	require.EqualError(t, err, "blsBn254: invalid signature")
}

func TestFromRaw(t *testing.T) {
	_, err := FromRaw(nil)
	require.EqualError(t, err, "blsBn254: nil raw key")

	_, err = FromRaw([]byte("invalid public key"))
	require.EqualError(t, err, "blsBn254: invalid raw key length, expected 96, got 18")

	_, err = FromRaw(make([]byte, 96))
	require.EqualError(t, err, "blsBn254: failed to unmarshal G1 pubkey: short buffer")
}

func randData(t *testing.T) []byte {
	t.Helper()
	data := make([]byte, 32)
	_, err := rand.Read(data)
	require.NoError(t, err)
	return data
}
