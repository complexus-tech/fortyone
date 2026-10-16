//go:build integration

package commentsrepository

import (
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"

	commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestExpandedReactionsRemainCanonicalScopedAndIdempotent(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspaceID, _, storyID := insertCommentTestWorkspace(t, ctx, postgres.Pool, "reaction-catalog")
	firstID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "first")
	secondID := insertCommentTestUser(t, ctx, postgres.Pool, workspaceID, "second")
	first := commentIntegrationActor(t, firstID, workspaceID)
	second := commentIntegrationActor(t, secondID, workspaceID)
	service := comments.New(New(logger.NewWithText(io.Discard, slog.LevelError, "reaction-test"), postgres.Pool))
	root, err := service.CreateComment(ctx, comments.CreateCommentCommand{WorkspaceID: workspaceID, StoryID: storyID, Actor: first, Content: "React here"})
	require.NoError(t, err)
	firstScope := comments.AuthorScope{CommentID: root.ID, WorkspaceID: workspaceID, Actor: first}
	secondScope := comments.AuthorScope{CommentID: root.ID, WorkspaceID: workspaceID, Actor: second}
	for _, emoji := range []string{
		// Expressions.
		"😀", "😃", "😄", "😁", "😆", "😅", "😂", "🤣", "😊", "😍", "🥰", "😎", "🤔", "🤨", "😐", "😬", "🙄", "😮", "😱", "😢", "😭", "😴", "🤯", "🥳", "🤩",
		// Feedback.
		"👍", "👎", "👏", "🙌", "🙏", "🤝", "💪", "👌", "✌️", "🤞", "🫡", "👀", "✅", "❌", "⚠️", "❤️", "💚", "💙", "💜", "💯",
		// Celebration.
		"🎉", "🎊", "🎈", "🎂", "🎁", "🏆", "🥇", "⭐", "🌟", "✨", "🔥", "💥", "🥂", "🍾", "🎶",
		// Work.
		"🚀", "💡", "🎯", "📌", "📍", "📝", "📋", "📚", "🔍", "🔎", "🛠️", "🔧", "⚙️", "🧪", "🐛", "💻", "📊", "📈", "⏳", "☕",
	} {
		state, err := service.SetReaction(ctx, comments.ReactionCommand{Scope: firstScope, Emoji: emoji, Active: true})
		require.NoError(t, err, emoji)
		require.Equal(t, []commentsdomain.Reaction{{Emoji: emoji, Count: 1, Reacted: true}}, state.Reactions)
		_, err = service.SetReaction(ctx, comments.ReactionCommand{Scope: firstScope, Emoji: emoji, Active: false})
		require.NoError(t, err)
	}
	for _, emoji := range []string{"arbitrary text", "❤", "👍🏽", "❤️❤️", "🛠", "⚙", "⚠", "✌"} {
		_, err := service.SetReaction(ctx, comments.ReactionCommand{Scope: firstScope, Emoji: emoji, Active: true})
		require.ErrorIs(t, err, comments.ErrInvalidComment)
		_, err = postgres.Pool.Exec(ctx, `INSERT INTO story_comment_reactions(comment_id,user_id,emoji) VALUES($1,$2,$3)`, root.ID, firstID, emoji)
		var constraint *pgconn.PgError
		require.True(t, errors.As(err, &constraint), "database must also reject %q", emoji)
		require.Equal(t, "23514", constraint.Code)
	}

	var group sync.WaitGroup
	failures := make(chan error, 8)
	for range 8 {
		group.Go(func() {
			_, err := service.SetReaction(ctx, comments.ReactionCommand{Scope: firstScope, Emoji: "❤️", Active: true})
			failures <- err
		})
	}
	group.Wait()
	close(failures)
	for err := range failures {
		require.NoError(t, err)
	}
	for range 2 {
		state, err := service.SetReaction(ctx, comments.ReactionCommand{Scope: secondScope, Emoji: "❤️", Active: true})
		require.NoError(t, err)
		require.Equal(t, []commentsdomain.Reaction{{Emoji: "❤️", Count: 2, Reacted: true}}, state.Reactions)
	}
	_, err = service.SetReaction(ctx, comments.ReactionCommand{Scope: firstScope, Emoji: "🎉", Active: true})
	require.NoError(t, err)
	for range 2 {
		state, err := service.SetReaction(ctx, comments.ReactionCommand{Scope: firstScope, Emoji: "❤️", Active: false})
		require.NoError(t, err)
		require.ElementsMatch(t, []commentsdomain.Reaction{{Emoji: "❤️", Count: 1, Reacted: false}, {Emoji: "🎉", Count: 1, Reacted: true}}, state.Reactions)
	}
	var count int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT count(*) FROM story_comment_reactions WHERE comment_id=$1 AND emoji='❤️'`, root.ID).Scan(&count))
	require.Equal(t, 1, count, "removing one user's emoji preserves the other user's reaction")
}
