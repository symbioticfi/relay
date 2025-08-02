package api_server

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// grpcMetrics holds Prometheus metrics for gRPC requests
type grpcMetrics struct {
	requestDuration  *prometheus.HistogramVec
	requestsTotal    *prometheus.CounterVec
	requestsInFlight *prometheus.GaugeVec
}

// newGRPCMetrics creates and registers gRPC metrics
func newGRPCMetrics() *grpcMetrics {
	// Duration buckets similar to the original HTTP server but adjusted for gRPC
	durationBuckets := []float64{.005, .01, .025, .05, .1, .25, .5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 5, 10, 15, 20, 40, 45, 60}

	metrics := &grpcMetrics{
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "grpc_server_request_duration_seconds",
				Help:    "Duration of gRPC requests in seconds.",
				Buckets: durationBuckets,
			},
			[]string{"method", "status_code"},
		),
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "grpc_server_requests_total",
				Help: "Total number of gRPC requests.",
			},
			[]string{"method", "status_code"},
		),
		requestsInFlight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "grpc_server_requests_in_flight",
				Help: "Current number of gRPC requests being processed.",
			},
			[]string{"method"},
		),
	}

	// Register metrics
	prometheus.DefaultRegisterer.MustRegister(
		metrics.requestDuration,
		metrics.requestsTotal,
		metrics.requestsInFlight,
	)

	return metrics
}

// UnaryServerInterceptor returns a gRPC unary interceptor for metrics collection
func (m *grpcMetrics) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		method := info.FullMethod

		// Track in-flight requests
		m.requestsInFlight.WithLabelValues(method).Inc()
		defer m.requestsInFlight.WithLabelValues(method).Dec()

		// Start timing
		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Determine status code
		statusCode := "OK"
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code().String()
			} else {
				statusCode = "Unknown"
			}
		}

		// Record metrics
		m.requestDuration.WithLabelValues(method, statusCode).Observe(duration)
		m.requestsTotal.WithLabelValues(method, statusCode).Inc()

		return resp, err
	}
}

// StreamServerInterceptor returns a gRPC stream interceptor for metrics collection
func (m *grpcMetrics) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		method := info.FullMethod

		// Track in-flight requests
		m.requestsInFlight.WithLabelValues(method).Inc()
		defer m.requestsInFlight.WithLabelValues(method).Dec()

		// Start timing
		start := time.Now()

		// Call the handler
		err := handler(srv, stream)

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Determine status code
		statusCode := "OK"
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code().String()
			} else {
				statusCode = "Unknown"
			}
		}

		// Record metrics
		m.requestDuration.WithLabelValues(method, statusCode).Observe(duration)
		m.requestsTotal.WithLabelValues(method, statusCode).Inc()

		return err
	}
}
