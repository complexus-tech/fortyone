package outboundwebhookshttp

import (
	"context"

	outboundwebhooksservice "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
)

// humanAccessFromContext preserves the browser-management boundary. Public API
// credentials use /api/v1; the first-party settings UI must authenticate with
// a current human session and current workspace membership instead.
func humanAccessFromContext(ctx context.Context) (outboundwebhooksservice.Access, error) {
	actor, err := platformauth.GetActor(ctx)
	if err != nil {
		return outboundwebhooksservice.Access{}, err
	}
	if actor.Kind != platformauth.PrincipalHumanUser {
		return outboundwebhooksservice.Access{}, errAccessDenied
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return outboundwebhooksservice.Access{}, err
	}
	if actor.WorkspaceID != workspace.ID {
		return outboundwebhooksservice.Access{}, errAccessDenied
	}
	role := authorization.WorkspaceRole(workspace.UserRole)
	if err := authorization.ValidateWorkspaceRole(role); err != nil {
		return outboundwebhooksservice.Access{}, err
	}

	return outboundwebhooksservice.Access{
		Actor: actor, WorkspaceID: workspace.ID, WorkspaceRole: role,
	}, nil
}
