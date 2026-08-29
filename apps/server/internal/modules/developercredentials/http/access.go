package developercredentialshttp

import (
	"context"
	"errors"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
)

// humanAccessFromContext is a deliberate adapter boundary. Management routes
// reject machine principals here even if a future credential gains a broad
// management scope; the service repeats the invariant for non-HTTP callers.
func humanAccessFromContext(ctx context.Context) (developercredentialsdomain.Access, error) {
	actor, err := platformauth.GetActor(ctx)
	if err != nil {
		return developercredentialsdomain.Access{}, err
	}
	if actor.Kind != platformauth.PrincipalHumanUser {
		return developercredentialsdomain.Access{}, developercredentialsdomain.ErrAccessDenied
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return developercredentialsdomain.Access{}, err
	}
	if actor.WorkspaceID != workspace.ID {
		return developercredentialsdomain.Access{}, errors.New("actor workspace does not match route workspace")
	}
	return developercredentialsdomain.Access{
		Actor: actor, WorkspaceID: workspace.ID, WorkspaceRole: mid.Role(workspace.UserRole),
	}, nil
}

func requestID(ctx context.Context) string {
	return web.GetRequestID(ctx)
}
