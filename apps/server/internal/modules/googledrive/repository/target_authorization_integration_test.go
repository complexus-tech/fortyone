//go:build integration

package googledriverepository

import (
	"context"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGoogleDriveTargetAuthorizationOnPostgres18(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	storyID := uuid.New()
	authorID := uuid.New()
	teammateID := uuid.New()
	commentID := uuid.New()

	_, err := postgres.Pool.Exec(ctx, `
		INSERT INTO public.workspaces (workspace_id, name, slug)
		VALUES ($1, 'Drive comment authorization', $2)
	`, workspaceID, "drive-comment-auth-"+uuid.NewString())
	require.NoError(t, err)

	_, err = postgres.Pool.Exec(ctx, `
		INSERT INTO public.teams (team_id, name, workspace_id, code, color)
		VALUES ($1, 'Drive comment authorization', $2, $3, '#000000')
	`, teamID, workspaceID, "GDA"+uuid.NewString()[:6])
	require.NoError(t, err)

	_, err = postgres.Pool.Exec(ctx, `
		INSERT INTO public.stories (id, team_id, title, workspace_id)
		VALUES ($1, $2, 'Drive comment authorization', $3)
	`, storyID, teamID, workspaceID)
	require.NoError(t, err)

	for label, userID := range map[string]uuid.UUID{
		"author":   authorID,
		"teammate": teammateID,
	} {
		_, err = postgres.Pool.Exec(ctx, `
			INSERT INTO public.users (user_id, username, email, full_name)
			VALUES ($1, $2, $3, $4)
		`, userID, "drive-comment-"+label+"-"+userID.String(), label+"-"+userID.String()+"@example.com", "Drive comment "+label)
		require.NoError(t, err)
		_, err = postgres.Pool.Exec(ctx, `
			INSERT INTO public.workspace_members (workspace_id, user_id, role)
			VALUES ($1, $2, 'member')
		`, workspaceID, userID)
		require.NoError(t, err)
		_, err = postgres.Pool.Exec(ctx, `
			INSERT INTO public.team_members (team_id, user_id)
			VALUES ($1, $2)
		`, teamID, userID)
		require.NoError(t, err)
	}

	_, err = postgres.Pool.Exec(ctx, `
		INSERT INTO public.story_comments (comment_id, content, story_id, commenter_id)
		VALUES ($1, 'Author-owned comment', $2, $3)
	`, commentID, storyID, authorID)
	require.NoError(t, err)

	repository := New(postgres.Pool)
	teammateCanRead, err := repository.TargetAccessible(
		ctx, workspaceID, teammateID, domain.TargetComment, commentID,
	)
	require.NoError(t, err)
	require.True(t, teammateCanRead)

	authorCanMutate, err := repository.TargetMutable(
		ctx, workspaceID, authorID, domain.TargetComment, commentID,
	)
	require.NoError(t, err)
	require.True(t, authorCanMutate)

	teammateCanMutate, err := repository.TargetMutable(
		ctx, workspaceID, teammateID, domain.TargetComment, commentID,
	)
	require.NoError(t, err)
	require.False(t, teammateCanMutate)

	documentID := uuid.New()
	_, err = postgres.Pool.Exec(ctx, `
		INSERT INTO public.documents (
			document_id, workspace_id, title, visibility, created_by, updated_by
		) VALUES ($1, $2, 'Private Drive reference', 'restricted', $3, $3)
	`, documentID, workspaceID, authorID)
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `
		INSERT INTO public.document_members (document_id, user_id, role)
		VALUES ($1, $2, 'viewer')
	`, documentID, teammateID)
	require.NoError(t, err)

	restrictedMemberCanRead, err := repository.TargetAccessible(
		ctx, workspaceID, teammateID, domain.TargetDocument, documentID,
	)
	require.NoError(t, err)
	require.True(t, restrictedMemberCanRead)

	// Defend against stale sharing rows after a visibility transition. Native
	// document authorization ignores document_members whenever a document is
	// private, so Drive references must apply the same rule.
	_, err = postgres.Pool.Exec(ctx, `
		UPDATE public.documents SET visibility = 'private' WHERE document_id = $1
	`, documentID)
	require.NoError(t, err)
	staleMemberCanRead, err := repository.TargetAccessible(
		ctx, workspaceID, teammateID, domain.TargetDocument, documentID,
	)
	require.NoError(t, err)
	require.False(t, staleMemberCanRead)

	seedGoogleDriveLifecycleAccount(
		t, ctx, postgres.Pool, authorID, []uuid.UUID{workspaceID},
	)
	_, err = postgres.Pool.Exec(ctx, `
		UPDATE public.workspaces SET deleted_at = CURRENT_TIMESTAMP
		WHERE workspace_id = $1
	`, workspaceID)
	require.NoError(t, err)

	deletedWorkspaceTargetReadable, err := repository.TargetAccessible(
		ctx, workspaceID, authorID, domain.TargetComment, commentID,
	)
	require.NoError(t, err)
	require.False(t, deletedWorkspaceTargetReadable)
	deletedWorkspaceTargetMutable, err := repository.TargetMutable(
		ctx, workspaceID, authorID, domain.TargetComment, commentID,
	)
	require.NoError(t, err)
	require.False(t, deletedWorkspaceTargetMutable)

	_, err = repository.GetConnection(ctx, workspaceID, authorID)
	require.ErrorIs(t, err, domain.ErrNotFound)
}
