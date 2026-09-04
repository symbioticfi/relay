package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetricsHandlerPprofExposure(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/pprof/goroutine?debug=1", nil)

		initMetricsHandler(MetricsConfig{}).ServeHTTP(recorder, request)

		require.Equal(t, http.StatusNotFound, recorder.Code)
	})

	t.Run("enabled", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/pprof/goroutine?debug=1", nil)

		initMetricsHandler(MetricsConfig{ServePprof: true}).ServeHTTP(recorder, request)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Contains(t, recorder.Body.String(), "goroutine profile")
	})
}
