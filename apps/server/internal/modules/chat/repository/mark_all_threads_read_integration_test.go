//go:build integration

package chatrepository

import (
	"context"
	"fmt"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatservice "github.com/complexus-tech/projects-api/internal/modules/chat/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type snapshotHookRepository struct {
	*Repository
	afterSnapshot func([]domain.ThreadReadTarget)
}

func (r snapshotHookRepository) UnreadThreadReadTargets(ctx context.Context, scope domain.ActorScope) ([]domain.ThreadReadTarget, error) {
	targets, err := r.Repository.UnreadThreadReadTargets(ctx, scope)
	if err == nil {
		r.afterSnapshot(targets)
	}
	return targets, err
}

func TestMarkAllThreadsReadIncludesEveryPageAndPreservesTimeline(t *testing.T) {
	f := newFixture(t)
	peer := humanScope(t, f.workspace, f.peer)
	const threads = 55
	for index := range threads {
		root := f.create(t, fmt.Sprintf("Root %d", index), nil)
		f.create(t, "Unread reply", &root.ID)
	}
	page, err := f.service.UnreadThreads(f.ctx, domain.UnreadThreadQuery{Scope: peer})
	require.NoError(t, err)
	require.Len(t, page.Items, 50)
	require.NotNil(t, page.NextCursor)

	require.NoError(t, f.service.MarkAllThreadsRead(f.ctx, peer))
	summary, err := f.service.Unread(f.ctx, peer)
	require.NoError(t, err)
	require.Equal(t, threads, summary.TotalUnread)
	require.Zero(t, summary.UnreadThreads)
	require.Equal(t, []domain.ConversationUnread{{ConversationID: f.room, UnreadCount: threads, UnreadRootCount: threads}}, summary.Conversations)
	var timelineReads, readEvents int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_conversation_reads WHERE user_id=$1`, f.peer).Scan(&timelineReads))
	require.Zero(t, timelineReads)
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox WHERE event_type='read.updated' AND actor_id=$1`, f.peer).Scan(&readEvents))
	require.Equal(t, threads, readEvents)
	require.NoError(t, f.service.MarkAllThreadsRead(f.ctx, peer))
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox WHERE event_type='read.updated' AND actor_id=$1`, f.peer).Scan(&readEvents))
	require.Equal(t, threads, readEvents, "an unchanged cursor does not emit another read event")
}

func TestMarkAllThreadsReadPreservesRepliesArrivingAfterSnapshot(t *testing.T) {
	f := newFixture(t)
	peer := humanScope(t, f.workspace, f.peer)
	root := f.create(t, "Original root", nil)
	lastObserved := f.create(t, "Original unread reply", &root.ID)
	var newReply, newRoot domain.Message
	service := chatservice.New(snapshotHookRepository{Repository: f.repo, afterSnapshot: func(targets []domain.ThreadReadTarget) {
		require.Equal(t, []domain.ThreadReadTarget{{ConversationID: f.room, ThreadRootID: root.ID, ThroughMessageID: lastObserved.ID}}, targets)
		newReply = f.create(t, "Reply after snapshot", &root.ID)
		newRoot = f.create(t, "Thread after snapshot", nil)
		f.create(t, "New thread reply", &newRoot.ID)
	}})
	require.NoError(t, service.MarkAllThreadsRead(f.ctx, peer))
	page, err := f.service.UnreadThreads(f.ctx, domain.UnreadThreadQuery{Scope: peer})
	require.NoError(t, err)
	require.Len(t, page.Items, 2)
	for _, thread := range page.Items {
		require.Equal(t, 1, thread.UnreadReplyCount)
		if thread.ID == root.ID {
			require.Equal(t, newReply.ID, thread.FirstUnreadReplyID)
		} else {
			require.Equal(t, newRoot.ID, thread.ID)
		}
	}
	var readEventMessage uuid.UUID
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT message_id FROM chat_outbox WHERE event_type='read.updated' AND actor_id=$1`, f.peer).Scan(&readEventMessage))
	require.Equal(t, lastObserved.ID, readEventMessage)
}

