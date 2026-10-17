//go:build integration

package commentsrepository

import (
	"testing"
	"time"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMarkSeenRequiresExactVisibleStoryAndKeepsFirstAcknowledgment(t *testing.T) {
	t.Parallel()
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspaceID, teamID, storyID := insertCommentTestWorkspace(t, ctx, postgres.Pool, "seen")
	foreignWorkspace, foreignTeam, foreignStory := insertCommentTestWorkspace(t, ctx, postgres.Pool, "seen-foreign")
	userID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "reader")
	actor := commentIntegrationActor(t, userID, workspaceID)
	service := comments.New(New(nil, postgres.Pool))
	comment, err := service.CreateComment(ctx, comments.CreateCommentCommand{WorkspaceID: workspaceID, StoryID: storyID, Actor: actor, Content: "Exact reply"})
	require.NoError(t, err)
	scope := comments.StoryScope{WorkspaceID: workspaceID, StoryID: storyID, Actor: actor}
	require.NoError(t, service.MarkSeen(ctx, scope, comment.ID))
	var first, replay time.Time
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT seen_at FROM story_comment_seen WHERE comment_id=$1 AND user_id=$2`, comment.ID, userID).Scan(&first))
	require.NoError(t, service.MarkSeen(ctx, scope, comment.ID))
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT seen_at FROM story_comment_seen WHERE comment_id=$1 AND user_id=$2`, comment.ID, userID).Scan(&replay))
	require.Equal(t, first, replay)

	wrongStory := scope
	wrongStory.StoryID = foreignStory
	require.ErrorIs(t, service.MarkSeen(ctx, wrongStory, comment.ID), comments.ErrNotFound)
	wrongWorkspace := scope
	wrongWorkspace.WorkspaceID = foreignWorkspace
	wrongWorkspace.Actor = commentIntegrationActor(t, userID, foreignWorkspace)
	require.ErrorIs(t, service.MarkSeen(ctx, wrongWorkspace, comment.ID), comments.ErrNotFound)
	restricted := scope
	restricted.Actor = restrictedCommentActor(t, userID, workspaceID, foreignTeam)
	require.ErrorIs(t, service.MarkSeen(ctx, restricted, comment.ID), comments.ErrNotFound)
	require.ErrorIs(t, service.MarkSeen(ctx, scope, uuid.New()), comments.ErrNotFound)

	_, err = postgres.Pool.Exec(ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, teamID, userID)
	require.NoError(t, err)
	require.ErrorIs(t, service.MarkSeen(ctx, scope, comment.ID), comments.ErrNotFound, "a previous receipt cannot bypass revoked access")
	_, err = postgres.Pool.Exec(ctx, `INSERT INTO team_members(team_id,user_id) VALUES($1,$2)`, teamID, userID)
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `UPDATE stories SET deleted_at=clock_timestamp() WHERE id=$1`, storyID)
	require.NoError(t, err)
	require.ErrorIs(t, service.MarkSeen(ctx, scope, comment.ID), comments.ErrNotFound)
	_, err = postgres.Pool.Exec(ctx, `UPDATE stories SET deleted_at=NULL WHERE id=$1`, storyID)
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `UPDATE users SET is_active=FALSE WHERE user_id=$1`, userID)
	require.NoError(t, err)
	require.ErrorIs(t, service.MarkSeen(ctx, scope, comment.ID), comments.ErrNotFound)
	var count int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT count(*) FROM story_comment_seen`).Scan(&count))
	require.Equal(t, 1, count)
}
