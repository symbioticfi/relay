package syncer

import (
	"context"

	"github.com/RoaringBitmap/roaring"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

// BuildResyncRequest creates a ResyncSignatureRequest for incomplete signature requests
// in a specific epoch. Only includes requests that haven't reached quorum threshold.
func BuildResyncRequest(
	ctx context.Context,
	repo repo,
	epoch entity.Epoch,
) (*ResyncSignatureRequest, error) {
	// Get validator set for this epoch to enable consistent indexing
	validatorSet, err := repo.GetValidatorSetByEpoch(ctx, uint64(epoch))
	if err != nil {
		return nil, errors.Errorf("failed to get validator set for epoch %d: %w", epoch, err)
	}

	// Get all signature requests for this epoch (paginated)
	signatureRequests, err := getAllSignatureRequestsForEpoch(ctx, repo, epoch)
	if err != nil {
		return nil, errors.Errorf("failed to get signature requests for epoch %d: %w", epoch, err)
	}

	// Filter requests that haven't reached quorum threshold
	incompleteRequests, err := FilterRequestsByQuorum(ctx, repo, validatorSet, signatureRequests)
	if err != nil {
		return nil, errors.Errorf("failed to filter incomplete requests for epoch %d: %w", epoch, err)
	}

	// Build missing signature bitmaps for each incomplete request
	requestSignatures := make(map[common.Hash]*roaring.Bitmap)
	for _, incompleteReq := range incompleteRequests {
		requestSignatures[incompleteReq.RequestHash] = incompleteReq.MissingValidatorsBitmap
	}

	return &ResyncSignatureRequest{
		Epoch:             epoch,
		RequestSignatures: requestSignatures,
	}, nil
}

// getAllSignatureRequestsForEpoch retrieves all signature requests for a specific epoch
// using pagination to handle large numbers of requests efficiently.
func getAllSignatureRequestsForEpoch(
	ctx context.Context,
	repo repo,
	epoch entity.Epoch,
) ([]entity.SignatureRequest, error) {
	const pageSize = 100 // Reasonable page size for memory efficiency

	var allRequests []entity.SignatureRequest
	var lastHash common.Hash

	for {
		// Get next page of signature requests
		requests, err := repo.GetSignatureRequestsByEpoch(ctx, epoch, pageSize, lastHash)
		if err != nil {
			return nil, errors.Errorf("failed to get signature requests page: %w", err)
		}

		// If no more requests, we're done
		if len(requests) == 0 {
			break
		}

		allRequests = append(allRequests, requests...)

		// If we got less than pageSize, we've reached the end
		if len(requests) < pageSize {
			break
		}

		// Update lastHash for next page
		lastHash = requests[len(requests)-1].Hash()
	}

	return allRequests, nil
}
