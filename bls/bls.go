package bls

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// Constants from the Solidity BN254 library
var (
	// FpModulus is the modulus for the underlying field F_p of the elliptic curve
	FpModulus, _ = new(big.Int).SetString("21888242871839275222246405745257275088696311157297823662689037894645226208583", 10)

	// FrModulus is the modulus for the underlying field F_r of the elliptic curve
	FrModulus, _ = new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
)

type KeyPair struct {
	SecretKey   SecretKey
	PublicKeyG1 G1
	PublicKeyG2 G2
}

// SecretKey represents a BLS secret key
type SecretKey struct {
	*fr.Element
}

// PublicKeyG1 represents a BLS public key (a point on G1)
type G1 struct {
	*bn254.G1Affine
}

// PublicKeyG2 represents a BLS public key (a point on G2)
type G2 struct {
	*bn254.G2Affine
}

// GenerateKeyOrLoad generates a new BLS key pair or loads an existing one from the specified path
func GenerateKeyOrLoad(path string) (*KeyPair, error) {
	// Try to load the key from file if it exists
	if keyPair, err := loadKeyFromFile(path); err == nil {
		return keyPair, nil
	}

	// Generate a new key if loading failed or file doesn't exist
	return generateAndSaveKey(path)
}

// loadKeyFromFile attempts to load a key from the given file path
func loadKeyFromFile(path string) (*KeyPair, error) {
	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	// Read the file
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse the key
	var sk SecretKey
	if err := sk.SetBytes(keyData); err != nil {
		return nil, fmt.Errorf("failed to parse secret key: %w", err)
	}

	// Compute public keys from secret key
	return computeKeyPair(&sk), nil
}

// generateAndSaveKey creates a new random key and saves it to the specified path
func generateAndSaveKey(path string) (*KeyPair, error) {
	// Generate random bytes for secret key
	var skBytes [32]byte
	if _, err := rand.Read(skBytes[:]); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Create a secret key from the random bytes
	var sk SecretKey
	if err := sk.SetBytes(skBytes[:]); err != nil {
		return nil, fmt.Errorf("failed to create secret key: %w", err)
	}

	// Compute the key pair
	keyPair := computeKeyPair(&sk)

	// Save the key to file
	keyData := sk.Marshal()
	if err := os.WriteFile(path, keyData, 0600); err != nil {
		return nil, fmt.Errorf("failed to save key to file: %w", err)
	}

	return keyPair, nil
}

// computeKeyPair derives the public keys from a secret key
func computeKeyPair(sk *SecretKey) *KeyPair {
	var pkG1 G1
	var pkG2 G2
	var skBig big.Int
	sk.BigInt(&skBig)

	// Get the generators for G1 and G2
	_, _, g1, g2 := bn254.Generators()

	// Compute public keys by scalar multiplication with generators
	pkG1.ScalarMultiplication(&g1, &skBig)
	pkG2.ScalarMultiplication(&g2, &skBig)

	return &KeyPair{
		SecretKey:   *sk,
		PublicKeyG1: pkG1,
		PublicKeyG2: pkG2,
	}
}

// Sign creates a BLS signature on a message using the secret key
func (kp *KeyPair) Sign(msgHash []byte) (*G1, error) {
	// Hash the message to a point on G1
	h1, err := hashToG1(msgHash)
	if err != nil {
		return nil, fmt.Errorf("failed to hash message to G1: %w", err)
	}

	// Convert secret key to big.Int
	var skBig big.Int
	kp.SecretKey.BigInt(&skBig)

	// Compute signature = h1 * sk
	var sig G1
	sig.ScalarMultiplication(h1.G1Affine, &skBig)

	return &sig, nil
}

