package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// config holds the configuration for tracing. This is used to initialize the
// tracing provider.
type config struct {
	service string
	version string
	environ string
}

// New returns a new config instance. This is used to initialize the tracing
// provider.
func New(service, version, environ string) *config {
	return &config{
		service: service,
		version: version,
		environ: environ,
	}
}

// StartTracing starts the tracing provider and returns the tracer provider.
func (c *config) StartTracing() (*trace.TracerProvider, error) {
	ctx := context.Background()
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithInsecure())
	if err != nil {

		return nil, err
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"",
			semconv.ServiceName(c.service),
			semconv.ServiceVersion(c.version),
			attribute.String("environment", c.environ),
		),
	)
	if err != nil {
		return nil, err
	}

	provider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)
	// Set the TracerProvider globally.
	otel.SetTracerProvider(provider)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return provider, nil
}
