//go:build integration

package chatrepository

import (
	"context"
	"errors"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	service "github.com/complexus-tech/projects-api/internal/modules/chat/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMarkUnreadKeepsTimelineAndThreadsIndependent(t *testing.T) {
	f := newFixture(t)
	peer := humanScope(t, f.workspace, f.peer)
	first := f.create(t, "First root", nil)
	firstReply := f.create(t, "First reply", &first.ID)
	second := f.create(t, "Second root", nil)
	lastReply := f.create(t, "Last reply", &first.ID)
	last := f.create(t, "Last root", nil)
	_, err := f.service.MarkRead(f.ctx, domain.ReadCommand{Scope: peer, ConversationID: f.room, ThroughMessageID: last.ID})
	require.NoError(t, err)
	_, err = f.service.MarkRead(f.ctx, domain.ReadCommand{Scope: peer, ConversationID: f.room, ThreadRootID: &first.ID, ThroughMessageID: lastReply.ID})
	require.NoError(t, err)

	state, err := f.service.MarkUnread(f.ctx, domain.MarkUnreadCommand{Scope: peer, ConversationID: f.room, FromMessageID: second.ID})
	require.NoError(t, err)
	require.Nil(t, state.ThreadRootID)
	require.Equal(t, second.Sequence-1, state.LastReadSequence)
	require.Equal(t, 2, state.UnreadCount)
	require.Equal(t, 2, state.Summary.TotalUnread)
	require.Zero(t, state.Summary.UnreadThreads)
	require.Nil(t, state.FirstUnreadReplyID)
	require.Equal(t, []domain.ConversationUnread{{ConversationID: f.room, UnreadCount: 2, UnreadRootCount: 2}}, state.Summary.Conversations)

	state, err = f.service.MarkUnread(f.ctx, domain.MarkUnreadCommand{Scope: peer, ConversationID: f.room, ThreadRootID: &first.ID, FromMessageID: lastReply.ID})
	require.NoError(t, err)
	require.Equal(t, lastReply.Sequence-1, state.LastReadSequence)
	require.Equal(t, 1, state.UnreadCount)
	require.Equal(t, &lastReply.ID, state.FirstUnreadReplyID)
	require.Equal(t, 3, state.Summary.TotalUnread)
	require.Equal(t, 1, state.Summary.UnreadThreads)

	state, err = f.service.MarkUnread(f.ctx, domain.MarkUnreadCommand{Scope: peer, ConversationID: f.room, ThreadRootID: &first.ID, FromMessageID: firstReply.ID})
	require.NoError(t, err)
	require.Equal(t, 2, state.UnreadCount)
	require.Equal(t, &firstReply.ID, state.FirstUnreadReplyID)
	require.Equal(t, 4, state.Summary.TotalUnread)
	require.Equal(t, []domain.ConversationUnread{{ConversationID: f.room, UnreadCount: 4, UnreadRootCount: 2, UnreadReplyCount: 2}}, state.Summary.Conversations)

	var eventsBefore int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox WHERE event_type='read.updated' AND actor_id=$1`, f.peer).Scan(&eventsBefore))
	repeated, err := f.service.MarkUnread(f.ctx, domain.MarkUnreadCommand{Scope: peer, ConversationID: f.room, ThreadRootID: &first.ID, FromMessageID: lastReply.ID})
	require.NoError(t, err)
	require.Equal(t, state, repeated, "marking a later message must preserve earlier unread messages")
	var eventsAfter int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox WHERE event_type='read.updated' AND actor_id=$1`, f.peer).Scan(&eventsAfter))
	require.Equal(t, eventsBefore, eventsAfter, "an unchanged read cursor must not emit another event")

	var event domain.Event
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT id, workspace_id, conversation_id, message_id, actor_id, thread_root_id, event_type FROM chat_outbox WHERE event_type='read.updated' AND actor_id=$1 AND message_id=$2 ORDER BY created_at DESC LIMIT 1`, f.peer, firstReply.ID).
		Scan(&event.ID, &event.WorkspaceID, &event.ConversationID, &event.MessageID, &event.ActorID, &event.ThreadRootID, &event.Type))
	recipients, err := f.repo.ReadRecipientIDs(f.ctx, event)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{f.peer}, recipients, "unread changes are private to the current reader's SSE sessions")
	var notificationRecipients int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox_recipients WHERE event_id=$1`, event.ID).Scan(&notificationRecipients))
	require.Zero(t, notificationRecipients)
}

