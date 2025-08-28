package syncer

import (
	"github.com/RoaringBitmap/roaring"
	"github.com/ethereum/go-ethereum/common"

	"github.com/symbioticfi/relay/core/entity"
)

// ResyncSignatureRequest represents a request to resync signatures for a specific epoch.
// Contains missing validator indices for each incomplete signature request.
type ResyncSignatureRequest struct {
	Epoch             entity.Epoch                    // Target epoch for resync
	RequestSignatures map[common.Hash]*roaring.Bitmap // reqHash -> missing validator indices bitmap
}

// ResyncSignatureResponse contains signatures grouped by request hash.
// Each signature includes the validator index for consistent mapping.
type ResyncSignatureResponse struct {
	Epoch      entity.Epoch                         // Response epoch
	Signatures map[common.Hash][]ValidatorSignature // grouped by request hash
}

// ValidatorSignature pairs a signature with its validator index in the active validator set.
// The validator index corresponds to the position in ValidatorSet.Validators.GetActiveValidators().
type ValidatorSignature struct {
	ValidatorIndex uint32                   // Index in active validator set
	Signature      entity.SignatureExtended // The actual signature data
}

// MissingSignatureInfo tracks which validator signatures are missing for a signature request.
// Used internally for building resync requests and processing responses.
type MissingSignatureInfo struct {
	RequestHash             common.Hash             // Hash of the signature request
	SignatureRequest        entity.SignatureRequest // The original request
	MissingValidatorsBitmap *roaring.Bitmap         // Bitmap of missing validator indices
	CurrentVotingPower      entity.VotingPower      // Current accumulated voting power
	RequiredVotingPower     entity.VotingPower      // Required voting power for quorum
}
