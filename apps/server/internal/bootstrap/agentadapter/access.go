package agentadapter

import (
	"context"

	"github.com/google/uuid"
)

type CommentAccess struct {
	Stories    CommentStories
	Workspace  WorkspaceAccess
	Configured bool
}

func (a CommentAccess) CanRead(ctx context.Context, storyID, workspaceID uuid.UUID) error {
	_, err := a.Stories.Get(ctx, storyID, workspaceID)
	return err
}

func (a CommentAccess) Enabled(ctx context.Context, workspaceID uuid.UUID) (bool, error) {
	if !a.Configured || a.Workspace == nil {
		return false, nil
	}
	return a.Workspace.WorkspaceCanUseMaya(ctx, workspaceID)
}
