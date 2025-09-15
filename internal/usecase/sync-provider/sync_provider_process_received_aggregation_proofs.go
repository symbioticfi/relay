package sync_provider

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
)

// ProcessReceivedAggregationProofs processes aggregation proofs received from peers
func (s *Syncer) ProcessReceivedAggregationProofs(ctx context.Context, response entity.WantAggregationProofsResponse) (entity.AggregationProofProcessingStats, error) {
	stats := entity.AggregationProofProcessingStats{}

	// Process each received aggregation proof
	for reqHash, proof := range response.Proofs {
		s.processSingleAggregationProof(ctx, reqHash, proof, &stats)
	}

	return stats, nil
}

// processSingleAggregationProof processes a single aggregation proof
func (s *Syncer) processSingleAggregationProof(ctx context.Context, reqHash common.Hash, proof entity.AggregationProof, stats *entity.AggregationProofProcessingStats) {
	// Get the signature request to verify the aggregation proof belongs to a valid request
	signatureRequest, err := s.cfg.Repo.GetSignatureRequest(ctx, reqHash)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			// Signature request not found - this might be an unrequested proof
			stats.UnrequestedProofCount++
			return
		}
		stats.SignatureRequestErrorCount++
		return // Continue processing other proofs
	}

	// Create aggregated signature message for processing
	msg := entity.AggregatedSignatureMessage{
		RequestHash:      reqHash,
		KeyTag:           signatureRequest.KeyTag,
		Epoch:            signatureRequest.RequiredEpoch,
		AggregationProof: proof,
	}

	// Process the aggregation proof using the signature processor
	if err := s.cfg.SignatureProcessor.ProcessAggregationProof(ctx, msg); err != nil {
		if errors.Is(err, entity.ErrEntityAlreadyExist) {
			stats.AlreadyExistCount++
			return
		}
		stats.ProcessingErrorCount++
		return // Continue processing other proofs
	}

	stats.ProcessedCount++
}
