//go:build integration

package chatrepository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	service "github.com/complexus-tech/projects-api/internal/modules/chat/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type fixture struct {
	ctx                                           context.Context
	pool                                          *pgxpool.Pool
	repo                                          *Repository
	service                                       *service.Service
	workspace, team, room, author, peer, outsider uuid.UUID
	scope                                         domain.ActorScope
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	postgres := testkit.NewPostgres(t)
	f := fixture{ctx: t.Context(), pool: postgres.Pool, workspace: uuid.New(), team: uuid.New(), author: uuid.New(), peer: uuid.New(), outsider: uuid.New()}
	_, err := f.pool.Exec(f.ctx, `INSERT INTO workspaces(workspace_id,name,slug) VALUES($1,'Chat',$2)`, f.workspace, f.workspace.String())
	require.NoError(t, err)
	for _, id := range []uuid.UUID{f.author, f.peer, f.outsider} {
		_, err = f.pool.Exec(f.ctx, `INSERT INTO users(user_id,username,email,full_name) VALUES($1,$2,$3,$4)`, id, id.String(), id.String()+"@example.test", "Person "+id.String())
		require.NoError(t, err)
		_, err = f.pool.Exec(f.ctx, `INSERT INTO workspace_members(workspace_id,user_id,role) VALUES($1,$2,'admin')`, f.workspace, id)
		require.NoError(t, err)
	}
	_, err = f.pool.Exec(f.ctx, `INSERT INTO teams(team_id,workspace_id,name,code,color) VALUES($1,$2,'Product','PROD','#111111')`, f.team, f.workspace)
	require.NoError(t, err)
	for _, id := range []uuid.UUID{f.author, f.peer} {
		_, err = f.pool.Exec(f.ctx, `INSERT INTO team_members(team_id,user_id) VALUES($1,$2)`, f.team, id)
		require.NoError(t, err)
	}
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT id FROM chat_conversations WHERE workspace_id=$1 AND team_id=$2`, f.workspace, f.team).Scan(&f.room))
	f.scope = humanScope(t, f.workspace, f.author)
	f.repo = New(nil, f.pool)
	f.service = service.New(f.repo, service.WithMayaActorID(uuid.MustParse("00000000-0000-0000-0000-000000000001")))
	return f
}

func humanScope(t *testing.T, workspace, user uuid.UUID) domain.ActorScope {
	t.Helper()
	actor, err := auth.NewHumanActor(user).WithWorkspace(workspace)
	require.NoError(t, err)
	return domain.ActorScope{WorkspaceID: workspace, Actor: actor}
}

func (f fixture) create(t *testing.T, text string, root *uuid.UUID) domain.Message {
	t.Helper()
	message, err := f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{Scope: f.scope, ConversationID: f.room, CreationID: uuid.New(), ThreadRootID: root, Text: text})
	require.NoError(t, err)
	return message
}

func (f fixture) eventFor(t *testing.T, id uuid.UUID) domain.Event {
	t.Helper()
	var event domain.Event
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT id,workspace_id,conversation_id,message_id,event_type FROM chat_outbox WHERE message_id=$1 AND event_type='message.created'`, id).Scan(&event.ID, &event.WorkspaceID, &event.ConversationID, &event.MessageID, &event.Type))
	return event
}

