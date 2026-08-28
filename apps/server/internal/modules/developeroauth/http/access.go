package developeroauthhttp

import (
	"context"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
)

const maximumManagementJSONBytes int64 = 32 << 10

// humanAccessFromContext is a deliberate management boundary. Machine actors
// cannot manage OAuth applications even if a future credential is accidentally
// granted integrations:manage. ApplicationManager repeats every authorization
// invariant so non-HTTP callers cannot bypass this adapter.
func humanAccessFromContext(ctx context.Context) (developeroauthdomain.ManagementAccess, error) {
	actor, err := platformauth.GetActor(ctx)
	if err != nil {
		return developeroauthdomain.ManagementAccess{}, err
	}
	if actor.Kind != platformauth.PrincipalHumanUser {
		return developeroauthdomain.ManagementAccess{}, developeroauthdomain.ErrAccessDenied
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return developeroauthdomain.ManagementAccess{}, err
	}
	if actor.WorkspaceID != workspace.ID {
		return developeroauthdomain.ManagementAccess{}, developeroauthdomain.ErrAccessDenied
	}

	return developeroauthdomain.ManagementAccess{
		Actor: actor, WorkspaceID: workspace.ID, WorkspaceRole: mid.Role(workspace.UserRole),
	}, nil
}

func requestID(ctx context.Context) string {
	return web.GetRequestID(ctx)
}
