package apiApp

import (
	"context"

	"github.com/symbiotic/relay/core/entity"
	"github.com/symbiotic/relay/internal/gen/api"
)

func (h *handler) SignMessagePost(ctx context.Context, reqRaw *api.SignMessagePostReq) (*api.SignMessagePostOK, error) {
	requiredEpoch := reqRaw.RequiredEpoch.Value
	if !reqRaw.RequiredEpoch.IsSet() {
		// later to optimize: fetch only epoch instead of entire set
		latestValSet, err := h.cfg.Repo.GetLatestValidatorSet(ctx)
		if err != nil {
			return nil, err
		}
		requiredEpoch = latestValSet.Epoch
	}
	req := entity.SignatureRequest{
		KeyTag:        entity.KeyTag(reqRaw.KeyTag),
		Message:       reqRaw.Message,
		RequiredEpoch: entity.Epoch(requiredEpoch),
	}

	err := h.cfg.Signer.Sign(ctx, req)
	if err != nil {
		return nil, err
	}

	return &api.SignMessagePostOK{
		RequestHash: req.Hash().Hex(),
		Epoch:       requiredEpoch,
	}, nil
}
