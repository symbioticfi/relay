package apiApp

import (
	"context"

	api2 "github.com/symbioticfi/relay/core/api/gen/api"

	"github.com/ethereum/go-ethereum/common"
)

func (h *handler) GetSignatureRequestGet(ctx context.Context, params api2.GetSignatureRequestGetParams) (*api2.SignatureRequest, error) {
	signatureRequest, err := h.cfg.Repo.GetSignatureRequest(ctx, common.HexToHash(params.RequestHash))
	if err != nil {
		return nil, err
	}

	return &api2.SignatureRequest{
		KeyTag:        uint8(signatureRequest.KeyTag),
		Message:       signatureRequest.Message,
		RequiredEpoch: uint64(signatureRequest.RequiredEpoch),
	}, nil
}
