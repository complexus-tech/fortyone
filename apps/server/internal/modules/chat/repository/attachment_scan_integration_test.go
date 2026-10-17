//go:build integration

package chatrepository

import (
	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestChatAttachmentAvailabilityAtLinkClaimAndRead(t *testing.T) {
	for _, allowed := range []string{"unscanned", "clean"} {
		t.Run(allowed, func(t *testing.T) {
			f := newFixture(t)
			id := uuid.New()
			scope := domain.AttachmentScope{ActorScope: f.scope, ConversationID: f.room}
			peer := domain.AttachmentScope{ActorScope: humanScope(t, f.workspace, f.peer), ConversationID: f.room}
			foreign := domain.AttachmentScope{ActorScope: humanScope(t, uuid.New(), f.author), ConversationID: f.room}
			_, err := f.pool.Exec(f.ctx, `INSERT INTO attachments(attachment_id,workspace_id,uploaded_by,filename,size,mime_type,blob_name) VALUES($1,$2,$3,'brief.pdf',128,'application/pdf',$4)`, id, f.workspace, f.author, id.String())
			require.NoError(t, err)
			setStatus := func(status string) {
				_, err := f.pool.Exec(f.ctx, `UPDATE attachments SET scan_status=$2 WHERE attachment_id=$1`, id, status)
				require.NoError(t, err)
			}
			for _, denied := range []string{"pending", "infected", "failed"} {
				setStatus(denied)
				require.ErrorIs(t, f.service.LinkPendingAttachment(f.ctx, scope, id), domain.ErrConflict, denied)
			}
			var references int
			require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_attachments WHERE attachment_id=$1`, id).Scan(&references))
			require.Zero(t, references)
			setStatus(allowed)
			foreignWorkspace, foreignAttachment := uuid.New(), uuid.New()
			_, err = f.pool.Exec(f.ctx, `INSERT INTO workspaces(workspace_id,name,slug) VALUES($1,'Other',$2)`, foreignWorkspace, foreignWorkspace.String())
			require.NoError(t, err)
			_, err = f.pool.Exec(f.ctx, `INSERT INTO attachments(attachment_id,workspace_id,uploaded_by,filename,size,mime_type,blob_name,scan_status) VALUES($1,$2,$3,'foreign.pdf',128,'application/pdf',$4,$5)`, foreignAttachment, foreignWorkspace, f.author, foreignAttachment.String(), allowed)
			require.NoError(t, err)
			require.ErrorIs(t, f.service.LinkPendingAttachment(f.ctx, scope, foreignAttachment), domain.ErrConflict, "an uploader cannot link a file from another workspace")
			require.ErrorIs(t, f.service.LinkPendingAttachment(f.ctx, foreign, id), domain.ErrNotFound)
			require.ErrorIs(t, f.service.LinkPendingAttachment(f.ctx, peer, id), domain.ErrConflict, "only the uploader may link an attachment")
			require.NoError(t, f.service.LinkPendingAttachment(f.ctx, scope, id))
			require.NoError(t, f.service.AuthorizeAttachment(f.ctx, scope, id))
			require.ErrorIs(t, f.service.AuthorizeAttachment(f.ctx, peer, id), domain.ErrNotFound, "pending uploads remain private")
			require.ErrorIs(t, f.service.AuthorizeAttachment(f.ctx, foreign, id), domain.ErrNotFound)
			for _, denied := range []string{"pending", "infected", "failed"} {
				setStatus(denied)
				require.ErrorIs(t, f.service.AuthorizeAttachment(f.ctx, scope, id), domain.ErrNotFound, denied)
				_, err := f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{Scope: f.scope, ConversationID: f.room, CreationID: uuid.New(), AttachmentIDs: []uuid.UUID{id}})
				require.ErrorIs(t, err, domain.ErrInvalid, denied)
			}
			setStatus(allowed)
			message, err := f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{Scope: f.scope, ConversationID: f.room, CreationID: uuid.New(), AttachmentIDs: []uuid.UUID{id}})
			require.NoError(t, err)
			require.Len(t, message.Attachments, 1)
			require.NoError(t, f.service.AuthorizeAttachment(f.ctx, peer, id))
			var status string
			require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT scan_status FROM attachments WHERE attachment_id=$1`, id).Scan(&status))
			require.Equal(t, allowed, status, "linking and sending must not invent a clean scan")
			for _, denied := range []string{"pending", "infected", "failed"} {
				setStatus(denied)
				require.ErrorIs(t, f.service.AuthorizeAttachment(f.ctx, peer, id), domain.ErrNotFound, denied)
			}
		})
	}
}

