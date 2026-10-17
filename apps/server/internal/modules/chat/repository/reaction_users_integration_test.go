//go:build integration

package chatrepository

import (
	"testing"

	"github.com/complexus-tech/projects-api/internal/migrations"
	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestReactionUsersMigrationRoundTripPreservesReactions(t *testing.T) {
	f := newFixture(t)
	message := f.create(t, "Keep this reaction", nil)
	_, err := f.service.SetReaction(f.ctx, f.scope, message.ID, "👍", true)
	require.NoError(t, err)
	for _, direction := range []string{"down", "up"} {
		migration, err := migrations.FS.ReadFile("000236_chat_reaction_users." + direction + ".sql")
		require.NoError(t, err)
		_, err = f.pool.Exec(f.ctx, string(migration))
		require.NoError(t, err)
		updated, err := f.service.GetMessage(f.ctx, f.scope, message.ID)
		require.NoError(t, err)
		require.Len(t, updated.Reactions, 1)
		require.Equal(t, 1, updated.Reactions[0].Count)
		require.True(t, updated.Reactions[0].Reacted)
		if direction == "down" {
			require.Empty(t, updated.Reactions[0].Users)
		} else {
			require.Equal(t, []domain.Person{{ID: f.author, FullName: "Person " + f.author.String()}}, updated.Reactions[0].Users)
		}
	}
}

func TestReactionUsersFollowEmojiMembershipAndMessageAccess(t *testing.T) {
	f := newFixture(t)
	_, err := f.pool.Exec(f.ctx, `UPDATE users SET full_name=CASE user_id WHEN $1 THEN 'Zara' ELSE 'Alex' END WHERE user_id IN ($1,$2)`, f.author, f.peer)
	require.NoError(t, err)
	message := f.create(t, "React to this", nil)
	peer := humanScope(t, f.workspace, f.peer)
	_, err = f.service.SetReaction(f.ctx, f.scope, message.ID, "👍", true)
	require.NoError(t, err)
	_, err = f.service.SetReaction(f.ctx, peer, message.ID, "👍", true)
	require.NoError(t, err)
	_, err = f.service.SetReaction(f.ctx, peer, message.ID, "❤️", true)
	require.NoError(t, err)
	updated, err := f.service.GetMessage(f.ctx, peer, message.ID)
	require.NoError(t, err)
	byEmoji := make(map[string]domain.Reaction)
	for _, reaction := range updated.Reactions {
		byEmoji[reaction.Emoji] = reaction
	}
	require.Equal(t, []domain.Person{{ID: f.peer, FullName: "Alex"}, {ID: f.author, FullName: "Zara"}}, byEmoji["👍"].Users)
	require.Equal(t, 2, byEmoji["👍"].Count)
	require.True(t, byEmoji["👍"].Reacted)
	require.Equal(t, []domain.Person{{ID: f.peer, FullName: "Alex"}}, byEmoji["❤️"].Users)

	updated, err = f.service.SetReaction(f.ctx, peer, message.ID, "👍", false)
	require.NoError(t, err)
	for _, reaction := range updated.Reactions {
		if reaction.Emoji == "👍" {
			require.Equal(t, []domain.Person{{ID: f.author, FullName: "Zara"}}, reaction.Users)
			require.Equal(t, 1, reaction.Count)
			require.False(t, reaction.Reacted)
		}
	}
	_, err = f.service.GetMessage(f.ctx, humanScope(t, f.workspace, f.outsider), message.ID)
	require.ErrorIs(t, err, domain.ErrNotFound)
	_, err = f.service.GetMessage(f.ctx, humanScope(t, uuid.New(), f.peer), message.ID)
	require.ErrorIs(t, err, domain.ErrNotFound)
	_, err = f.pool.Exec(f.ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.team, f.peer)
	require.NoError(t, err)
	_, err = f.service.GetMessage(f.ctx, peer, message.ID)
	require.ErrorIs(t, err, domain.ErrNotFound)
	_, err = f.service.DeleteMessage(f.ctx, domain.DeleteMessageCommand{Scope: f.scope, MessageID: message.ID, Version: message.Version})
	require.NoError(t, err)
	_, err = f.service.GetMessage(f.ctx, f.scope, message.ID)
	require.ErrorIs(t, err, domain.ErrNotFound)
	var remaining int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT jsonb_array_length(public.chat_message_json($1,$2)->'reactions')`, message.ID, f.author).Scan(&remaining))
	require.Zero(t, remaining)
}
