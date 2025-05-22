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

	headerHash1 []byte
	headerHash2 []byte
	headerHash3 []byte
}

func TestCommitValsetHeader(t *testing.T) {
	waitCommitPhase(t)

	svc := initValsetTestServices(t)

	validatorSet, err := svc.deriver1.GetValidatorSet(t.Context(), new(big.Int).SetInt64(time.Now().Unix()))
	require.NoError(t, err)
	_ = validatorSet

	if !bytes.Equal(svc.headerHash1, svc.headerHash2) || !bytes.Equal(svc.headerHash1, svc.headerHash3) {
		t.Fatal("headerHash1 and headerHash2 are not equal")
	}

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

	proofData, err := proof.DoProve(proof.RawProveInput{
		AllValidators:    validatorSet.Validators,
		SignerValidators: validatorSet.Validators,
		RequiredKeyTag:   15,
		Message:          svc.headerHash1,
		Signature:        *aggSignature,
		SignersAggKeyG2:  *aggPublicKeyG2,
	})
	require.NoError(t, err)

	//proofData := decodeHex(t, "01c70454a912bf226d9b0a7b38dfef9319f92b893115fd5b168f0061c56a11e30d8c66ed7585aafd81e6c20cdbe81d385ee13871ef1da2b041e218076fbfd88e0cf9e7f5e7b25f241973a4a4ae6a7f29d430af5c243cd254d5035e7ad1883d9d1f0c0f76011867ebb6115185f5b9fe538de1181a39cd9e5efa03046b031c64df0b86f9a8fcb28738e82eabe0237a57bc47a02158841039f0d12ec3abb3d9ee2d12185fed2304764a978ba5405c684093479e18d934c7c8ba9e031981c836d4ff028f842b327dd18be5ba410bc423ce989f6807f1766acdae5669dea546d8cd591f98ba029d5e2a77520b2639234354c2e3983ce9590efbee7b293a8ee32bfbf80000000124997c0ef7b3e53580aaa97c84ae4682a7a7ec617110c5790ce06ca6bf837600114ec9b6c4503e96f11bcdb0e4601fecd83b5b8e4c7d9df6204aea2a4b7617471e9bada9e6dd91bf84b89967925bf1a90aa162f5f2883c4713522263f983f5ec101464ab309ff2a609396c898689eb0e5e4f703d350adc6ed69d6dfdc1a5bbbd")

	header, err := svc.generator1.GenerateValidatorSetHeaderOnCapture(t.Context())
	require.NoError(t, err)

	fmt.Println("G2>>>", hex.EncodeToString(aggPublicKeyG2.Marshal()))
	fmt.Println("aggSigG1>>>", hex.EncodeToString(aggSignature.Marshal()))
	fmt.Println("proof_>>>", hex.EncodeToString(proofData.Proof))
	fmt.Println("Commitments>>>", hex.EncodeToString(proofData.Commitments))
	fmt.Println("commitmentPok>>>", hex.EncodeToString(proofData.CommitmentPok))
	fmt.Println("validatorSet.TotalActiveVotingPower.String()>>>", validatorSet.TotalActiveVotingPower.String())

	fmt.Println("fullProof>>>", hex.EncodeToString(proofData.Marshall()))

	err = svc.eth1.CommitValsetHeader(t.Context(), header, proofData.Marshall())
	require.NoError(t, err)
}

func initValsetTestServices(t *testing.T) *valsetTestServices {
	t.Helper()
	pk1 := "87191036493798670866484781455694320176667203290824056510541300741498740913410"
	pk2 := "26972876870930381973856869753776124637336739336929668162870464864826929175089"
	pk3 := "11008377096554045051122023680185802911050337017631086444859313200352654461863"

	keyPair1 := bls.ComputeKeyPair(bytesFromPK(t, pk1))
	keyPair2 := bls.ComputeKeyPair(bytesFromPK(t, pk2))
	keyPair3 := bls.ComputeKeyPair(bytesFromPK(t, pk3))

	eth1, err := eth.NewEthClient(eth.Config{
		MasterRPCURL:   "http://localhost:8545",
		MasterAddress:  "0x04C89607413713Ec9775E14b954286519d836FEf",
		PrivateKey:     bytesFromPK(t, pk1),
		RequestTimeout: time.Minute,
	})
	require.NoError(t, err)

	eth2, err := eth.NewEthClient(eth.Config{
		MasterRPCURL:   "http://localhost:8545",
		MasterAddress:  "0x04C89607413713Ec9775E14b954286519d836FEf",
		PrivateKey:     bytesFromPK(t, pk2),
		RequestTimeout: time.Minute,
	})
	require.NoError(t, err)

	eth3, err := eth.NewEthClient(eth.Config{
		MasterRPCURL:   "http://localhost:8545",
		MasterAddress:  "0x04C89607413713Ec9775E14b954286519d836FEf",
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

	header1, err := generator1.GenerateValidatorSetHeaderOnCapture(t.Context())
	require.NoError(t, err)

	fmt.Println("header1>>>", hex.EncodeToString(header1.ExtraData))

	headerHash1, err := generator1.GenerateValidatorSetHeaderHash(header1)
	require.NoError(t, err)

	fmt.Println(hex.EncodeToString(headerHash1))

	headerSignature1, err := keyPair1.Sign(headerHash1)
	require.NoError(t, err)

	header2, err := generator2.GenerateValidatorSetHeaderOnCapture(t.Context())
	require.NoError(t, err)

	headerHash2, err := generator2.GenerateValidatorSetHeaderHash(header2)
	require.NoError(t, err)

	headerSignature2, err := keyPair2.Sign(headerHash2)
	require.NoError(t, err)

	header3, err := generator3.GenerateValidatorSetHeaderOnCapture(t.Context())
	require.NoError(t, err)

	headerHash3, err := generator3.GenerateValidatorSetHeaderHash(header3)
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
		headerHash1:      headerHash1,
		headerHash2:      headerHash2,
		headerHash3:      headerHash3,
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
	eth1, err := eth.NewEthClient(eth.Config{
		MasterRPCURL:   "http://localhost:8545",
		MasterAddress:  "0x04C89607413713Ec9775E14b954286519d836FEf",
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
