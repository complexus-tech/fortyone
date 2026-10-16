//go:build integration

package commentsrepository

import (
	"io"
	"log/slog"
	"testing"
	"time"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestThreadReadMarkersPreserveUnopenedRepliesAcrossVisits(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspaceID, _, storyID := insertCommentTestWorkspace(t, ctx, postgres.Pool, "thread-reads")
	authorID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "author")
	readerID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "reader")
	author := commentIntegrationActor(t, authorID, workspaceID)
	reader := commentIntegrationActor(t, readerID, workspaceID)
	service := comments.New(New(logger.NewWithText(io.Discard, slog.LevelError, "thread-reads-test"), postgres.Pool))
	create := func(parentID *uuid.UUID, content string) comments.CoreComment {
		t.Helper()
		comment, err := service.CreateComment(ctx, comments.CreateCommentCommand{WorkspaceID: workspaceID, StoryID: storyID, Actor: author, ParentID: parentID, Content: content})
		require.NoError(t, err)
		return comment
	}
	first := create(nil, "First thread")
	firstReply := create(&first.ID, "First reply")
	nested := create(&firstReply.ID, "Nested reply")
	second := create(nil, "Unopened thread")
	create(&second.ID, "Unopened reply")
	_, err := service.CreateComment(ctx, comments.CreateCommentCommand{WorkspaceID: workspaceID, StoryID: storyID, Actor: reader, ParentID: &first.ID, Content: "My own reply"})
	require.NoError(t, err)
	scope := comments.StoryScope{StoryID: storyID, WorkspaceID: workspaceID, Actor: reader}
	initial, err := service.GetStoryState(ctx, scope)
	require.NoError(t, err)
	require.Equal(t, int64(5), initial.UnreadCount, "roots and every foreign descendant count once")
	require.Empty(t, initial.ThreadReads)

	rootRead, err := service.MarkRead(ctx, comments.ReadCommand{Scope: scope, ObservedAt: initial.ObservedAt})
	require.NoError(t, err)
	require.Equal(t, int64(3), rootRead.UnreadCount, "reading the story must leave collapsed replies unread")
	require.Empty(t, rootRead.ThreadReads)
	opened, err := service.MarkRead(ctx, comments.ReadCommand{Scope: scope, ThreadID: &first.ID, ObservedAt: initial.ObservedAt})
	require.NoError(t, err)
	require.Equal(t, int64(1), opened.UnreadCount, "opening one thread includes its deep descendants but not another thread")
	require.Len(t, opened.ThreadReads, 1)
	require.WithinDuration(t, initial.ObservedAt, opened.ThreadReads[first.ID], time.Microsecond)
	visited, err := service.GetStoryState(ctx, scope)
	require.NoError(t, err)
	require.Equal(t, opened.ThreadReads, visited.ThreadReads)
	require.Equal(t, int64(1), visited.UnreadCount, "closing and revisiting cannot resurrect already viewed replies")

	create(&nested.ID, "Arrived after opening")
	afterArrival, err := service.GetStoryState(ctx, scope)
	require.NoError(t, err)
	require.Equal(t, int64(2), afterArrival.UnreadCount)
	stale, err := service.MarkRead(ctx, comments.ReadCommand{Scope: scope, ThreadID: &first.ID, ObservedAt: initial.ObservedAt.Add(-time.Hour)})
	require.NoError(t, err)
	require.Equal(t, opened.ThreadReads, stale.ThreadReads, "old devices cannot move a thread marker backwards")
	require.Equal(t, int64(2), stale.UnreadCount, "the observed cutoff cannot consume a later reply")
	opened, err = service.MarkRead(ctx, comments.ReadCommand{Scope: scope, ThreadID: &first.ID, ObservedAt: afterArrival.ObservedAt})
	require.NoError(t, err)
	require.Equal(t, int64(1), opened.UnreadCount)
	require.Equal(t, rootRead.LastReadAt, opened.LastReadAt, "opening replies does not advance top-level reads")
	otherReader, err := service.GetStoryState(ctx, comments.StoryScope{StoryID: storyID, WorkspaceID: workspaceID, Actor: author})
	require.NoError(t, err)
	require.Empty(t, otherReader.ThreadReads, "thread markers belong to one reader")
	require.Nil(t, otherReader.LastReadAt)
	require.Equal(t, int64(1), otherReader.UnreadCount, "the reader's own reply remains unread for the other author")
	allRead, err := service.MarkRead(ctx, comments.ReadCommand{Scope: scope, ThreadID: &second.ID, ObservedAt: time.Now().Add(24 * time.Hour)})
	require.NoError(t, err)
	require.Equal(t, int64(0), allRead.UnreadCount)
	require.False(t, allRead.ThreadReads[second.ID].After(allRead.ObservedAt), "future client times are bounded by the server")
	create(&second.ID, "Arrived after future cutoff was clamped")
	newState, err := service.GetStoryState(ctx, scope)
	require.NoError(t, err)
	require.Equal(t, int64(1), newState.UnreadCount)

	require.NoError(t, service.DeleteComment(ctx, comments.AuthorScope{CommentID: first.ID, WorkspaceID: workspaceID, Actor: author}))
	deleted, err := service.GetStoryState(ctx, scope)
	require.NoError(t, err)
	require.NotContains(t, deleted.ThreadReads, first.ID)
	require.Contains(t, deleted.ThreadReads, second.ID)
}

