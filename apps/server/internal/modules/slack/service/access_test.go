package slack

import (
	"context"
	"errors"
	"testing"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceAccessPolicyRejectsRevokedAndDowngradedActors(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	actorID := uuid.New()
	tests := []struct {
		name string
		role authorization.WorkspaceRole
		err  error
	}{
		{name: "active admin", role: authorization.WorkspaceRoleAdmin},
		{name: "member downgrade", role: authorization.WorkspaceRoleMember, err: slackdomain.ErrForbidden},
		{name: "guest downgrade", role: authorization.WorkspaceRoleGuest, err: slackdomain.ErrForbidden},
		{name: "revoked membership", err: slackdomain.ErrForbidden},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := &mockRepo{workspaceRole: test.role}
			if test.name == "revoked membership" {
				repository.workspaceRoleErr = slackdomain.ErrForbidden
			}
			service := New(nil, repository, nil, nil, Config{})

			err := service.requireWorkspaceAdmin(context.Background(), workspaceID, actorID)

			if test.err == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, test.err)
		})
	}
}

func TestWorkspaceMemberPolicyRejectsUnknownRoleAndInfrastructureFailure(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	actorID := uuid.New()

	unknown := New(nil, &mockRepo{workspaceRole: authorization.WorkspaceRole("owner")}, nil, nil, Config{})
	require.ErrorIs(t, unknown.requireWorkspaceMember(context.Background(), workspaceID, actorID), slackdomain.ErrForbidden)

	databaseErr := errors.New("database unavailable")
	unavailable := New(nil, &mockRepo{workspaceRoleErr: databaseErr}, nil, nil, Config{})
	err := unavailable.requireWorkspaceMember(context.Background(), workspaceID, actorID)
	require.ErrorIs(t, err, databaseErr)
	require.NotErrorIs(t, err, slackdomain.ErrForbidden)
}

func TestWorkspaceAccessFailsClosedWithoutRepositoryCapability(t *testing.T) {
	t.Parallel()

	service := New(nil, nil, nil, nil, Config{})
	err := service.requireWorkspaceAdmin(context.Background(), uuid.New(), uuid.New())
	require.ErrorContains(t, err, "authorization is not configured")
}
