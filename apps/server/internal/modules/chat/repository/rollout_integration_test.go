//go:build integration

package chatrepository

import (
	"context"
	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func setChatEnabled(t *testing.T, f fixture, enabled bool) {
	t.Helper()
	_, err := f.pool.Exec(f.ctx, `INSERT INTO workspace_settings(workspace_id,chat_enabled) VALUES($1,$2) ON CONFLICT(workspace_id) DO UPDATE SET chat_enabled=EXCLUDED.chat_enabled`, f.workspace, enabled)
	require.NoError(t, err)
}

func TestWorkspaceSettingLeavesOutboxPendingAndResumesWithoutRestart(t *testing.T) {
	f := newFixture(t)
	message := f.create(t, "Hold delivery while workspace disabled", nil)
	setChatEnabled(t, f, false)
	events, err := f.repo.ClaimEvents(f.ctx, 5)
	require.NoError(t, err)
	require.Empty(t, events)
	var attempts int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT attempts FROM chat_outbox WHERE message_id=$1 AND event_type='message.created'`, message.ID).Scan(&attempts))
	require.Zero(t, attempts)
	setChatEnabled(t, f, true)
	events, err = f.repo.ClaimEvents(f.ctx, 5)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, message.ID, events[0].MessageID)
	setChatEnabled(t, f, false)
	require.ErrorIs(t, f.repo.CompleteEvent(f.ctx, events[0].ID, events[0].LeaseToken), domain.ErrConflict)
	require.NoError(t, f.repo.ReleaseEvent(f.ctx, events[0].ID, events[0].LeaseToken))
	var delivered bool
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT delivered_at IS NOT NULL FROM chat_outbox WHERE id=$1`, events[0].ID).Scan(&delivered))
	require.False(t, delivered)
}

func TestWorkspaceSettingPausesPendingTranscriptions(t *testing.T) {
	f := newFixture(t)
	attachment, _ := addVoiceNote(t, f, true)
	_, err := f.repo.RequestTranscription(f.ctx, domain.AttachmentScope{ActorScope: f.scope, ConversationID: f.room}, attachment, false)
	require.NoError(t, err)
	setChatEnabled(t, f, false)
	jobs, err := f.repo.ClaimTranscriptions(f.ctx, 2)
	require.NoError(t, err)
	require.Empty(t, jobs)
	var status string
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT status FROM chat_audio_transcriptions WHERE attachment_id=$1`, attachment).Scan(&status))
	require.Equal(t, "pending", status)
	setChatEnabled(t, f, true)
	jobs, err = f.repo.ClaimTranscriptions(f.ctx, 2)
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	setChatEnabled(t, f, false)
	_, err = f.repo.ReadTranscriptionAudio(f.ctx, jobs[0])
	require.ErrorIs(t, err, domain.ErrNotFound)
	require.NoError(t, f.repo.RequeueTranscription(f.ctx, jobs[0]))
	setChatEnabled(t, f, true)
	jobs, err = f.repo.ClaimTranscriptions(f.ctx, 2)
	require.NoError(t, err)
	require.Len(t, jobs, 1)
}

func TestWorkspaceSettingBlocksRepositoryReadsAndWrites(t *testing.T) {
	f := newFixture(t)
	message := f.create(t, "Available by default", nil)
	setChatEnabled(t, f, false)
	_, err := f.repo.GetMessage(f.ctx, f.scope, message.ID)
	require.ErrorIs(t, err, domain.ErrNotFound)
	_, err = f.service.CreateDirect(f.ctx, f.scope, f.peer)
	require.ErrorIs(t, err, domain.ErrNotFound)
	_, err = f.service.SetReaction(f.ctx, f.scope, message.ID, "👍", true)
	require.ErrorIs(t, err, domain.ErrNotFound)
	page, err := f.service.ListConversations(f.ctx, domain.ListQuery{Scope: f.scope})
	require.NoError(t, err)
	require.Empty(t, page.Items)
	setChatEnabled(t, f, true)
	got, err := f.repo.GetMessage(f.ctx, f.scope, message.ID)
	require.NoError(t, err)
	require.Equal(t, message.ID, got.ID)
}

func TestOldSerializableSnapshotRetriesAndRejectsDisabledWorkspace(t *testing.T) {
	f := newFixture(t)
	settingsRepo := workspacesrepository.New(f.pool)
	settings, err := settingsRepo.GetOrCreateWorkspaceSettings(f.ctx, f.workspace)
	require.NoError(t, err)
	attempts := 0
	creation := uuid.New()
	err = f.repo.WithinUnit(f.ctx, func(unit context.Context) error {
		attempts++
		tx, ok := platformdatabase.UnitTransaction(unit)
		require.True(t, ok)
		var snapshot int
		if err := tx.QueryRow(unit, `SELECT 1 FROM workspaces WHERE workspace_id=$1`, f.workspace).Scan(&snapshot); err != nil {
			return err
		}
		if attempts == 1 {
			settings.ChatEnabled = false
			if _, err := settingsRepo.UpdateWorkspaceSettings(f.ctx, f.workspace, settings); err != nil {
				return err
			}
		}
		_, err := f.repo.CreateMessage(unit, domain.CreateMessageCommand{Scope: f.scope, ConversationID: f.room, CreationID: creation, Text: "Must not commit after disable"}, f.author)
		return err
	})
	require.ErrorIs(t, err, domain.ErrNotFound)
	require.Equal(t, 2, attempts, "the stale snapshot must receive40001 then retry with current settings")
	var count int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_conversation_messages WHERE creation_id=$1`, creation).Scan(&count))
	require.Zero(t, count)
}
