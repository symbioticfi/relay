package bls12381

import (
	"fmt"
	"math/big"

	bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"

	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type Signature = symbiotic.RawSignature
type Message = []byte
type MessageHash = symbiotic.RawMessageHash
type CompactPublicKey = symbiotic.CompactPublicKey
type RawPublicKey = symbiotic.RawPublicKey

const (
	RawKeyLength      int = bls12381.SizeOfG1AffineCompressed + bls12381.SizeOfG2AffineCompressed
	MessageHashLength int = 32
	hashToG1Domain        = "BLS_SIG_BLS12381G1_XMD:SHA-256_SSWU_RO_NUL_"
)

type PublicKey struct {
	g1PubKey bls12381.G1Affine
	g2PubKey bls12381.G2Affine
}

type PrivateKey struct {
	privateKey *big.Int
}

func NewPrivateKey(b []byte) (*PrivateKey, error) {
	return &PrivateKey{privateKey: new(big.Int).SetBytes(b)}, nil
}

func HashMessage(msg []byte) MessageHash {
	return crypto.Keccak256(msg)
}

func GenerateKey() (*PrivateKey, error) {
	sk := new(fr.Element)
	var err error
	sk, err = sk.SetRandom()
	if err != nil {
		return nil, errors.Errorf("bls12381: failed to generate key: %w", err)
	}
	return &PrivateKey{privateKey: sk.BigInt(new(big.Int))}, nil
}

func (k *PrivateKey) Bytes() []byte {
	return k.privateKey.Bytes()
}

func (k *PrivateKey) Sign(msg []byte) (Signature, MessageHash, error) {
	// symbiotic using keccak256 for hashing in bls12381-bn254
	hash := HashMessage(msg)
	g1Hash, err := HashToG1(hash)
	if err != nil {
		return nil, nil, errors.Errorf("bls12381: failed to map hash to G1: %w", err)
	}

	var g1Sig bls12381.G1Affine
	g1Sig.ScalarMultiplication(g1Hash, k.privateKey)

	return g1Sig.Marshal(), hash, nil
}

func HashToG1(data []byte) (*bls12381.G1Affine, error) {
	point, err := bls12381.HashToG1(data, []byte(hashToG1Domain))
	if err != nil {
		return nil, errors.Errorf("bls12381: failed to hash to G1: %w", err)
	}
	return &point, nil
}

func (k *PrivateKey) PublicKey() symbiotic.PublicKey {
	g1PubKey := bls12381.G1Affine{}
	g2PubKey := bls12381.G2Affine{}

	_, _, g1Gen, g2Gen := bls12381.Generators()
	g1PubKey.ScalarMultiplication(&g1Gen, k.privateKey)
	g2PubKey.ScalarMultiplication(&g2Gen, k.privateKey)

	return &PublicKey{g1PubKey: g1PubKey, g2PubKey: g2PubKey}
}

func NewPublicKey(g1PubKey bls12381.G1Affine, g2PubKey bls12381.G2Affine) *PublicKey {
	return &PublicKey{g1PubKey: g1PubKey, g2PubKey: g2PubKey}
}

func (k *PublicKey) Verify(msg Message, sig Signature) error {
	msgHash := HashMessage(msg)
	return k.VerifyWithHash(msgHash, sig)
}

func (k *PublicKey) VerifyWithHash(msgHash MessageHash, sig Signature) error {
	if len(msgHash) != MessageHashLength {
		return errors.Errorf("bls12381: invalid message hash length")
	}

	g1Hash, err := HashToG1(msgHash)
	if err != nil {
		return errors.Errorf("bls12381: failed to hash message to G1: %w", err)
	}

	var g1Sig bls12381.G1Affine
	if _, err = g1Sig.SetBytes(sig); err != nil {
		return errors.Errorf("bls12381: failed to set big into G1: %w", err)
	}

	_, _, _, g2Gen := bls12381.Generators()
	var negSig bls12381.G1Affine
	negSig.Neg(&g1Sig)

	g1P := [2]bls12381.G1Affine{*g1Hash, negSig}
	g1Q := [2]bls12381.G2Affine{k.g2PubKey, g2Gen}

	ok, err := bls12381.PairingCheck(g1P[:], g1Q[:])
	if err != nil {
		return errors.Errorf("bls12381: pairing check failed: %w", err)
	}
	if !ok {
		return errors.Errorf("bls12381: invalid signature")
	}
	return nil
}

// OnChain might be one way operation, meaning that it's impossible to reconstruct PublicKey from compact
func (k *PublicKey) OnChain() CompactPublicKey {
	// DEV: g1PubKey Marshalled is 96 bytes in total, x and y each 48bytes
	// but onchain we need to pad each field to 64 bytes, hence we manually pad it here
	pk := k.g1PubKey.Marshal()
	paddedPk := make([]byte, 128)
	copy(paddedPk[16:64], pk[0:48])   // x coordinate
	copy(paddedPk[80:128], pk[48:96]) // y coordinate
	return paddedPk
}

func (k *PublicKey) Raw() RawPublicKey {
	g1Bytes := k.g1PubKey.Bytes()
	g2Bytes := k.g2PubKey.Bytes()

	// combined g1 and g2 [compressed]
	return append(g1Bytes[:], g2Bytes[:]...)
}

func (k *PublicKey) G2() *bls12381.G2Affine {
	return &k.g2PubKey
}

func (k *PublicKey) MarshalText() ([]byte, error) {
	g1Bytes := k.g1PubKey.Bytes()
	g2Bytes := k.g2PubKey.Bytes()
	return fmt.Appendf(nil, "G1/%s;G2/%s", hexutil.Encode(g1Bytes[:]), hexutil.Encode(g2Bytes[:])), nil
}

func FromRaw(rawKey RawPublicKey) (*PublicKey, error) {
	if rawKey == nil {
		return nil, errors.New("bls12381: nil raw key")
	}
	if len(rawKey) != RawKeyLength {
		return nil, errors.Errorf("bls12381: invalid raw key length, expected %d, got %d", RawKeyLength, len(rawKey))
	}

	var g1 bls12381.G1Affine
	var g2 bls12381.G2Affine

	if _, err := g1.SetBytes(rawKey[:bls12381.SizeOfG1AffineCompressed]); err != nil {
		return nil, errors.Errorf("bls12381: failed to unmarshal G1 pubkey: %w", err)
	}
	if _, err := g2.SetBytes(rawKey[bls12381.SizeOfG1AffineCompressed:]); err != nil {
		return nil, errors.Errorf("bls12381: failed to unmarshal G2 pubkey: %w", err)
	}

	return &PublicKey{g1PubKey: g1, g2PubKey: g2}, nil
}

func FromPrivateKey(privateKey *PrivateKey) symbiotic.PublicKey {
	return privateKey.PublicKey()
}
