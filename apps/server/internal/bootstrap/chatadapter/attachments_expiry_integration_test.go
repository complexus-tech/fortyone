//go:build integration

package chatadapter

import (
	"testing"
	"time"

	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	attachmentsrepository "github.com/complexus-tech/projects-api/internal/modules/attachments/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPendingChatUploadExpiryIsBoundedAndPreservesRecentSentSharedAndClaimedFiles(t *testing.T) {
	f := newAttachmentFixture(t)
	now := time.Now().UTC()
	old, shared, sent, recent, claiming := f.upload(t, true), f.upload(t, true), f.upload(t, true), f.upload(t, true), f.upload(t, true)
	message := f.message(t, sent)
	f.exec(t, `UPDATE chat_attachments SET created_at=$1 WHERE attachment_id=ANY($2)`, now.Add(-48*time.Hour), []uuid.UUID{old, sent, claiming})
	f.exec(t, `UPDATE chat_attachments SET created_at=$1 WHERE attachment_id=$2`, now.Add(-47*time.Hour), shared)
	f.exec(t, `INSERT INTO document_attachments(document_id,attachment_id,created_by) VALUES ($1,$2,$3)`, f.workspaceDocument, shared, f.scope.Actor.PrincipalID)

	// A concurrent sender holds the association row while claiming its upload.
	tx, err := f.pool.Begin(t.Context())
	require.NoError(t, err)
	defer tx.Rollback(t.Context())
	_, err = tx.Exec(t.Context(), `UPDATE chat_attachments SET message_id=$1 WHERE attachment_id=$2`, message.ID, claiming)
	require.NoError(t, err)
	store := attachmentsrepository.New(f.pool)
	batch := attachmentdomain.PendingChatUploadRetention{CreatedBefore: now.Add(-24 * time.Hour), RetiredAt: now, Limit: 1, StorageProvider: "aws", ContainerName: "chat-test"}
	count, err := store.RetirePendingChatUploads(t.Context(), batch)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	f.count(t, 0, `SELECT count(*) FROM attachments WHERE attachment_id=$1`, old)
	f.count(t, 1, `SELECT count(*) FROM chat_attachments WHERE attachment_id=$1`, shared)
	batch.Limit = 100
	count, err = store.RetirePendingChatUploads(t.Context(), batch)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	f.count(t, 0, `SELECT count(*) FROM chat_attachments WHERE attachment_id=$1`, shared)
	require.NoError(t, tx.Commit(t.Context()))
	count, err = store.RetirePendingChatUploads(t.Context(), batch)
	require.NoError(t, err)
	require.Zero(t, count)
	for _, id := range []uuid.UUID{shared, sent, recent, claiming} {
		f.count(t, 1, `SELECT count(*) FROM attachments WHERE attachment_id=$1`, id)
		f.count(t, 0, `SELECT count(*) FROM attachment_object_deletion_outbox WHERE attachment_id=$1`, id)
	}
	f.count(t, 1, `SELECT count(*) FROM attachment_object_deletion_outbox WHERE attachment_id=$1 AND container_name='chat-test'`, old)
	f.count(t, 1, `SELECT count(*) FROM chat_attachments WHERE attachment_id=$1 AND message_id=$2`, claiming, message.ID)
}

func TestPendingChatUploadExpiryRollsBackAssociationAndMetadataIfEnqueueFails(t *testing.T) {
	f := newAttachmentFixture(t)
	id := f.upload(t, true)
	now := time.Now().UTC()
	f.exec(t, `UPDATE chat_attachments SET created_at=$1 WHERE attachment_id=$2`, now.Add(-48*time.Hour), id)
	f.exec(t, `CREATE FUNCTION reject_pending_upload_retirement() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'retirement unavailable'; END $$; CREATE TRIGGER reject_pending_upload_retirement BEFORE INSERT ON attachment_object_deletion_outbox FOR EACH ROW EXECUTE FUNCTION reject_pending_upload_retirement()`)
	count, err := attachmentsrepository.New(f.pool).RetirePendingChatUploads(t.Context(), attachmentdomain.PendingChatUploadRetention{CreatedBefore: now.Add(-24 * time.Hour), RetiredAt: now, Limit: 100, StorageProvider: "aws", ContainerName: "chat-test"})
	require.ErrorContains(t, err, "retirement unavailable")
	require.Zero(t, count)
	f.count(t, 1, `SELECT count(*) FROM chat_attachments WHERE attachment_id=$1 AND message_id IS NULL`, id)
	f.count(t, 1, `SELECT count(*) FROM attachments WHERE attachment_id=$1`, id)
	f.count(t, 0, `SELECT count(*) FROM attachment_object_deletion_outbox WHERE attachment_id=$1`, id)
}
