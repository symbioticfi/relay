package p2p

import (
	"context"

	v1 "github.com/symbioticfi/relay/internal/client/p2p/proto/v1"
)

type p2pHandler struct {
	v1.UnimplementedSymbioticP2PServiceServer
}

func (h *p2pHandler) WantSignatures(ctx context.Context, req *v1.WantSignaturesRequest) (*v1.WantSignaturesResponse, error) {
	return &v1.WantSignaturesResponse{}, nil
}
