package entity

type P2PMessageType string

const (
	P2PMessageTypeSignatureHash        P2PMessageType = "signature_hash_generated"
	P2PMessageTypeSignaturesAggregated P2PMessageType = "signatures_aggregated"
)

type SignatureHashMessage struct {
	Request   SignatureRequest
	Signature []byte
	PublicKey []byte
}

type SenderInfo struct {
	Type      P2PMessageType
	Sender    string
	Timestamp int64
}

type P2PSignatureHashMessage struct {
	Message SignatureMessage
	Info    SenderInfo
}

type P2PSignaturesAggregatedMessage struct {
	Message AggregatedSignatureMessage
	Info    SenderInfo
}
