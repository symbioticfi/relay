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
	keyPairs []bls.KeyPair

	eths []*eth.Client

	derivers   []*valset.Deriver
	generators []*valset.Generator

	headerSignatures []*bls.G1
	headerHashes     [][]byte
	extraDatas       [][]entity.ExtraData
	headers          []entity.ValidatorSetHeader
}

func TestCommitValsetHeader(t *testing.T) {
	waitCommitPhase(t)

	svc := initValsetTestServices(t)

	validatorSet, err := svc.derivers[0].GetValidatorSet(t.Context(), new(big.Int).SetInt64(time.Now().Unix()))
	require.NoError(t, err)
	_ = validatorSet

	for i := 0; i < len(svc.headerHashes); i++ {
		for j := i + 1; j < len(svc.headerHashes); j++ {
			if !bytes.Equal(svc.headerHashes[i], svc.headerHashes[j]) {
				t.Fatal("headerHashes are not equal")
			}
		}
	}

	aggSignature := bls.ZeroG1()
	for _, headerSignature := range svc.headerSignatures {
		aggSignature = aggSignature.Add(headerSignature)
	}

	aggPublicKeyG1 := bls.ZeroG1()
	for _, keyPair := range svc.keyPairs {
		aggPublicKeyG1 = aggPublicKeyG1.Add(&keyPair.PublicKeyG1)
	}

	aggPublicKeyG2 := bls.ZeroG2()
	for _, keyPair := range svc.keyPairs {
		aggPublicKeyG2 = aggPublicKeyG2.Add(&keyPair.PublicKeyG2)
	}

	proofData, err := proof.DoProve(proof.RawProveInput{
		AllValidators:    validatorSet.Validators,
		SignerValidators: validatorSet.Validators,
		RequiredKeyTag:   15,
		Message:          svc.headerHashes[0],
		Signature:        *aggSignature,
		SignersAggKeyG2:  *aggPublicKeyG2,
	})
	require.NoError(t, err)

	//proofData := decodeHex(t, "01c70454a912bf226d9b0a7b38dfef9319f92b893115fd5b168f0061c56a11e30d8c66ed7585aafd81e6c20cdbe81d385ee13871ef1da2b041e218076fbfd88e0cf9e7f5e7b25f241973a4a4ae6a7f29d430af5c243cd254d5035e7ad1883d9d1f0c0f76011867ebb6115185f5b9fe538de1181a39cd9e5efa03046b031c64df0b86f9a8fcb28738e82eabe0237a57bc47a02158841039f0d12ec3abb3d9ee2d12185fed2304764a978ba5405c684093479e18d934c7c8ba9e031981c836d4ff028f842b327dd18be5ba410bc423ce989f6807f1766acdae5669dea546d8cd591f98ba029d5e2a77520b2639234354c2e3983ce9590efbee7b293a8ee32bfbf80000000124997c0ef7b3e53580aaa97c84ae4682a7a7ec617110c5790ce06ca6bf837600114ec9b6c4503e96f11bcdb0e4601fecd83b5b8e4c7d9df6204aea2a4b7617471e9bada9e6dd91bf84b89967925bf1a90aa162f5f2883c4713522263f983f5ec101464ab309ff2a609396c898689eb0e5e4f703d350adc6ed69d6dfdc1a5bbbd")

	header, err := svc.generators[0].GenerateValidatorSetHeaderOnCapture(t.Context())
	require.NoError(t, err)

	extraData, err := svc.generators[0].GenerateExtraData(t.Context(), header, valset.ZkVerificationType)
	require.NoError(t, err)

	fmt.Println("G2>>>", hex.EncodeToString(aggPublicKeyG2.Marshal()))
	fmt.Println("aggSigG1>>>", hex.EncodeToString(aggSignature.Marshal()))
	fmt.Println("proof_>>>", hex.EncodeToString(proofData.Proof))
	fmt.Println("Commitments>>>", hex.EncodeToString(proofData.Commitments))
	fmt.Println("commitmentPok>>>", hex.EncodeToString(proofData.CommitmentPok))
	fmt.Println("validatorSet.TotalActiveVotingPower.String()>>>", validatorSet.TotalActiveVotingPower.String())

	fmt.Println("fullProof>>>", hex.EncodeToString(proofData.Marshall()))

	err = svc.eths[0].CommitValsetHeader(t.Context(), header, extraData, proofData.Marshall(), []byte{})
	require.NoError(t, err)
}

