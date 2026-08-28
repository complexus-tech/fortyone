package tracing

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestAddSpanFromContextUsesParentProviderAndTrace(t *testing.T) {
	t.Parallel()

	provider := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	t.Cleanup(func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Fatalf("shutdown tracing provider: %v", err)
		}
	})
	parentContext, parent := provider.Tracer("test-parent").Start(context.Background(), "parent")
	t.Cleanup(func() { parent.End() })

	childContext, child := AddSpanFromContext(parentContext, "child", attribute.String("safe", "value"))
	t.Cleanup(func() { child.End() })

	parentSpanContext := parent.SpanContext()
	childSpanContext := child.SpanContext()
	if !parentSpanContext.IsValid() || !childSpanContext.IsValid() {
		t.Fatalf("span contexts must be valid: parent=%v child=%v", parentSpanContext, childSpanContext)
	}
	if childSpanContext.TraceID() != parentSpanContext.TraceID() {
		t.Fatalf("child trace ID = %s, want %s", childSpanContext.TraceID(), parentSpanContext.TraceID())
	}
	if childSpanContext.SpanID() == parentSpanContext.SpanID() {
		t.Fatal("child span must have its own span ID")
	}
	if got := trace.SpanFromContext(childContext).SpanContext(); !got.Equal(childSpanContext) {
		t.Fatalf("returned context span = %v, want %v", got, childSpanContext)
	}
}

func TestAddSpanFromContextWithoutActiveSpanIsSafe(t *testing.T) {
	t.Parallel()

	ctx, span := AddSpanFromContext(context.Background(), "detached")
	span.End()
	if ctx == nil {
		t.Fatal("returned context is nil")
	}
}
