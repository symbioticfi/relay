package entity

type P2PMessageType string

const (
	SignatureGenerated P2PMessageType = "valset_generated"
)

// P2PMessage is the basic unit of communication between peers
type P2PMessage struct {
	Type      P2PMessageType `json:"type"`
	Sender    string         `json:"sender"`
	Timestamp int64          `json:"timestamp"`
	Data      []byte         `json:"data"`
}

type SignatureMessage struct {
	MessageHash []byte
	Signature   []byte
	PublicKeyG1 []byte
	PublicKeyG2 []byte
	KeyTag      uint64
}
