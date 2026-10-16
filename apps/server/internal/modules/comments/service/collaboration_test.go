package comments

import (
	"testing"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCollaborationRejectsInvalidIntentBeforePersistence(t *testing.T) {
	t.Parallel()
	workspaceID, actorID := uuid.New(), uuid.New()
	actor := commentTestActor(t, actorID, workspaceID)
	scope := AuthorScope{CommentID: uuid.New(), WorkspaceID: workspaceID, Actor: actor}
	service := New(serviceRepositoryFake{})
	_, err := service.SetReaction(t.Context(), ReactionCommand{Scope: scope, Emoji: "unsupported", Active: true})
	require.ErrorIs(t, err, ErrInvalidComment)
	_, err = service.MarkRead(t.Context(), ReadCommand{Scope: StoryScope{StoryID: uuid.New(), WorkspaceID: workspaceID, Actor: actor}})
	require.ErrorIs(t, err, ErrInvalidComment)
	readOnly, err := platformauth.NewActor(actorID, platformauth.PrincipalOAuthUser, uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeCommentsRead), platformauth.UnrestrictedTeamAccess())
	require.NoError(t, err)
	readOnly, err = readOnly.WithWorkspace(workspaceID)
	require.NoError(t, err)
	scope.Actor = readOnly
	_, err = service.SetReaction(t.Context(), ReactionCommand{Scope: scope, Emoji: "👍", Active: true})
	require.ErrorIs(t, err, ErrForbidden)
	_, err = service.SetResolution(t.Context(), ResolutionCommand{Scope: scope, Resolved: true})
	require.ErrorIs(t, err, ErrForbidden)
	_, err = service.GetStoryState(t.Context(), StoryScope{StoryID: uuid.New(), WorkspaceID: uuid.New(), Actor: actor})
	require.ErrorIs(t, err, ErrForbidden)
}
