package p2p

import (
	"context"

	"github.com/go-errors/errors"

	"github.com/symbioticfi/relay/core/entity"
	p2pv1 "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
	"github.com/symbioticfi/relay/pkg/log"
)

// signatureRequestHandler defines the interface for handling signature requests
type signatureRequestHandler interface {
	HandleWantSignaturesRequest(ctx context.Context, request entity.WantSignaturesRequest) (entity.WantSignaturesResponse, error)
}

type GRPCHandler struct {
	p2pv1.UnimplementedSymbioticP2PServiceServer

	handler signatureRequestHandler
}

func NewP2PHandler(handler signatureRequestHandler) *GRPCHandler {
	return &GRPCHandler{
		handler: handler,
	}
}

// WantSignatures handles incoming signature requests from peers
func (h *GRPCHandler) WantSignatures(ctx context.Context, req *p2pv1.WantSignaturesRequest) (*p2pv1.WantSignaturesResponse, error) {
	ctx = log.WithComponent(ctx, "p2p-grpc-handler")

	entityReq, err := protoToEntityRequest(req)
	if err != nil {
		return &p2pv1.WantSignaturesResponse{}, errors.Errorf("failed to convert request: %w", err)
	}

	response, err := h.handler.HandleWantSignaturesRequest(ctx, entityReq)
	if err != nil {
		return &p2pv1.WantSignaturesResponse{}, errors.Errorf("failed to handle request: %w", err)
	}

	return entityToProtoResponse(response), nil
}
