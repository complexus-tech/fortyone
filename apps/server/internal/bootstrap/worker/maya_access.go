package workerbootstrap

import (
	"context"

	"github.com/google/uuid"
)

// mayaWorkspaceAccess is the worker composition root's minimal entitlement
// dependency. One Maya repository instance is shared by every worker surface.
type mayaWorkspaceAccess interface {
	WorkspaceCanUseMaya(context.Context, uuid.UUID) (bool, error)
}

// workspaceAssistantAccess translates Maya's repository capability to the
// consumer-owned Slack access interface without introducing SQL at bootstrap.
type workspaceAssistantAccess struct {
	access mayaWorkspaceAccess
}

func (a workspaceAssistantAccess) CanUseAssistant(
	ctx context.Context,
	workspaceID uuid.UUID,
) (bool, error) {
	return a.access.WorkspaceCanUseMaya(ctx, workspaceID)
}