func initValsetTestServices(t *testing.T) *valsetTestServices {
	t.Helper()

	base := new(big.Int)
	base.SetString("1000000000000000000", 10)

	n := 3

	pks := make([]*big.Int, 0, n)
	for i := 0; i < n; i++ {
		n := new(big.Int).Add(base, big.NewInt(int64(i)))
		pks = append(pks, n)
	}

	keyPairs := make([]bls.KeyPair, 0, len(pks))
	for _, pk := range pks {
		pkBytes := [32]byte{}
		pk.FillBytes(pkBytes[:])
		keyPairs = append(keyPairs, bls.ComputeKeyPair(pkBytes[:]))
	}

	eths := make([]*eth.Client, 0, len(keyPairs))
	for i, _ := range keyPairs {
		pk := new(big.Int).Add(base, big.NewInt(int64(i)))
		pkBytes := [32]byte{}
		pk.FillBytes(pkBytes[:])
		eth_, err := eth.NewEthClient(eth.Config{
			MasterRPCURL:   "http://localhost:8545",
			MasterAddress:  "0xF91E4B4166AD3eafDE95FeB6402560FCAb881690",
			PrivateKey:     pkBytes[:],
			RequestTimeout: time.Minute,
		})
		require.NoError(t, err)
		eths = append(eths, eth_)
	}

	phase, err := eths[0].GetCurrentPhase(t.Context())
	require.NoError(t, err)
	require.Equal(t, entity.COMMIT, phase)

	slog.InfoContext(t.Context(), "current phase", "phase", phase)

	derivers := make([]*valset.Deriver, 0, len(eths))
	for _, eth := range eths {
		deriver, err := valset.NewDeriver(eth)
		require.NoError(t, err)
		derivers = append(derivers, deriver)
	}

	generators := make([]*valset.Generator, 0, len(derivers))
	for i, deriver := range derivers {
		generator, err := valset.NewGenerator(deriver, eths[i])
		require.NoError(t, err)
		generators = append(generators, generator)
	}

	headers := make([]entity.ValidatorSetHeader, 0, len(generators))
	for _, generator := range generators {
		header, err := generator.GenerateValidatorSetHeaderOnCapture(t.Context())
		require.NoError(t, err)
		headers = append(headers, header)
	}

	extraDatas := make([][]entity.ExtraData, 0, len(generators))
	for i, generator := range generators {
		extraData_, err := generator.GenerateExtraData(t.Context(), headers[i], valset.ZkVerificationType)
		require.NoError(t, err)
		extraDatas = append(extraDatas, extraData_)
	}

	jsonHeader, err := headers[0].EncodeJSON()
	require.NoError(t, err)
	fmt.Println("header1>>>", string(jsonHeader))

	jsonExtraData, err := entity.ExtraDataList(extraDatas[0]).EncodeJSON()
	require.NoError(t, err)
	fmt.Println("extraData1>>>", string(jsonExtraData))

	headerHashes := make([][]byte, 0, len(generators))
	for i, generator := range generators {
		headerHash, err := generator.GenerateValidatorSetHeaderHash(headers[i], extraDatas[i])
		require.NoError(t, err)
		headerHashes = append(headerHashes, headerHash)
	}

	fmt.Println(hex.EncodeToString(headerHashes[0]))

	headerSignatures := make([]*bls.G1, 0, len(generators))
	for i, keyPair := range keyPairs {
		headerSignature, err := keyPair.Sign(headerHashes[i])
		require.NoError(t, err)
		headerSignatures = append(headerSignatures, headerSignature)
	}

	return &valsetTestServices{
		keyPairs:         keyPairs,
		eths:             eths,
		derivers:         derivers,
		generators:       generators,
		headerSignatures: headerSignatures,
		headerHashes:     headerHashes,
		extraDatas:       extraDatas,
		headers:          headers,
	}
}

func bytesFromPK(t *testing.T, pk1 string) []byte {
	t.Helper()
	b, ok := new(big.Int).SetString(pk1, 10)
	require.True(t, ok)
	return b.Bytes()
}

func waitCommitPhase(t *testing.T) {
	t.Helper()

	base := new(big.Int)
	base.SetString("1000000000000000000", 10)
	zeroPk := new(big.Int).Add(base, big.NewInt(int64(0)))
	pkBytes := [32]byte{}
	zeroPk.FillBytes(pkBytes[:])
	eth1, err := eth.NewEthClient(eth.Config{
		MasterRPCURL:   "http://localhost:8545",
		MasterAddress:  "0xF91E4B4166AD3eafDE95FeB6402560FCAb881690",
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
			if entity.COMMIT == phase {
				return
			}
			slog.InfoContext(t.Context(), "waiting for commit phase", "phase", phase)
		case <-t.Context().Done():
			return
		}
	}
}
