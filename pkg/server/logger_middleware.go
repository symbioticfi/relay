package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type logFormatter struct {
}

// NewLogEntry creates a new LogEntry for the request.
func (l logFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &logEntry{request: r}
}

type logEntry struct {
	request *http.Request
}

func (l *logEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	scheme := "http"
	if l.request.TLS != nil {
		scheme = "https"
	}
	action := fmt.Sprintf("%s %s://%s%s", l.request.Method, scheme, l.request.Host, l.request.RequestURI)

	slog.DebugContext(l.request.Context(), "Request served",
		slog.String("action", action),
		slog.Int("status", status),
		slog.String("elapsed", elapsed.String()),
	)
}

func (l *logEntry) Panic(v interface{}, stack []byte) {
	middleware.PrintPrettyStack(v)
}
