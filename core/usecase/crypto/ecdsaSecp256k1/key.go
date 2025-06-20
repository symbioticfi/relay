package ecdsaSecp256k1

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"
	"middleware-offchain/core/entity"
	symbKeys "middleware-offchain/core/usecase/crypto/key-types"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
)

type Signature = entity.RawSignature
type Message = []byte
type MessageHash = entity.RawMessageHash
type CompactPublicKey = entity.CompactPublicKey
type RawPublicKey = entity.RawPublicKey

const (
	RawKeyLength      int = 33
	MessageHashLength int = 32
)

type PublicKey struct {
	pubKey ecdsa.PublicKey
}

type PrivateKey struct {
	privateKey ecdsa.PrivateKey
}

func Hash(msg []byte) MessageHash {
	return crypto.Keccak256(msg)
}

func NewPrivateKey(b []byte) (*PrivateKey, error) {
	k, err := crypto.ToECDSA(b)
	if err != nil {
		return nil, errors.Errorf("ecdsaSecp256l1: failed to parse private key: %v", err)
	}
	return &PrivateKey{privateKey: *k}, nil
}

func GenerateKey() (*PrivateKey, error) {
	pk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, errors.Errorf("ecdsaSecp256l1: failed to generate key %w", err)
	}
	return &PrivateKey{privateKey: *pk}, nil
}

func (k *PrivateKey) Bytes() []byte {
	return k.privateKey.D.Bytes()
}

func (k *PrivateKey) Sign(msg []byte) (Signature, MessageHash, error) {
	// symbiotic using keccak256 for hashing in bls-bn254
	hash := Hash(msg)

	sig, err := crypto.Sign(hash, &k.privateKey)
	if err != nil {
		return nil, nil, errors.Errorf("ecdsaSecp256l1: failed to sign %w", err)
	}

	return sig, hash, nil
}

func (k *PrivateKey) PublicKey() symbKeys.PublicKey {
	pub := PublicKey{}
	pub.pubKey = k.privateKey.PublicKey
	return &pub
}

func NewPublicKey(x *big.Int, y *big.Int) *PublicKey {
	return &PublicKey{
		pubKey: ecdsa.PublicKey{
			X: x,
			Y: y,
		},
	}
}

func (k *PublicKey) Verify(msg Message, sig Signature) error {
	return k.VerifyWithHash(Hash(msg), sig)
}

func (k *PublicKey) VerifyWithHash(msgHash MessageHash, sig Signature) error {
	if len(msgHash) != MessageHashLength {
		return errors.Errorf("blsBn254: invalid message hash length")
	}

	ok := crypto.VerifySignature(crypto.FromECDSAPub(&k.pubKey), msgHash, sig)
	if !ok {
		return errors.Errorf("ecdsaSecp256l1: failed to verify signature %s", sig)
	}
	return nil
}

// OnChain might be one way operation, meaning that it's impossible to reconstruct PublicKey from compact
func (k *PublicKey) OnChain() CompactPublicKey {
	// returns eth address in this case
	return crypto.PubkeyToAddress(k.pubKey).Bytes()
}

func (k *PublicKey) Raw() RawPublicKey {
	// returns 33 bytes compressed pubKey
	return crypto.CompressPubkey(&k.pubKey)
}

func (k *PublicKey) MarshalText() (text []byte, err error) {
	return []byte(hexutil.Encode(k.Raw())), nil
}

func FromRaw(rawKey RawPublicKey) (*PublicKey, error) {
	if rawKey == nil {
		return nil, errors.New("blsBn254: nil raw key")
	}
	if len(rawKey) != RawKeyLength {
		return nil, fmt.Errorf("blsBn254: invalid raw key length, expected %d, got %d", RawKeyLength, len(rawKey))
	}
	pk, err := crypto.DecompressPubkey(rawKey)
	if err != nil {
		return nil, fmt.Errorf("blsBn254: failed to decompress public key %w", err)
	}
	return &PublicKey{pubKey: *pk}, nil
}

func FromPrivateKey(privateKey *PrivateKey) symbKeys.PublicKey {
	return privateKey.PublicKey()
}
