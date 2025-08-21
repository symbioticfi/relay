package api_server

import (
	"context"

	"github.com/symbioticfi/relay/core/entity"
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
)

// SignMessage handles the gRPC SignMessage request
func (h *grpcHandler) SignMessage(ctx context.Context, req *apiv1.SignMessageRequest) (*apiv1.SignMessageResponse, error) {
	requiredEpoch := req.RequiredEpoch
	if req.RequiredEpoch == nil {
		// later to optimize: fetch only epoch instead of entire set
		latestValsetMeta, err := h.cfg.Repo.GetLatestValidatorSetMeta(ctx)
		if err != nil {
			return nil, err
		}
		requiredEpoch = &latestValsetMeta.Epoch
	}

	signReq := entity.SignatureRequest{
		KeyTag:        entity.KeyTag(req.GetKeyTag()),
		Message:       req.GetMessage(),
		RequiredEpoch: entity.Epoch(*requiredEpoch),
	}

	err := h.cfg.Signer.Sign(ctx, signReq)
	if err != nil {
		return nil, err
	}

	return &apiv1.SignMessageResponse{
		RequestHash: signReq.Hash().Hex(),
		Epoch:       *requiredEpoch,
	}, nil
}
