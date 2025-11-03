package api_server

import (
	"context"

	"github.com/samber/lo"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"github.com/symbioticfi/relay/symbiotic/entity"
)

// GetSignatureRequestIDsByEpoch handles the gRPC GetSignatureRequestIDsByEpoch request
func (h *grpcHandler) GetSignatureRequestIDsByEpoch(ctx context.Context, req *apiv1.GetSignatureRequestIDsByEpochRequest) (*apiv1.GetSignatureRequestIDsByEpochResponse, error) {
	epoch := req.GetEpoch()

	requestIDs, err := h.cfg.Repo.GetSignatureRequestIDsByEpoch(ctx, entity.Epoch(epoch))
	if err != nil {
		return nil, errors.Errorf("failed to get signature request IDs by epoch: %w", err)
	}

	return &apiv1.GetSignatureRequestIDsByEpochResponse{
		RequestIds: lo.Map(requestIDs, func(requestID common.Hash, _ int) string { return requestID.Hex() }),
	}, nil
}
