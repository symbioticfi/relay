package key_types

import "github.com/symbioticfi/relay/core/entity"

type Signature = entity.RawSignature
type Message = []byte
type MessageHash = entity.RawMessageHash
type CompactPublicKey = entity.CompactPublicKey
type RawPublicKey = entity.RawPublicKey

type PublicKey interface {
	Verify(msg Message, sig Signature) error
	VerifyWithHash(msgHash MessageHash, sig Signature) error
	OnChain() CompactPublicKey
	Raw() RawPublicKey
	MarshalText() (text []byte, err error)
}

type PrivateKey interface {
	Bytes() []byte
	Sign(msg []byte) (Signature, MessageHash, error)
	PublicKey() PublicKey
}
