//go:build integration

package chatrepository

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/complexus-tech/projects-api/internal/migrations"
	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatsql "github.com/complexus-tech/projects-api/internal/modules/chat/repository/sqlc"
	chatservice "github.com/complexus-tech/projects-api/internal/modules/chat/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestChatUnreadCountsKeepTimelineAndThreadsIndependent(t *testing.T) {
	f := newFixture(t)
	peer := humanScope(t, f.workspace, f.peer)
	first := f.create(t, "First root", nil)
	second := f.create(t, "Second root", nil)
	firstReply := f.create(t, "First thread reply", &first.ID)
	secondReply := f.create(t, "Second thread reply", &second.ID)
	deleted := f.create(t, "Deleted reply", &first.ID)
	_, err := f.service.DeleteMessage(f.ctx, domain.DeleteMessageCommand{Scope: f.scope, MessageID: deleted.ID, Version: deleted.Version})
	require.NoError(t, err)
	for _, root := range []*uuid.UUID{nil, &first.ID} {
		_, err = f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{Scope: peer, ConversationID: f.room, ThreadRootID: root, CreationID: uuid.New(), Text: "My own message"})
		require.NoError(t, err)
	}

	summary, err := f.service.Unread(f.ctx, peer)
	require.NoError(t, err)
	require.Equal(t, 4, summary.TotalUnread)
	require.Equal(t, 2, summary.UnreadThreads)
	require.Equal(t, []domain.ConversationUnread{{ConversationID: f.room, UnreadCount: 4, UnreadRootCount: 2, UnreadReplyCount: 2}}, summary.Conversations)
	conversation, err := f.service.GetConversation(f.ctx, peer, f.room)
	require.NoError(t, err)
	require.Equal(t, 4, conversation.UnreadCount)
	require.Equal(t, 2, conversation.UnreadRootCount)
	require.Equal(t, 2, conversation.UnreadReplyCount)
	root, err := f.service.GetMessage(f.ctx, peer, first.ID)
	require.NoError(t, err)
	require.Equal(t, 2, root.ReplyCount)
	require.Equal(t, 1, root.UnreadReplyCount)

	read, err := f.service.MarkRead(f.ctx, domain.ReadCommand{Scope: peer, ConversationID: f.room, ThroughMessageID: second.ID})
	require.NoError(t, err)
	require.Zero(t, read.UnreadCount, "read response is scoped to the timeline")
	summary, err = f.service.Unread(f.ctx, peer)
	require.NoError(t, err)
	require.Equal(t, 2, summary.TotalUnread, "unopened threads remain unread")
	require.Equal(t, 2, summary.UnreadThreads)
	_, err = f.service.MarkRead(f.ctx, domain.ReadCommand{Scope: peer, ConversationID: f.room, ThreadRootID: &first.ID, ThroughMessageID: firstReply.ID})
	require.NoError(t, err)
	summary, err = f.service.Unread(f.ctx, peer)
	require.NoError(t, err)
	require.Equal(t, 1, summary.TotalUnread)
	require.Equal(t, 1, summary.UnreadThreads)
	_, err = f.service.MarkRead(f.ctx, domain.ReadCommand{Scope: peer, ConversationID: f.room, ThreadRootID: &second.ID, ThroughMessageID: secondReply.ID})
	require.NoError(t, err)
	summary, err = f.service.Unread(f.ctx, peer)
	require.NoError(t, err)
	require.Zero(t, summary.TotalUnread)
	require.Zero(t, summary.UnreadThreads)
	require.Empty(t, summary.Conversations)
}

