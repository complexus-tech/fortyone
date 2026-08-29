package maya

import "go.opentelemetry.io/otel"

var mayaServiceTracer = otel.Tracer("github.com/complexus-tech/projects-api/internal/modules/maya/service")
