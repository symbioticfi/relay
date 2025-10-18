package api_server

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/symbioticfi/relay/internal/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetAggregationProof handles the gRPC GetAggregationProof request
func (h *grpcHandler) GetAggregationProof(ctx context.Context, req *apiv1.GetAggregationProofRequest) (*apiv1.GetAggregationProofResponse, error) {
	requestID := common.HexToHash(req.GetRequestId())

	signatureRequest, err := h.cfg.Repo.GetSignatureRequest(ctx, requestID)
	if err != nil {
		return nil, errors.Errorf("failed to get signature request: %w", err)
	}

	if !signatureRequest.KeyTag.Type().AggregationKey() {
		return nil, errors.Errorf("key tag %s is not an aggregation key", signatureRequest.KeyTag)
	}

	proof, err := h.cfg.Repo.GetAggregationProof(ctx, requestID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil, status.Errorf(codes.NotFound, "aggregation proof for request %s not found", req.GetRequestId())
		}
		return nil, errors.Errorf("failed to get aggregation proof: %w", err)
	}

	return &apiv1.GetAggregationProofResponse{
		AggregationProof: &apiv1.AggregationProof{
			MessageHash: proof.MessageHash,
			Proof:       proof.Proof,
		},
	}, nil
}
