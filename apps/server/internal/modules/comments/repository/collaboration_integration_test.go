//go:build integration

package commentsrepository

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCommentCollaborationIsScopedAndIdempotent(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	workspaceID, teamID, storyID := insertCommentTestWorkspace(t, ctx, postgres.Pool, "collaboration")
	otherWorkspace, otherTeam, _ := insertCommentTestWorkspace(t, ctx, postgres.Pool, "foreign")
	authorID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "author")
	readerID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "reader")
	foreignID := insertCommentTestUser(t, ctx, postgres.Pool, otherWorkspace, "foreign")
	service := comments.New(New(logger.NewWithText(io.Discard, slog.LevelError, "collaboration-test"), postgres.Pool))
	author := commentIntegrationActor(t, authorID, workspaceID)
	reader := commentIntegrationActor(t, readerID, workspaceID)
	root, err := service.CreateComment(ctx, comments.CreateCommentCommand{WorkspaceID: workspaceID, StoryID: storyID, Actor: author, Content: "Thread"})
	require.NoError(t, err)
	reply, err := service.CreateComment(ctx, comments.CreateCommentCommand{WorkspaceID: workspaceID, StoryID: storyID, ParentID: &root.ID, Actor: reader, Content: "Reply"})
	require.NoError(t, err)
	authorScope := comments.AuthorScope{CommentID: root.ID, WorkspaceID: workspaceID, Actor: author}
	readerScope := comments.AuthorScope{CommentID: root.ID, WorkspaceID: workspaceID, Actor: reader}

	for range 2 {
		state, err := service.SetReaction(ctx, comments.ReactionCommand{Scope: authorScope, Emoji: "👍", Active: true})
		require.NoError(t, err)
		require.Len(t, state.Reactions, 1)
		require.Equal(t, int64(1), state.Reactions[0].Count)
		require.True(t, state.Reactions[0].Reacted)
	}
	state, err := service.SetReaction(ctx, comments.ReactionCommand{Scope: readerScope, Emoji: "👍", Active: true})
	require.NoError(t, err)
	require.Equal(t, int64(2), state.Reactions[0].Count)
	for range 2 {
		state, err = service.SetReaction(ctx, comments.ReactionCommand{Scope: authorScope, Emoji: "👍", Active: false})
		require.NoError(t, err)
		require.Equal(t, int64(1), state.Reactions[0].Count)
		require.False(t, state.Reactions[0].Reacted)
	}
	state, err = service.SetResolution(ctx, comments.ResolutionCommand{Scope: readerScope, Resolved: true})
	require.NoError(t, err)
	require.NotNil(t, state.ResolvedAt)
	require.Equal(t, &readerID, state.ResolvedBy)
	resolvedAt := *state.ResolvedAt
	state, err = service.SetResolution(ctx, comments.ResolutionCommand{Scope: authorScope, Resolved: true})
	require.NoError(t, err)
	require.Equal(t, resolvedAt, *state.ResolvedAt, "retries must preserve the original resolver")
	require.Equal(t, &readerID, state.ResolvedBy)
	state, err = service.SetResolution(ctx, comments.ResolutionCommand{Scope: authorScope, Resolved: false})
	require.NoError(t, err)
	require.Nil(t, state.ResolvedAt)
	require.Nil(t, state.ResolvedBy)
	replyScope := authorScope
	replyScope.CommentID = reply.ID
	_, err = service.SetResolution(ctx, comments.ResolutionCommand{Scope: replyScope, Resolved: true})
	require.ErrorIs(t, err, comments.ErrNotFound, "only roots may be resolved")

	storyScope := comments.StoryScope{StoryID: storyID, WorkspaceID: workspaceID, Actor: author}
	initial, err := service.GetStoryState(ctx, storyScope)
	require.NoError(t, err)
	require.Nil(t, initial.LastReadAt)
	require.Equal(t, int64(1), initial.UnreadCount, "own comments are not unread")
	require.Equal(t, &reply.ID, initial.LatestCommentID, "latest comment includes replies, independently of root pagination")
	newest, err := service.CreateComment(ctx, comments.CreateCommentCommand{WorkspaceID: workspaceID, StoryID: storyID, Actor: reader, Content: "Arrived after initial read"})
	require.NoError(t, err)
	read, err := service.MarkRead(ctx, comments.ReadCommand{Scope: storyScope, ObservedAt: initial.ObservedAt})
	require.NoError(t, err)
	require.Equal(t, int64(2), read.UnreadCount, "root cutoff must preserve newly arrived roots and unopened replies")
	read, err = service.MarkRead(ctx, comments.ReadCommand{Scope: storyScope, ObservedAt: initial.ObservedAt, ThreadID: &root.ID})
	require.NoError(t, err)
	require.Equal(t, int64(1), read.UnreadCount, "opening the thread clears only replies observed before its cutoff")
	require.Equal(t, &newest.ID, read.LatestCommentID)
	require.WithinDuration(t, initial.ObservedAt, *read.LastReadAt, time.Microsecond)
	read, err = service.MarkRead(ctx, comments.ReadCommand{Scope: storyScope, ObservedAt: initial.ObservedAt.Add(-time.Hour)})
	require.NoError(t, err)
	require.WithinDuration(t, initial.ObservedAt, *read.LastReadAt, time.Microsecond, "older devices must not move read markers backwards")
	readerState, err := service.GetStoryState(ctx, comments.StoryScope{StoryID: storyID, WorkspaceID: workspaceID, Actor: reader})
	require.NoError(t, err)
	require.Nil(t, readerState.LastReadAt, "read state belongs to one user")
	read, err = service.MarkRead(ctx, comments.ReadCommand{Scope: storyScope, ObservedAt: time.Now().Add(24 * time.Hour)})
	require.NoError(t, err)
	require.False(t, read.LastReadAt.After(read.ObservedAt), "untrusted future cutoffs must be clamped")

	foreignScope := comments.AuthorScope{CommentID: root.ID, WorkspaceID: otherWorkspace, Actor: commentIntegrationActor(t, foreignID, otherWorkspace)}
	_, err = service.SetReaction(ctx, comments.ReactionCommand{Scope: foreignScope, Emoji: "✅", Active: true})
	require.ErrorIs(t, err, comments.ErrNotFound)
	restrictedScope := authorScope
	restrictedScope.Actor = restrictedCommentActor(t, authorID, workspaceID, otherTeam)
	_, err = service.SetResolution(ctx, comments.ResolutionCommand{Scope: restrictedScope, Resolved: true})
	require.ErrorIs(t, err, comments.ErrNotFound)
	_, err = postgres.Pool.Exec(ctx, "DELETE FROM public.team_members WHERE team_id = $1 AND user_id = $2", teamID, authorID)
	require.NoError(t, err)
	_, err = service.GetStoryState(ctx, storyScope)
	require.ErrorIs(t, err, comments.ErrNotFound)
	_, err = service.MarkRead(ctx, comments.ReadCommand{Scope: storyScope, ObservedAt: initial.ObservedAt})
	require.ErrorIs(t, err, comments.ErrNotFound)
	_, err = service.SetReaction(ctx, comments.ReactionCommand{Scope: authorScope, Emoji: "👀", Active: true})
	require.ErrorIs(t, err, comments.ErrNotFound)

	_, err = postgres.Pool.Exec(ctx, "INSERT INTO public.team_members (team_id, user_id) VALUES ($1, $2)", teamID, authorID)
	require.NoError(t, err)
	err = service.DeleteComment(ctx, authorScope)
	require.NoError(t, err)
	var reactionCount int
	require.NoError(t, postgres.Pool.QueryRow(ctx, "SELECT count(*) FROM public.story_comment_reactions WHERE comment_id = $1", root.ID).Scan(&reactionCount))
	require.Zero(t, reactionCount)
}

