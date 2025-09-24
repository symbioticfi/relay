package entity

import (
	"github.com/ethereum/go-ethereum/common"
)

// WantSignaturesRequest represents a request to resync signatures for a specific epoch.
// Contains missing validator indices for each incomplete signature request.
type WantSignaturesRequest struct {
	WantSignatures map[common.Hash]Bitmap // signatureTargetID -> missing validator indices bitmap
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
	ProcessingErrorCount       int // Failed to process signature
	AlreadyExistCount          int // Signature already exists (ErrEntityAlreadyExist)
}

// TotalErrors returns the total number of errors encountered
func (s SignatureProcessingStats) TotalErrors() int {
	return s.UnrequestedSignatureCount + s.UnrequestedHashCount + s.SignatureRequestErrorCount +
		s.ProcessingErrorCount + s.AlreadyExistCount
}

// WantAggregationProofsRequest represents a request to resync aggregation proofs for specific signature requests.
// Contains request hashes for which aggregation proofs are needed.
type WantAggregationProofsRequest struct {
	SignatureTargetIDs []common.Hash // signatureTargetID list for missing aggregation proofs
}

// WantAggregationProofsResponse contains aggregation proofs grouped by request hash.
// Each aggregation proof corresponds to a complete signature aggregation for a request.
type WantAggregationProofsResponse struct {
	Proofs map[common.Hash]AggregationProof // signatureTargetID -> aggregation proof
}

// AggregationProofProcessingStats contains detailed statistics for processing received aggregation proofs
type AggregationProofProcessingStats struct {
	ProcessedCount         int // Successfully processed aggregation proofs
	UnrequestedProofCount  int // Proofs for hashes we didn't request
	VerificationErrorCount int // Failed to verify aggregation proof
	ProcessingErrorCount   int // Failed to process aggregation proof
	AlreadyExistCount      int // Aggregation proof already exists (ErrEntityAlreadyExist)
}

// TotalErrors returns the total number of errors encountered
func (s AggregationProofProcessingStats) TotalErrors() int {
	return s.UnrequestedProofCount + s.VerificationErrorCount + s.ProcessingErrorCount + s.AlreadyExistCount
}
