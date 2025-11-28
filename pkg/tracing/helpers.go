package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/symbioticfi/relay/pkg/log"
)

func StartSpan(ctx context.Context, spanName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return startSpanWithKind(ctx, spanName, trace.SpanKindInternal, attrs...)
}

func StartServerSpan(ctx context.Context, spanName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return startSpanWithKind(ctx, spanName, trace.SpanKindServer, attrs...)
}

func StartClientSpan(ctx context.Context, spanName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return startSpanWithKind(ctx, spanName, trace.SpanKindClient, attrs...)
}

func StartProducerSpan(ctx context.Context, spanName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return startSpanWithKind(ctx, spanName, trace.SpanKindProducer, attrs...)
}

func StartConsumerSpan(ctx context.Context, spanName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return startSpanWithKind(ctx, spanName, trace.SpanKindConsumer, attrs...)
}

func StartDBSpan(ctx context.Context, operation, statement string) (context.Context, trace.Span) {
	tracer := otel.Tracer(ServiceName)
	ctx, span := tracer.Start(ctx, fmt.Sprintf("db.%s", operation),
		trace.WithSpanKind(trace.SpanKindClient),
	)

	if span.IsRecording() {
		span.SetAttributes(
			semconv.DBSystemKey.String("badger"),
			semconv.DBOperationName(operation),
			semconv.DBQueryText(statement),
		)
	}

	ctx = log.WithTraceContext(ctx)

	return ctx, span
}

func startSpanWithKind(ctx context.Context, spanName string, kind trace.SpanKind, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	tracer := otel.Tracer(ServiceName)
	ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(kind), trace.WithAttributes(attrs...))

	ctx = log.WithTraceContext(ctx)

	return ctx, span
}

func RecordError(span trace.Span, err error) {
	if err == nil || !span.IsRecording() {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

func AddEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	if !span.IsRecording() {
		return
	}
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

func SetAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	if !span.IsRecording() {
		return
	}
	span.SetAttributes(attrs...)
}

// Common attribute keys for consistency across the codebase
var (
	// Chain and network attributes
	AttrChainID = attribute.Key("chain.id")
	AttrEpoch   = attribute.Key("epoch")
	AttrAddress = attribute.Key("address")

	// Signature and validation attributes
	AttrKeyTag           = attribute.Key("key.tag")
	AttrValidatorIndex   = attribute.Key("validator.index")
	AttrValidatorCount   = attribute.Key("validator.count")
	AttrValidatorAddress = attribute.Key("validator.address")
	AttrSignatureCount   = attribute.Key("signature.count")
	AttrQuorumThreshold  = attribute.Key("quorum.threshold")

	// P2P attributes
	AttrPeerID      = attribute.Key("peer.id")
	AttrMessageType = attribute.Key("message.type")

	// Database attributes
	AttrQueryName = attribute.Key("query.name")
	AttrCacheHit  = attribute.Key("cache.hit")

	// EVM attributes
	AttrMethodName = attribute.Key("evm.method")
	AttrTxHash     = attribute.Key("tx.hash")
	AttrGasUsed    = attribute.Key("gas.used")

	// Aggregation attributes
	AttrProofSize      = attribute.Key("proof.size")
	AttrProofType      = attribute.Key("proof.type")
	AttrAggregationID  = attribute.Key("aggregation.id")
	AttrRequestID      = attribute.Key("request.id")
	AttrVerifyDuration = attribute.Key("verify.duration_ms")
)
