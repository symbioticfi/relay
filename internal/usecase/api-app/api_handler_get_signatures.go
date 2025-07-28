package apiApp

import (
	"context"

	api2 "github.com/symbioticfi/relay/core/api/gen/api"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"github.com/samber/lo"

	"github.com/symbioticfi/relay/core/entity"
)

func (h *handler) GetSignaturesGet(ctx context.Context, params api2.GetSignaturesGetParams) ([]api2.Signature, error) {
	signatures, err := h.cfg.Repo.GetAllSignatures(ctx, common.HexToHash(params.RequestHash))
	if err != nil {
		return nil, errors.Errorf("failed to get signatures: %w", err)
	}

	return convertSignaturesToAPI(signatures), nil
}

func convertSignaturesToAPI(signatures []entity.SignatureExtended) []api2.Signature {
	return lo.Map(signatures, func(sig entity.SignatureExtended, _ int) api2.Signature {
		return convertSignatureToAPI(sig)
	})
}

func convertSignatureToAPI(sig entity.SignatureExtended) api2.Signature {
	return api2.Signature{
		Signature:   sig.Signature,
		MessageHash: sig.MessageHash,
		PublicKey:   sig.PublicKey,
	}
}
