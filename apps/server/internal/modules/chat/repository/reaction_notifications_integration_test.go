//go:build integration

package chatrepository

import (
	"testing"

	"github.com/complexus-tech/projects-api/internal/migrations"
	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func reactionEventFor(t *testing.T, f fixture, messageID, actorID uuid.UUID, emoji string) domain.Event {
	t.Helper()
	var event domain.Event
	require.NoError(t, f.pool.QueryRow(f.ctx, `
		SELECT id,workspace_id,conversation_id,message_id,actor_id,thread_root_id,event_type,created_at
		FROM chat_outbox WHERE message_id=$1 AND actor_id=$2 AND reaction_emoji=$3
		ORDER BY created_at DESC,id DESC LIMIT 1`, messageID, actorID, emoji).
		Scan(&event.ID, &event.WorkspaceID, &event.ConversationID, &event.MessageID, &event.ActorID, &event.ThreadRootID, &event.Type, &event.CreatedAt))
	return event
}

func TestReactionNotificationsTargetAuthorForReadMessagesAndReplies(t *testing.T) {
	f := newFixture(t)
	peer := humanScope(t, f.workspace, f.peer)
	root := f.create(t, "Root message", nil)
	reply := f.create(t, "Reply message", &root.ID)
	for _, message := range []domain.Message{root, reply} {
		_, err := f.service.MarkRead(f.ctx, domain.ReadCommand{
			Scope: f.scope, ConversationID: f.room, ThreadRootID: message.ThreadRootID, ThroughMessageID: message.ID,
		})
		require.NoError(t, err)
		_, err = f.service.SetReaction(f.ctx, peer, message.ID, "👍", true)
		require.NoError(t, err)
		event := reactionEventFor(t, f, message.ID, f.peer, "👍")
		require.Equal(t, message.ThreadRootID, event.ThreadRootID)
		recipients, err := f.repo.NotificationRecipients(f.ctx, event)
		require.NoError(t, err)
		require.Equal(t, []domain.NotificationRecipient{{UserID: f.author, Reason: "reaction"}}, recipients,
			"new reactions notify only the message author even when that message is already read")
		content, err := f.repo.NotificationContent(f.ctx, event)
		require.NoError(t, err)
		require.Equal(t, "👍", content.ReactionEmoji)
		require.Equal(t, "Person "+f.peer.String(), content.ActorName)
		require.Equal(t, "Person "+f.author.String(), content.AuthorName)
		require.Equal(t, message.Text, content.Text)
	}
}

func TestReactionNotificationIntentExcludesSelfDuplicateAndRemovedOccurrences(t *testing.T) {
	f := newFixture(t)
	peer := humanScope(t, f.workspace, f.peer)
	message := f.create(t, "Message", nil)
	_, err := f.service.SetReaction(f.ctx, f.scope, message.ID, "👍", true)
	require.NoError(t, err)
	self := reactionEventFor(t, f, message.ID, f.author, "👍")
	recipients, err := f.repo.NotificationRecipients(f.ctx, self)
	require.NoError(t, err)
	require.Empty(t, recipients)
	_, err = f.repo.NotificationContent(f.ctx, self)
	require.ErrorIs(t, err, domain.ErrNotFound)

	for range 2 {
		_, err = f.service.SetReaction(f.ctx, peer, message.ID, "👍", true)
		require.NoError(t, err)
	}
	first := reactionEventFor(t, f, message.ID, f.peer, "👍")
	var additions int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox WHERE actor_id=$1 AND reaction_emoji='👍'`, f.peer).Scan(&additions))
	require.Equal(t, 1, additions, "idempotent PUT must not create another notification intent")
	_, err = f.service.SetReaction(f.ctx, peer, message.ID, "👍", false)
	require.NoError(t, err)
	recipients, err = f.repo.NotificationRecipients(f.ctx, first)
	require.NoError(t, err)
	require.Empty(t, recipients, "removal before dispatch suppresses the original addition")
	_, err = f.repo.NotificationContent(f.ctx, first)
	require.ErrorIs(t, err, domain.ErrNotFound)
	var removalIntents int
	require.NoError(t, f.pool.QueryRow(f.ctx, `
		SELECT count(*) FROM chat_outbox event JOIN chat_outbox_recipients recipient ON recipient.event_id=event.id
		WHERE event.event_type='reaction.updated' AND event.reaction_emoji IS NULL`).Scan(&removalIntents))
	require.Zero(t, removalIntents)

	_, err = f.service.SetReaction(f.ctx, peer, message.ID, "👍", true)
	require.NoError(t, err)
	latest := reactionEventFor(t, f, message.ID, f.peer, "👍")
	require.NotEqual(t, first.ID, latest.ID)
	recipients, err = f.repo.NotificationRecipients(f.ctx, first)
	require.NoError(t, err)
	require.Empty(t, recipients, "a later re-add must not revive an earlier delivery intent")
	_, err = f.repo.NotificationContent(f.ctx, first)
	require.ErrorIs(t, err, domain.ErrNotFound)
	recipients, err = f.repo.NotificationRecipients(f.ctx, latest)
	require.NoError(t, err)
	require.Equal(t, []domain.NotificationRecipient{{UserID: f.author, Reason: "reaction"}}, recipients)
}

func TestReactionNotificationDeliveryRechecksAccessAndLegacyIntent(t *testing.T) {
	for _, scenario := range []string{"actor removed", "author removed", "message deleted", "legacy event", "foreign workspace"} {
		t.Run(scenario, func(t *testing.T) {
			f := newFixture(t)
			message := f.create(t, "Message", nil)
			_, err := f.service.SetReaction(f.ctx, humanScope(t, f.workspace, f.peer), message.ID, "👍", true)
			require.NoError(t, err)
			event := reactionEventFor(t, f, message.ID, f.peer, "👍")
			switch scenario {
			case "actor removed":
				_, err = f.pool.Exec(f.ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.team, f.peer)
			case "author removed":
				_, err = f.pool.Exec(f.ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.team, f.author)
			case "message deleted":
				_, err = f.service.DeleteMessage(f.ctx, domain.DeleteMessageCommand{Scope: f.scope, MessageID: message.ID, Version: message.Version})
			case "legacy event":
				_, err = f.pool.Exec(f.ctx, `UPDATE chat_outbox SET reaction_emoji=NULL WHERE id=$1`, event.ID)
			case "foreign workspace":
				event.WorkspaceID = uuid.New()
			}
			require.NoError(t, err)
			recipients, err := f.repo.NotificationRecipients(f.ctx, event)
			require.NoError(t, err)
			require.Empty(t, recipients)
			_, err = f.repo.NotificationContent(f.ctx, event)
			require.ErrorIs(t, err, domain.ErrNotFound)
		})
	}
}

func TestReactionNotificationFailureRollsBackReaction(t *testing.T) {
	f := newFixture(t)
	message := f.create(t, "Message", nil)
	_, err := f.pool.Exec(f.ctx, `
		CREATE FUNCTION reject_reaction_intent() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN RAISE EXCEPTION 'test outbox unavailable'; END; $$;
		CREATE TRIGGER reject_reaction_intent BEFORE INSERT ON chat_outbox_recipients
		FOR EACH ROW WHEN (NEW.reason='reaction') EXECUTE FUNCTION reject_reaction_intent();`)
	require.NoError(t, err)
	_, err = f.service.SetReaction(f.ctx, humanScope(t, f.workspace, f.peer), message.ID, "👍", true)
	require.ErrorContains(t, err, "test outbox unavailable")
	var reactions, events int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_message_reactions WHERE message_id=$1`, message.ID).Scan(&reactions))
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox WHERE message_id=$1 AND event_type='reaction.updated'`, message.ID).Scan(&events))
	require.Zero(t, reactions, "reaction and durable notification intent must commit together")
	require.Zero(t, events)
}

func TestDirectReactionNotificationAndMigrationRecovery(t *testing.T) {
	f := newFixture(t)
	dm, err := f.service.CreateDirect(f.ctx, f.scope, f.peer)
	require.NoError(t, err)
	message, err := f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{
		Scope: f.scope, ConversationID: dm.ID, CreationID: uuid.New(), Text: "Direct message",
	})
	require.NoError(t, err)
	_, err = f.service.SetReaction(f.ctx, humanScope(t, f.workspace, f.peer), message.ID, "👍", true)
	require.NoError(t, err)
	event := reactionEventFor(t, f, message.ID, f.peer, "👍")
	recipients, err := f.repo.NotificationRecipients(f.ctx, event)
	require.NoError(t, err)
	require.Equal(t, []domain.NotificationRecipient{{UserID: f.author, Reason: "reaction"}}, recipients)
	content, err := f.repo.NotificationContent(f.ctx, event)
	require.NoError(t, err)
	require.Equal(t, "dm", content.ConversationKind)

	down, err := migrations.FS.ReadFile("000237_chat_reaction_notifications.down.sql")
	require.NoError(t, err)
	_, err = f.pool.Exec(f.ctx, string(down))
	require.NoError(t, err)
	var reactions, events, intents int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_message_reactions WHERE message_id=$1`, message.ID).Scan(&reactions))
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox WHERE id=$1`, event.ID).Scan(&events))
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox_recipients WHERE event_id=$1`, event.ID).Scan(&intents))
	require.Equal(t, 1, reactions, "schema rollback preserves user reactions")
	require.Equal(t, 1, events, "schema rollback preserves realtime events")
	require.Zero(t, intents)

	up, err := migrations.FS.ReadFile("000237_chat_reaction_notifications.up.sql")
	require.NoError(t, err)
	_, err = f.pool.Exec(f.ctx, string(up))
	require.NoError(t, err)
	recipients, err = f.repo.NotificationRecipients(f.ctx, event)
	require.NoError(t, err)
	require.Empty(t, recipients, "reapplying the schema cannot create retroactive reaction notifications")
}