func TestMarkAllThreadsReadHonorsMembershipAndCredentialRestrictions(t *testing.T) {
	f := newFixture(t)
	peer := humanScope(t, f.workspace, f.peer)
	root := f.create(t, "Team root", nil)
	f.create(t, "Team reply", &root.ID)
	dm, err := f.service.CreateDirect(f.ctx, f.scope, f.peer)
	require.NoError(t, err)
	dmRoot, err := f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{Scope: f.scope, ConversationID: dm.ID, CreationID: uuid.New(), Text: "DM root"})
	require.NoError(t, err)
	_, err = f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{Scope: f.scope, ConversationID: dm.ID, CreationID: uuid.New(), ThreadRootID: &dmRoot.ID, Text: "DM reply"})
	require.NoError(t, err)
	require.NoError(t, f.service.MarkAllThreadsRead(f.ctx, humanScope(t, f.workspace, f.outsider)))
	var reads int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_thread_reads WHERE user_id=$1`, f.outsider).Scan(&reads))
	require.Zero(t, reads, "workspace admin membership cannot grant conversation access")

	restricted := peer
	restricted.Actor.TeamAccess, err = auth.RestrictedTeamAccess(uuid.New())
	require.NoError(t, err)
	require.NoError(t, f.service.MarkAllThreadsRead(f.ctx, restricted))
	summary, err := f.service.Unread(f.ctx, peer)
	require.NoError(t, err)
	require.Equal(t, 2, summary.UnreadThreads)
	restricted.Actor.TeamAccess, err = auth.RestrictedTeamAccess(f.team)
	require.NoError(t, err)
	require.NoError(t, f.service.MarkAllThreadsRead(f.ctx, restricted))
	page, err := f.service.UnreadThreads(f.ctx, domain.UnreadThreadQuery{Scope: peer})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, dmRoot.ID, page.Items[0].ID, "team restrictions must preserve independent DM read state")
	require.NoError(t, f.service.MarkAllThreadsRead(f.ctx, peer))
	summary, err = f.service.Unread(f.ctx, peer)
	require.NoError(t, err)
	require.Zero(t, summary.UnreadThreads)
	require.Equal(t, 2, summary.TotalUnread, "both conversation timelines remain unread")
}

func TestMarkAllThreadsReadRollsBackWhenSnapshotAccessIsRevoked(t *testing.T) {
	f := newFixture(t)
	peer := humanScope(t, f.workspace, f.peer)
	root := f.create(t, "First room root", nil)
	f.create(t, "First room reply", &root.ID)
	team := uuid.New()
	_, err := f.pool.Exec(f.ctx, `INSERT INTO teams(team_id,workspace_id,name,code,color) VALUES($1,$2,'Second','SECOND','#111111')`, team, f.workspace)
	require.NoError(t, err)
	for _, user := range []uuid.UUID{f.author, f.peer} {
		_, err = f.pool.Exec(f.ctx, `INSERT INTO team_members(team_id,user_id) VALUES($1,$2)`, team, user)
		require.NoError(t, err)
	}
	var room uuid.UUID
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT id FROM chat_conversations WHERE team_id=$1`, team).Scan(&room))
	secondRoot, err := f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{Scope: f.scope, ConversationID: room, CreationID: uuid.New(), Text: "Second room root"})
	require.NoError(t, err)
	_, err = f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{Scope: f.scope, ConversationID: room, CreationID: uuid.New(), ThreadRootID: &secondRoot.ID, Text: "Second room reply"})
	require.NoError(t, err)
	service := chatservice.New(snapshotHookRepository{Repository: f.repo, afterSnapshot: func(targets []domain.ThreadReadTarget) {
		require.Len(t, targets, 2)
		_, err := f.pool.Exec(f.ctx, `DELETE FROM team_members WHERE user_id=$1 AND team_id=(SELECT team_id FROM chat_conversations WHERE id=$2)`, f.peer, targets[1].ConversationID)
		require.NoError(t, err)
	}})
	require.ErrorIs(t, service.MarkAllThreadsRead(f.ctx, peer), domain.ErrNotFound)
	var reads, readEvents int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_thread_reads WHERE user_id=$1`, f.peer).Scan(&reads))
	require.Zero(t, reads, "a later authorization failure rolls back earlier cursor changes")
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox WHERE event_type='read.updated' AND actor_id=$1`, f.peer).Scan(&readEvents))
	require.Zero(t, readEvents, "rolled-back read state must not publish success events")
}
