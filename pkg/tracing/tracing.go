package tracing

import (
	"context"
	"time"

	"github.com/go-errors/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials/insecure"
)

const ServiceName = "symbiotic-relay"

type Config struct {
	Enabled    bool
	Endpoint   string
	SampleRate float64
	InstanceID string
	Version    string
}

type Tracer struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
}

// New creates a new tracer with OTLP exporter for Jaeger
func New(ctx context.Context, cfg Config) (*Tracer, error) {
	if !cfg.Enabled {
		// Return a no-op tracer when tracing is disabled
		return &Tracer{
			provider: sdktrace.NewTracerProvider(),
			tracer:   otel.Tracer(ServiceName),
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, errors.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource with service name and instance-specific attributes
	resourceAttrs := []resource.Option{
		resource.WithAttributes(
			semconv.ServiceNameKey.String(ServiceName),
		),
		resource.WithAttributes(
			semconv.ServiceInstanceIDKey.String(cfg.InstanceID),
		),
		resource.WithAttributes(
			semconv.ServiceVersionKey.String(cfg.Version),
		),
	}

	resourceAttrs = append(resourceAttrs, resource.WithHost())

	res, err := resource.New(ctx, resourceAttrs...)
	if err != nil {
		return nil, errors.Errorf("failed to create resource: %w", err)
	}

	// Determine sampler based on sample rate
	var sampler sdktrace.Sampler
	if cfg.SampleRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if cfg.SampleRate <= 0.0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(cfg.SampleRate)
	}

	// Create trace provider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global trace provider
	otel.SetTracerProvider(provider)

	// Set global propagator for context propagation (W3C Trace Context)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Tracer{
		provider: provider,
		tracer:   otel.Tracer(ServiceName),
	}, nil
}

// Shutdown gracefully shuts down the tracer provider
func (t *Tracer) Shutdown(ctx context.Context) error {
	if t.provider == nil {
		return nil
	}
	return t.provider.Shutdown(ctx)
}
