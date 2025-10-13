package key_types

import (
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

type Signature = symbiotic.RawSignature
type Message = []byte
type MessageHash = symbiotic.RawMessageHash
type CompactPublicKey = symbiotic.CompactPublicKey
type RawPublicKey = symbiotic.RawPublicKey

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
