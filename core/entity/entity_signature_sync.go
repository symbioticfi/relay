package entity

import (
	"github.com/ethereum/go-ethereum/common"
)

// WantSignaturesRequest represents a request to resync signatures for a specific epoch.
// Contains missing validator indices for each incomplete signature request.
type WantSignaturesRequest struct {
	WantSignatures map[common.Hash]SignatureBitmap // reqHash -> missing validator indices bitmap
}

// WantSignaturesResponse contains signatures grouped by request hash.
// Each signature includes the validator index for consistent mapping.
type WantSignaturesResponse struct {
	Signatures map[common.Hash][]ValidatorSignature // grouped by request hash
}

// ValidatorSignature pairs a signature with its validator index in the active validator set.
// The validator index corresponds to the position in ValidatorSet.Validators.GetActiveValidators().
type ValidatorSignature struct {
	ValidatorIndex uint32            // Index in active validator set
	Signature      SignatureExtended // The actual signature data
}
