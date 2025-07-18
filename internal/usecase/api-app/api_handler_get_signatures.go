package apiApp

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"github.com/symbioticfi/relay/core/entity"
	"github.com/symbioticfi/relay/internal/gen/api"
)

func (h *handler) GetSignaturesGet(ctx context.Context, params api.GetSignaturesGetParams) ([]api.Signature, error) {
	signatures, err := h.cfg.Repo.GetAllSignatures(ctx, common.HexToHash(params.RequestHash))
	if err != nil {
		return nil, errors.Errorf("failed to get signatures: %w", err)
	}

	return convertSignaturesToAPI(signatures), nil
}

func convertSignaturesToAPI(signatures []entity.SignatureExtended) []api.Signature {
	return lo.Map(signatures, func(sig entity.SignatureExtended, _ int) api.Signature {
		return convertSignatureToAPI(sig)
	})
}

func convertSignatureToAPI(sig entity.SignatureExtended) api.Signature {
	return api.Signature{
		Signature:   sig.Signature,
		MessageHash: sig.MessageHash,
		PublicKey:   sig.PublicKey,
	}
}
