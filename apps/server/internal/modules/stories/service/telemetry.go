package stories

import "go.opentelemetry.io/otel"

var storyServiceTracer = otel.Tracer("github.com/complexus-tech/projects-api/internal/modules/stories/service")
