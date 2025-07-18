package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/symbiotic/relay/pkg/log"
)

func RequestToSlog(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		reqID := middleware.GetReqID(ctx)
		ctx = log.WithAttrs(ctx,
			slog.String("request_id", reqID),
			slog.String("request_url", r.RequestURI),
		)
		ctx = log.WithComponent(ctx, "api")
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