func TestChatUnreadSummaryRechecksMembershipAndCredentialScope(t *testing.T) {
	f := newFixture(t)
	peer := humanScope(t, f.workspace, f.peer)
	root := f.create(t, "Team root", nil)
	f.create(t, "Team reply", &root.ID)
	dm, err := f.service.CreateDirect(f.ctx, f.scope, f.peer)
	require.NoError(t, err)
	_, err = f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{Scope: f.scope, ConversationID: dm.ID, CreationID: uuid.New(), Text: "Private message"})
	require.NoError(t, err)
	list, err := f.service.ListConversations(f.ctx, domain.ListQuery{Scope: peer, Limit: 1})
	require.NoError(t, err)
	require.Len(t, list.Items, 1)
	require.NotNil(t, list.NextCursor)
	summary, err := f.service.Unread(f.ctx, peer)
	require.NoError(t, err)
	require.Equal(t, 3, summary.TotalUnread, "summary includes conversations outside the loaded page")
	require.Len(t, summary.Conversations, 2)

	restricted := peer
	restricted.Actor.TeamAccess, err = auth.RestrictedTeamAccess(f.team)
	require.NoError(t, err)
	summary, err = f.service.Unread(f.ctx, restricted)
	require.NoError(t, err)
	require.Equal(t, 2, summary.TotalUnread, "team-restricted credentials cannot count DMs")
	outsider := humanScope(t, f.workspace, f.outsider)
	summary, err = f.service.Unread(f.ctx, outsider)
	require.NoError(t, err)
	require.Zero(t, summary.TotalUnread, "workspace admins do not bypass chat membership")
	_, err = f.service.UnreadThreads(f.ctx, domain.UnreadThreadQuery{Scope: outsider, ConversationID: &f.room})
	require.ErrorIs(t, err, domain.ErrNotFound)

	_, err = f.pool.Exec(f.ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.team, f.peer)
	require.NoError(t, err)
	summary, err = f.service.Unread(f.ctx, peer)
	require.NoError(t, err)
	require.Equal(t, 1, summary.TotalUnread, "team revocation preserves independent DM access")
	threads, err := f.service.UnreadThreads(f.ctx, domain.UnreadThreadQuery{Scope: peer})
	require.NoError(t, err)
	require.Empty(t, threads.Items)
	_, err = f.pool.Exec(f.ctx, `UPDATE users SET is_active=FALSE WHERE user_id=$1`, f.peer)
	require.NoError(t, err)
	summary, err = f.service.Unread(f.ctx, peer)
	require.NoError(t, err)
	require.Zero(t, summary.TotalUnread)
}

func TestChatUnreadThreadsFindOldRootsAndPaginateTombstones(t *testing.T) {
	f := newFixture(t)
	peer := humanScope(t, f.workspace, f.peer)
	first := f.create(t, "Old first root", nil)
	second := f.create(t, "Old second root", nil)
	firstReply := f.create(t, "First thread unread", &first.ID)
	nextReply := f.create(t, "First thread next unread", &first.ID)
	f.create(t, "Second thread unread", &second.ID)
	_, err := f.service.DeleteMessage(f.ctx, domain.DeleteMessageCommand{Scope: f.scope, MessageID: second.ID, Version: second.Version})
	require.NoError(t, err)
	f.create(t, "Latest timeline message", nil)
	latest, err := f.service.ListMessages(f.ctx, domain.MessageQuery{Scope: peer, ConversationID: f.room, Limit: 1})
	require.NoError(t, err)
	require.Len(t, latest.Items, 1)
	require.NotEqual(t, first.ID, latest.Items[0].ID)
	page, err := f.service.UnreadThreads(f.ctx, domain.UnreadThreadQuery{Scope: peer, Limit: 1})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, second.ID, page.Items[0].ID)
	require.NotNil(t, page.Items[0].DeletedAt)
	require.Empty(t, page.Items[0].Text)
	require.Equal(t, 1, page.Items[0].UnreadReplyCount)
	require.Equal(t, "Product", page.Items[0].ConversationName)
	require.Equal(t, "team", page.Items[0].ConversationKind)
	require.NotNil(t, page.NextCursor)
	next, err := f.service.UnreadThreads(f.ctx, domain.UnreadThreadQuery{Scope: peer, Cursor: *page.NextCursor, Limit: 1})
	require.NoError(t, err)
	require.Len(t, next.Items, 1)
	require.Equal(t, first.ID, next.Items[0].ID)
	require.Equal(t, firstReply.ID, next.Items[0].FirstUnreadReplyID)
	require.Nil(t, next.NextCursor)
	_, err = f.service.UnreadThreads(f.ctx, domain.UnreadThreadQuery{Scope: peer, ConversationID: &f.room, Cursor: *page.NextCursor})
	require.ErrorIs(t, err, domain.ErrInvalid, "cursor cannot cross a conversation filter")
	_, err = f.service.MarkRead(f.ctx, domain.ReadCommand{Scope: peer, ConversationID: f.room, ThreadRootID: &first.ID, ThroughMessageID: firstReply.ID})
	require.NoError(t, err)
	remaining, err := f.service.UnreadThreads(f.ctx, domain.UnreadThreadQuery{Scope: peer, ConversationID: &f.room})
	require.NoError(t, err)
	for _, item := range remaining.Items {
		if item.ID == first.ID {
			require.Equal(t, nextReply.ID, item.FirstUnreadReplyID, "jump target advances past the thread read cursor")
		}
	}
}

