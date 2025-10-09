package ecdsaSecp256k1

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestECDSAKeys(t *testing.T) {
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
		newPublic := NewPublicKey(key.pubKey.X, key.pubKey.Y)

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

func TestMakePrivateFromNumber(t *testing.T) {
	keyBytes := big.NewInt(1000000000000).FillBytes(make([]byte, 32))
	newPrivateKey, err := NewPrivateKey(keyBytes)
	require.NoError(t, err)

	_, err = crypto.ToECDSA(newPrivateKey.Bytes())
	require.NoError(t, err)
}

func TestMarshallText(t *testing.T) {
	key, err := NewPrivateKey(big.NewInt(123).FillBytes(make([]byte, 32)))
	require.NoError(t, err)

	text, err := key.PublicKey().MarshalText()
	require.NoError(t, err)

	require.Equal(t, "0x03a598a8030da6d86c6bc7f2f5144ea549d28211ea58faa70ebf4c1e665c1fe9b5", string(text))
}

func TestInvalidVerification(t *testing.T) {
	private, err := GenerateKey()
	require.NoError(t, err)
	public := private.PublicKey()

	err = public.VerifyWithHash([]byte{1}, nil)
	require.Error(t, err)
	require.EqualError(t, err, "ecdsaSecp256k1: invalid message hash length")
	err = public.VerifyWithHash(make([]byte, 32), nil)
	require.Error(t, err)
	require.EqualError(t, err, "ecdsaSecp256k1: invalid signature length, expected 65 bytes, got 0")
	err = public.VerifyWithHash(make([]byte, 32), make([]byte, 65))
	require.Error(t, err)
	require.EqualError(t, err, "ecdsaSecp256k1: failed to verify signature 0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
}

func TestFromRaw(t *testing.T) {
	_, err := FromRaw(nil)
	require.EqualError(t, err, "ecdsaSecp256k1: nil raw key")

	_, err = FromRaw([]byte("invalid public key"))
	require.EqualError(t, err, "ecdsaSecp256k1: invalid raw key length, expected 33, got 18")

	_, err = FromRaw(make([]byte, 33))
	require.EqualError(t, err, "ecdsaSecp256k1: failed to decompress public key invalid public key")
}

func randData(t *testing.T) []byte {
	t.Helper()
	data := make([]byte, 32)
	_, err := rand.Read(data)
	require.NoError(t, err)
	return data
}
