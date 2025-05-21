package test

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"middleware-offchain/internal/client/eth"
	"middleware-offchain/internal/client/valset"
	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
	"middleware-offchain/pkg/proof"
)

func Test_VerifyQuorumSig(t *testing.T) {
	pk1 := "87191036493798670866484781455694320176667203290824056510541300741498740913410"
	pk2 := "26972876870930381973856869753776124637336739336929668162870464864826929175089"
	pk3 := "11008377096554045051122023680185802911050337017631086444859313200352654461863"
	keyPair1 := bls.ComputeKeyPair(bytesFromPK(t, pk1))
	keyPair2 := bls.ComputeKeyPair(bytesFromPK(t, pk2))
	keyPair3 := bls.ComputeKeyPair(bytesFromPK(t, pk3))

	message := []byte("Hello, World!")
	messageHash := crypto.Keccak256(message)

	signature1, err := keyPair1.Sign(messageHash)
	require.NoError(t, err)
	signature2, err := keyPair2.Sign(messageHash)
	require.NoError(t, err)
	signature3, err := keyPair3.Sign(messageHash)
	require.NoError(t, err)

	eth1, err := eth.NewEthClient(eth.Config{
		MasterRPCURL:   "http://localhost:8545",
		MasterAddress:  "0x04C89607413713Ec9775E14b954286519d836FEf",
		PrivateKey:     bytesFromPK(t, pk1),
		RequestTimeout: time.Minute,
	})

	phase, err := eth1.GetCurrentPhase(t.Context())
	require.NoError(t, err)
	require.Equal(t, entity.IDLE, phase)

	require.NoError(t, err)
	deriver1, err := valset.NewDeriver(eth1)
	require.NoError(t, err)

	validatorSet, err := deriver1.GetValidatorSet(t.Context(), new(big.Int).SetInt64(time.Now().Unix()))
	require.NoError(t, err)
	_ = validatorSet
	epoch, err := eth1.GetCurrentEpoch(t.Context()) // todo current valset epoch method
	require.NoError(t, err)

	aggSignature := bls.ZeroG1().
		Add(signature1).
		Add(signature2).
		Add(signature3)
	//aggPublicKeyG1 := bls.ZeroG1().
	//	Add(&svc.keyPair1.PublicKeyG1).
	//	Add(&svc.keyPair2.PublicKeyG1).
	//	Add(&svc.keyPair3.PublicKeyG1)
	aggPublicKeyG2 := bls.ZeroG2().
		Add(&keyPair1.PublicKeyG2).
		Add(&keyPair2.PublicKeyG2).
		Add(&keyPair3.PublicKeyG2)

	proofData, err := proof.DoProve(proof.RawProveInput{
		AllValidators:    validatorSet.Validators,
		SignerValidators: validatorSet.Validators,
		RequiredKeyTag:   15,
		Message:          messageHash,
		Signature:        *aggSignature,
		SignersAggKeyG2:  *aggPublicKeyG2,
	})
	require.NoError(t, err)

	fmt.Println(">>> proofData: ", hex.EncodeToString(proofData.Marshall()))

	marshall := proofData.Marshall()
	//marshall = decodeHex(t, "2aa7ad49204d25ed5ba4cd6a358d1dc9cbf02c16730da4c1ac274e3fd80af4f816d93ad4be7d52da1b815058bfb5ce92bcf837773fe84588d980f959ce5adc5320a9ac5fdb1720ed31325796e4fffec52c1cb3977abae58b23d3a9e9469821e50a931a790ffc232f95248a2f70ada7adb18c0aff9ebda8b277d37dc42031c7b91ac956f6c0677e74986985885bbb4e947476925066dcecf1e7260c909a7f789314f0fe13c939302085ec398ddb8ffe65330e1a3c1b9f6925d183cb71aa0fe9e31ee24c8653d310de152a73fc01a651ae0732a7145bc25bc05a4f802732281ae622c2f1ea54899e76710d67ca1f8049539be02a5013d2ac9edf6a8178404cd1d2283c0bc04119d80a3931c32024d5733673ea86e31ce560f8076b0082070ec639258520fa397318b3639fd7a98704bd8df125a8bb2ff20a66612f3def4e1529a314dc70a5ec5edbb9bab8006772b44b10f01db98b790ad24096276810c683d02f014ca1ef05888ed653675eb270ee0f4f4d4d87bc52e72b2a2dfce9e91933366e0000000000000000000000000000000000000000000000000000000000000000")
	ok, err := eth1.VerifyQuorumSig(t.Context(), epoch, messageHash, 15, new(big.Int).SetInt64(1e18), marshall)
	require.NoError(t, err)

	require.True(t, ok)
}

func decodeHex(t *testing.T, s string) []byte { //nolint:unused // will be used later
	t.Helper()
	b, err := hex.DecodeString(s)
	require.NoError(t, err)
	return b
}