func TestChatUnreadMigrationRestoresProjectionDependenciesOnRollback(t *testing.T) {
	f := newFixture(t)
	peer := humanScope(t, f.workspace, f.peer)
	root := f.create(t, "Root", nil)
	f.create(t, "Unread reply", &root.ID)
	down, err := migrations.FS.ReadFile("000235_chat_unread_counts.down.sql")
	require.NoError(t, err)
	_, err = f.pool.Exec(f.ctx, string(down))
	require.NoError(t, err, "old projections must be restored before their helper is dropped")
	conversation, err := f.service.GetConversation(f.ctx, peer, f.room)
	require.NoError(t, err)
	require.Equal(t, 1, conversation.UnreadCount, "rollback restores root-only projection")
	var unsupportedReasons int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox_recipients WHERE reason='message'`).Scan(&unsupportedReasons))
	require.Zero(t, unsupportedReasons)
	up, err := migrations.FS.ReadFile("000235_chat_unread_counts.up.sql")
	require.NoError(t, err)
	_, err = f.pool.Exec(f.ctx, string(up))
	require.NoError(t, err)
	summary, err := f.service.Unread(f.ctx, peer)
	require.NoError(t, err)
	require.Equal(t, 2, summary.TotalUnread)
	require.Equal(t, 1, summary.UnreadThreads)
}

func TestChatUnreadQueriesUseReadCursorsWithLargeHistory(t *testing.T) {
	f := newFixture(t)
	_, err := f.pool.Exec(f.ctx, `INSERT INTO chat_conversation_messages(workspace_id,conversation_id,sequence,author_id,creation_id,creation_hash,text)
        SELECT $1,$2,n,$3,gen_random_uuid(),decode(repeat('ab',32),'hex'),'Historical root ' || n FROM generate_series(1,10000) n`, f.workspace, f.room, f.author)
	require.NoError(t, err)
	var root uuid.UUID
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT id FROM chat_conversation_messages WHERE conversation_id=$1 AND sequence=1`, f.room).Scan(&root))
	_, err = f.pool.Exec(f.ctx, `INSERT INTO chat_conversation_messages(workspace_id,conversation_id,sequence,thread_root_id,author_id,creation_id,creation_hash,text)
        SELECT $1,$2,10000+n,$3,$4,gen_random_uuid(),decode(repeat('cd',32),'hex'),'Historical reply ' || n FROM generate_series(1,100) n`, f.workspace, f.room, root, f.author)
	require.NoError(t, err)
	_, err = f.pool.Exec(f.ctx, `INSERT INTO chat_conversation_reads(conversation_id,user_id,last_read_sequence) VALUES($1,$2,10000)`, f.room, f.peer)
	require.NoError(t, err)
	_, err = f.pool.Exec(f.ctx, `INSERT INTO chat_thread_reads(thread_root_id,user_id,last_read_sequence) VALUES($1,$2,10050)`, root, f.peer)
	require.NoError(t, err)
	_, err = f.pool.Exec(f.ctx, `ANALYZE public.chat_conversation_messages`)
	require.NoError(t, err)
	tx, err := f.pool.Begin(f.ctx)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback(f.ctx) }()
	_, err = tx.Exec(f.ctx, `SET LOCAL plan_cache_mode='force_generic_plan'`)
	require.NoError(t, err)
	recorder := &historyPlanRecorder{Tx: tx}
	service := chatservice.New(&Repository{pool: f.pool, queries: chatsql.New(recorder)})
	peer := humanScope(t, f.workspace, f.peer)
	summary, err := service.Unread(f.ctx, peer)
	require.NoError(t, err)
	require.Equal(t, 50, summary.TotalUnread)
	require.Equal(t, 1, summary.UnreadThreads)
	threads, err := service.UnreadThreads(f.ctx, domain.UnreadThreadQuery{Scope: peer})
	require.NoError(t, err)
	require.Len(t, threads.Items, 1)
	require.Equal(t, 50, threads.Items[0].UnreadReplyCount)
	var firstUnreadSequence int64
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT sequence FROM chat_conversation_messages WHERE id=$1`, threads.Items[0].FirstUnreadReplyID).Scan(&firstUnreadSequence))
	require.EqualValues(t, 10051, firstUnreadSequence)
	for _, entry := range recorder.plans {
		var plans []struct {
			Plan          map[string]any `json:"Plan"`
			ExecutionTime float64        `json:"Execution Time"`
		}
		require.NoError(t, json.Unmarshal(entry.plan, &plans))
		require.Len(t, plans, 1)
		indexed := false
		var inspect func(map[string]any)
		inspect = func(node map[string]any) {
			if name, ok := node["Index Name"].(string); ok && (strings.Contains(name, "chat_roots_sequence") || strings.Contains(name, "chat_replies_sequence") || strings.Contains(name, "conversation_id_sequence")) {
				indexed = true
			}
			if children, ok := node["Plans"].([]any); ok {
				for _, child := range children {
					inspect(child.(map[string]any))
				}
			}
		}
		inspect(plans[0].Plan)
		require.True(t, indexed, "unread queries should use existing message sequence indexes")
		t.Logf("%s over 10,100 messages: %.2f ms", strings.SplitN(entry.sql, "\n", 2)[0], plans[0].ExecutionTime)
	}
}
