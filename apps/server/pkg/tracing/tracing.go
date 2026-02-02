package tracing

import (
	"context"
	"net/url"
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

	endpoint := strings.TrimSpace(c.endpoint)
	if strings.Contains(endpoint, "%2F") || strings.Contains(endpoint, "%2f") {
		if decoded, err := url.PathUnescape(endpoint); err == nil {
			endpoint = decoded
		}
	}

	switch {
	case strings.HasPrefix(endpoint, "https://") || strings.HasPrefix(endpoint, "http://"):
		parsed, err := url.Parse(endpoint)
		if err != nil {
			return nil, err
		}
		if parsed.Scheme == "http" {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		opts = append(opts, otlptracehttp.WithEndpoint(parsed.Host))
		if parsed.Path != "" && parsed.Path != "/" {
			opts = append(opts, otlptracehttp.WithURLPath(parsed.Path))
		}
	default:
		opts = append(opts, otlptracehttp.WithInsecure())
		host := endpoint
		path := ""
		if idx := strings.Index(host, "/"); idx != -1 {
			path = host[idx:]
			host = host[:idx]
		}
		opts = append(opts, otlptracehttp.WithEndpoint(host))
		if path != "" {
			opts = append(opts, otlptracehttp.WithURLPath(path))
		}
	}

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
