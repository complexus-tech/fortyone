package stories

import (
	"context"
	"testing"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type credentialStoryReferenceRepositoryStub struct {
	scope      StoryReadScope
	teamCode   string
	sequenceID int
	story      CoreSingleStory
}

func (r *credentialStoryReferenceRepositoryStub) QueryCredentialVisibleStoryByRef(
	_ context.Context,
	scope StoryReadScope,
	teamCode string,
	sequenceID int,
) (CoreSingleStory, error) {
	r.scope = scope
	r.teamCode = teamCode
	r.sequenceID = sequenceID
	return r.story, nil
}

func TestQueryByRefForIntegrationUsesRestrictedCredentialScope(t *testing.T) {
	t.Parallel()

	workspaceID, credentialID, allowedTeamID := uuid.New(), uuid.New(), uuid.New()
	teamAccess, err := platformauth.RestrictedTeamAccess(allowedTeamID)
	require.NoError(t, err)
	actor, err := platformauth.NewActor(
		credentialID,
		platformauth.PrincipalServiceAccount,
		credentialID,
		platformauth.MustScopeSet(platformauth.ScopeStoriesRead),
		teamAccess,
	)
	require.NoError(t, err)
	actor, err = actor.WithWorkspace(workspaceID)
	require.NoError(t, err)
	ctx, err := platformauth.SetActor(context.Background(), actor)
	require.NoError(t, err)
	repo := &credentialStoryReferenceRepositoryStub{story: CoreSingleStory{
		ID: uuid.New(), Team: allowedTeamID, Workspace: workspaceID,
	}}
	service := &Service{repo: repo}

	story, err := service.QueryByRefForIntegration(ctx, workspaceID, "PRD-727")

	require.NoError(t, err)
	require.Equal(t, repo.story.ID, story.ID)
	require.Equal(t, credentialID, repo.scope.ActorID)
	require.Equal(t, workspaceID, repo.scope.WorkspaceID)
	require.False(t, repo.scope.UnrestrictedTeamAccess)
	require.Equal(t, []uuid.UUID{allowedTeamID}, repo.scope.AllowedTeamIDs)
	require.Equal(t, "PRD", repo.teamCode)
	require.Equal(t, 727, repo.sequenceID)
}

func TestQueryByRefForIntegrationRejectsHumanActor(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	actor, err := platformauth.NewHumanActor(uuid.New()).WithWorkspace(workspaceID)
	require.NoError(t, err)
	ctx, err := platformauth.SetActor(context.Background(), actor)
	require.NoError(t, err)
	repo := &credentialStoryReferenceRepositoryStub{}
	service := &Service{repo: repo}

	_, err = service.QueryByRefForIntegration(ctx, workspaceID, "PRD-727")

	require.ErrorIs(t, err, ErrStoryReadForbidden)
	require.Empty(t, repo.teamCode)
}
