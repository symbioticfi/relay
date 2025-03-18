package bls

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Constants from the Solidity BN254 library
var (
	// FpModulus is the modulus for the underlying field F_p of the elliptic curve
	FpModulus, _ = new(big.Int).SetString("21888242871839275222246405745257275088696311157297823662689037894645226208583", 10)

	// FrModulus is the modulus for the underlying field F_r of the elliptic curve
	FrModulus, _ = new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
)

// G1Point represents a point on the G1 curve (similar to the Solidity struct)
type G1Point struct {
	X *big.Int
	Y *big.Int
}

// G2Point represents a point on the G2 curve (similar to the Solidity struct)
type G2Point struct {
	X [2]*big.Int
	Y [2]*big.Int
}

// SecretKey represents a BLS secret key
type SecretKey struct {
	value fr.Element
}

// PublicKey represents a BLS public key (a point on G2)
type PublicKey struct {
	point bn254.G2Affine
}

// Signature represents a BLS signature (a point on G1)
type Signature struct {
	point bn254.G1Affine
}

// GenerateKey generates a new BLS key pair
func GenerateKey() (*SecretKey, *PublicKey, error) {
	// Generate a random secret key
	var skBytes [32]byte
	_, err := rand.Read(skBytes[:])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Create a secret key from the random bytes
	var sk SecretKey
	sk.value.SetBytes(skBytes[:])

	// Compute the public key
	var pk PublicKey
	var skBig big.Int
	sk.value.ToBigIntRegular(&skBig)

	// Compute g2 * sk where g2 is the generator of G2
	_, _, _, g2 := bn254.Generators()
	pk.point.ScalarMultiplication(&g2, &skBig)

	return &sk, &pk, nil
}

// Sign creates a BLS signature on a message using the secret key
func (sk *SecretKey) Sign(message []byte) (*Signature, error) {
	// Hash the message to a point on G1
	h1, err := hashToG1(message)
	if err != nil {
		return nil, fmt.Errorf("failed to hash message to G1: %w", err)
	}

	// Convert secret key to big.Int
	var skBig big.Int
	sk.value.ToBigIntRegular(&skBig)

	// Compute signature = h1 * sk
	var sig Signature
	sig.point.ScalarMultiplication(h1, &skBig)

	return &sig, nil
}

// Public returns the public key corresponding to the secret key
func (sk *SecretKey) Public() *PublicKey {
	var pk PublicKey
	var skBig big.Int
	sk.value.ToBigIntRegular(&skBig)

	// Compute the public key as g2 * sk
	_, _, _, g2 := bn254.Generators()
	pk.point.ScalarMultiplication(&g2, &skBig)

	return &pk
}

// Verify checks if a signature is valid for a message and public key
func (pk *PublicKey) Verify(signature *bn254.G1Affine, pubkey *bn254.G2Affine, message []byte) (bool, error) {
	// Hash the message to a point on G1
	h1, err := hashToG1(message)
	if err != nil {
		return false, fmt.Errorf("failed to hash message to G1: %w", err)
	}

	// Get the G2 generator
	_, _, _, g2 := bn254.Generators()

	var negSig bn254.G1Affine
	negSig.Neg((*bn254.G1Affine)(signature))

	P := [2]bn254.G1Affine{*h1, negSig}
	Q := [2]bn254.G2Affine{*pubkey, g2}

	ok, err := bn254.PairingCheck(P[:], Q[:])
	if err != nil {
		return false, nil
	}
	return ok, nil
}