func TestStoryCommentStateLatestHasStableTieBreakAndEmptyState(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	workspaceID, _, storyID := insertCommentTestWorkspace(t, ctx, postgres.Pool, "latest")
	actorID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "latest-reader")
	service := comments.New(New(logger.NewWithText(io.Discard, slog.LevelError, "latest-comment-test"), postgres.Pool))
	scope := comments.StoryScope{StoryID: storyID, WorkspaceID: workspaceID, Actor: commentIntegrationActor(t, actorID, workspaceID)}
	state, err := service.GetStoryState(ctx, scope)
	require.NoError(t, err)
	require.Nil(t, state.LatestCommentID)
	require.NotNil(t, state.Comments)
	require.NotNil(t, state.ThreadReads)
	require.Empty(t, state.ThreadReads)

	firstID, secondID := uuid.New(), uuid.New()
	_, err = postgres.Pool.Exec(ctx, `
		INSERT INTO public.story_comments (comment_id, story_id, commenter_id, content, created_at)
		VALUES ($1, $3, $4, 'First', $5), ($2, $3, $4, 'Second', $5)
	`, firstID, secondID, storyID, actorID, time.Now().UTC())
	require.NoError(t, err)
	state, err = service.GetStoryState(ctx, scope)
	require.NoError(t, err)
	want := firstID
	if strings.Compare(secondID.String(), firstID.String()) > 0 {
		want = secondID
	}
	require.Equal(t, &want, state.LatestCommentID)
}
