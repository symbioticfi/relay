//go:build manual

package test

import (
	"bytes"
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

type valsetTestServices struct {
	keyPair1 bls.KeyPair
	keyPair2 bls.KeyPair
	keyPair3 bls.KeyPair

	eth1 *eth.Client
	eth2 *eth.Client
	eth3 *eth.Client

	generator1 *valset.Generator
	generator2 *valset.Generator
	generator3 *valset.Generator

	headerSignature1 *bls.G1
	headerSignature2 *bls.G1
	headerSignature3 *bls.G1
	deriver1         *valset.Deriver
}

func TestCommitValsetHeader(t *testing.T) {
	waitCommitPhase(t)

	svc := initValsetTestServices(t)

	validatorSet, err := svc.deriver1.GetValidatorSet(t.Context(), new(big.Int).SetInt64(time.Now().Unix()))
	require.NoError(t, err)

	aggSignature := bls.ZeroG1().
		Add(svc.headerSignature1).
		Add(svc.headerSignature2).
		Add(svc.headerSignature3)
	//aggPublicKeyG1 := bls.ZeroG1().
	//	Add(&svc.keyPair1.PublicKeyG1).
	//	Add(&svc.keyPair2.PublicKeyG1).
	//	Add(&svc.keyPair3.PublicKeyG1)
	aggPublicKeyG2 := bls.ZeroG2().
		Add(&svc.keyPair1.PublicKeyG2).
		Add(&svc.keyPair2.PublicKeyG2).
		Add(&svc.keyPair3.PublicKeyG2)

	proofData, err := proof.DoProve(validatorSet.Validators, 15)
	require.NoError(t, err)

	header, err := svc.generator1.GenerateValidatorSetHeader(t.Context())
	require.NoError(t, err)

	var result bytes.Buffer

	fmt.Println("G2>>>", hex.EncodeToString(aggPublicKeyG2.Marshal()))
	fmt.Println("aggSigG1>>>", hex.EncodeToString(aggSignature.Marshal()))
	fmt.Println("proof_>>>", hex.EncodeToString(proofData[:256]))
	fmt.Println("Commitments>>>", hex.EncodeToString(proofData[260:324]))
	fmt.Println("commitmentPok>>>", hex.EncodeToString(proofData[324:388]))

	result.Write(aggSignature.Marshal())   // abi.encode(aggSigG1)
	result.Write(aggPublicKeyG2.Marshal()) // abi.encode(aggKeyG2)
	result.Write(proofData[:256])          // slice(proof_, 0, 256)
	result.Write(proofData[260:324])       // slice(commitments, 260, 324)
	result.Write(proofData[324:388])       // slice(commitmentPok, 324, 388)
	result.Write(inputs(t))                // zkProof.input

	err = svc.eth1.CommitValsetHeader(t.Context(), header, result.Bytes())
	require.NoError(t, err)
}

func inputs(t *testing.T) []byte {
	in := []string{"0", "0", "0", "0", "0", "0", "0", "0", "17452784377140135873242247846499243451530443834097508626974155003329264289405", "0"}
	var result bytes.Buffer
	for _, s := range in {
		b, ok := new(big.Int).SetString(s, 10)
		require.True(t, ok)
		result.Write(b.Bytes())
	}
	return result.Bytes()
}

func initValsetTestServices(t *testing.T) *valsetTestServices {
	pk1 := "87191036493798670866484781455694320176667203290824056510541300741498740913410"
	pk2 := "26972876870930381973856869753776124637336739336929668162870464864826929175089"
	pk3 := "11008377096554045051122023680185802911050337017631086444859313200352654461863"

	keyPair1 := bls.ComputeKeyPair(bytesFromPK(t, pk1))
	keyPair2 := bls.ComputeKeyPair(bytesFromPK(t, pk2))
	keyPair3 := bls.ComputeKeyPair(bytesFromPK(t, pk3))

	eth1, err := eth.NewEthClient(eth.Config{
		MasterRPCURL:   "http://localhost:8545",
		MasterAddress:  "0x5081a39b8A5f0E35a8D959395a630b68B74Dd30f",
		PrivateKey:     bytesFromPK(t, pk1),
		RequestTimeout: time.Minute,
	})
	require.NoError(t, err)

	eth2, err := eth.NewEthClient(eth.Config{
		MasterRPCURL:   "http://localhost:8545",
		MasterAddress:  "0x5081a39b8A5f0E35a8D959395a630b68B74Dd30f",
		PrivateKey:     bytesFromPK(t, pk2),
		RequestTimeout: time.Minute,
	})
	require.NoError(t, err)

	eth3, err := eth.NewEthClient(eth.Config{
		MasterRPCURL:   "http://localhost:8545",
		MasterAddress:  "0x5081a39b8A5f0E35a8D959395a630b68B74Dd30f",
		PrivateKey:     bytesFromPK(t, pk3),
		RequestTimeout: time.Minute,
	})
	require.NoError(t, err)

	phase, err := eth1.GetCurrentPhase(t.Context())
	require.NoError(t, err)
	require.Equal(t, entity.COMMIT, phase)

	slog.InfoContext(t.Context(), "current phase", "phase", phase)

	deriver1, err := valset.NewDeriver(eth1)
	require.NoError(t, err)

	generator1, err := valset.NewGenerator(deriver1, eth1)
	require.NoError(t, err)

	deriver2, err := valset.NewDeriver(eth2)
	require.NoError(t, err)

	generator2, err := valset.NewGenerator(deriver2, eth2)
	require.NoError(t, err)

	deriver3, err := valset.NewDeriver(eth3)
	require.NoError(t, err)

	generator3, err := valset.NewGenerator(deriver3, eth3)
	require.NoError(t, err)

	header1, err := generator1.GenerateValidatorSetHeader(t.Context())
	require.NoError(t, err)

	headerHash1, err := generator1.GenerateValidatorSetHeaderHash(t.Context(), header1)
	require.NoError(t, err)

	fmt.Println(hex.EncodeToString(headerHash1))

	headerSignature1, err := keyPair1.Sign(headerHash1)
	require.NoError(t, err)

	header2, err := generator2.GenerateValidatorSetHeader(t.Context())
	require.NoError(t, err)

	headerHash2, err := generator2.GenerateValidatorSetHeaderHash(t.Context(), header2)
	require.NoError(t, err)

	headerSignature2, err := keyPair2.Sign(headerHash2)
	require.NoError(t, err)

	header3, err := generator3.GenerateValidatorSetHeader(t.Context())
	require.NoError(t, err)

	headerHash3, err := generator3.GenerateValidatorSetHeaderHash(t.Context(), header3)
	require.NoError(t, err)

	headerSignature3, err := keyPair3.Sign(headerHash3)
	require.NoError(t, err)

	return &valsetTestServices{
		keyPair1:         keyPair1,
		keyPair2:         keyPair2,
		keyPair3:         keyPair3,
		eth1:             eth1,
		eth2:             eth2,
		eth3:             eth3,
		deriver1:         deriver1,
		generator1:       generator1,
		generator2:       generator2,
		generator3:       generator3,
		headerSignature1: headerSignature1,
		headerSignature2: headerSignature2,
		headerSignature3: headerSignature3,
	}
}

func bytesFromPK(t *testing.T, pk1 string) []byte {
	t.Helper()
	b, ok := new(big.Int).SetString(pk1, 10)
	require.True(t, ok)
	return b.Bytes()
}

func waitCommitPhase(t *testing.T) {
	eth1, err := eth.NewEthClient(eth.Config{
		MasterRPCURL:   "http://localhost:8545",
		MasterAddress:  "0x5081a39b8A5f0E35a8D959395a630b68B74Dd30f",
		PrivateKey:     bytesFromPK(t, "87191036493798670866484781455694320176667203290824056510541300741498740913410"),
		RequestTimeout: time.Minute,
	})
	require.NoError(t, err)

	tick := time.NewTicker(time.Second)
	for {
		select {
		case <-tick.C:
			phase, err := eth1.GetCurrentPhase(t.Context())
			require.NoError(t, err)
			if entity.COMMIT == phase {
				return
			}
			slog.InfoContext(t.Context(), "waiting for commit phase", "phase", phase)
		case <-t.Context().Done():
			return
		}
	}
}
