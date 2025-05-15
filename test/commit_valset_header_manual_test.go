//go:build manual

package test

import (
	"log/slog"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"middleware-offchain/internal/client/eth"
	"middleware-offchain/internal/client/valset"
	"middleware-offchain/internal/entity"
	"middleware-offchain/pkg/bls"
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
	svc := initValsetTestServices(t)

	//validatorSet, err := svc.deriver1.GetValidatorSet(t.Context(), new(big.Int).SetInt64(time.Now().Unix()))
	//require.NoError(t, err)

	//proofData, err := proof.DoProve(validatorSet.Validators, 15)
	//require.NoError(t, err)

	header, err := svc.generator1.GenerateValidatorSetHeader(t.Context())
	require.NoError(t, err)

	err = svc.eth1.CommitValsetHeader(t.Context(), header, []byte{1, 2, 3})
	require.NoError(t, err)
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
