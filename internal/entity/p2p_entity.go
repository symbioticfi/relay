package entity

import (
	"middleware-offchain/pkg/bls"
)

type P2PMessageType string

const (
	P2PMessageTypeSignatureHash        P2PMessageType = "signature_hash_generated"
	P2PMessageTypeSignaturesAggregated P2PMessageType = "signatures_aggregated"
)

type HashType string

const (
	HashTypeValsetHeader HashType = "valset_header"
	HashTypeMessage      HashType = "message"
)

type SignatureHashMessage struct {
	Request   SignatureRequest
	Signature []byte
	PublicKey []byte
	HashType  HashType
}

type SignaturesAggregatedMessage struct {
	Request     SignatureRequest
	Proof       AggregationProof
	PublicKeyG1 *bls.G1
	HashType    HashType
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
