package signals

import (
	"context"
	"log/slog"

	"github.com/maniartech/signals"
)

// Signal is a wrapper around library signals.Signal that provides a type-safe way to emit and listen to events.
// It allows to get rid of vendor lock and provides ability to change the underlying implementation if needed.
type Signal[T any] struct {
	sig signals.Signal[T]
}

type SignalListener[T any] func(context.Context, T)

func New[T any]() *Signal[T] {
	return &Signal[T]{
		sig: signals.New[T](),
	}
}

func (s *Signal[T]) Emit(ctx context.Context, payload T) {
	s.sig.Emit(ctx, payload)
}

func (s *Signal[T]) AddListener(listener func(context.Context, T) error, key string) int {
	return s.sig.AddListener(func(ctx context.Context, t T) {
		err := listener(ctx, t)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to handle event in listener", "key", key, "event", t, "err", err)
		}
	}, key)
}
