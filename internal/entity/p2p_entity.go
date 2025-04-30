package entity

import (
	"encoding/json"
)

// P2PMessage is the basic unit of communication between peers
type P2PMessage struct {
	Type      string          `json:"type"`
	Sender    string          `json:"sender"`
	Timestamp int64           `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}
