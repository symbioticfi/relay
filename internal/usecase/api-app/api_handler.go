package apiApp

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-errors/errors"
	"github.com/go-faster/jx"
	ht "github.com/ogen-go/ogen/http"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/validate"
	"github.com/samber/lo"

	"middleware-offchain/core/entity"
	"middleware-offchain/internal/gen/api"
)

type handler struct {
	cfg Config
}

func (h *handler) NewError(ctx context.Context, err error) *api.ErrorStatusCode {
	return errorWithHTTPCode(ctx, err)
}

func errorHandler(ctx context.Context, w http.ResponseWriter, _ *http.Request, err error) {
	code := errorWithHTTPCode(ctx, err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code.StatusCode)

	e := jx.GetEncoder()
	lo.ToPtr(code.GetResponse()).Encode(e)

	_, _ = w.Write(e.Bytes())
}

func errorWithHTTPCode(ctx context.Context, err error) *api.ErrorStatusCode {
	resp := &api.ErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response: api.Error{
			ErrorMessage: "",
			ErrorCode:    api.ErrorErrorCodeOoops,
		},
	}

	var ctError *validate.InvalidContentTypeError
	var ogenErr ogenerrors.Error
	var decodeParamError *ogenerrors.DecodeParamError
	switch {
	case errors.Is(err, ht.ErrNotImplemented):
		resp.StatusCode = http.StatusNotImplemented
	case errors.Is(err, entity.ErrNotAnAggregator):
		resp = &api.ErrorStatusCode{
			StatusCode: http.StatusMethodNotAllowed,
			Response: api.Error{
				ErrorMessage: "Not an aggregator",
				ErrorCode:    api.ErrorErrorCodeNotAnAggregator,
			},
		}
	case errors.As(err, &ctError):
		resp.StatusCode = http.StatusUnsupportedMediaType
	case errors.As(err, &decodeParamError):
		resp = &api.ErrorStatusCode{
			StatusCode: http.StatusBadRequest,
			Response: api.Error{
				ErrorMessage: decodeParamError.Error(),
				ErrorCode:    api.ErrorErrorCodeNoData,
			},
		}
	case errors.As(err, &ogenErr):
		resp.StatusCode = ogenErr.Code()
	case errors.Is(err, entity.ErrEntityNotFound):
		resp = &api.ErrorStatusCode{
			StatusCode: http.StatusNotFound,
			Response: api.Error{
				ErrorMessage: "Entity not found",
				ErrorCode:    api.ErrorErrorCodeNoData,
			},
		}
	case errors.Is(err, context.Canceled):
		resp = &api.ErrorStatusCode{
			StatusCode: 499, // nginx uses this code when client cancelled request
			Response: api.Error{
				ErrorMessage: "Cancelled",
				ErrorCode:    api.ErrorErrorCodeOoops,
			},
		}
	}

	if resp.StatusCode > 499 {
		slog.ErrorContext(ctx, "Failed to serve http request with error", "err", err)
	} else {
		slog.DebugContext(ctx, "Failed to serve http request", "err", err)
	}

	return resp
}
