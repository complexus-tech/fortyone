//go:build integration

package chatrepository

import (
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

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