func TestMarkUnreadPreservesExistingUnreadAndSkipsOwnAndDeletedMessages(t *testing.T) {
	f := newFixture(t)
	peer := humanScope(t, f.workspace, f.peer)
	root := f.create(t, "Root", nil)
	first := f.create(t, "Received reply", &root.ID)
	_, err := f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{Scope: peer, ConversationID: f.room, ThreadRootID: &root.ID, CreationID: uuid.New(), Text: "Own reply"})
	require.NoError(t, err)
	deleted := f.create(t, "Deleted reply", &root.ID)
	_, err = f.service.DeleteMessage(f.ctx, domain.DeleteMessageCommand{Scope: f.scope, MessageID: deleted.ID, Version: deleted.Version})
	require.NoError(t, err)
	last := f.create(t, "Last reply", &root.ID)

	state, err := f.service.MarkUnread(f.ctx, domain.MarkUnreadCommand{Scope: peer, ConversationID: f.room, ThreadRootID: &root.ID, FromMessageID: last.ID})
	require.NoError(t, err)
	require.Zero(t, state.LastReadSequence, "an absent read cursor is already entirely unread")
	require.Equal(t, 2, state.UnreadCount, "own messages and tombstones do not count as unread")
	require.Equal(t, &first.ID, state.FirstUnreadReplyID)
	require.Equal(t, 3, state.Summary.TotalUnread, "the root remains independently unread")
	var readEvents int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox WHERE event_type='read.updated'`).Scan(&readEvents))
	require.Zero(t, readEvents, "an absent cursor is already unread and needs no write or event")
}

func TestMarkUnreadRejectsUnauthorizedAndMismatchedTargets(t *testing.T) {
	f := newFixture(t)
	peer := humanScope(t, f.workspace, f.peer)
	root := f.create(t, "Root", nil)
	reply := f.create(t, "Reply", &root.ID)
	otherRoot := f.create(t, "Other root", nil)
	deleted := f.create(t, "Deleted", nil)
	_, err := f.service.DeleteMessage(f.ctx, domain.DeleteMessageCommand{Scope: f.scope, MessageID: deleted.ID, Version: deleted.Version})
	require.NoError(t, err)
	dm, err := f.service.CreateDirect(f.ctx, peer, f.author)
	require.NoError(t, err)
	restricted := peer
	restricted.Actor.TeamAccess, err = auth.RestrictedTeamAccess(uuid.New())
	require.NoError(t, err)

	cases := []struct {
		name    string
		command domain.MarkUnreadCommand
		want    error
	}{
		{"foreign workspace", domain.MarkUnreadCommand{Scope: humanScope(t, uuid.New(), f.peer), ConversationID: f.room, FromMessageID: root.ID}, domain.ErrNotFound},
		{"nonmember", domain.MarkUnreadCommand{Scope: humanScope(t, f.workspace, f.outsider), ConversationID: f.room, FromMessageID: root.ID}, domain.ErrNotFound},
		{"restricted team", domain.MarkUnreadCommand{Scope: restricted, ConversationID: f.room, FromMessageID: root.ID}, domain.ErrNotFound},
		{"wrong conversation", domain.MarkUnreadCommand{Scope: peer, ConversationID: dm.ID, FromMessageID: root.ID}, domain.ErrInvalid},
		{"reply as timeline", domain.MarkUnreadCommand{Scope: peer, ConversationID: f.room, FromMessageID: reply.ID}, domain.ErrInvalid},
		{"root as reply", domain.MarkUnreadCommand{Scope: peer, ConversationID: f.room, ThreadRootID: &root.ID, FromMessageID: root.ID}, domain.ErrInvalid},
		{"wrong thread", domain.MarkUnreadCommand{Scope: peer, ConversationID: f.room, ThreadRootID: &otherRoot.ID, FromMessageID: reply.ID}, domain.ErrInvalid},
		{"deleted target", domain.MarkUnreadCommand{Scope: peer, ConversationID: f.room, FromMessageID: deleted.ID}, domain.ErrNotFound},
		{"missing target", domain.MarkUnreadCommand{Scope: peer, ConversationID: f.room, FromMessageID: uuid.New()}, domain.ErrNotFound},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := f.service.MarkUnread(f.ctx, test.command)
			require.ErrorIs(t, err, test.want)
		})
	}
	var readRows, unreadEvents int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_conversation_reads`).Scan(&readRows))
	require.Zero(t, readRows)
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox WHERE event_type='read.updated'`).Scan(&unreadEvents))
	require.Zero(t, unreadEvents)

	_, err = f.pool.Exec(f.ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.team, f.peer)
	require.NoError(t, err)
	_, err = f.service.MarkUnread(f.ctx, domain.MarkUnreadCommand{Scope: peer, ConversationID: f.room, FromMessageID: root.ID})
	require.ErrorIs(t, err, domain.ErrNotFound, "membership is checked at mutation time")
}

type failedUnreadSummaryRepository struct {
	*Repository
	failure error
}

func (r failedUnreadSummaryRepository) Unread(context.Context, domain.ActorScope) (domain.UnreadSummary, error) {
	return domain.UnreadSummary{}, r.failure
}

func TestMarkUnreadRollsBackCursorAndOutboxWhenReconciliationFails(t *testing.T) {
	f := newFixture(t)
	peer := humanScope(t, f.workspace, f.peer)
	first := f.create(t, "First", nil)
	last := f.create(t, "Last", nil)
	_, err := f.service.MarkRead(f.ctx, domain.ReadCommand{Scope: peer, ConversationID: f.room, ThroughMessageID: last.ID})
	require.NoError(t, err)
	failure := errors.New("summary unavailable")
	failingService := service.New(failedUnreadSummaryRepository{Repository: f.repo, failure: failure})
	_, err = failingService.MarkUnread(f.ctx, domain.MarkUnreadCommand{Scope: peer, ConversationID: f.room, FromMessageID: first.ID})
	require.ErrorIs(t, err, failure)
	var sequence int64
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT last_read_sequence FROM chat_conversation_reads WHERE conversation_id=$1 AND user_id=$2`, f.room, f.peer).Scan(&sequence))
	require.Equal(t, last.Sequence, sequence)
	var events int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox WHERE event_type='read.updated' AND message_id=$1`, first.ID).Scan(&events))
	require.Zero(t, events)
}
