package syncer

import (
	"context"

	"github.com/RoaringBitmap/roaring"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/core/usecase/crypto"
)

// ValidatorIndexMapping creates a mapping from validator public keys to their indices
// in the active validator set. Uses GetActiveValidators() for consistent ordering.
func ValidatorIndexMapping(validatorSet entity.ValidatorSet, keyTag entity.KeyTag) (map[string]uint32, error) {
	activeValidators := validatorSet.Validators.GetActiveValidators()
	mapping := make(map[string]uint32)

	for i, validator := range activeValidators {
		publicKey, found := validator.FindKeyByKeyTag(keyTag)
		if !found {
			// Skip validators without the required key tag
			continue
		}

		// Use public key bytes as the key for consistent mapping
		mapping[string(publicKey)] = uint32(i)
	}

	return mapping, nil
}

// BuildMissingSignatureBitmap creates a roaring bitmap of missing validator indices
// for a specific signature request. Returns bitmap with bits set for validators
// that haven't provided signatures yet.
func BuildMissingSignatureBitmap(
	ctx context.Context,
	repo repo,
	validatorSet entity.ValidatorSet,
	signatureRequest entity.SignatureRequest,
) (*roaring.Bitmap, error) {
	// Get validator index mapping for this key tag
	validatorIndexMapping, err := ValidatorIndexMapping(validatorSet, signatureRequest.KeyTag)
	if err != nil {
		return nil, errors.Errorf("failed to create validator index mapping: %w", err)
	}

	// Get all existing signatures for this request
	signatures, err := repo.GetAllSignatures(ctx, signatureRequest.Hash())
	if err != nil {
		return nil, errors.Errorf("failed to get signatures for request %s: %w", signatureRequest.Hash().Hex(), err)
	}

	// Create bitmap with all active validator indices
	missingBitmap := roaring.New()
	for _, validatorIndex := range validatorIndexMapping {
		missingBitmap.Add(validatorIndex)
	}

	// Remove indices for validators who have already provided signatures
	for _, signature := range signatures {
		// Convert signature public key to the format used in mapping
		pk, err := crypto.NewPublicKey(signatureRequest.KeyTag.Type(), signature.PublicKey)
		if err != nil {
			// Skip invalid public keys rather than failing the whole operation
			continue
		}

		onChainKey := pk.OnChain()
		if validatorIndex, found := validatorIndexMapping[string(onChainKey)]; found {
			missingBitmap.Remove(validatorIndex)
		}
	}

	return missingBitmap, nil
}

// FilterRequestsByQuorum filters signature requests that haven't reached quorum threshold.
// Uses the same quorum logic as aggregator-app: currentVotingPower.Cmp(quorumThreshold) >= 0
func FilterRequestsByQuorum(
	ctx context.Context,
	repo repo,
	validatorSet entity.ValidatorSet,
	signatureRequests []entity.SignatureRequest,
) ([]MissingSignatureInfo, error) {
	var incompleteRequests []MissingSignatureInfo

	for _, req := range signatureRequests {
		// Get all signatures for this request
		signatures, err := repo.GetAllSignatures(ctx, req.Hash())
		if err != nil {
			return nil, errors.Errorf("failed to get signatures for request %s: %w", req.Hash().Hex(), err)
		}

		// Extract public keys and find corresponding validators
		publicKeys, err := extractPublicKeys(req.KeyTag, signatures)
		if err != nil {
			return nil, errors.Errorf("failed to extract public keys for request %s: %w", req.Hash().Hex(), err)
		}

		validators := validatorSet.FindValidatorsBySignatures(req.KeyTag, publicKeys)
		currentVotingPower := validators.GetTotalActiveVotingPower()

		// Check if quorum is reached (same logic as aggregator-app)
		quorumReached := currentVotingPower.Cmp(validatorSet.QuorumThreshold.Int) >= 0
		if quorumReached {
			// Skip requests that have already reached quorum
			continue
		}

		// Build missing signature bitmap for incomplete requests
		missingBitmap, err := BuildMissingSignatureBitmap(ctx, repo, validatorSet, req)
		if err != nil {
			return nil, errors.Errorf("failed to build missing signature bitmap for request %s: %w", req.Hash().Hex(), err)
		}

		incompleteRequests = append(incompleteRequests, MissingSignatureInfo{
			RequestHash:             req.Hash(),
			SignatureRequest:        req,
			MissingValidatorsBitmap: missingBitmap,
			CurrentVotingPower:      currentVotingPower,
			RequiredVotingPower:     validatorSet.QuorumThreshold,
		})
	}

	return incompleteRequests, nil
}

// extractPublicKeys converts signatures to compact public keys, reusing aggregator-app logic
func extractPublicKeys(keyTag entity.KeyTag, signatures []entity.SignatureExtended) ([]entity.CompactPublicKey, error) {
	publicKeys := make([]entity.CompactPublicKey, 0, len(signatures))
	for _, signature := range signatures {
		pk, err := crypto.NewPublicKey(keyTag.Type(), signature.PublicKey)
		if err != nil {
			return nil, errors.Errorf("failed to get public key: %w", err)
		}
		publicKeys = append(publicKeys, pk.OnChain())
	}
	return publicKeys, nil
}
