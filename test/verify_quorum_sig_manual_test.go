package test

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"middleware-offchain/internal/client/eth"
	"middleware-offchain/internal/client/valset"
	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
	"middleware-offchain/pkg/proof"
)

func Test_VerifyQuorumSig(t *testing.T) {
	waitIdleOrCommitPhase(t)

	base := new(big.Int)
	base.SetString("1000000000000000000", 10)

	// message := []byte("Hello, World!")
	// messageHash := crypto.Keccak256(message)
	messageHash, err := hex.DecodeString("cca0534ef01f2606de9b6c90df9f0a2e1a18fb5ce4d1f9cf1f94d35b398ebce4")
	require.NoError(t, err)

	fmt.Println("messageHash>>>", hex.EncodeToString(messageHash))

	zeroPk := new(big.Int).Add(base, big.NewInt(int64(0)))
	pkBytes := [32]byte{}
	zeroPk.FillBytes(pkBytes[:])
	eth1, err := eth.NewEthClient(eth.Config{
		MasterRPCURL:   "http://localhost:8545",
		MasterAddress:  "0x63d855589514F1277527f4fD8D464836F8Ca73Ba",
		PrivateKey:     pkBytes[:],
		RequestTimeout: time.Minute,
	})

	require.NoError(t, err)
	deriver1, err := valset.NewDeriver(eth1)
	require.NoError(t, err)

	validatorSet, err := deriver1.GetValidatorSet(t.Context(), new(big.Int).SetInt64(time.Now().Unix()))
	require.NoError(t, err)
	_ = validatorSet

	epoch, err := eth1.GetCurrentEpoch(t.Context()) // todo current valset epoch method
	require.NoError(t, err)

	generator, err := valset.NewGenerator(deriver1, eth1)
	require.NoError(t, err)
	header, err := generator.GenerateCurrentValidatorSetHeader(t.Context())
	require.NoError(t, err)
	fmt.Println("header>>>", header)
	extraData, err := generator.GenerateExtraData(t.Context(), header, valset.ZkVerificationType)
	require.NoError(t, err)
	fmt.Println("extraData>>>", extraData)

	pks := make([]*big.Int, 0, len(validatorSet.Validators))
	for i := 0; i < len(validatorSet.Validators); i++ {
		n := new(big.Int).Add(base, big.NewInt(int64(i)))
		pks = append(pks, n)
	}

	keyPairs := make([]bls.KeyPair, 0, len(validatorSet.Validators))
	for _, pk := range pks {
		pkBytes := [32]byte{}
		pk.FillBytes(pkBytes[:])
		keyPairs = append(keyPairs, bls.ComputeKeyPair(pkBytes[:]))
	}

	signatures := make([]*bls.G1, 0, len(validatorSet.Validators))
	for _, keyPair := range keyPairs {
		signature, err := keyPair.Sign(messageHash)
		require.NoError(t, err)
		signatures = append(signatures, signature)
	}

	aggSignature := bls.ZeroG1()
	for _, signature := range signatures {
		aggSignature = aggSignature.Add(signature)
	}

	aggPublicKeyG1 := bls.ZeroG1()
	for _, keyPair := range keyPairs {
		aggPublicKeyG1 = aggPublicKeyG1.Add(&keyPair.PublicKeyG1)
	}

	aggPublicKeyG2 := bls.ZeroG2()
	for _, keyPair := range keyPairs {
		aggPublicKeyG2 = aggPublicKeyG2.Add(&keyPair.PublicKeyG2)
	}

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
	ok, err := eth1.VerifyQuorumSig(t.Context(), epoch, messageHash, 15, new(big.Int).Div(new(big.Int).Mul(validatorSet.TotalActiveVotingPower, big.NewInt(2)), big.NewInt(3)), marshall, []byte{})
	require.NoError(t, err)

	require.True(t, ok)
}

func decodeHex(t *testing.T, s string) []byte { //nolint:unused // will be used later
	t.Helper()
	b, err := hex.DecodeString(s)
	require.NoError(t, err)
	return b
}

func waitIdleOrCommitPhase(t *testing.T) {
	t.Helper()
	base := new(big.Int)
	base.SetString("1000000000000000000", 10)
	zeroPk := new(big.Int).Add(base, big.NewInt(int64(0)))
	pkBytes := [32]byte{}
	zeroPk.FillBytes(pkBytes[:])
	eth1, err := eth.NewEthClient(eth.Config{
		MasterRPCURL:   "http://localhost:8545",
		MasterAddress:  "0x63d855589514F1277527f4fD8D464836F8Ca73Ba",
		PrivateKey:     pkBytes[:],
		RequestTimeout: time.Minute,
	})
	require.NoError(t, err)

	tick := time.NewTicker(time.Second)
	for {
		select {
		case <-tick.C:
			phase, err := eth1.GetCurrentPhase(t.Context())
			require.NoError(t, err)
			if entity.IDLE == phase || entity.COMMIT == phase {
				return
			}
			slog.InfoContext(t.Context(), "waiting for idle or commit phase", "phase", phase)
		case <-t.Context().Done():
			return
		}
	}
}
