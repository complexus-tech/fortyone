package workerbootstrap

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceAssistantAccessDelegatesToMayaRepositoryCapability(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	wantErr := errors.New("database unavailable")
	access := &mayaWorkspaceAccessStub{err: wantErr}

	allowed, err := (workspaceAssistantAccess{access: access}).CanUseAssistant(
		context.Background(),
		workspaceID,
	)

	require.False(t, allowed)
	require.ErrorIs(t, err, wantErr)
	require.Equal(t, []uuid.UUID{workspaceID}, access.workspaceIDs)
}

type mayaWorkspaceAccessStub struct {
	allowed      bool
	err          error
	workspaceIDs []uuid.UUID
}

func (access *mayaWorkspaceAccessStub) WorkspaceCanUseMaya(
	_ context.Context,
	workspaceID uuid.UUID,
) (bool, error) {
	access.workspaceIDs = append(access.workspaceIDs, workspaceID)
	return access.allowed, access.err
}
