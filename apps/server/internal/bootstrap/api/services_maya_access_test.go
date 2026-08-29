package api

import (
	"context"
	"errors"
	"testing"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEnsureBackgroundMayaEnabledPreservesStoryEntitlementErrors(t *testing.T) {
	t.Parallel()

	databaseErr := errors.New("database unavailable")
	tests := []struct {
		name       string
		allowed    bool
		accessErr  error
		wantErr    error
		wantString string
	}{
		{name: "allowed", allowed: true},
		{
			name:       "not entitled",
			wantErr:    stories.ErrAutoSchedulingUnavailable,
			wantString: stories.ErrAutoSchedulingUnavailable.Error(),
		},
		{
			name:       "access unavailable",
			accessErr:  databaseErr,
			wantErr:    stories.ErrAutoSchedulingAccessCheckFailed,
			wantString: "auto-scheduling access could not be verified: check background Maya access: database unavailable",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			workspaceID := uuid.New()
			access := &apiMayaWorkspaceAccessStub{
				allowed: test.allowed,
				err:     test.accessErr,
			}

			err := ensureBackgroundMayaEnabled(context.Background(), access, workspaceID)

			if test.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, test.wantErr)
				require.EqualError(t, err, test.wantString)
			}
			require.Equal(t, []uuid.UUID{workspaceID}, access.workspaceIDs)
		})
	}
}

type apiMayaWorkspaceAccessStub struct {
	allowed      bool
	err          error
	workspaceIDs []uuid.UUID
}

func (access *apiMayaWorkspaceAccessStub) WorkspaceCanUseMaya(
	_ context.Context,
	workspaceID uuid.UUID,
) (bool, error) {
	access.workspaceIDs = append(access.workspaceIDs, workspaceID)
	return access.allowed, access.err
}
