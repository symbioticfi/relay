package api_server

import (
	"context"

	"github.com/ethereum/go-ethereum/common"

	"github.com/symbioticfi/relay/internal/gen/api/v1"
)

// GetSignatureRequest handles the gRPC GetSignatureRequest request
func (h *grpcHandler) GetSignatureRequest(ctx context.Context, req *v1.GetSignatureRequestRequest) (*v1.SignatureRequest, error) {
	signatureRequest, err := h.cfg.Repo.GetSignatureRequest(ctx, common.HexToHash(req.RequestHash))
	if err != nil {
		return nil, err
	}

	return &v1.SignatureRequest{
		KeyTag:        uint32(signatureRequest.KeyTag),
		Message:       signatureRequest.Message,
		RequiredEpoch: uint64(signatureRequest.RequiredEpoch),
	}, nil
}
