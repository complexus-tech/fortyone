//go:build integration

package chatadapter

import (
	"context"
	"testing"

	chatdomain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatrepository "github.com/complexus-tech/projects-api/internal/modules/chat/repository"
	chatservice "github.com/complexus-tech/projects-api/internal/modules/chat/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

type attachmentFixture struct {
	pool              *pgxpool.Pool
	chat              *chatservice.Service
	retirer           *AttachmentRetirer
	scope             chatdomain.AttachmentScope
	peer              uuid.UUID
	workspaceDocument uuid.UUID
}

func newAttachmentFixture(t *testing.T) attachmentFixture {
	t.Helper()
	pool := testkit.NewPostgres(t).Pool
	workspace, team, user, peer, document := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	f := attachmentFixture{pool: pool, peer: peer, workspaceDocument: document}
	f.exec(t, `INSERT INTO users(user_id,username,email) VALUES ($1,$2,$3),($4,$5,$6)`, user, user.String(), user.String()+"@example.test", peer, peer.String(), peer.String()+"@example.test")
	f.exec(t, `INSERT INTO workspaces(workspace_id,name,slug) VALUES ($1,'Attachment lifecycle',$2)`, workspace, workspace.String())
	f.exec(t, `INSERT INTO workspace_members(workspace_id,user_id,role) VALUES ($1,$2,'admin'),($1,$3,'member')`, workspace, user, peer)
	f.exec(t, `INSERT INTO teams(team_id,workspace_id,name,code,color) VALUES ($1,$2,'Product','PROD','#777777')`, team, workspace)
	f.exec(t, `INSERT INTO team_members(team_id,user_id) VALUES ($1,$2),($1,$3)`, team, user, peer)
	f.exec(t, `INSERT INTO documents(document_id,workspace_id,title,visibility,created_by,updated_by) VALUES ($1,$2,'Shared document','workspace',$3,$3)`, document, workspace, user)
	actor, err := auth.NewHumanActor(user).WithWorkspace(workspace)
	require.NoError(t, err)
	f.scope.ActorScope = chatdomain.ActorScope{WorkspaceID: workspace, Actor: actor}
	require.NoError(t, pool.QueryRow(t.Context(), `SELECT id FROM chat_conversations WHERE team_id=$1`, team).Scan(&f.scope.ConversationID))
	f.retirer = NewAttachmentRetirer(pool, "aws", "chat-test")
	f.chat = chatservice.New(chatrepository.New(nil, pool, chatrepository.WithAttachmentRetirement(f.retirer.Retire)))
	return f
}

func (f attachmentFixture) exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	_, err := f.pool.Exec(t.Context(), sql, args...)
	require.NoError(t, err)
}

func (f attachmentFixture) count(t *testing.T, want int, sql string, args ...any) {
	t.Helper()
	var count int
	require.NoError(t, f.pool.QueryRow(t.Context(), sql, args...).Scan(&count))
	require.Equal(t, want, count, sql)
}

func (f attachmentFixture) upload(t *testing.T, pending bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	f.exec(t, `INSERT INTO attachments(attachment_id,workspace_id,uploaded_by,filename,size,mime_type,blob_name,scan_status) VALUES ($1,$2,$3,'upload.txt',10,'text/plain',$4,'clean')`, id, f.scope.WorkspaceID, f.scope.Actor.PrincipalID, id.String())
	if pending {
		require.NoError(t, f.chat.LinkPendingAttachment(t.Context(), f.scope, id))
	}
	return id
}

func (f attachmentFixture) message(t *testing.T, ids ...uuid.UUID) chatdomain.Message {
	t.Helper()
	message, err := f.chat.CreateMessage(t.Context(), chatdomain.CreateMessageCommand{Scope: f.scope.ActorScope, ConversationID: f.scope.ConversationID, CreationID: uuid.New(), Text: "Shared design files", AttachmentIDs: ids})
	require.NoError(t, err)
	return message
}

