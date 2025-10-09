package entity

type SenderInfo struct {
	// Sender is a p2p peer id
	Sender    string
	PublicKey []byte
}

// P2PMessage is a generic message structure for P2P communication, containing SenderInfo and a message of type T.
type P2PMessage[T any] struct {
	SenderInfo   SenderInfo
	Message      T
	TraceContext map[string]string `json:"trace_context,omitempty"`
}