func TestThreadReadMarkersRequireCanonicalVisibleStoryRoot(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspaceID, teamID, storyID := insertCommentTestWorkspace(t, ctx, postgres.Pool, "thread-scope")
	foreignWorkspace, foreignTeam, foreignStory := insertCommentTestWorkspace(t, ctx, postgres.Pool, "thread-foreign")
	authorID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "author")
	foreignID := insertCommentTestUser(t, ctx, postgres.Pool, foreignWorkspace, "foreign")
	author := commentIntegrationActor(t, authorID, workspaceID)
	foreign := commentIntegrationActor(t, foreignID, foreignWorkspace)
	service := comments.New(New(logger.NewWithText(io.Discard, slog.LevelError, "thread-scope-test"), postgres.Pool))
	root, err := service.CreateComment(ctx, comments.CreateCommentCommand{WorkspaceID: workspaceID, StoryID: storyID, Actor: author, Content: "Root"})
	require.NoError(t, err)
	reply, err := service.CreateComment(ctx, comments.CreateCommentCommand{WorkspaceID: workspaceID, StoryID: storyID, Actor: author, ParentID: &root.ID, Content: "Reply"})
	require.NoError(t, err)
	foreignRoot, err := service.CreateComment(ctx, comments.CreateCommentCommand{WorkspaceID: foreignWorkspace, StoryID: foreignStory, Actor: foreign, Content: "Foreign"})
	require.NoError(t, err)
	otherStory := uuid.New()
	_, err = postgres.Pool.Exec(ctx, `INSERT INTO stories(id,workspace_id,team_id,title,sequence_id) VALUES($1,$2,$3,'Other story',2)`, otherStory, workspaceID, teamID)
	require.NoError(t, err)
	otherRoot, err := service.CreateComment(ctx, comments.CreateCommentCommand{WorkspaceID: workspaceID, StoryID: otherStory, Actor: author, Content: "Other story"})
	require.NoError(t, err)
	scope := comments.StoryScope{StoryID: storyID, WorkspaceID: workspaceID, Actor: author}
	for _, threadID := range []uuid.UUID{reply.ID, foreignRoot.ID, otherRoot.ID, uuid.New()} {
		_, err := service.MarkRead(ctx, comments.ReadCommand{Scope: scope, ThreadID: &threadID, ObservedAt: time.Now()})
		require.ErrorIs(t, err, comments.ErrNotFound, "only this visible story's canonical roots may be marked")
	}
	restricted := scope
	restricted.Actor = restrictedCommentActor(t, authorID, workspaceID, foreignTeam)
	_, err = service.MarkRead(ctx, comments.ReadCommand{Scope: restricted, ThreadID: &root.ID, ObservedAt: time.Now()})
	require.ErrorIs(t, err, comments.ErrNotFound)
	_, err = postgres.Pool.Exec(ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, teamID, authorID)
	require.NoError(t, err)
	_, err = service.MarkRead(ctx, comments.ReadCommand{Scope: scope, ThreadID: &root.ID, ObservedAt: time.Now()})
	require.ErrorIs(t, err, comments.ErrNotFound)
	_, err = postgres.Pool.Exec(ctx, `INSERT INTO team_members(team_id,user_id) VALUES($1,$2)`, teamID, authorID)
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `UPDATE stories SET deleted_at=clock_timestamp() WHERE id=$1`, storyID)
	require.NoError(t, err)
	_, err = service.MarkRead(ctx, comments.ReadCommand{Scope: scope, ThreadID: &root.ID, ObservedAt: time.Now()})
	require.ErrorIs(t, err, comments.ErrNotFound)
	var count int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT count(*) FROM story_comment_thread_reads WHERE user_id=$1`, authorID).Scan(&count))
	require.Zero(t, count, "rejected requests must not persist read markers")
}
