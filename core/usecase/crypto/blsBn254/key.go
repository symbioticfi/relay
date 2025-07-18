package blsBn254

import (
	"fmt"
	"math/big"

	"github.com/symbioticfi/relay/core/entity"
	symbKeys "github.com/symbioticfi/relay/core/usecase/crypto/key-types"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
)

type Signature = entity.RawSignature
type Message = []byte
type MessageHash = entity.RawMessageHash
type CompactPublicKey = entity.CompactPublicKey
type RawPublicKey = entity.RawPublicKey

const (
	RawKeyLength      int = 96
	MessageHashLength int = 32
)

type PublicKey struct {
	g1PubKey bn254.G1Affine
	g2PubKey bn254.G2Affine
}

type PrivateKey struct {
	privateKey *big.Int
}

func Hash(msg []byte) MessageHash {
	return crypto.Keccak256(msg)
}

func NewPrivateKey(b []byte) (*PrivateKey, error) {
	return &PrivateKey{
		privateKey: new(big.Int).SetBytes(b),
	}, nil
}

func GenerateKey() (*PrivateKey, error) {
	sk := new(fr.Element)
	var err error

	sk, err = sk.SetRandom()
	if err != nil {
		return nil, errors.Errorf("blsBn254: failed to generate key: %w", err)
	}

	return &PrivateKey{
		privateKey: sk.BigInt(new(big.Int)),
	}, nil
}

func (k *PrivateKey) Bytes() []byte {
	return k.privateKey.Bytes()
}

func (k *PrivateKey) Sign(msg []byte) (Signature, MessageHash, error) {
	// symbiotic using keccak256 for hashing in bls-bn254
	hash := Hash(msg)

	g1Hash, err := hashToG1(hash)
	if err != nil {
		return nil, nil, errors.Errorf("blsBn254: failed to map hash to G1: %w", err)
	}

	g1Sig := bn254.G1Affine{}
	g1Sig.ScalarMultiplication(g1Hash, k.privateKey)

	// returns non compressed G1 point
	return g1Sig.Marshal(), hash, nil
}

func hashToG1(data []byte) (*bn254.G1Affine, error) {
	// Convert data to a big integer
	x := new(big.Int).SetBytes(data)

	fpModulus := fp.Modulus()
	// Ensure x is within the field
	x.Mod(x, fpModulus)

	for {
		// Find y coordinate for the current x
		beta, y, err := findYFromX(x)
		if err != nil {
			return nil, err
		}

		// Check if y^2 == beta
		y2 := new(big.Int).Mul(y, y)
		y2.Mod(y2, fpModulus)

		if y2.Cmp(beta) == 0 {
			// Create a G1 point with the found coordinates
			var point bn254.G1Affine
			point.X.SetBigInt(x)
			point.Y.SetBigInt(y)

			return &point, nil
		}

		// Increment x and try again
		x.Add(x, big.NewInt(1))
		x.Mod(x, fpModulus)
	}
}

// FindYFromX calculates the y coordinate for a given x on the BN254 curve
// Returns (beta, y) where beta = x^3 + 3 (mod p) and y = sqrt(beta) if it exists
func findYFromX(x *big.Int) (beta *big.Int, y *big.Int, err error) {
	fpModulus := fp.Modulus()

	// Calculate beta = x^3 + 3 mod p
	beta = new(big.Int).Exp(x, big.NewInt(3), fpModulus) // x^3
	beta.Add(beta, big.NewInt(3))                        // x^3 + 3
	beta.Mod(beta, fpModulus)                            // (x^3 + 3) mod p

	// Calculate y = beta^((p+1)/4) mod p
	// The exponent (p+1)/4 for BN254 is 0xc19139cb84c680a6e14116da060561765e05aa45a1c72a34f082305b61f3f52
	exponent, success := new(big.Int).SetString("c19139cb84c680a6e14116da060561765e05aa45a1c72a34f082305b61f3f52", 16)
	if !success {
		return nil, nil, errors.New("blsBn254: failed to set exponent")
	}

	y = new(big.Int).Exp(beta, exponent, fpModulus)

	return beta, y, nil
}

func (k *PrivateKey) PublicKey() symbKeys.PublicKey {
	g1PubKey := bn254.G1Affine{}
	g2PubKey := bn254.G2Affine{}

	// Get the generators for G1 and G2
	_, _, g1Gen, g2Gen := bn254.Generators()

	// Compute public keys by scalar multiplication with generators
	g1PubKey.ScalarMultiplication(&g1Gen, k.privateKey)
	g2PubKey.ScalarMultiplication(&g2Gen, k.privateKey)

	return &PublicKey{
		g1PubKey: g1PubKey,
		g2PubKey: g2PubKey,
	}
}

