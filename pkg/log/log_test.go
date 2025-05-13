package log

import (
	"log/slog"
	"testing"
	"time"

	"github.com/go-errors/errors"
)

func TestLog(t *testing.T) {
	t.Setenv("LOG_MODE", "local")
	Init("debug")

	slog.Info("calculations completed")
	slog.Error("oh, no", "err", myFund(), "someTime", time.Now(), "result", 42, "duration", time.Second)
	slog.Info("calculations completed", "someTime", time.Now(), "result", 42, "duration", time.Second)
	slog.Error("oh, no", "err", errors.Errorf("hello error: %w", myFund()))
}

func myFund() error {
	return errors.New("my error")
}

func TestSplit(t *testing.T) {
	Init("debug")
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