// Verify checks if a signature is valid for a message and public key
func (pubkey *G2) Verify(signature *G1, msgHash []byte) (bool, error) {
	// Hash the message to a point on G1
	h1, err := hashToG1(msgHash)
	if err != nil {
		return false, fmt.Errorf("failed to hash message to G1: %w", err)
	}

	// Get the G2 generator
	_, _, _, g2 := bn254.Generators()

	var negSig bn254.G1Affine
	negSig.Neg(signature.G1Affine)

	P := [2]bn254.G1Affine{*h1.G1Affine, negSig}
	Q := [2]bn254.G2Affine{*pubkey.G2Affine, g2}

	ok, err := bn254.PairingCheck(P[:], Q[:])
	if err != nil {
		return false, nil
	}
	return ok, nil
}

// HashToPoint hashes data to a point on the BN254 curve
func hashToG1(data []byte) (*G1, error) {
	// Convert data to a big integer
	x := new(big.Int).SetBytes(data)

	// Ensure x is within the field
	x.Mod(x, FpModulus)

	for {
		// Find y coordinate for the current x
		beta, y, err := findYFromX(x)
		if err != nil {
			return nil, err
		}

		// Check if y^2 == beta
		y2 := new(big.Int).Mul(y, y)
		y2.Mod(y2, FpModulus)

		if y2.Cmp(beta) == 0 {
			// Create a G1 point with the found coordinates
			var point bn254.G1Affine
			point.X.SetBigInt(x)
			point.Y.SetBigInt(y)

			return &G1{G1Affine: &point}, nil
		}

		// Increment x and try again
		x.Add(x, big.NewInt(1))
		x.Mod(x, FpModulus)
	}
}

// findYFromX calculates the y coordinate for a given x on the BN254 curve
// Returns (beta, y) where beta = x^3 + 3 (mod p) and y = sqrt(beta) if it exists
func findYFromX(x *big.Int) (*big.Int, *big.Int, error) {
	// Calculate beta = x^3 + 3 mod p
	beta := new(big.Int).Exp(x, big.NewInt(3), FpModulus) // x^3
	beta.Add(beta, big.NewInt(3))                         // x^3 + 3
	beta.Mod(beta, FpModulus)                             // (x^3 + 3) mod p

	// Calculate y = beta^((p+1)/4) mod p
	// The exponent (p+1)/4 for BN254 is 0xc19139cb84c680a6e14116da060561765e05aa45a1c72a34f082305b61f3f52
	exponent, success := new(big.Int).SetString("c19139cb84c680a6e14116da060561765e05aa45a1c72a34f082305b61f3f52", 16)
	if !success {
		return nil, nil, fmt.Errorf("failed to set exponent")
	}

	y := new(big.Int).Exp(beta, exponent, FpModulus)

	return beta, y, nil
}

// Add adds two G1 public keys together
func (p *G1) Add(other *G1) *G1 {
	var result bn254.G1Affine
	result.Add(p.G1Affine, other.G1Affine)
	p.G1Affine = &result
	return p
}

// Sub subtracts another G1 public key from this one
func (p *G1) Sub(other *G1) *G1 {
	var result bn254.G1Affine
	var negOther bn254.G1Affine
	negOther.Neg(other.G1Affine)
	result.Add(p.G1Affine, &negOther)
	p.G1Affine = &result
	return p
}

// Add adds two G2 public keys together
func (p *G2) Add(other *G2) *G2 {
	var result bn254.G2Affine
	result.Add(p.G2Affine, other.G2Affine)
	p.G2Affine = &result
	return p
}

// Sub subtracts another G2 public key from this one
func (p *G2) Sub(other *G2) *G2 {
	var result bn254.G2Affine
	var negOther bn254.G2Affine
	negOther.Neg(other.G2Affine)
	result.Add(p.G2Affine, &negOther)
	p.G2Affine = &result
	return p
}

func SerializeG1(g1 *G1) []byte {
	bytes := g1.G1Affine.RawBytes()
	return bytes[:]
}

func SerializeG2(g2 *G2) []byte {
	bytes := g2.G2Affine.RawBytes()
	return bytes[:]
}

func DeserializeG1(bytes []byte) *G1 {
	var g1 G1
	g1.G1Affine.SetBytes(bytes)
	return &g1
}

func DeserializeG2(bytes []byte) *G2 {
	var g2 G2
	g2.G2Affine.SetBytes(bytes)
	return &g2
}
