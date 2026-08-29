package tracing

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const contextTracerName = "github.com/complexus-tech/projects-api/pkg/tracing"

// StartSpan starts a span with the provided tracer and attributes.
func StartSpan(ctx context.Context, tracer trace.Tracer, spanName string, keyValues ...attribute.KeyValue) (context.Context, trace.Span) {
	if tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := tracer.Start(ctx, spanName)
	for _, kv := range keyValues {
		span.SetAttributes(kv)
	}

	return ctx, span
}

// AddSpan is kept as the primary helper name for span creation in services.
func AddSpan(ctx context.Context, tracer trace.Tracer, spanName string, keyValues ...attribute.KeyValue) (context.Context, trace.Span) {
	return StartSpan(ctx, tracer, spanName, keyValues...)
}

// AddSpanFromContext starts a child span on the same provider as the current
// span. Service, repository, and worker code use this helper without depending
// on HTTP request-context implementation details.
func AddSpanFromContext(ctx context.Context, spanName string, keyValues ...attribute.KeyValue) (context.Context, trace.Span) {
	provider := trace.SpanFromContext(ctx).TracerProvider()
	return StartSpan(ctx, provider.Tracer(contextTracerName), spanName, keyValues...)
}

// StartHTTPSpan starts a standard HTTP handler span and sets request attributes.
func StartHTTPSpan(ctx context.Context, tracer trace.Tracer, spanName, method, endpoint string) (context.Context, trace.Span) {
	ctx, span := StartSpan(ctx, tracer, spanName)
	if span != nil {
		span.SetAttributes(
			attribute.String("http.method", method),
			attribute.String("http.endpoint", endpoint),
		)
	}

	return ctx, span
}

// InjectTraceHeaders injects the trace context into response headers.
func InjectTraceHeaders(ctx context.Context, headers http.Header) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(headers))
}
