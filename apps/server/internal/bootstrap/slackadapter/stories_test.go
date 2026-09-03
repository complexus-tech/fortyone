package slackadapter

import (
	"context"
	"testing"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type storyBackendActorCapture struct {
	actor       platformauth.Actor
	actorErr    error
	queryCalls  int
	workspaceID uuid.UUID
	reference   string
	integration bool
}

func (backend *storyBackendActorCapture) QueryByRefForIntegration(
	ctx context.Context,
	workspaceID uuid.UUID,
	reference string,
) (stories.CoreSingleStory, error) {
	backend.integration = true
	return backend.QueryByRef(ctx, workspaceID, reference)
}

func (backend *storyBackendActorCapture) Create(
	context.Context,
	stories.CoreNewStory,
	uuid.UUID,
) (stories.CoreSingleStory, error) {
	return stories.CoreSingleStory{}, nil
}

func (backend *storyBackendActorCapture) QueryByRef(
	ctx context.Context,
	workspaceID uuid.UUID,
	reference string,
) (stories.CoreSingleStory, error) {
	backend.queryCalls++
	backend.actor, backend.actorErr = platformauth.GetActor(ctx)
	backend.workspaceID = workspaceID
	backend.reference = reference
	return stories.CoreSingleStory{Workspace: workspaceID}, nil
}

func (backend *storyBackendActorCapture) UpdateExternalUserActionIfUnchanged(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	uuid.UUID,
	time.Time,
	map[string]any,
) error {
	return nil
}

func TestStoryServiceQueryByRefBindsLinkedSlackActor(t *testing.T) {
	t.Parallel()

	workspaceID, actorID := uuid.New(), uuid.New()
	backend := &storyBackendActorCapture{}
	adapter := NewStoryService(backend)

	story, err := adapter.QueryByRef(context.Background(), workspaceID, actorID, "PRD-727")

	require.NoError(t, err)
	require.NoError(t, backend.actorErr)
	require.Equal(t, workspaceID, story.Workspace)
	require.False(t, backend.integration)
	require.Equal(t, 1, backend.queryCalls)
	require.Equal(t, workspaceID, backend.workspaceID)
	require.Equal(t, "PRD-727", backend.reference)
	require.Equal(t, actorID, backend.actor.PrincipalID)
	require.Equal(t, platformauth.PrincipalHumanUser, backend.actor.Kind)
	require.Equal(t, workspaceID, backend.actor.WorkspaceID)
	require.True(t, backend.actor.Scopes.Has(platformauth.ScopeStoriesRead))
}

func TestStoryServiceQueryByRefRejectsMissingSlackActor(t *testing.T) {
	t.Parallel()

	backend := &storyBackendActorCapture{}
	adapter := NewStoryService(backend)

	_, err := adapter.QueryByRef(context.Background(), uuid.New(), uuid.Nil, "PRD-727")

	require.EqualError(t, err, "slack story actor is required")
	require.Zero(t, backend.queryCalls)
}

func TestStoryServiceQueryByRefForInstallationBindsRestrictedSlackActor(t *testing.T) {
	t.Parallel()

	workspaceID, installationID := uuid.New(), uuid.New()
	allowedTeamID, deniedTeamID := uuid.New(), uuid.New()
	backend := &storyBackendActorCapture{}
	adapter := NewStoryService(backend)

	story, err := adapter.QueryByRefForInstallation(
		context.Background(),
		workspaceID,
		installationID,
		[]uuid.UUID{allowedTeamID},
		"PRD-727",
	)

	require.NoError(t, err)
	require.NoError(t, backend.actorErr)
	require.Equal(t, workspaceID, story.Workspace)
	require.True(t, backend.integration)
	require.Equal(t, installationID, backend.actor.PrincipalID)
	require.Equal(t, installationID, backend.actor.CredentialID)
	require.Equal(t, platformauth.PrincipalServiceAccount, backend.actor.Kind)
	require.Equal(t, workspaceID, backend.actor.WorkspaceID)
	require.True(t, backend.actor.Scopes.Has(platformauth.ScopeStoriesRead))
	require.True(t, backend.actor.TeamAccess.Allows(allowedTeamID))
	require.False(t, backend.actor.TeamAccess.Allows(deniedTeamID))
}