func TestChatMembershipDMAndCredentialBoundaries(t *testing.T) {
	f := newFixture(t)
	conversation, err := f.service.GetConversation(f.ctx, f.scope, f.room)
	require.NoError(t, err)
	require.Equal(t, "Product", conversation.Name)
	require.Equal(t, 3, conversation.MemberCount)
	outsider := humanScope(t, f.workspace, f.outsider)
	_, err = f.service.GetConversation(f.ctx, outsider, f.room)
	require.ErrorIs(t, err, domain.ErrNotFound, "workspace admin is not an implicit team chat member")
	_, err = f.service.ListMessages(f.ctx, domain.MessageQuery{Scope: outsider, ConversationID: f.room})
	require.ErrorIs(t, err, domain.ErrNotFound)
	foreign := humanScope(t, uuid.New(), f.author)
	_, err = f.service.GetConversation(f.ctx, foreign, f.room)
	require.ErrorIs(t, err, domain.ErrNotFound)
	restricted := f.scope
	restricted.Actor.TeamAccess, err = auth.RestrictedTeamAccess(uuid.New())
	require.NoError(t, err)
	_, err = f.service.GetConversation(f.ctx, restricted, f.room)
	require.ErrorIs(t, err, domain.ErrNotFound)
	_, err = f.service.CreateDirect(f.ctx, restricted, f.peer)
	require.ErrorIs(t, err, domain.ErrForbidden)
	dm, err := f.service.CreateDirect(f.ctx, f.scope, f.peer)
	require.NoError(t, err)
	require.Equal(t, "dm", dm.Kind)
	peer := humanScope(t, f.workspace, f.peer)
	reverse, err := f.service.CreateDirect(f.ctx, peer, f.author)
	require.NoError(t, err)
	require.Equal(t, dm.ID, reverse.ID)
	_, err = f.service.GetConversation(f.ctx, outsider, dm.ID)
	require.ErrorIs(t, err, domain.ErrNotFound)
	_, err = f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{Scope: f.scope, ConversationID: dm.ID, CreationID: uuid.New(), Text: "Maya", AgentMentions: []string{"maya"}})
	require.ErrorIs(t, err, domain.ErrInvalid)
	list, err := f.service.ListConversations(f.ctx, domain.ListQuery{Scope: f.scope, Search: "Product"})
	require.NoError(t, err)
	require.Len(t, list.Items, 1)
	require.Equal(t, f.room, list.Items[0].ID)
	_, err = f.pool.Exec(f.ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.team, f.author)
	require.NoError(t, err)
	_, err = f.service.GetConversation(f.ctx, f.scope, f.room)
	require.ErrorIs(t, err, domain.ErrNotFound)
	_, err = f.service.GetConversation(f.ctx, f.scope, dm.ID)
	require.NoError(t, err, "leaving a team does not revoke a separate DM")
	_, err = f.pool.Exec(f.ctx, `DELETE FROM workspace_members WHERE workspace_id=$1 AND user_id=$2`, f.workspace, f.author)
	require.NoError(t, err)
	_, err = f.service.GetConversation(f.ctx, f.scope, dm.ID)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestChatCreationReplayPagingAndContentConflicts(t *testing.T) {
	f := newFixture(t)
	command := domain.CreateMessageCommand{Scope: f.scope, ConversationID: f.room, CreationID: uuid.New(), Text: "Identical request"}
	const callers = 4
	results := make(chan domain.Message, callers)
	failures := make(chan error, callers)
	var group sync.WaitGroup
	start := make(chan struct{})
	for range callers {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			message, err := f.service.CreateMessage(f.ctx, command)
			results <- message
			failures <- err
		}()
	}
	close(start)
	group.Wait()
	close(results)
	close(failures)
	for err := range failures {
		require.NoError(t, err)
	}
	var root domain.Message
	fresh := 0
	for result := range results {
		if root.ID == uuid.Nil {
			root = result
		}
		require.Equal(t, root.ID, result.ID)
		if !result.Replayed {
			fresh++
		}
	}
	require.Equal(t, 1, fresh)
	var count int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox WHERE message_id=$1`, root.ID).Scan(&count))
	require.Equal(t, 1, count)
	command.Text = "Different request"
	_, err := f.service.CreateMessage(f.ctx, command)
	require.ErrorIs(t, err, domain.ErrConflict)
	roots := []domain.Message{root}
	for i := range 8 {
		roots = append(roots, f.create(t, fmt.Sprintf("Root %d onboarding", i), nil))
	}
	first, err := f.service.ListMessages(f.ctx, domain.MessageQuery{Scope: f.scope, ConversationID: f.room, Limit: 3})
	require.NoError(t, err)
	require.Len(t, first.Items, 3)
	require.True(t, first.HasOlder)
	require.False(t, first.HasNewer)
	require.Equal(t, roots[6].ID, first.Items[0].ID)
	older, err := f.service.ListMessages(f.ctx, domain.MessageQuery{Scope: f.scope, ConversationID: f.room, Before: *first.OlderCursor, Limit: 3})
	require.NoError(t, err)
	require.Equal(t, roots[3].ID, older.Items[0].ID)
	require.True(t, older.HasNewer)
	newer, err := f.service.ListMessages(f.ctx, domain.MessageQuery{Scope: f.scope, ConversationID: f.room, After: *older.NewerCursor, Limit: 3})
	require.NoError(t, err)
	require.Equal(t, first.Items, newer.Items)
	replies := []domain.Message{}
	for i := range 8 {
		replies = append(replies, f.create(t, fmt.Sprintf("Reply %d ownership", i), &root.ID))
	}
	around, err := f.service.ListMessages(f.ctx, domain.MessageQuery{Scope: f.scope, ConversationID: f.room, ThreadRootID: &root.ID, Around: &replies[4].ID, Limit: 3})
	require.NoError(t, err)
	require.Len(t, around.Items, 3)
	require.Equal(t, replies[4].ID, around.Items[1].ID)
	require.True(t, around.HasOlder)
	require.True(t, around.HasNewer)
	rootWindow, err := f.service.ListMessages(f.ctx, domain.MessageQuery{Scope: f.scope, ConversationID: f.room, Around: &replies[4].ID, Limit: 3})
	require.NoError(t, err)
	require.Equal(t, root.ID, rootWindow.Items[0].ID)
	matches, err := f.service.Search(f.ctx, domain.SearchQuery{Scope: f.scope, ConversationID: f.room, Query: "ownership", Limit: 3})
	require.NoError(t, err)
	require.Len(t, matches.Items, 3)
	require.NotNil(t, matches.NextCursor)
	edited, err := f.service.UpdateMessage(f.ctx, domain.UpdateMessageCommand{Scope: f.scope, MessageID: root.ID, Version: root.Version, Text: "Edited"})
	require.NoError(t, err)
	require.Equal(t, 2, edited.Version)
	_, err = f.service.UpdateMessage(f.ctx, domain.UpdateMessageCommand{Scope: f.scope, MessageID: root.ID, Version: root.Version, Text: "Lost update"})
	require.ErrorIs(t, err, domain.ErrConflict)
	_, err = f.service.DeleteMessage(f.ctx, domain.DeleteMessageCommand{Scope: f.scope, MessageID: root.ID, Version: root.Version})
	require.ErrorIs(t, err, domain.ErrConflict)
	deleted, err := f.service.DeleteMessage(f.ctx, domain.DeleteMessageCommand{Scope: f.scope, MessageID: root.ID, Version: edited.Version})
	require.NoError(t, err)
	require.Equal(t, 3, deleted.Version)
	tombstone, err := f.service.GetMessage(f.ctx, f.scope, root.ID)
	require.NoError(t, err)
	require.Empty(t, tombstone.Text)
	require.JSONEq(t, "null", string(tombstone.RichContent))
	require.Empty(t, tombstone.MentionedUserIDs)
	require.Empty(t, tombstone.Attachments)
	require.Equal(t, 8, tombstone.ReplyCount)
	command.Text = "Identical request"
	replay, err := f.service.CreateMessage(f.ctx, command)
	require.NoError(t, err)
	require.True(t, replay.Replayed)
	require.Equal(t, root.ID, replay.ID)
	require.NotNil(t, replay.DeletedAt)
	leaf := roots[1]
	_, err = f.service.DeleteMessage(f.ctx, domain.DeleteMessageCommand{Scope: f.scope, MessageID: leaf.ID, Version: leaf.Version})
	require.NoError(t, err)
	_, err = f.service.GetMessage(f.ctx, f.scope, leaf.ID)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestChatReadReactionAndDelayedNotificationSemantics(t *testing.T) {
	f := newFixture(t)
	peer := humanScope(t, f.workspace, f.peer)
	created, err := f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{Scope: f.scope, ConversationID: f.room, CreationID: uuid.New(), Text: "Ping", MentionedUserIDs: []uuid.UUID{f.peer}})
	require.NoError(t, err)
	event := f.eventFor(t, created.ID)
	recipients, err := f.repo.NotificationRecipients(f.ctx, event)
	require.NoError(t, err)
	require.Equal(t, []domain.NotificationRecipient{{UserID: f.peer, Reason: "mention"}}, recipients)
	_, err = f.service.UpdateMessage(f.ctx, domain.UpdateMessageCommand{Scope: f.scope, MessageID: created.ID, Version: created.Version, Text: "Mention removed"})
	require.NoError(t, err)
	recipients, err = f.repo.NotificationRecipients(f.ctx, event)
	require.NoError(t, err)
	require.Len(t, recipients, 1, "creation recipient snapshot survives a subsequent edit")
	_, err = f.pool.Exec(f.ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.team, f.peer)
	require.NoError(t, err)
	recipients, err = f.repo.NotificationRecipients(f.ctx, event)
	require.NoError(t, err)
	require.Empty(t, recipients, "delivery rechecks membership")
	_, err = f.pool.Exec(f.ctx, `INSERT INTO team_members(team_id,user_id) VALUES($1,$2)`, f.team, f.peer)
	require.NoError(t, err)
	second := f.create(t, "Second root", nil)
	reply := f.create(t, "Thread reply", &created.ID)
	state, err := f.service.MarkRead(f.ctx, domain.ReadCommand{Scope: peer, ConversationID: f.room, ThroughMessageID: second.ID})
	require.NoError(t, err)
	require.Zero(t, state.UnreadCount)
	stale, err := f.service.MarkRead(f.ctx, domain.ReadCommand{Scope: peer, ConversationID: f.room, ThroughMessageID: created.ID})
	require.NoError(t, err)
	require.Equal(t, state.LastReadSequence, stale.LastReadSequence)
	readErrors := make(chan error, 2)
	var readGroup sync.WaitGroup
	for _, id := range []uuid.UUID{created.ID, second.ID} {
		readGroup.Add(1)
		go func(id uuid.UUID) {
			defer readGroup.Done()
			_, err := f.service.MarkRead(f.ctx, domain.ReadCommand{Scope: peer, ConversationID: f.room, ThroughMessageID: id})
			readErrors <- err
		}(id)
	}
	readGroup.Wait()
	close(readErrors)
	for err := range readErrors {
		require.NoError(t, err)
	}
	conversation, err := f.service.GetConversation(f.ctx, peer, f.room)
	require.NoError(t, err)
	require.Equal(t, second.Sequence, conversation.LastReadSequence)
	var readEvents int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox WHERE conversation_id=$1 AND event_type='read.updated'`, f.room).Scan(&readEvents))
	require.Equal(t, 1, readEvents, "unchanged or stale cursors must not trigger an SSE read feedback loop")

	var threadReads int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_thread_reads WHERE thread_root_id=$1 AND user_id=$2`, created.ID, f.peer).Scan(&threadReads))
	require.Zero(t, threadReads, "main timeline read does not mark thread seen")
	_, err = f.service.MarkRead(f.ctx, domain.ReadCommand{Scope: peer, ConversationID: f.room, ThroughMessageID: reply.ID})
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = f.service.MarkRead(f.ctx, domain.ReadCommand{Scope: peer, ConversationID: f.room, ThreadRootID: &created.ID, ThroughMessageID: reply.ID})
	require.NoError(t, err)
	const callers = 4
	errorsCh := make(chan error, callers)
	var group sync.WaitGroup
	for range callers {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := f.service.SetReaction(f.ctx, peer, second.ID, "👍", true)
			errorsCh <- err
		}()
	}
	group.Wait()
	close(errorsCh)
	for err := range errorsCh {
		require.NoError(t, err)
	}
	reacted, err := f.service.GetMessage(f.ctx, peer, second.ID)
	require.NoError(t, err)
	require.Equal(t, 1, reacted.Version)
	require.Equal(t, []domain.Reaction{{Emoji: "👍", Count: 1, Reacted: true, Users: []domain.Person{{ID: f.peer, FullName: "Person " + f.peer.String()}}}}, reacted.Reactions)
	var events int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_outbox WHERE message_id=$1 AND event_type='reaction.updated'`, second.ID).Scan(&events))
	require.Equal(t, 1, events)
	claimed, err := f.repo.ClaimEvents(f.ctx, 100)
	require.NoError(t, err)
	require.NotEmpty(t, claimed)
	require.ErrorIs(t, f.repo.CompleteEvent(f.ctx, claimed[0].ID, uuid.New()), domain.ErrConflict)
	require.NoError(t, f.repo.CompleteEvent(f.ctx, claimed[0].ID, claimed[0].LeaseToken))
}