// Serialize converts a signature to bytes
func (sig *Signature) Serialize() []byte {
	data := sig.point.Marshal()
	return data
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (pk *PublicKey) MarshalBinary() ([]byte, error) {
	return pk.point.Marshal(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (pk *PublicKey) UnmarshalBinary(data []byte) error {
	return pk.point.Unmarshal(data)
}

// AggregateSignatures combines multiple signatures into a single signature
func AggregateSignatures(signatures []bn254.G1Affine) (*bn254.G1Affine, error) {
	if len(signatures) == 0 {
		return nil, errors.New("no signatures to aggregate")
	}

	// Parse the first signature
	var aggSig *bn254.G1Affine

	// Add the remaining signatures
	for i := 1; i < len(signatures); i++ {
		aggSig = aggSig.Add(aggSig, &signatures[i])
	}

	return aggSig, nil
}

// hashToG1 hashes a message to a point on the G1 curve
func hashToG1(message []byte) (*bn254.G1Affine, error) {
	// Use mimc hash which is efficient for BN254
	mimcHash := mimc.NewMiMC()

	// Write the message to the hash
	_, err := mimcHash.Write(message)
	if err != nil {
		return nil, err
	}

	// Get the hash output
	hashOutput := mimcHash.Sum(nil)

	// Convert the hash to a scalar in Fr
	var scalar fr.Element
	scalar.SetBytes(hashOutput)

	// Get the generator of G1
	_, _, g1, _ := bn254.Generators()

	// Convert scalar to big.Int
	var scalarBig big.Int
	scalar.ToBigIntRegular(&scalarBig)

	// Create a point on G1 by multiplying the generator by the scalar
	var point bn254.G1Affine
	point.ScalarMultiplication(&g1, &scalarBig)

	return &point, nil
}

// EncodeToMessageHash prepares a message for signing by hashing it to a 32-byte value
func EncodeToMessageHash(message []byte) []byte {
	// Use Keccak256 for Ethereum compatibility
	return crypto.Keccak256(message)
}

// HashToField hashes the input to a field element (for testing/compatibility with Solidity)
func HashToField(input []byte) *big.Int {
	h := crypto.Keccak256(input)
	return new(big.Int).Mod(new(big.Int).SetBytes(h), FrModulus)
}

// FromSolidityG1Point converts a Solidity G1Point to gnark's G1Affine
func FromSolidityG1Point(x, y *big.Int) (bn254.G1Affine, error) {
	var p bn254.G1Affine

	// Check if x and y are valid field elements
	if x.Cmp(FpModulus) >= 0 || y.Cmp(FpModulus) >= 0 {
		return p, errors.New("coordinates out of range")
	}

	// Convert to bytes in the expected format
	xBytes := padBytes(x.Bytes(), 32)
	yBytes := padBytes(y.Bytes(), 32)

	// Combine x and y coordinates
	data := append(xBytes, yBytes...)

	// Unmarshal to G1Affine
	if err := p.Unmarshal(data); err != nil {
		return p, fmt.Errorf("failed to unmarshal G1 point: %w", err)
	}

	return p, nil
}

// ToSolidityG1Point converts gnark's G1Affine to Solidity G1Point coordinates
func ToSolidityG1Point(p bn254.G1Affine) (x, y *big.Int) {
	data := p.Marshal()

	x = new(big.Int).SetBytes(data[:32])
	y = new(big.Int).SetBytes(data[32:])

	return x, y
}

// FromSolidityG2Point converts Solidity G2Point to gnark's G2Affine
func FromSolidityG2Point(x [2]*big.Int, y [2]*big.Int) (bn254.G2Affine, error) {
	var p bn254.G2Affine

	// Check if coordinates are valid field elements
	for i := 0; i < 2; i++ {
		if x[i].Cmp(FpModulus) >= 0 || y[i].Cmp(FpModulus) >= 0 {
			return p, errors.New("coordinates out of range")
		}
	}

	// For BN254, G2 points are encoded as x0,x1,y0,y1 where
	// x = x0 + x1*u, y = y0 + y1*u (u is the quadratic non-residue)

	// Convert to bytes in the expected format
	x0Bytes := padBytes(x[0].Bytes(), 32)
	x1Bytes := padBytes(x[1].Bytes(), 32)
	y0Bytes := padBytes(y[0].Bytes(), 32)
	y1Bytes := padBytes(y[1].Bytes(), 32)

	// Combine all coordinates in the order expected by gnark
	data := make([]byte, 128)
	copy(data[0:32], x0Bytes)
	copy(data[32:64], x1Bytes)
	copy(data[64:96], y0Bytes)
	copy(data[96:128], y1Bytes)

	// Unmarshal to G2Affine
	if err := p.Unmarshal(data); err != nil {
		return p, fmt.Errorf("failed to unmarshal G2 point: %w", err)
	}

	return p, nil
}

// ToSolidityG2Point converts gnark's G2Affine to Solidity G2Point coordinates
func ToSolidityG2Point(p bn254.G2Affine) (x [2]*big.Int, y [2]*big.Int) {
	data := p.Marshal()

	x = [2]*big.Int{
		new(big.Int).SetBytes(data[0:32]),
		new(big.Int).SetBytes(data[32:64]),
	}
	y = [2]*big.Int{
		new(big.Int).SetBytes(data[64:96]),
		new(big.Int).SetBytes(data[96:128]),
	}

	return x, y
}

// padBytes pads a byte slice to the specified length
func padBytes(input []byte, length int) []byte {
	if len(input) >= length {
		return input
	}

	result := make([]byte, length)
	copy(result[length-len(input):], input)
	return result
}

// HashG1Point hashes a G1 point to a 32-byte value (equivalent to Solidity's hashG1Point)
func HashG1Point(p bn254.G1Affine) [32]byte {
	data := p.Marshal()
	return common.BytesToHash(crypto.Keccak256(data))
}

// HashG2Point hashes a G2 point to a 32-byte value (equivalent to Solidity's hashG2Point)
func HashG2Point(p bn254.G2Affine) [32]byte {
	data := p.Marshal()
	return common.BytesToHash(crypto.Keccak256(data))
}
