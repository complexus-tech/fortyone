package tracing

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// config holds the configuration for tracing. This is used to initialize the
// tracing provider.
type config struct {
	service  string
	version  string
	environ  string
	endpoint string
	headers  map[string]string
}

// New returns a new config instance. This is used to initialize the tracing
// provider.
func New(service, version, environ, endpoint string, headers map[string]string) *config {
	return &config{
		service:  service,
		version:  version,
		environ:  environ,
		endpoint: endpoint,
		headers:  headers,
	}
}

// StartTracing starts the tracing provider and returns the tracer provider.
func (c *config) StartTracing() (*sdktrace.TracerProvider, error) {
	opts := []otlptracehttp.Option{}

	// Determine if TLS should be used based on endpoint prefix
	endpoint := c.endpoint
	if after, ok := strings.CutPrefix(endpoint, "https://"); ok {
		endpoint = after
	} else if after, ok := strings.CutPrefix(endpoint, "http://"); ok {
		endpoint = after
		opts = append(opts, otlptracehttp.WithInsecure())
	} else {
		// No protocol prefix - assume insecure
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	opts = append(opts, otlptracehttp.WithEndpoint(endpoint))

	if len(c.headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(c.headers))
	}

	exporter, err := otlptracehttp.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	// exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	// if err != nil {
	// 	return nil, err
	// }

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

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
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
