//go:build integration

package useruow

import (
	"testing"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAccountDeletionPreservesChatHistoryWithCollidingAuthorCreationKeys(t *testing.T) {
	f := newDeletionFixture(t)
	ctx := t.Context()
	room := deletionTeamConversation(t, f)
	creationID := uuid.New()
	first := insertDeletionChatMessage(t, f, room, f.user, creationID, 1, nil)
	second := insertDeletionChatMessage(t, f, room, f.other, creationID, 2, nil)
	reply := insertDeletionChatMessage(t, f, room, f.other, uuid.New(), 3, &first)
	remaining := uuid.New()
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO users(user_id,username,email) VALUES ($1,'remaining-admin','remaining-admin@example.com')`, remaining)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO workspace_members(workspace_id,user_id,role) VALUES ($1,$2,'admin')`, f.workspace, remaining)

	_, err := f.manager.Delete(ctx, f.command())
	require.NoError(t, err)
	_, err = f.manager.Delete(ctx, usersdomain.AccountDeletion{UserID: f.other, RequestedAt: time.Now().UTC()})
	require.NoError(t, err)

	deletionCount(t, f, `SELECT count(*) FROM chat_conversation_messages WHERE conversation_id=$1 AND author_id=$2 AND creation_id=id AND text='Retained team discussion'`, 3, room, usersdomain.DeletedUserID)
	deletionCount(t, f, `SELECT count(*) FROM chat_conversation_messages WHERE id=$1 AND thread_root_id=$2`, 1, reply, first)
	deletionCount(t, f, `SELECT count(*) FROM chat_conversation_messages WHERE id=$1 AND thread_root_id IS NULL`, 1, second)
}

