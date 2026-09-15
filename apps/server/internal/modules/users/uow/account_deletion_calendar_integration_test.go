//go:build integration

package useruow

import (
	"testing"

	calendarrepository "github.com/complexus-tech/projects-api/internal/modules/calendar/repository"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAccountDeletionPreservesRealCalendarCleanupUntilProviderDrain(t *testing.T) {
	f := newDeletionFixture(t)
	ctx := t.Context()
	connection, block := uuid.New(), uuid.New()
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO calendar_connections(connection_id,workspace_id,user_id,provider,connected_email,provider_account_id,token_payload,scopes) VALUES ($1,$2,$3,'google','private-calendar@example.com','personal-provider-id','sealed-cleanup-token',ARRAY['https://www.googleapis.com/auth/calendar.events.owned'])`, connection, f.workspace, f.user)
	execDeletion(t, ctx, f.db.Pool, `INSERT INTO calendar_schedule_blocks(block_id,workspace_id,user_id,story_id,block_type,title,start_at,end_at,source,external_provider,external_calendar_id,external_event_id) VALUES ($1,$2,$3,$4,'work','Private schedule title',now(),now()+interval '1 hour','maya','google','primary','owned-provider-event')`, block, f.workspace, f.user, f.story)
	repo, err := usersrepository.NewAccountDeletionRepository(f.db.Pool, "aws", "attachments", "profiles")
	require.NoError(t, err)
	f.manager, err = New(f.db.Pool, repo, calendarrepository.New(f.db.Pool))
	require.NoError(t, err)

	pending, err := f.manager.Delete(ctx, f.command())
	require.NoError(t, err)
	require.True(t, pending)
	deletionCount(t, f, `SELECT count(*) FROM users WHERE user_id=$1 AND NOT is_active AND email<> 'departing@example.com' AND full_name IS NULL`, 1, f.user)
	deletionCount(t, f, `SELECT count(*) FROM workspace_members WHERE user_id=$1`, 0, f.user)
	deletionCount(t, f, `SELECT count(*) FROM calendar_schedule_blocks WHERE user_id=$1`, 0, f.user)
	deletionCount(t, f, `SELECT count(*) FROM calendar_connections WHERE connection_id=$1 AND cleanup_pending_at IS NOT NULL AND connected_email='' AND provider_account_id='' AND token_payload='sealed-cleanup-token'`, 1, connection)
	deletionCount(t, f, `SELECT count(*) FROM calendar_schedule_event_outbox WHERE user_id=$1 AND operation='delete' AND provider_event_id='owned-provider-event' AND payload=CAST('{}' AS jsonb) AND processed_at IS NULL AND dead_lettered_at IS NULL`, 1, f.user)
	deletionCount(t, f, `SELECT count(*) FROM account_deletion_requests WHERE user_id=$1`, 1, f.user)
	completed, err := f.manager.FinalizePending(ctx, 50)
	require.NoError(t, err)
	require.Zero(t, completed)

	// Simulate the provider dispatcher's confirmed delivery and credential purge.
	// This test exercises database sequencing; no provider request is performed.
	execDeletion(t, ctx, f.db.Pool, `UPDATE calendar_schedule_event_outbox SET processed_at=now() WHERE user_id=$1`, f.user)
	execDeletion(t, ctx, f.db.Pool, `DELETE FROM calendar_connections WHERE connection_id=$1`, connection)
	completed, err = f.manager.FinalizePending(ctx, 50)
	require.NoError(t, err)
	require.Equal(t, 1, completed)
	deletionCount(t, f, `SELECT count(*) FROM users WHERE user_id=$1`, 0, f.user)
	deletionCount(t, f, `SELECT count(*) FROM calendar_schedule_event_outbox WHERE user_id=$1`, 0, f.user)
}
