package entity

import (
	"middleware-offchain/core/entity"
)

type SenderInfo struct {
	// Sender is a p2p peer id
	Sender string
}

type P2PSignatureMessage struct {
	SenderInfo SenderInfo
	Message    entity.SignatureMessage
}

type P2PAggregatedSignatureMessage struct {
	SenderInfo SenderInfo
	Message    entity.AggregatedSignatureMessage
}
