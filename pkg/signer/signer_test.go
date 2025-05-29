package signer

import (
	"bytes"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/crypto"
	"middleware-offchain/pkg/bls"
	"testing"
)

type MockKeyProvider struct {
}

func (*MockKeyProvider) GetPrivateKey(keyTag uint8) ([]byte, error) {
	return []byte("testrandomkey"), nil
}

func (*MockKeyProvider) HasKey(keyTag uint8) (bool, error) {
	return true, nil
}

func TestBLSBn254(t *testing.T) {
	signer := NewSigner(&MockKeyProvider{})

	msg := []byte("message")
	signature, err := signer.Sign(15, msg)
	if err != nil {
		t.Fatalf("signer returned an error: %v", err)
	}

	keyPair := bls.ComputeKeyPair([]byte("testrandomkey"))

	g1Sig := bls.G1{
		G1Affine: new(bn254.G1Affine),
	}
	_, err = g1Sig.SetBytes(signature.Signature)
	if err != nil {
		t.Fatalf("SetBytes returned an error: %v", err)
	}
	result, err := keyPair.PublicKeyG2.Verify(&g1Sig, crypto.Keccak256(msg))
	if err != nil {
		t.Fatalf("Verifying returned an error: %v", err)
	}
	if !result {
		t.Fatalf("Verifying returned false")
	}

	if !bytes.Equal(signature.MessageHash, crypto.Keccak256(msg)) {
		t.Fatalf("Signature returned wrong message")
	}

	if len(signature.PublicKey) != 96 {
		t.Fatalf("Public key length is incorrect")
	}

	g1Pubkey := bls.G1{
		G1Affine: new(bn254.G1Affine),
	}
	g2Pubkey := bls.G2{
		G2Affine: new(bn254.G2Affine),
	}
	g1Pubkey.SetBytes(signature.PublicKey[:32])
	g2Pubkey.SetBytes(signature.PublicKey[32:])

	if !keyPair.PublicKeyG1.Equal(g1Pubkey.G1Affine) {
		t.Fatalf("PublicKeyG1 returned wrong public key")
	}
	if !keyPair.PublicKeyG2.Equal(g2Pubkey.G2Affine) {
		t.Fatalf("PublicKeyG2 returned wrong public key")
	}
}
