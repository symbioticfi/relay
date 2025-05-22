package signer_app

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *SignerApp) Handler() http.Handler {
	r := chi.NewRouter()

	r.Post("/signMessage", s.signMessageHandler)

	return r
}

type signMessageRequest struct {
	Data []byte `json:"data"`
}

func (s *SignerApp) signMessageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req signMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(ctx, w, err)
		return
	}

	if err := s.signMessage(ctx, req.Data); err != nil {
		handleError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

type errorStatusCode struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func handleError(ctx context.Context, w http.ResponseWriter, err error) {
	resp := &errorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Message:    "internal error",
	}

	switch {
	case errors.Is(err, context.Canceled):
		resp = &errorStatusCode{
			StatusCode: 499,
			Message:    "request cancelled",
		}
	case errors.Is(err, context.DeadlineExceeded):
		resp = &errorStatusCode{
			StatusCode: 500,
			Message:    "deadline exceeded",
		}
	}

	if resp.StatusCode > 499 {
		slog.ErrorContext(ctx, "failed to serve http request with error", "err", err)
	} else {
		slog.DebugContext(ctx, "failed to serve http request", "err", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.WarnContext(ctx, "failed to encode response", "err", err)
		return
	}
}
