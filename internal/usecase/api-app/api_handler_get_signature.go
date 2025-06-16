package apiApp

import (
	"context"

	"github.com/go-errors/errors"

	"middleware-offchain/internal/gen/api"
)

func (h *handler) GetSignatureGet(ctx context.Context, params api.GetSignatureGetParams) (*api.Signature, error) {
	// todo ilya implement
	return nil, errors.New("not yet implemented")
}
