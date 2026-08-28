package invitations

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var invitationTracer = otel.Tracer("github.com/complexus-tech/projects-api/internal/modules/invitations/service")

func startInvitationSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return invitationTracer.Start(ctx, name)
}
