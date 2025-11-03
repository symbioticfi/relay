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

// GetSignatureRequest handles the gRPC GetSignatureRequest request
func (h *grpcHandler) GetSignatureRequest(ctx context.Context, req *apiv1.GetSignatureRequestRequest) (*apiv1.GetSignatureRequestResponse, error) {
	requestID := common.HexToHash(req.GetRequestId())

	signatureRequest, err := h.cfg.Repo.GetSignatureRequest(ctx, requestID)
	if err != nil {
		if errors.Is(err, entity.ErrEntityNotFound) {
			return nil, status.Errorf(codes.NotFound, "signature request %s not found", req.GetRequestId())
		}
		return nil, err
	}

	return &apiv1.GetSignatureRequestResponse{
		SignatureRequest: &apiv1.SignatureRequest{
			RequestId:     requestID.Hex(),
			KeyTag:        uint32(signatureRequest.KeyTag),
			Message:       signatureRequest.Message,
			RequiredEpoch: uint64(signatureRequest.RequiredEpoch),
		},
	}, nil
}