func NewPublicKey(g1PubKey bn254.G1Affine, g2PubKey bn254.G2Affine) *PublicKey {
	return &PublicKey{
		g1PubKey: g1PubKey,
		g2PubKey: g2PubKey,
	}
}

func (k *PublicKey) Verify(msg Message, sig Signature) error {
	msgHash := Hash(msg)

	// Hash the message to a point on G1
	g1Hash, err := hashToG1(msgHash)
	if err != nil {
		return errors.Errorf("blsBn254: failed to hash message to G1: %w", err)
	}

	g1Sig := bn254.G1Affine{}
	_, err = g1Sig.SetBytes(sig)
	if err != nil {
		return errors.Errorf("blsBn254: failed to set big into G1: %w", err)
	}

	// Get the G2 generator
	_, _, _, g2Gen := bn254.Generators()

	var negSig bn254.G1Affine
	negSig.Neg(&g1Sig)

	g1P := [2]bn254.G1Affine{*g1Hash, negSig}
	g1Q := [2]bn254.G2Affine{k.g2PubKey, g2Gen}

	ok, err := bn254.PairingCheck(g1P[:], g1Q[:])
	if err != nil {
		return errors.Errorf("blsBn254: pairing check failed: %w", err)
	}
	if !ok {
		return errors.Errorf("blsBn254: invalid signature")
	}
	return nil
}

func (k *PublicKey) VerifyWithHash(msgHash MessageHash, sig Signature) error {
	if len(msgHash) != MessageHashLength {
		return errors.Errorf("blsBn254: invalid message hash length")
	}

	// Hash the message to a point on G1
	g1Hash, err := hashToG1(msgHash)
	if err != nil {
		return errors.Errorf("blsBn254: failed to hash message to G1: %w", err)
	}

	g1Sig := bn254.G1Affine{}
	_, err = g1Sig.SetBytes(sig)
	if err != nil {
		return errors.Errorf("blsBn254: failed to set big into G1: %w", err)
	}

	// Get the G2 generator
	_, _, _, g2Gen := bn254.Generators()

	var negSig bn254.G1Affine
	negSig.Neg(&g1Sig)

	g1P := [2]bn254.G1Affine{*g1Hash, negSig}
	g1Q := [2]bn254.G2Affine{k.g2PubKey, g2Gen}

	ok, err := bn254.PairingCheck(g1P[:], g1Q[:])
	if err != nil {
		return errors.Errorf("blsBn254: pairing check failed: %w", err)
	}
	if !ok {
		return errors.Errorf("blsBn254: invalid signature")
	}
	return nil
}

// OnChain might be one way operation, meaning that it's impossible to reconstruct PublicKey from compact
func (k *PublicKey) OnChain() CompactPublicKey {
	return k.g1PubKey.Marshal()
}

func (k *PublicKey) Raw() RawPublicKey {
	g1Bytes := k.g1PubKey.Bytes()
	g2Bytes := k.g2PubKey.Bytes()

	// combined g1 and g2 [compressed]
	return append(g1Bytes[:], g2Bytes[:]...)
}

func (k *PublicKey) MarshalText() (text []byte, err error) {
	g1Bytes := k.g1PubKey.Bytes()
	g2Bytes := k.g2PubKey.Bytes()
	return []byte(fmt.Sprintf("G1/%s;G2/%s", hexutil.Encode(g1Bytes[:]), hexutil.Encode(g2Bytes[:]))), nil
}

func FromRaw(rawKey RawPublicKey) (*PublicKey, error) {
	if rawKey == nil {
		return nil, errors.New("blsBn254: nil raw key")
	}
	if len(rawKey) != RawKeyLength {
		return nil, fmt.Errorf("blsBn254: invalid raw key length, expected %d, got %d", RawKeyLength, len(rawKey))
	}
	g1 := bn254.G1Affine{}
	g2 := bn254.G2Affine{}

	err := g1.Unmarshal(rawKey[:32])
	if err != nil {
		return nil, fmt.Errorf("blsBn254: failed to unmarshal G1 pubkey: %w", err)
	}
	err = g2.Unmarshal(rawKey[32:])
	if err != nil {
		return nil, fmt.Errorf("blsBn254: failed to unmarshal G2 pubkey: %w", err)
	}
	return &PublicKey{g1, g2}, nil
}

func FromPrivateKey(privateKey *PrivateKey) symbKeys.PublicKey {
	return privateKey.PublicKey()
}
