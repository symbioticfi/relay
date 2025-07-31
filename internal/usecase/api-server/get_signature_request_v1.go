package api_server

import (
	"context"

	"github.com/ethereum/go-ethereum/common"

	"github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetSignatureRequest handles the gRPC GetSignatureRequest request
func (h *grpcHandler) GetSignatureRequest(ctx context.Context, req *v1.GetSignatureRequestRequest) (*v1.GetSignatureRequestResponse, error) {
	signatureRequest, err := h.cfg.Repo.GetSignatureRequest(ctx, common.HexToHash(req.GetRequestHash()))
	if err != nil {
		return nil, err
	}

	return &v1.GetSignatureRequestResponse{
		KeyTag:        uint32(signatureRequest.KeyTag),
		Message:       signatureRequest.Message,
		RequiredEpoch: uint64(signatureRequest.RequiredEpoch),
	}, nil
}
