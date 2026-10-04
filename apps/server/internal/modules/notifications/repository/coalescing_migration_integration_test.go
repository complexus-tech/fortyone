//go:build integration

package notificationsrepository

import (
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/migrations"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestUnreadStoryCoalescingUpgradeRetainsLatestInboxAndEveryOriginalPayload(t *testing.T) {
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	pool := f.postgres.Pool
	// Recreate the pre-212 table shape in this disposable database. Production
	// recovery remains forward-only; its refusing down migration is not used.
	mustNotificationExec(t, ctx, pool, `DROP INDEX public.idx_notifications_unread_story_update`)
	mustNotificationExec(t, ctx, pool, `DROP TABLE public.notification_event_receipts`)
	mustNotificationExec(t, ctx, pool, `ALTER TABLE public.notifications DROP COLUMN event_at, DROP COLUMN coalescing_key`)

	base := time.Date(2026, time.October, 3, 9, 0, 0, 0, time.UTC)
	readTime := base.Add(3 * time.Minute)
	commentID := uuid.New()
	type legacyEvent struct {
		id, recipient, entity, actor uuid.UUID
		key, kind, entityType        string
		createdAt                    time.Time
		readAt                       *time.Time
	}
	events := []legacyEvent{
		{uuid.New(), f.recipientA, f.storyA, f.actorB, "older-status", "story_update", "story", base, nil},
		{uuid.New(), f.recipientA, f.storyA, f.actorA, "latest-priority", "story_update", "story", base.Add(time.Minute), nil},
		{uuid.New(), f.recipientA, f.storyA, f.actorB, "read-history", "story_update", "story", base.Add(2 * time.Minute), &readTime},
		{uuid.New(), f.recipientA, commentID, f.actorA, "first-comment", "story_comment", "comment", base, nil},
		{uuid.New(), f.recipientA, commentID, f.actorB, "second-comment", "story_comment", "comment", base.Add(time.Minute), nil},
		{uuid.New(), f.recipientA, f.storyA, f.actorA, "mention", "mention", "story", base, nil},
		{uuid.New(), f.recipientB, f.storyA, f.actorA, "other-recipient", "story_update", "story", base, nil},
		{uuid.New(), f.recipientA, f.storyB, f.actorA, "other-story", "story_update", "story", base, nil},
	}
	originalPayloads := make(map[string][]byte, len(events))
	for _, event := range events {
		mustNotificationExec(t, ctx, pool, `
			INSERT INTO public.notifications (
				notification_id, recipient_id, workspace_id, type, entity_type,
				entity_id, actor_id, title, created_at, read_at, message, dedupe_key
			)
			VALUES ($1,$2,$3,CAST($4 AS notification_type),CAST($5 AS entity_type),$6,$7,$8,$9,$10,
				jsonb_build_object('template',CAST($8 AS text),'variables',jsonb_build_object('value',
					jsonb_build_object('type','text','value',CAST($11 AS text)))),$11)`,
			event.id, event.recipient, f.workspaceA, event.kind, event.entityType,
			event.entity, event.actor, "Legacy "+event.key, event.createdAt, event.readAt, event.key)
		var payload []byte
		require.NoError(t, pool.QueryRow(ctx, `
			SELECT jsonb_build_array(type, entity_type, entity_id, actor_id, title, message)
			FROM public.notifications WHERE notification_id=$1`, event.id).Scan(&payload))
		originalPayloads[event.key] = payload
	}

	script, err := migrations.FS.ReadFile("000212_unread_story_notification_coalescing.up.sql")
	require.NoError(t, err)
	mustNotificationExec(t, ctx, pool, string(script))

	var retained []uuid.UUID
	rows, err := pool.Query(ctx, `SELECT notification_id FROM public.notifications`)
	require.NoError(t, err)
	for rows.Next() {
		var id uuid.UUID
		require.NoError(t, rows.Scan(&id))
		retained = append(retained, id)
	}
	require.NoError(t, rows.Err())
	rows.Close()
	expected := make([]uuid.UUID, 0, len(events)-1)
	for _, event := range events[1:] {
		expected = append(expected, event.id)
	}
	require.ElementsMatch(t, expected, retained)

	for index, event := range events {
		var payloadMatches bool
		var actor uuid.UUID
		var occurredAt time.Time
		var notificationID *uuid.UUID
		require.NoError(t, pool.QueryRow(ctx, `
			SELECT payload=CAST($2 AS jsonb), actor_id, occurred_at, notification_id
			FROM public.notification_event_receipts WHERE dedupe_key=$1`,
			event.key, originalPayloads[event.key]).Scan(&payloadMatches, &actor, &occurredAt, &notificationID))
		require.True(t, payloadMatches, event.key)
		require.Equal(t, event.actor, actor, event.key)
		require.True(t, event.createdAt.Equal(occurredAt), event.key)
		if index == 0 {
			require.Nil(t, notificationID, "the removed inbox row retains an unlinked original receipt")
		} else {
			require.NotNil(t, notificationID, event.key)
			require.Equal(t, event.id, *notificationID, event.key)
			var unchanged bool
			require.NoError(t, pool.QueryRow(ctx, `
				SELECT jsonb_build_array(type,entity_type,entity_id,actor_id,title,message)=CAST($2 AS jsonb)
					AND created_at=$3 AND event_at=$3
				FROM public.notifications WHERE notification_id=$1`,
				event.id, originalPayloads[event.key], event.createdAt).Scan(&unchanged))
			require.True(t, unchanged, event.key)
		}
	}
	var readAt *time.Time
	require.NoError(t, pool.QueryRow(ctx, `SELECT read_at FROM public.notifications WHERE notification_id=$1`, events[2].id).Scan(&readAt))
	require.NotNil(t, readAt)
	require.True(t, readTime.Equal(*readAt), "the read-history timestamp must remain unchanged")

	_, err = pool.Exec(ctx, `
		INSERT INTO public.notifications (recipient_id,workspace_id,type,entity_type,entity_id,actor_id,title,message)
		VALUES ($1,$2,'story_update','story',$3,$4,'Duplicate unread',CAST('{}' AS jsonb))`,
		f.recipientA, f.workspaceA, f.storyA, f.actorA)
	var conflict *pgconn.PgError
	require.ErrorAs(t, err, &conflict)
	require.Equal(t, "23505", conflict.Code)
	require.Equal(t, "idx_notifications_unread_story_update", conflict.ConstraintName)
}
