package feedbackrepository

import (
	"context"
	"testing"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestScopedCoreWritesRejectMismatchedActorBeforeDatabaseAccess(t *testing.T) {
	t.Parallel()
	repo := &Repo{}
	scope := feedback.CoreAccessScope{WorkspaceID: uuid.New(), ActorID: uuid.New(), AllTeams: true}

	_, err := repo.CreateCommentScoped(context.Background(), scope, feedback.CoreCommentInput{
		WorkspaceID: scope.WorkspaceID,
		ItemID:      uuid.New(),
		AuthorID:    uuid.New(),
		Body:        "hello",
	})
	require.ErrorIs(t, err, feedback.ErrForbidden)

	_, err = repo.LinkStoryScoped(context.Background(), scope, feedback.CoreStoryLinkInput{
		WorkspaceID:     scope.WorkspaceID,
		ItemID:          uuid.New(),
		StoryID:         uuid.New(),
		CreatedByUserID: uuid.New(),
	})
	require.ErrorIs(t, err, feedback.ErrForbidden)
}

func TestListItemsScopedRejectsWorkspaceMismatchBeforeDatabaseAccess(t *testing.T) {
	t.Parallel()
	repo := &Repo{}
	scope := feedback.CoreAccessScope{WorkspaceID: uuid.New(), ActorID: uuid.New(), AllTeams: true}

	_, err := repo.ListItemsScoped(context.Background(), scope, feedback.CoreListItemsInput{
		WorkspaceID: uuid.New(),
		Page:        1,
		PageSize:    20,
	})

	require.ErrorIs(t, err, feedback.ErrForbidden)
}
