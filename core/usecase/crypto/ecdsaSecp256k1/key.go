package ecdsaSecp256k1

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/symbioticfi/relay/core/entity"
	symbKeys "github.com/symbioticfi/relay/core/usecase/crypto/key-types"

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
	kInt := big.NewInt(0).SetBytes(b)
	if kInt.BitLen() > 32*8 { // 32 bytes = 256 bits
		return nil, errors.Errorf("ecdsaSecp256k1: private key too long, expected 32 bytes, got %d bytes", len(b))
	}
	fixedLengthBytes := kInt.FillBytes(make([]byte, 32))

	k, err := crypto.ToECDSA(fixedLengthBytes)
	if err != nil {
		return nil, errors.Errorf("ecdsaSecp256k1: failed to parse private key: %v", err)
	}
	return &PrivateKey{privateKey: *k}, nil
}

func GenerateKey() (*PrivateKey, error) {
	pk, err := crypto.GenerateKey()
	if err != nil {
		return nil, errors.Errorf("ecdsaSecp256k1: failed to generate key %w", err)
	}
	return &PrivateKey{privateKey: *pk}, nil
}

func (k *PrivateKey) Bytes() []byte {
	return k.privateKey.D.FillBytes(make([]byte, 32))
}

func (k *PrivateKey) Sign(msg []byte) (Signature, MessageHash, error) {
	// symbiotic using keccak256 for hashing in ecdsaSecp256k1
	hash := Hash(msg)

	sig, err := crypto.Sign(hash, &k.privateKey)
	if err != nil {
		return nil, nil, errors.Errorf("ecdsaSecp256k1: failed to sign %w", err)
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
		return errors.Errorf("ecdsaSecp256k1: invalid message hash length")
	}
	if len(sig) != 65 {
		return errors.Errorf("ecdsaSecp256k1: invalid signature length, expected 65 bytes, got %d", len(sig))
	}

	// Remove recovery ID from signature for verification
	sigWithoutRecoveryId := sig[:64]
	ok := crypto.VerifySignature(crypto.FromECDSAPub(&k.pubKey), msgHash, sigWithoutRecoveryId)
	if !ok {
		sigStr, _ := sig.MarshalText()
		return errors.Errorf("ecdsaSecp256k1: failed to verify signature %s", sigStr)
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
		return nil, errors.New("ecdsaSecp256k1: nil raw key")
	}
	if len(rawKey) != RawKeyLength {
		return nil, errors.Errorf("ecdsaSecp256k1: invalid raw key length, expected %d, got %d", RawKeyLength, len(rawKey))
	}

	pk, err := crypto.DecompressPubkey(rawKey)
	if err != nil {
		return nil, errors.Errorf("ecdsaSecp256k1: failed to decompress public key %w", err)
	}

	return &PublicKey{pubKey: *pk}, nil
}

func FromPrivateKey(privateKey *PrivateKey) symbKeys.PublicKey {
	return privateKey.PublicKey()
}
