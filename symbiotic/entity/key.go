package entity

type Message = []byte
type MessageHash = RawMessageHash

type PublicKey interface {
	Verify(msg Message, sig RawSignature) error
	VerifyWithHash(msgHash MessageHash, sig RawSignature) error
	OnChain() CompactPublicKey
	Raw() RawPublicKey
	MarshalText() (text []byte, err error)
}

type PrivateKey interface {
	Bytes() []byte
	Sign(msg []byte) (RawSignature, MessageHash, error)
	PublicKey() PublicKey
}
