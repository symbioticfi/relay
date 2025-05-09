package entity

type P2PMessageType string

const (
	P2PMessageTypeSignatureHash P2PMessageType = "signature_hash_generated"
)

type SignatureHashMessage struct {
	MessageHash []byte
	Signature   []byte
	PublicKeyG1 []byte
	PublicKeyG2 []byte
	KeyTag      uint8
}

type SenderInfo struct {
	Type      P2PMessageType
	Sender    string
	Timestamp int64
}

type P2PSignatureHashMessage struct {
	Message SignatureHashMessage
	Info    SenderInfo
}
