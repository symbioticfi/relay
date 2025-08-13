package log

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-errors/errors"
	slogmulti "github.com/samber/slog-multi"
)

var once sync.Once

func Init(levelStr, mode string) {
	level := parseLogLevel(levelStr)

	once.Do(func() {
		internalInit(level, mode)
	})
}

func internalInit(level slog.Level, mode string) {
	switch mode {
	case "pretty":
		initPretty(level)
	case "text":
		initText(level)
	case "json":
		initJson(level)
	default:
		initJson(level)
	}
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

	handler := handlerRoutes.Handler()

	slog.SetDefault(slog.New(handler))
}

func initJson(level slog.Level) {
	infoAndBelowHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,
		Level:       level,
		ReplaceAttr: replaceAttr,
	})
	errorHandler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelError,
		ReplaceAttr: replaceAttr,
	})

	handlerRoutes := slogmulti.Router().
		Add(ContextHandler{Handler: infoAndBelowHandler}, notErrorLevel).
		Add(ContextHandler{Handler: errorHandler}, errorLevel)

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
		} else {
			// Normalize big numbers to strings to avoid scientific notation in pretty logger
			switch v := val.(type) {
			case *big.Int:
				if v == nil {
					a.Value = slog.StringValue("nil")
				} else {
					a.Value = slog.StringValue(v.String())
				}
			case big.Int:
				a.Value = slog.StringValue(v.String())
			}
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
