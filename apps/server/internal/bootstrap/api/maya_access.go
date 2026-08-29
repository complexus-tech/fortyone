package api

import (
	"context"
	"fmt"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

// mayaWorkspaceAccess is the API composition root's minimal entitlement
// dependency. The Maya repository satisfies it with its generated SQLC query.
type mayaWorkspaceAccess interface {
	WorkspaceCanUseMaya(context.Context, uuid.UUID) (bool, error)
}

func ensureBackgroundMayaEnabled(
	ctx context.Context,
	access mayaWorkspaceAccess,
	workspaceID uuid.UUID,
) error {
	hasAccess, err := access.WorkspaceCanUseMaya(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("%w: check background Maya access: %v", stories.ErrAutoSchedulingAccessCheckFailed, err)
	}
	if !hasAccess {
		return stories.ErrAutoSchedulingUnavailable
	}
	return nil
}
