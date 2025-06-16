package apiApp

import (
	"context"

	"middleware-offchain/core/entity"
	"middleware-offchain/internal/gen/api"
)

func (h *handler) SignMessagePost(ctx context.Context, reqRaw *api.SignatureRequest) (*api.SignMessagePostOK, error) {
	req := entity.SignatureRequest{
		KeyTag:        entity.KeyTag(reqRaw.KeyTag),
		RequiredEpoch: reqRaw.Epoch,
		Message:       reqRaw.Data,
	}

	err := h.cfg.Signer.Sign(ctx, req)
	if err != nil {
		return nil, err
	}

	return &api.SignMessagePostOK{
		RequestHash: req.Hash().Hex(),
	}, nil
}
