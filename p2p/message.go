package p2p

import (
	"encoding/json"
	"math/big"
)

// Message types
const (
	TypeSignatureRequest = "signature"
)

// Message is the basic unit of communication between peers
type Message struct {
	Type      string          `json:"type"`
	Sender    string          `json:"sender"`
	Timestamp int64           `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// SignatureMessage contains a peer's signature for a message
type SignatureMessage struct {
	Epoch       *big.Int `json:"epoch"`
	MessageHash string   `json:"message_hash"`
	Signature   []byte   `json:"signature"`
	PublicKey   []byte   `json:"public_key"`
}
