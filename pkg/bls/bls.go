package bls

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
)

// Constants from the Solidity BN254 library
var (
	// FpModulus is the modulus for the underlying field F_p of the elliptic curve
	FpModulus, _ = new(big.Int).SetString("21888242871839275222246405745257275088696311157297823662689037894645226208583", 10)

	// FrModulus is the modulus for the underlying field F_r of the elliptic curve
	FrModulus, _ = new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
)

// G1 represents a BLS public key (a point on G1)
type G1 struct {
	*bn254.G1Affine
}

// HashToG1 hashes data to a point on the BN254 curve
func HashToG1(data []byte) (*G1, error) {
	// Convert data to a big integer
	x := new(big.Int).SetBytes(data)

	// Ensure x is within the field
	x.Mod(x, FpModulus)

	for {
		// Find y coordinate for the current x
		beta, y, err := FindYFromX(x)
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

// FindYFromX calculates the y coordinate for a given x on the BN254 curve
// Returns (beta, y) where beta = x^3 + 3 (mod p) and y = sqrt(beta) if it exists
func FindYFromX(x *big.Int) (beta *big.Int, y *big.Int, err error) {
	// Calculate beta = x^3 + 3 mod p
	beta = new(big.Int).Exp(x, big.NewInt(3), FpModulus) // x^3
	beta.Add(beta, big.NewInt(3))                        // x^3 + 3
	beta.Mod(beta, FpModulus)                            // (x^3 + 3) mod p

	// Calculate y = beta^((p+1)/4) mod p
	// The exponent (p+1)/4 for BN254 is 0xc19139cb84c680a6e14116da060561765e05aa45a1c72a34f082305b61f3f52
	exponent, success := new(big.Int).SetString("c19139cb84c680a6e14116da060561765e05aa45a1c72a34f082305b61f3f52", 16)
	if !success {
		return nil, nil, errors.New("failed to set exponent")
	}

	y = new(big.Int).Exp(beta, exponent, FpModulus)

	return beta, y, nil
}

func Compress(g1 *G1) (common.Hash, error) {
	x := g1.X.BigInt(new(big.Int))
	y := g1.Y.BigInt(new(big.Int))
	_, derivedY, err := FindYFromX(x)
	if err != nil {
		return common.Hash{}, errors.New("failed to find Y from X")
	}

	flag := y.Cmp(derivedY) != 0
	compressedKeyG1 := new(big.Int).Mul(x, big.NewInt(2))
	if flag {
		compressedKeyG1.Add(compressedKeyG1, big.NewInt(1))
	}

	compressedKeyG1Bytes := [32]byte{}
	compressedKeyG1.FillBytes(compressedKeyG1Bytes[:])

	return common.Hash(compressedKeyG1Bytes), nil
}

func Decompress(compressed [32]byte) (*G1, error) {
	x, flag := new(big.Int).DivMod(new(big.Int).SetBytes(compressed[:32]), big.NewInt(2), big.NewInt(2))
	_, y, err := FindYFromX(x)
	if err != nil {
		return nil, err
	}
	g1 := ZeroG1()
	g1.X.SetBigInt(x)
	g1.Y.SetBigInt(y)
	if flag.Cmp(big.NewInt(1)) == 0 {
		g1.Neg(g1.G1Affine)
	}

	return g1, nil
}

func ZeroG1() *G1 {
	return &G1{G1Affine: new(bn254.G1Affine)}
}
