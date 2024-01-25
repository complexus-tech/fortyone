package web

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type contextKey int

const key contextKey = 1

type Values struct {
	TraceID    string
	Tracer     trace.Tracer
	Now        time.Time
	StatusCode int
}

// SetValues sets the web values on the context.
func SetValues(ctx context.Context, v *Values) context.Context {
	return context.WithValue(ctx, key, v)
}

// GetValues returns the web values from the context.
func GetValues(ctx context.Context) *Values {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return &Values{
			TraceID: "None",
			Tracer:  noop.NewTracerProvider().Tracer(""),
			Now:     time.Now(),
		}
	}
	return v
}

// GetTraceID returns the trace id from the web values.
func GetTraceID(ctx context.Context) string {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return "None"
	}
	return v.TraceID
}

func GetTime(ctx context.Context) time.Time {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return time.Now()
	}
	return v.Now
}

func AddSpan(ctx context.Context, spanName string, keyValues ...attribute.KeyValue) (context.Context, trace.Span) {
	v, ok := ctx.Value(key).(*Values)
	if !ok || v.Tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	ctx, span := v.Tracer.Start(ctx, spanName)
	for _, kv := range keyValues {
		span.SetAttributes(kv)
	}
	return ctx, span
}

// SetStatusCode sets the status code on the web values.
func SetStatusCode(ctx context.Context, statusCode int) error {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return errors.New("web value missing from context")
	}
	v.StatusCode = statusCode
	return nil
}
