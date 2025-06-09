package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	sentryslog "github.com/getsentry/sentry-go/slog"
	"github.com/go-errors/errors"
	slogmulti "github.com/samber/slog-multi"
)

var once sync.Once

func Init(levelStr, mode string) {
	level := parseLogLevel(levelStr)

	if mode != "text" && mode != "pretty" {
		mode = "text"
	}

	once.Do(func() {
		internalInit(level, mode)
	})
}

func internalInit(level slog.Level, mode string) {
	if mode == "pretty" {
		initPretty(level)
		return
	}
	initText(level)
}

func parseLogLevel(levelStr string) slog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}

func initPretty(level slog.Level) {
	prettyHandler := NewHandler(&slog.HandlerOptions{
		AddSource:   false,
		Level:       level,
		ReplaceAttr: replaceAttr,
	})

	handler := ContextHandler{Handler: prettyHandler}

	slog.SetDefault(slog.New(handler))
}

func initText(level slog.Level) {
	infoAndBelowHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,
		Level:       level,
		ReplaceAttr: replaceAttr,
	})
	errorHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelError,
		ReplaceAttr: replaceAttr,
	})

	handlerRoutes := slogmulti.Router().
		Add(ContextHandler{Handler: infoAndBelowHandler}, notErrorLevel).
		Add(ContextHandler{Handler: errorHandler}, errorLevel)

	if sentryDSN := os.Getenv("SENTRY_DSN"); sentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{ //nolint:exhaustruct
			Dsn:         sentryDSN,
			Environment: os.Getenv("SENTRY_ENVIRONMENT"),
			Release:     os.Getenv("SENTRY_RELEASE"),
		})
		if err != nil {
			panic("failed to init sentry: " + err.Error())
		}

		sentryHandler := sentryslog.Option{
			Level:           slog.LevelError,
			ReplaceAttr:     replaceAttr,
			AttrFromContext: []func(ctx context.Context) []slog.Attr{getAttrs},
			AddSource:       false,
			Converter:       nil,
			Hub:             nil,
		}.NewSentryHandler()

		handlerRoutes = handlerRoutes.Add(sentryHandler, errorLevel)
	}

	handler := handlerRoutes.Handler()

	slog.SetDefault(slog.New(handler))
}

func errorLevel(_ context.Context, r slog.Record) bool {
	return r.Level == slog.LevelError
}

func notErrorLevel(_ context.Context, r slog.Record) bool {
	return r.Level != slog.LevelError
}

type humanDuration time.Duration

func (d humanDuration) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, time.Duration(d).String())), nil
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Value.Kind() == slog.KindAny {
		val := a.Value.Any()
		if errVal, ok := val.(error); ok {
			a.Value = fmtErr(errVal)
		}
	}
	if a.Value.Kind() == slog.KindDuration {
		a.Value = slog.AnyValue(humanDuration(a.Value.Duration()))
	}

	return a
}

// fmtErr returns a slog.GroupValue with keys "msg" and "trace".
// If the error does not implement interface { StackTrace() errors.StackTrace },
// the "trace" key is omitted.
func fmtErr(err error) slog.Value {
	var groupValues []slog.Attr

	groupValues = append(groupValues, slog.String("err", err.Error()))

	type StackTracer interface {
		Callers() []uintptr
	}
	// Find the trace to the location of the first errors.New,
	// errors.Wrap, or errors.WithStack call.
	var st StackTracer
	for err := err; err != nil; err = errors.Unwrap(err) {
		if x, ok := err.(StackTracer); ok {
			st = x
		}
	}

	if st != nil {
		for i, caller := range st.Callers() {
			fileName, fileLine := runtime.FuncForPC(caller).FileLine(caller - 1)
			groupValues = append(groupValues,
				slog.String("trace"+strconv.Itoa(i), fileName+":"+strconv.Itoa(fileLine)),
			)
		}
	}

	return slog.GroupValue(groupValues...)
}
