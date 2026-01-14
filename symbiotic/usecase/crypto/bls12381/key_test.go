package bls12381

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBLS12381Keys(t *testing.T) {
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

func TestMarshalText(t *testing.T) {
	key, err := NewPrivateKey(big.NewInt(123).FillBytes(make([]byte, 32)))
	require.NoError(t, err)

	text, err := key.PublicKey().MarshalText()
	require.NoError(t, err)

	require.Equal(t, "G1/0xa0ec3e71a719a25208adc97106b122809210faf45a17db24f10ffb1ac014fac1ab95a4a1967e55b185d4df622685b9e8;G2/0x95e18bbdb8b7bd39ea677ee923d7e87af449c45209e635907a4a8a2e4c65fff97c46d038cff53a994da273310ac85866096a5e13fd3ebf4e140e26f6ddfac66651e04e530e6045572acab753bb1bcef990fe14b4426caee41016af69d313750d", string(text))
}

func TestInvalidVerification(t *testing.T) {
	private, err := GenerateKey()
	require.NoError(t, err)
	public := private.PublicKey()

	err = public.VerifyWithHash([]byte{1}, nil)
	require.EqualError(t, err, "bls12381: invalid message hash length")

	err = public.VerifyWithHash(make([]byte, 32), nil)
	require.EqualError(t, err, "bls12381: failed to set big into G1: short buffer")

	err = public.VerifyWithHash(make([]byte, 32), make([]byte, 65))
	require.EqualError(t, err, "bls12381: failed to set big into G1: short buffer")
}

func TestFromRaw(t *testing.T) {
	_, err := FromRaw(nil)
	require.EqualError(t, err, "bls12381: nil raw key")

	_, err = FromRaw([]byte("invalid public key"))
	require.EqualError(t, err, "bls12381: invalid raw key length, expected 144, got 18")

	_, err = FromRaw(make([]byte, 144))
	require.EqualError(t, err, "bls12381: failed to unmarshal G1 pubkey: short buffer")
}

func randData(t *testing.T) []byte {
	t.Helper()
	data := make([]byte, 32)
	_, err := rand.Read(data)
	require.NoError(t, err)
	return data
}
