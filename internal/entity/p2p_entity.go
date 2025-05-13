package entity

import (
	"middleware-offchain/pkg/bls"
)

type P2PMessageType string

const (
	P2PMessageTypeSignatureHash        P2PMessageType = "signature_hash_generated"
	P2PMessageTypeSignaturesAggregated P2PMessageType = "signatures_aggregated"
)

type SignatureHashMessage struct {
	MessageHash []byte
	Signature   []byte
	PublicKeyG1 []byte
	PublicKeyG2 []byte
	KeyTag      uint8
}

type SignaturesAggregatedMessage struct {
	PublicKeyG1 *bls.G1
	Proof       []byte
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

type P2PSignaturesAggregatedMessage struct {
	Message SignaturesAggregatedMessage
	Info    SenderInfo
}
