package api_server

import (
	"context"

	"github.com/go-errors/errors"
	"github.com/samber/lo"
	"github.com/symbioticfi/relay/internal/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// GetSignatureRequestsByEpoch handles the gRPC GetSignatureRequestsByEpoch request
func (h *grpcHandler) GetSignatureRequestsByEpoch(ctx context.Context, req *apiv1.GetSignatureRequestsByEpochRequest) (*apiv1.GetSignatureRequestsByEpochResponse, error) {
	epoch := req.GetEpoch()

	signatureRequestsWithID, err := h.cfg.Repo.GetSignatureRequestsWithIDByEpoch(ctx, symbiotic.Epoch(epoch))
	if err != nil {
		return nil, errors.Errorf("failed to get signature requests by epoch: %w", err)
	}

	return &apiv1.GetSignatureRequestsByEpochResponse{
		SignatureRequests: lo.Map(signatureRequestsWithID, func(reqWithID entity.SignatureRequestWithID, _ int) *apiv1.SignatureRequest {
			return &apiv1.SignatureRequest{
				RequestId:     reqWithID.RequestID.Hex(),
				KeyTag:        uint32(reqWithID.SignatureRequest.KeyTag),
				Message:       reqWithID.SignatureRequest.Message,
				RequiredEpoch: uint64(reqWithID.SignatureRequest.RequiredEpoch),
			}
		}),
	}, nil
}
