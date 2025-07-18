package apiApp

import (
	"context"

	"github.com/ethereum/go-ethereum/common"

	"github.com/symbioticfi/relay/internal/gen/api"
)

func (h *handler) GetSignatureRequestGet(ctx context.Context, params api.GetSignatureRequestGetParams) (*api.SignatureRequest, error) {
	signatureRequest, err := h.cfg.Repo.GetSignatureRequest(ctx, common.HexToHash(params.RequestHash))
	if err != nil {
		return nil, err
	}

	return &api.SignatureRequest{
		KeyTag:        uint8(signatureRequest.KeyTag),
		Message:       signatureRequest.Message,
		RequiredEpoch: uint64(signatureRequest.RequiredEpoch),
	}, nil
}