func TestChatAttachmentClaimsAndCreationHookRollback(t *testing.T) {
	f := newFixture(t)
	attachment := uuid.New()
	_, err := f.pool.Exec(f.ctx, `INSERT INTO attachments(attachment_id,workspace_id,uploaded_by,filename,size,mime_type,blob_name,scan_status) VALUES($1,$2,$3,'brief.pdf',128,'application/pdf','fixture.pdf','clean')`, attachment, f.workspace, f.author)
	require.NoError(t, err)
	scope := domain.AttachmentScope{ActorScope: f.scope, ConversationID: f.room}
	require.NoError(t, f.service.LinkPendingAttachment(f.ctx, scope, attachment))
	peer := domain.AttachmentScope{ActorScope: humanScope(t, f.workspace, f.peer), ConversationID: f.room}
	require.ErrorIs(t, f.service.AuthorizeAttachment(f.ctx, peer, attachment), domain.ErrNotFound)
	_, err = f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{Scope: peer.ActorScope, ConversationID: f.room, CreationID: uuid.New(), Text: "Steal pending upload", AttachmentIDs: []uuid.UUID{attachment}})
	require.ErrorIs(t, err, domain.ErrInvalid)
	command := domain.CreateMessageCommand{Scope: f.scope, ConversationID: f.room, CreationID: uuid.New(), Text: "Brief https://fortyone.app", AttachmentIDs: []uuid.UUID{attachment}}
	failed := errors.New("hook failed")
	aborting := service.New(f.repo, service.WithMessageCreatedHook(func(context.Context, domain.CreateMessageCommand, domain.Message) error { return failed }))
	_, err = aborting.CreateMessage(f.ctx, command)
	require.ErrorIs(t, err, failed)
	var pending bool
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT message_id IS NULL FROM chat_attachments WHERE attachment_id=$1`, attachment).Scan(&pending))
	require.True(t, pending)
	created, err := f.service.CreateMessage(f.ctx, command)
	require.NoError(t, err)
	require.Len(t, created.Attachments, 1)
	require.Equal(t, "brief.pdf", created.Attachments[0].FileName)
	require.NoError(t, f.service.AuthorizeAttachment(f.ctx, peer, attachment))
	_, err = f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{Scope: f.scope, ConversationID: f.room, CreationID: uuid.New(), Text: "Reuse upload", AttachmentIDs: []uuid.UUID{attachment}})
	require.ErrorIs(t, err, domain.ErrInvalid)
	resources, err := f.service.Resources(f.ctx, domain.SearchQuery{Scope: f.scope, ConversationID: f.room, Kind: "all"})
	require.NoError(t, err)
	require.Len(t, resources.Items, 2)
	_, err = f.service.DeleteMessage(f.ctx, domain.DeleteMessageCommand{Scope: f.scope, MessageID: created.ID, Version: created.Version})
	require.NoError(t, err)
	require.ErrorIs(t, f.service.AuthorizeAttachment(f.ctx, peer, attachment), domain.ErrNotFound)
}
