package log

import (
	"log/slog"
	"testing"
	"time"

	"github.com/go-errors/errors"
)

func TestLog(t *testing.T) {
	t.Setenv("LOG_MODE", "local")
	Init("debug", "pretty")

	slog.Info("calculations completed")
	slog.Error("oh, no", "err", errorFunc(), "someTime", time.Now(), "result", 42, "duration", time.Second)
	slog.Info("calculations completed", "someTime", time.Now(), "result", 42, "duration", time.Second)
	slog.Error("oh, no", "err", errors.Errorf("hello error: %w", errorFunc()))
}

func TestComponent(t *testing.T) {
	Init("debug", "pretty")

	ctx := WithComponent(t.Context(), "aggregator")
	slog.InfoContext(ctx, "calculations completed")
	slog.ErrorContext(ctx, "oh, no", "err", errorFunc(), "someTime", time.Now(), "result", 42, "duration", time.Second)
	slog.InfoContext(ctx, "calculations completed", "someTime", time.Now(), "result", 42, "duration", time.Second)
	slog.ErrorContext(ctx, "oh, no", "err", errors.Errorf("hello error: %w", errorFunc()))

	slog.InfoContext(WithComponent(ctx, "anotherComponent"), "with 2 components")
}

func TestComponentJSON(t *testing.T) {
	Init("debug", "json")

	ctx := WithComponent(t.Context(), "aggregator")
	slog.InfoContext(WithComponent(ctx, "anotherComponent"), "with 2 components")
}

func errorFunc() error {
	return errors.New("my error")
}

func TestSplit(t *testing.T) {
	Init("debug", "pretty")
	slog.Info("hello", "there", "world")
	slog.DebugContext(t.Context(), "hello debug", "there1", "world1")
	slog.Error("hello error", "there2", "world2")

	ctx := WithAttrs(t.Context(), slog.String("key", "value"))

	time.Sleep(time.Second)

	slog.InfoContext(ctx, "hello4", "thee4", "world4")
	slog.DebugContext(ctx, "hello debug5", "there5", "world5")
	slog.ErrorContext(ctx, "hello error6", "there6", "world6")

	time.Sleep(time.Second)

	withGroupLogger := slog.Default().WithGroup("helloGroup")
	withGroupLogger = withGroupLogger.With(slog.Int("iiii", 7777))

	withGroupLogger.InfoContext(ctx, "hello4", "thee4", "world4")
	withGroupLogger.DebugContext(ctx, "hello debug5", "there5", "world5")
	withGroupLogger.ErrorContext(ctx, "hello error6", "there6", "world6")
}
