package bls

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/stretchr/testify/require"
)

// X 3540889048368762839810957002310395959297514716659952953090020851261640281949
//
// Y 3538373235723608870338803792065563630735759205971295744131778317708515222458

func TestComputeKeyPair(t *testing.T) {
	b, ok := new(big.Int).SetString("11008377096554045051122023680185802911050337017631086444859313200352654461863", 10)
	require.True(t, ok)
	keyPair := ComputeKeyPair(b.Bytes())

	fmt.Println(keyPair.PublicKeyG1.X.BigInt(new(big.Int)).String())
	fmt.Println(keyPair.PublicKeyG1.Y.BigInt(new(big.Int)).String())

	aaa(t, keyPair.PublicKeyG1)
}

func aaa(t *testing.T, affine G1) {
	arguments := abi.Arguments{
		{
			Name: "X",
			Type: abi.Type{
				T:    abi.UintTy,
				Size: 256,
			},
		},
		{
			Name: "Y",
			Type: abi.Type{
				T:    abi.UintTy,
				Size: 256,
			},
		},
	}

	// Кодируем значение ключа
	keyBytes, err := arguments.Pack(affine.G1Affine.X.BigInt(new(big.Int)), affine.G1Affine.Y.BigInt(new(big.Int)))
	require.NoError(t, err)

	h := affine.Bytes()

	fmt.Println(hex.EncodeToString(h[:]))     // TODO remove
	fmt.Println(hex.EncodeToString(keyBytes)) // TODO remove
}

//!!! 0x1740403ebEe1CF24fFb3Dc4De0C9f0225DfF5a71
//1ce60f16dce18584d1ba5f2367b25ba1e8bed720531f02ab5d2c74e07e4f42f913f857ca2ee059372a492f8f9cb6ef753f0a50bddfd4ed627bcd81530b5465ee
//0000000000000000000000001740403ebee1cf24ffb3dc4de0c9f0225dff5a71
//!!! 0x440a6eb7537dccD1Ed5028228a6aa1DEC607B377
//0db65d1f45b767890071e2f5b97fbb3379f52a94876cd3edd61d95ad5ee08bb60cb08f820079520583e2d4828e3b9a3076dabbfc45f407e149d1311081d12855
//000000000000000000000000440a6eb7537dccd1ed5028228a6aa1dec607b377
//!!! 0x75C070599402f65AD6f1165073105b2351C027d5
//07d4127a5f0129083bfd00521e2c36e072b431cefc43626362ddb9a635f80f5d07d2a5f5b198ca1782eca763eb6f14957230d83667a52f2890c33ed6729e23ba
//00000000000000000000000075c070599402f65ad6f1165073105b2351c027d5
