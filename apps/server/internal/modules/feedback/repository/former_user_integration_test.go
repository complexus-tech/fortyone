//go:build integration

package feedbackrepository

import (
	"testing"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFormerFeedbackAuthorReadPathsPreserveAttributionWithoutPrivateIdentity(t *testing.T) {
	ctx := t.Context()
	postgres := testkit.NewPostgres(t)
	f := newFeedbackIntegrationFixture(t, ctx, postgres.Pool)
	repo := New(nil, postgres.Pool)
	formerID := uuid.MustParse("ffffffff-ffff-4fff-8fff-ffffffffffff")
	feedbackIntegrationExec(t, ctx, postgres.Pool,
		`UPDATE feedback_contributors SET user_id=NULL, kind='anonymous', display_name=NULL, email=NULL, avatar_url=NULL WHERE id=(SELECT contributor_id FROM feedback_items WHERE id=$1)`, f.itemA)
	feedbackIntegrationExec(t, ctx, postgres.Pool,
		`UPDATE feedback_items SET author_id=$2 WHERE id=$1`, f.itemA, formerID)
	scope := feedback.CoreAccessScope{WorkspaceID: f.workspaceA, ActorID: f.actorA, AllTeams: true}
	author, err := repo.GetPrivateAuthorScoped(ctx, scope, f.itemA)
	require.NoError(t, err)
	require.Equal(t, "Former user", author.DisplayName)
	require.Nil(t, author.UserID)
	require.Nil(t, author.Email)
	require.Nil(t, author.AvatarURL)
	items, err := repo.ListSimilarItems(ctx, f.portalA, "Feedback item a", "", 10)
	require.NoError(t, err)
	require.NotEmpty(t, items)
	require.Equal(t, "Former user", items[0].AuthorName)
	require.Nil(t, items[0].AuthorID)
	require.Nil(t, items[0].AuthorAvatar)

	// A genuine anonymous contribution is not a deleted-account placeholder.
	feedbackIntegrationExec(t, ctx, postgres.Pool, `UPDATE feedback_items SET author_id=NULL WHERE id=$1`, f.itemA)
	author, err = repo.GetPrivateAuthorScoped(ctx, scope, f.itemA)
	require.NoError(t, err)
	require.Equal(t, "Anonymous", author.DisplayName)
}
