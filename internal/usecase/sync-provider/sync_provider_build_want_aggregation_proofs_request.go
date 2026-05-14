package sync_provider

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"go.opentelemetry.io/otel/attribute"

	"github.com/symbioticfi/relay/internal/entity"
	"github.com/symbioticfi/relay/pkg/tracing"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// BuildWantAggregationProofsRequest builds a request for missing aggregation proofs from recent epochs
func (s *Syncer) BuildWantAggregationProofsRequest(ctx context.Context) (entity.WantAggregationProofsRequest, error) {
	ctx, span := tracing.StartSpan(ctx, "sync-provider.BuildWantAggregationProofsRequest")
	defer span.End()

	latestEpoch, err := s.cfg.Repo.GetLatestValidatorSetEpoch(ctx)
	if err != nil {
		tracing.RecordError(span, err)
		return entity.WantAggregationProofsRequest{}, errors.Errorf("failed to get latest epoch: %w", err)
	}

	startEpoch := symbiotic.Epoch(0)
	if latestEpoch >= symbiotic.Epoch(s.cfg.EpochsToSync) {
		startEpoch = latestEpoch - symbiotic.Epoch(s.cfg.EpochsToSync)
	}

	tracing.SetAttributes(span,
		tracing.AttrEpoch.Int64(int64(latestEpoch)),
		attribute.Int64("start_epoch", int64(startEpoch)),
	)

	var allRequestIDs []common.Hash
	totalRequests := 0

	// Iterate through epochs from newest to oldest to prioritize recent requests
	for epoch := latestEpoch; epoch >= startEpoch && totalRequests < s.cfg.MaxAggProofRequestsPerSync; epoch-- {
		var lastHash common.Hash
		remaining := s.cfg.MaxAggProofRequestsPerSync - totalRequests

		// Paginate through signature requests without aggregation proofs for this epoch
		for remaining > 0 {
			requests, err := s.cfg.Repo.GetSignatureRequestsWithoutAggregationProof(ctx, epoch, remaining, lastHash)
			if err != nil {
				return entity.WantAggregationProofsRequest{}, errors.Errorf("failed to get signature requests without aggregation proof for epoch %d: %w", epoch, err)
			}

			if len(requests) == 0 {
				break // No more requests for this epoch
			}

			for _, req := range requests {
				if !req.KeyTag.Type().AggregationKey() {
					continue
				}
				// Check if proof already exists — clean up stale pending marker
				_, err := s.cfg.Repo.GetAggregationProof(ctx, req.RequestID)
				if err == nil {
					_ = s.cfg.Repo.RemoveAggregationProofPending(ctx, req.RequiredEpoch, req.RequestID)
					continue
				}
				allRequestIDs = append(allRequestIDs, req.RequestID)
				totalRequests++
			}

			lastHash = requests[len(requests)-1].RequestID // Update for pagination
			remaining = s.cfg.MaxAggProofRequestsPerSync - totalRequests
		}

		// Handle epoch == 0 to avoid underflow in unsigned arithmetic
		if epoch == 0 {
			break
		}
	}

	tracing.SetAttributes(span,
		attribute.Int("response.request_ids_count", len(allRequestIDs)),
	)

	return entity.WantAggregationProofsRequest{
		RequestIDs: allRequestIDs,
	}, nil
}
