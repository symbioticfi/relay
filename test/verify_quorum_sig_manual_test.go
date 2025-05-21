package test

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"middleware-offchain/internal/client/eth"
	"middleware-offchain/internal/client/valset"
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
		MasterAddress:  "0x5081a39b8A5f0E35a8D959395a630b68B74Dd30f",
		PrivateKey:     bytesFromPK(t, pk1),
		RequestTimeout: time.Minute,
	})
	require.NoError(t, err)
	deriver1, err := valset.NewDeriver(eth1)
	require.NoError(t, err)

	validatorSet, err := deriver1.GetValidatorSet(t.Context(), new(big.Int).SetInt64(time.Now().Unix()))
	require.NoError(t, err)
	_ = validatorSet

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

	ok, err := eth1.VerifyQuorumSig(t.Context(), messageHash, 15, new(big.Int).SetInt64(1e18), proofData.Marshall())
	require.NoError(t, err)
	require.True(t, ok)
}
