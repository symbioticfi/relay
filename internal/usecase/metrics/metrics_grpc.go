package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// GRPCMetrics holds Prometheus metrics for gRPC requests
type GRPCMetrics struct {
	requestDuration  *prometheus.HistogramVec
	requestsTotal    *prometheus.CounterVec
	requestsInFlight *prometheus.GaugeVec
}

// UnaryServerInterceptor returns a gRPC unary interceptor for metrics collection
func (m *Metrics) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
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
func (m *Metrics) StreamServerInterceptor() grpc.StreamServerInterceptor {
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
