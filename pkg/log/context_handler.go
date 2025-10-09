//nolint:wrapcheck // this is the library code, don't need to wrap it
package log

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
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

func WithComponent(ctx context.Context, component string) context.Context {
	return WithAttrs(ctx, slog.String("component", component))
}

func WithTraceContext(ctx context.Context) context.Context {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if !spanCtx.IsValid() {
		return ctx
	}

	return WithAttrs(ctx,
		slog.String("trace_id", spanCtx.TraceID().String()),
		slog.String("span_id", spanCtx.SpanID().String()),
	)
}

func WithAttrs(ctx context.Context, as ...slog.Attr) context.Context {
	if len(as) == 0 {
		return ctx
	}

	parentAttrs := getAttrs(ctx)
	newAttrs := copyAttrs(parentAttrs, len(as))

	for _, a := range as {
		replaced := false
		for i := range newAttrs {
			if newAttrs[i].Key == a.Key {
				newAttrs[i] = a
				replaced = true
				break
			}
		}
		if !replaced {
			newAttrs = append(newAttrs, a)
		}
	}

	return context.WithValue(ctx, attrsKeyValue, newAttrs)
}

func copyAttrs(parentAttrs []slog.Attr, addCapacity int) []slog.Attr {
	if parentAttrs == nil {
		return nil
	}

	newAttrs := make([]slog.Attr, len(parentAttrs), len(parentAttrs)+addCapacity)
	copy(newAttrs, parentAttrs)
	return newAttrs
}

func getAttrs(ctx context.Context) []slog.Attr {
	if attrs, ok := ctx.Value(attrsKeyValue).([]slog.Attr); ok {
		return attrs
	}

	return nil
}