func TestAccountDeletionRetiresPrivateChatUploadsBeforeDeferredFinalization(t *testing.T) {
	f := newDeletionFixture(t)
	f.cleanup.pending = true
	ctx := t.Context()
	teamRoom := deletionTeamConversation(t, f)
	directRoom := uuid.New()
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO chat_conversations(id,workspace_id,kind,dm_user_low,dm_user_high) VALUES ($1,$2,'dm',LEAST(CAST($3 AS uuid),CAST($4 AS uuid)),GREATEST(CAST($3 AS uuid),CAST($4 AS uuid)))`, directRoom, f.workspace, f.user, f.other)
	teamMessage := insertDeletionChatMessage(t, f, teamRoom, f.user, uuid.New(), 1, nil)
	directMessage := insertDeletionChatMessage(t, f, directRoom, f.user, uuid.New(), 1, nil)

	pending := insertDeletionChatAttachment(t, f, teamRoom, f.user, nil)
	otherPending := insertDeletionChatAttachment(t, f, teamRoom, f.other, nil)
	shared := insertDeletionChatAttachment(t, f, teamRoom, f.user, &teamMessage)
	directOwn := insertDeletionChatAttachment(t, f, directRoom, f.user, &directMessage)
	directPeer := insertDeletionChatAttachment(t, f, directRoom, f.other, &directMessage)
	// This upload also belongs to a retained shared document and must survive.
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO chat_attachments(attachment_id,workspace_id,conversation_id,uploaded_by,message_id) VALUES ($1,$2,$3,$4,$5)`, f.shared, f.workspace, directRoom, f.user, directMessage)

	_, err := f.manager.Delete(ctx, f.command())
	require.NoError(t, err)
	deletionCount(t, f, `SELECT count(*) FROM users WHERE user_id=$1 AND NOT is_active`, 1, f.user)
	deletionCount(t, f, `SELECT count(*) FROM chat_conversations WHERE id=$1`, 0, directRoom)
	for _, id := range []uuid.UUID{pending, directOwn, directPeer} {
		deletionCount(t, f, `SELECT count(*) FROM attachments WHERE attachment_id=$1`, 0, id)
		deletionCount(t, f, `SELECT count(*) FROM attachment_object_deletion_outbox WHERE attachment_id=$1 AND container_name='attachments'`, 1, id)
	}
	deletionCount(t, f, `SELECT count(*) FROM chat_attachments WHERE attachment_id=$1 AND message_id IS NULL AND uploaded_by=$2`, 1, otherPending, f.other)
	deletionCount(t, f, `SELECT count(*) FROM chat_attachments WHERE attachment_id=$1 AND message_id=$2 AND uploaded_by=$3`, 1, shared, teamMessage, usersdomain.DeletedUserID)
	for _, id := range []uuid.UUID{otherPending, shared, f.shared} {
		deletionCount(t, f, `SELECT count(*) FROM attachments WHERE attachment_id=$1`, 1, id)
		deletionCount(t, f, `SELECT count(*) FROM attachment_object_deletion_outbox WHERE attachment_id=$1`, 0, id)
	}
	f.cleanup.pending = false
	count, err := f.manager.FinalizePending(ctx, 50)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestAccountDeletionRollsBackChatRetirementOnLateFailure(t *testing.T) {
	f := newDeletionFixture(t)
	ctx := t.Context()
	room := deletionTeamConversation(t, f)
	message := insertDeletionChatMessage(t, f, room, f.user, uuid.New(), 1, nil)
	pending := insertDeletionChatAttachment(t, f, room, f.user, nil)
	execDeletion(t, ctx, f.db.Pool, `CREATE FUNCTION reject_chat_account_delete_test() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'forced late failure'; END $$; CREATE TRIGGER reject_chat_account_delete_test BEFORE DELETE ON users FOR EACH ROW EXECUTE FUNCTION reject_chat_account_delete_test()`)
	_, err := f.manager.Delete(ctx, f.command())
	require.ErrorContains(t, err, "forced late failure")
	deletionCount(t, f, `SELECT count(*) FROM chat_conversation_messages WHERE id=$1 AND author_id=$2`, 1, message, f.user)
	deletionCount(t, f, `SELECT count(*) FROM chat_attachments WHERE attachment_id=$1 AND uploaded_by=$2`, 1, pending, f.user)
	deletionCount(t, f, `SELECT count(*) FROM attachments WHERE attachment_id=$1`, 1, pending)
	deletionCount(t, f, `SELECT count(*) FROM attachment_object_deletion_outbox WHERE attachment_id=$1`, 0, pending)
}

func deletionTeamConversation(t *testing.T, f deletionFixture) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, f.db.Pool.QueryRow(t.Context(), `SELECT id FROM chat_conversations WHERE workspace_id=$1 AND kind='team'`, f.workspace).Scan(&id))
	return id
}

func insertDeletionChatMessage(t *testing.T, f deletionFixture, room, author, creationID uuid.UUID, sequence int, root *uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	execDeletion(t, t.Context(), f.db.Pool, `INSERT INTO chat_conversation_messages(id,workspace_id,conversation_id,sequence,thread_root_id,author_id,creation_id,creation_hash,text) VALUES ($1,$2,$3,$4,$5,$6,$7,decode(repeat('ab',32),'hex'),'Retained team discussion')`, id, f.workspace, room, sequence, root, author, creationID)
	return id
}

func insertDeletionChatAttachment(t *testing.T, f deletionFixture, room, author uuid.UUID, message *uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	execDeletion(t, t.Context(), f.db.Pool, `INSERT INTO attachments(attachment_id,workspace_id,uploaded_by,filename,size,mime_type,blob_name) VALUES ($1,$2,$3,'chat.txt',5,'text/plain',$4)`, id, f.workspace, author, id.String())
	execDeletion(t, t.Context(), f.db.Pool, `INSERT INTO chat_attachments(attachment_id,workspace_id,conversation_id,uploaded_by,message_id) VALUES ($1,$2,$3,$4,$5)`, id, f.workspace, room, author, message)
	return id
}
