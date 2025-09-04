package syncer

import (
	"github.com/RoaringBitmap/roaring/v2"
	"github.com/ethereum/go-ethereum/common"

	"github.com/symbioticfi/relay/core/entity"
)

// WantSignaturesRequest represents a request to resync signatures for a specific epoch.
// Contains missing validator indices for each incomplete signature request.
type WantSignaturesRequest struct {
	Epoch          entity.Epoch                    // Target epoch for resync
	WantSignatures map[common.Hash]*roaring.Bitmap // reqHash -> missing validator indices bitmap
}

// WantSignatureResponse contains signatures grouped by request hash.
// Each signature includes the validator index for consistent mapping.
type WantSignatureResponse struct {
	Epoch      entity.Epoch                         // Response epoch
	Signatures map[common.Hash][]ValidatorSignature // grouped by request hash
}

// ValidatorSignature pairs a signature with its validator index in the active validator set.
// The validator index corresponds to the position in ValidatorSet.Validators.GetActiveValidators().
type ValidatorSignature struct {
	ValidatorIndex uint32                   // Index in active validator set
	Signature      entity.SignatureExtended // The actual signature data
}
