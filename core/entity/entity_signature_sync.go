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

// SignatureProcessingStats contains detailed statistics for processing received signatures
type SignatureProcessingStats struct {
	ProcessedCount             int // Successfully processed signatures
	UnrequestedSignatureCount  int // Signatures for validators we didn't request
	UnrequestedHashCount       int // Signatures for hashes we didn't request
	SignatureRequestErrorCount int // Failed to get signature request
	PublicKeyErrorCount        int // Failed to create public key from signature
	ValidatorInfoErrorCount    int // Failed to get validator info
	ProcessingErrorCount       int // Failed to process signature
	AlreadyExistCount          int // Signature already exists (ErrEntityAlreadyExist)
}

// TotalErrors returns the total number of errors encountered
func (s SignatureProcessingStats) TotalErrors() int {
	return s.UnrequestedSignatureCount + s.UnrequestedHashCount + s.SignatureRequestErrorCount +
		s.PublicKeyErrorCount + s.ValidatorInfoErrorCount + s.ProcessingErrorCount + s.AlreadyExistCount
}
