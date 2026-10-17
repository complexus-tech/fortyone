//go:build integration

package chatrepository

import (
	"encoding/json"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNotificationContentUsesCurrentMessageAndSourceWithinWorkspace(t *testing.T) {
	f := newFixture(t)
	rich := json.RawMessage(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Please review","marks":[{"type":"bold"}]}]}]}`)
	message, err := f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{
		Scope: f.scope, ConversationID: f.room, CreationID: uuid.New(), Text: "Please review", RichContent: rich,
	})
	require.NoError(t, err)
	event := f.eventFor(t, message.ID)
	content, err := f.repo.NotificationContent(f.ctx, event)
	require.NoError(t, err)
	require.Equal(t, "team", content.ConversationKind)
	require.Equal(t, "Product", content.ConversationName)
	require.Equal(t, "Person "+f.author.String(), content.AuthorName)
	require.Equal(t, "Please review", content.Text)
	require.JSONEq(t, string(rich), string(content.RichContent))
	require.Empty(t, content.Attachments)

	foreign := event
	foreign.WorkspaceID = uuid.New()
	_, err = f.repo.NotificationContent(f.ctx, foreign)
	require.ErrorIs(t, err, domain.ErrNotFound)

	_, err = f.service.UpdateMessage(f.ctx, domain.UpdateMessageCommand{Scope: f.scope, MessageID: message.ID, Version: message.Version, Text: "Updated before delivery"})
	require.NoError(t, err)
	content, err = f.repo.NotificationContent(f.ctx, event)
	require.NoError(t, err)
	require.Equal(t, "Updated before delivery", content.Text)

	_, err = f.pool.Exec(f.ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.team, f.peer)
	require.NoError(t, err)
	_, err = f.repo.NotificationContent(f.ctx, event)
	require.ErrorIs(t, err, domain.ErrNotFound, "no remaining authorized recipient must suppress message content")
}

func TestEveryIncomingMessageSnapshotsMembersAndRechecksReadBoundaries(t *testing.T) {
	f := newFixture(t)
	root := f.create(t, "Ordinary team message", nil)
	reply := f.create(t, "Reply without prior participation", &root.ID)
	rootEvent, replyEvent := f.eventFor(t, root.ID), f.eventFor(t, reply.ID)
	for _, test := range []struct {
		event  domain.Event
		reason string
	}{{rootEvent, "message"}, {replyEvent, "message"}} {
		recipients, err := f.repo.NotificationRecipients(f.ctx, test.event)
		require.NoError(t, err)
		require.Equal(t, []domain.NotificationRecipient{{UserID: f.peer, Reason: test.reason}}, recipients)
	}
	_, err := f.pool.Exec(f.ctx, `INSERT INTO team_members(team_id,user_id) VALUES($1,$2)`, f.team, f.outsider)
	require.NoError(t, err)
	recipients, err := f.repo.NotificationRecipients(f.ctx, rootEvent)
	require.NoError(t, err)
	require.Equal(t, []domain.NotificationRecipient{{UserID: f.peer, Reason: "message"}}, recipients, "joining later must not replay earlier interruptions")
	peer := humanScope(t, f.workspace, f.peer)
	_, err = f.service.MarkRead(f.ctx, domain.ReadCommand{Scope: peer, ConversationID: f.room, ThroughMessageID: root.ID})
	require.NoError(t, err)
	recipients, err = f.repo.NotificationRecipients(f.ctx, rootEvent)
	require.NoError(t, err)
	require.Empty(t, recipients)
	recipients, err = f.repo.NotificationRecipients(f.ctx, replyEvent)
	require.NoError(t, err)
	require.Len(t, recipients, 1, "conversation reads cannot suppress an unread reply")
	_, err = f.service.MarkRead(f.ctx, domain.ReadCommand{Scope: peer, ConversationID: f.room, ThreadRootID: &root.ID, ThroughMessageID: reply.ID})
	require.NoError(t, err)
	recipients, err = f.repo.NotificationRecipients(f.ctx, replyEvent)
	require.NoError(t, err)
	require.Empty(t, recipients)
}

func TestMessageDeliverySuppressesDeletionAndOwnMessages(t *testing.T) {
	f := newFixture(t)
	message := f.create(t, "Delete before delivery", nil)
	event := f.eventFor(t, message.ID)
	_, err := f.service.DeleteMessage(f.ctx, domain.DeleteMessageCommand{Scope: f.scope, MessageID: message.ID, Version: message.Version})
	require.NoError(t, err)
	recipients, err := f.repo.NotificationRecipients(f.ctx, event)
	require.NoError(t, err)
	require.Empty(t, recipients)
	message = f.create(t, "Own message", nil)
	event = f.eventFor(t, message.ID)
	_, err = f.pool.Exec(f.ctx, `INSERT INTO chat_outbox_recipients(event_id,user_id,reason) VALUES($1,$2,'message')`, event.ID, f.author)
	require.NoError(t, err)
	recipients, err = f.repo.NotificationRecipients(f.ctx, event)
	require.NoError(t, err)
	for _, recipient := range recipients {
		require.NotEqual(t, f.author, recipient.UserID)
		require.NotEqual(t, uuid.Nil, recipient.UserID)
	}
}

func TestParticipantRepliesKeepTheirExistingTargetedDeliveryReason(t *testing.T) {
	f := newFixture(t)
	root := f.create(t, "Thread", nil)
	_, err := f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{
		Scope: humanScope(t, f.workspace, f.peer), ConversationID: f.room,
		CreationID: uuid.New(), ThreadRootID: &root.ID, Text: "Following this thread",
	})
	require.NoError(t, err)
	reply := f.create(t, "Reply to participant", &root.ID)
	recipients, err := f.repo.NotificationRecipients(f.ctx, f.eventFor(t, reply.ID))
	require.NoError(t, err)
	require.Equal(t, []domain.NotificationRecipient{{UserID: f.peer, Reason: "reply"}}, recipients)
}
