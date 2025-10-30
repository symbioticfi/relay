package entity

import (
	"github.com/ethereum/go-ethereum/common"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// SignatureRequestWithID represents a signature request with its request ID
type SignatureRequestWithID struct {
	RequestID        common.Hash
	SignatureRequest symbiotic.SignatureRequest
}
