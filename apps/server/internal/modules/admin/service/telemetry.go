package admin

import "go.opentelemetry.io/otel"

var adminTracer = otel.Tracer("github.com/complexus-tech/projects-api/internal/modules/admin/service")