func TestChatTranscriptionAllowsAvailableAudioStates(t *testing.T) {
	for _, allowed := range []string{"unscanned", "clean"} {
		t.Run(allowed, func(t *testing.T) {
			f := newFixture(t)
			id, _ := addVoiceNote(t, f, true)
			_, err := f.pool.Exec(f.ctx, `UPDATE attachments SET scan_status=$2 WHERE attachment_id=$1`, id, allowed)
			require.NoError(t, err)
			scope := domain.AttachmentScope{ActorScope: f.scope, ConversationID: f.room}
			foreign := domain.AttachmentScope{ActorScope: humanScope(t, uuid.New(), f.author), ConversationID: f.room}
			_, err = f.repo.RequestTranscription(f.ctx, foreign, id, false)
			require.ErrorIs(t, err, domain.ErrNotFound)
			_, err = f.repo.RequestTranscription(f.ctx, scope, id, false)
			require.NoError(t, err)
			jobs, err := f.repo.ClaimTranscriptions(f.ctx, 2)
			require.NoError(t, err)
			require.Len(t, jobs, 1)
			_, err = f.repo.ReadTranscriptionAudio(f.ctx, jobs[0])
			require.NoError(t, err)
			require.NoError(t, f.repo.CompleteTranscription(f.ctx, jobs[0], "A voice note.", "test-model"))
			result, err := f.repo.GetTranscription(f.ctx, scope, id)
			require.NoError(t, err)
			require.Equal(t, "ready", result.Status)
			require.Equal(t, "A voice note.", result.Text)
			_, err = f.repo.GetTranscription(f.ctx, foreign, id)
			require.ErrorIs(t, err, domain.ErrNotFound)
		})
	}
}

func TestChatTranscriptionCannotReadOrPublishAfterScanRevocation(t *testing.T) {
	f := newFixture(t)
	id, _ := addVoiceNote(t, f, true)
	scope := domain.AttachmentScope{ActorScope: f.scope, ConversationID: f.room}
	_, err := f.repo.RequestTranscription(f.ctx, scope, id, false)
	require.NoError(t, err)
	jobs, err := f.repo.ClaimTranscriptions(f.ctx, 2)
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	_, err = f.repo.ReadTranscriptionAudio(f.ctx, jobs[0])
	require.NoError(t, err)
	for _, status := range []string{"pending", "infected", "failed"} {
		_, err = f.pool.Exec(f.ctx, `UPDATE attachments SET scan_status=$2 WHERE attachment_id=$1`, id, status)
		require.NoError(t, err)
		_, err = f.repo.ReadTranscriptionAudio(f.ctx, jobs[0])
		require.ErrorIs(t, err, domain.ErrNotFound, status)
		_, err = f.repo.GetTranscription(f.ctx, scope, id)
		require.ErrorIs(t, err, domain.ErrNotFound, status)
		_, err = f.repo.RequestTranscription(f.ctx, scope, id, false)
		require.ErrorIs(t, err, domain.ErrNotFound, status)
		require.ErrorIs(t, f.repo.CompleteTranscription(f.ctx, jobs[0], "Must not publish", "test-model"), domain.ErrConflict, status)
	}
}
