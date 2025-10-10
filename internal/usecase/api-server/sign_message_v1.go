package api_server

import (
	"context"

	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

// SignMessage handles the gRPC SignMessage request
func (h *grpcHandler) SignMessage(ctx context.Context, req *apiv1.SignMessageRequest) (*apiv1.SignMessageResponse, error) {
	requiredEpoch := req.RequiredEpoch
	if req.RequiredEpoch == nil {
		latestEpoch, err := h.cfg.Repo.GetLatestValidatorSetEpoch(ctx)
		if err != nil {
			return nil, err
		}
		requiredEpoch = (*uint64)(&latestEpoch)
	}

	signReq := symbiotic.SignatureRequest{
		KeyTag:        symbiotic.KeyTag(req.GetKeyTag()),
		Message:       req.GetMessage(),
		RequiredEpoch: symbiotic.Epoch(*requiredEpoch),
	}

	reqID, err := h.cfg.Signer.RequestSignature(ctx, signReq)
	if err != nil {
		return nil, err
	}

	return &apiv1.SignMessageResponse{
		RequestId: reqID.Hex(),
		Epoch:     *requiredEpoch,
	}, nil
}
