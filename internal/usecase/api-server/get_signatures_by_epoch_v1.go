package api_server

import (
	"context"

	"github.com/go-errors/errors"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/symbiotic/entity"
)

// GetSignaturesByEpoch handles the gRPC GetSignaturesByEpoch request
func (h *grpcHandler) GetSignaturesByEpoch(ctx context.Context, req *apiv1.GetSignaturesByEpochRequest) (*apiv1.GetSignaturesByEpochResponse, error) {
	epoch := req.GetEpoch()

	signatures, err := h.cfg.Repo.GetSignaturesByEpoch(ctx, entity.Epoch(epoch))
	if err != nil {
		return nil, errors.Errorf("failed to get signatures by epoch: %w", err)
	}

	return &apiv1.GetSignaturesByEpochResponse{Signatures: convertSignaturesToPB(signatures)}, nil
}
