//nolint:wrapcheck // this is the library code, don't need to wrap it
package log

import (
	"context"
	"log/slog"
)

type ContextHandler struct {
	slog.Handler
}

func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(getAttrs(ctx)...)
	return h.Handler.Handle(ctx, r)
}

type attrsKey struct{}

var attrsKeyValue attrsKey

func WithAttrs(ctx context.Context, as ...slog.Attr) context.Context {
	if len(as) == 0 {
		return ctx
	}

	parentAttrs := getAttrs(ctx)
	newAttrs := copyAttrs(parentAttrs)

	newAttrs = append(newAttrs, as...)

	return context.WithValue(ctx, attrsKeyValue, newAttrs)
}

func copyAttrs(parentAttrs []slog.Attr) []slog.Attr {
	if parentAttrs == nil {
		return nil
	}

	newAttrs := make([]slog.Attr, len(parentAttrs))
	copy(newAttrs, parentAttrs)
	return newAttrs
}

func getAttrs(ctx context.Context) []slog.Attr {
	if attrs, ok := ctx.Value(attrsKeyValue).([]slog.Attr); ok {
		return attrs
	}

	return nil
}