func TestChatAttachmentRetirementOnEditDeleteAndPendingCancel(t *testing.T) {
	f := newAttachmentFixture(t)
	removed, retained := f.upload(t, true), f.upload(t, true)
	message := f.message(t, removed, retained)
	updated, err := f.chat.UpdateMessage(t.Context(), chatdomain.UpdateMessageCommand{Scope: f.scope.ActorScope, MessageID: message.ID, Version: message.Version, Text: "Keep one file", AttachmentIDs: []uuid.UUID{retained}})
	require.NoError(t, err)
	f.count(t, 0, `SELECT count(*) FROM attachments WHERE attachment_id=$1`, removed)
	f.count(t, 1, `SELECT count(*) FROM attachment_object_deletion_outbox WHERE attachment_id=$1 AND container_name='chat-test' AND status='pending'`, removed)
	f.count(t, 1, `SELECT count(*) FROM chat_attachments WHERE attachment_id=$1 AND message_id=$2`, retained, message.ID)
	_, err = f.chat.DeleteMessage(t.Context(), chatdomain.DeleteMessageCommand{Scope: f.scope.ActorScope, MessageID: message.ID, Version: updated.Version})
	require.NoError(t, err)
	_, err = f.chat.DeleteMessage(t.Context(), chatdomain.DeleteMessageCommand{Scope: f.scope.ActorScope, MessageID: message.ID, Version: updated.Version})
	require.NoError(t, err)
	f.count(t, 0, `SELECT count(*) FROM attachments WHERE attachment_id=$1`, retained)
	f.count(t, 1, `SELECT count(*) FROM attachment_object_deletion_outbox WHERE attachment_id=$1`, retained)

	pending := f.upload(t, true)
	peer, err := auth.NewHumanActor(f.peer).WithWorkspace(f.scope.WorkspaceID)
	require.NoError(t, err)
	peerScope := f.scope
	peerScope.Actor = peer
	require.NoError(t, f.chat.UnlinkPendingAttachment(t.Context(), peerScope, pending))
	f.count(t, 1, `SELECT count(*) FROM chat_attachments WHERE attachment_id=$1`, pending)
	f.count(t, 0, `SELECT count(*) FROM attachment_object_deletion_outbox WHERE attachment_id=$1`, pending)
	require.NoError(t, f.chat.UnlinkPendingAttachment(t.Context(), f.scope, pending))
	require.NoError(t, f.chat.UnlinkPendingAttachment(t.Context(), f.scope, pending))
	f.count(t, 0, `SELECT count(*) FROM attachments WHERE attachment_id=$1`, pending)
	f.count(t, 1, `SELECT count(*) FROM attachment_object_deletion_outbox WHERE attachment_id=$1`, pending)
}

func TestChatAttachmentRetirementPreservesSharedReferencesAndCompensatesFailedLinks(t *testing.T) {
	f := newAttachmentFixture(t)
	shared := f.upload(t, true)
	f.exec(t, `INSERT INTO document_attachments(document_id,attachment_id,created_by) VALUES ($1,$2,$3)`, f.workspaceDocument, shared, f.scope.Actor.PrincipalID)
	message := f.message(t, shared)
	_, err := f.chat.DeleteMessage(t.Context(), chatdomain.DeleteMessageCommand{Scope: f.scope.ActorScope, MessageID: message.ID, Version: message.Version})
	require.NoError(t, err)
	f.count(t, 1, `SELECT count(*) FROM attachments WHERE attachment_id=$1`, shared)
	f.count(t, 0, `SELECT count(*) FROM attachment_object_deletion_outbox WHERE attachment_id=$1`, shared)

	unlinked := f.upload(t, false)
	require.NoError(t, f.retirer.RetireUnlinkedUpload(t.Context(), f.scope.WorkspaceID, unlinked))
	require.NoError(t, f.retirer.RetireUnlinkedUpload(t.Context(), f.scope.WorkspaceID, unlinked))
	f.count(t, 1, `SELECT count(*) FROM attachment_object_deletion_outbox WHERE attachment_id=$1`, unlinked)
	linked := f.upload(t, true)
	require.NoError(t, f.retirer.RetireUnlinkedUpload(t.Context(), f.scope.WorkspaceID, linked))
	f.count(t, 1, `SELECT count(*) FROM chat_attachments WHERE attachment_id=$1`, linked)
	f.count(t, 0, `SELECT count(*) FROM attachment_object_deletion_outbox WHERE attachment_id=$1`, linked)
	require.ErrorContains(t, f.retirer.Retire(context.Background(), f.scope.WorkspaceID, []uuid.UUID{shared}), "requires the message transaction")
}

func TestChatAttachmentRetirementFailureRollsBackMessageAndReferences(t *testing.T) {
	f := newAttachmentFixture(t)
	id := f.upload(t, true)
	message := f.message(t, id)
	pending := f.upload(t, true)
	f.exec(t, `CREATE FUNCTION reject_chat_retirement() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'retirement unavailable'; END $$; CREATE TRIGGER reject_chat_retirement BEFORE INSERT ON attachment_object_deletion_outbox FOR EACH ROW EXECUTE FUNCTION reject_chat_retirement()`)
	_, err := f.chat.UpdateMessage(t.Context(), chatdomain.UpdateMessageCommand{Scope: f.scope.ActorScope, MessageID: message.ID, Version: message.Version, Text: "Attempt to remove file"})
	require.ErrorContains(t, err, "retirement unavailable")
	_, err = f.chat.DeleteMessage(t.Context(), chatdomain.DeleteMessageCommand{Scope: f.scope.ActorScope, MessageID: message.ID, Version: message.Version})
	require.ErrorContains(t, err, "retirement unavailable")
	require.ErrorContains(t, f.chat.UnlinkPendingAttachment(t.Context(), f.scope, pending), "retirement unavailable")
	f.count(t, 1, `SELECT count(*) FROM chat_conversation_messages WHERE id=$1 AND version=1 AND deleted_at IS NULL AND text='Shared design files'`, message.ID)
	for _, attachment := range []uuid.UUID{id, pending} {
		f.count(t, 1, `SELECT count(*) FROM attachments WHERE attachment_id=$1`, attachment)
		f.count(t, 1, `SELECT count(*) FROM chat_attachments WHERE attachment_id=$1`, attachment)
	}
	f.count(t, 1, `SELECT count(*) FROM chat_outbox WHERE message_id=$1`, message.ID)
	f.count(t, 0, `SELECT count(*) FROM attachment_object_deletion_outbox`)
}
