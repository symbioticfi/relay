package entity

import (
	"github.com/ethereum/go-ethereum/common"
)

// WantSignaturesRequest represents a request to resync signatures for a specific epoch.
// Contains missing validator indices for each incomplete signature request.
type WantSignaturesRequest struct {
	WantSignatures map[common.Hash]Bitmap // requestID -> missing validator indices bitmap
}

// WantSignaturesResponse contains signatures grouped by Request id.
// Each signature includes the validator index for consistent mapping.
type WantSignaturesResponse struct {
	Signatures map[common.Hash][]ValidatorSignature
}

// ValidatorSignature pairs a signature with its validator index in the active validator set.
// The validator index corresponds to the position in ValidatorSet.Validators.GetActiveValidators().
type ValidatorSignature struct {
	ValidatorIndex uint32            // Index in active validator set
	Signature      SignatureExtended // The actual signature data
}

// SignatureProcessingStats contains detailed statistics for processing received signatures
type SignatureProcessingStats struct {
	ProcessedCount            int // Successfully processed signatures
	UnrequestedSignatureCount int // Signatures for validators we didn't request
	UnrequestedHashCount      int // Signatures for hashes we didn't request
	SignatureRequestFailCount int // Failed to get signature request
	ProcessingFailCount       int // Failed to process signature
	AlreadyExistCount         int // Signature already exists (ErrEntityAlreadyExist)
}

// TotalErrors returns the total number of errors encountered
func (s SignatureProcessingStats) TotalErrors() int {
	return s.UnrequestedSignatureCount + s.UnrequestedHashCount + s.SignatureRequestFailCount +
		s.ProcessingFailCount + s.AlreadyExistCount
}

// WantAggregationProofsRequest represents a request to resync aggregation proofs for specific signature requests.
// Contains request ids for which aggregation proofs are needed.
type WantAggregationProofsRequest struct {
	RequestIDs []common.Hash // requestID list for missing aggregation proofs
}

// WantAggregationProofsResponse contains aggregation proofs grouped by request id.
// Each aggregation proof corresponds to a complete signature aggregation for a request.
type WantAggregationProofsResponse struct {
	Proofs map[common.Hash]AggregationProof // requestID -> aggregation proof
}

// AggregationProofProcessingStats contains detailed statistics for processing received aggregation proofs
type AggregationProofProcessingStats struct {
	ProcessedCount        int // Successfully processed aggregation proofs
	UnrequestedProofCount int // Proofs for hashes we didn't request
	VerificationFailCount int // Failed to verify aggregation proof
	ProcessingFailCount   int // Failed to process aggregation proof
	AlreadyExistCount     int // Aggregation proof already exists (ErrEntityAlreadyExist)
}

// TotalErrors returns the total number of errors encountered
func (s AggregationProofProcessingStats) TotalErrors() int {
	return s.UnrequestedProofCount + s.VerificationFailCount + s.ProcessingFailCount + s.AlreadyExistCount
}
