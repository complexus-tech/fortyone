//go:build integration

package notificationsrepository

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newMayaReplyNotification(t *testing.T, f notificationIntegrationFixture, recipient uuid.UUID) (uuid.UUID, domain.NewNotification) {
	t.Helper()
	id := uuid.New()
	input := f.storyNotification(recipient, "maya-reply:"+id.String()+":"+recipient.String())
	input.Type = domain.NotificationTypeCommentReply
	input.ActorID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	input.OccurredAt = time.Now().UTC().Truncate(time.Microsecond)
	input.Message.Variables["commentId"] = domain.Variable{Value: id.String(), Type: "comment"}
	mustNotificationExec(t, t.Context(), f.postgres.Pool, `INSERT INTO story_comments(comment_id,story_id,commenter_id,content) VALUES($1,$2,$3,'Maya response')`, id, f.storyA, input.ActorID)
	return id, input
}

func dueMayaReplies(t *testing.T, f notificationIntegrationFixture) {
	t.Helper()
	mustNotificationExec(t, t.Context(), f.postgres.Pool, `UPDATE maya_reply_notification_deliveries SET available_at=clock_timestamp()-INTERVAL '1 second' WHERE completed_at IS NULL`)
}

func createMayaNotification(f notificationIntegrationFixture) func(context.Context, domain.NewNotification) error {
	return func(ctx context.Context, input domain.NewNotification) error {
		_, _, err := f.repo.Create(ctx, input)
		return err
	}
}

func TestMayaReplyDecisionWaitsAndSuppressesOnlyExactSeenRecipient(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	id, input := newMayaReplyNotification(t, f, f.recipientA)
	require.NoError(t, f.repo.DeferMayaReply(ctx, input, id))
	var delay float64
	require.NoError(t, f.postgres.Pool.QueryRow(ctx, `SELECT EXTRACT(EPOCH FROM delivery.available_at-comment.created_at) FROM maya_reply_notification_deliveries delivery JOIN story_comments comment ON comment.comment_id=delivery.comment_id WHERE delivery.dedupe_key=$1`, input.DedupeKey).Scan(&delay))
	require.Equal(t, 5.0, delay)
	count, err := f.repo.DispatchMayaReplies(ctx, 25, createMayaNotification(f))
	require.NoError(t, err)
	require.Zero(t, count, "a newly published reply gets a visibility grace period")

	otherRecipient := input
	otherRecipient.RecipientID = f.guestA
	otherRecipient.DedupeKey = "maya-reply:" + id.String() + ":" + f.guestA.String()
	require.NoError(t, f.repo.DeferMayaReply(ctx, otherRecipient, id))
	unseenID, unseen := newMayaReplyNotification(t, f, f.recipientA)
	require.NoError(t, f.repo.DeferMayaReply(ctx, unseen, unseenID))
	mustNotificationExec(t, ctx, f.postgres.Pool, `INSERT INTO story_comment_seen(comment_id,user_id) VALUES($1,$2)`, id, f.recipientA)
	// Broad story/thread cutoffs do not prove that this individual reply was seen.
	mustNotificationExec(t, ctx, f.postgres.Pool, `INSERT INTO story_comment_reads(story_id,user_id,last_read_at) VALUES($1,$2,clock_timestamp()+INTERVAL '1 day')`, f.storyA, f.recipientA)
	mustNotificationExec(t, ctx, f.postgres.Pool, `INSERT INTO story_comment_thread_reads(thread_id,user_id,last_read_at) VALUES($1,$2,clock_timestamp()+INTERVAL '1 day')`, unseenID, f.recipientA)
	dueMayaReplies(t, f)
	count, err = f.repo.DispatchMayaReplies(ctx, 25, createMayaNotification(f))
	require.NoError(t, err)
	require.Equal(t, 3, count)
	var notifications, receipts int
	require.NoError(t, f.postgres.Pool.QueryRow(ctx, `SELECT count(*) FROM notifications WHERE workspace_id=$1`, f.workspaceA).Scan(&notifications))
	require.Equal(t, 2, notifications, "seen user/comment is quiet; another recipient and another reply still notify")
	require.NoError(t, f.postgres.Pool.QueryRow(ctx, `SELECT count(*) FROM notification_event_receipts WHERE dedupe_key=$1`, input.DedupeKey).Scan(&receipts))
	require.Zero(t, receipts, "suppression emits no inbox, push, or email work")
	require.NoError(t, f.repo.DeferMayaReply(ctx, input, id))
	count, err = f.repo.DispatchMayaReplies(ctx, 25, createMayaNotification(f))
	require.NoError(t, err)
	require.Zero(t, count, "completed suppression is durable on invocation replay")
	conflicting := input
	conflicting.Title = "Different payload"
	require.ErrorIs(t, f.repo.DeferMayaReply(ctx, conflicting, id), domain.ErrConflict)

	mention := input
	mention.Type = domain.NotificationTypeMention
	mention.DedupeKey = "maya-mention:" + id.String() + ":" + f.recipientA.String()
	row, created, err := f.repo.Create(ctx, mention)
	require.NoError(t, err)
	require.True(t, created)
	require.NotEqual(t, uuid.Nil, row.ID, "explicit mentions remain immediate even for a seen reply")
	human := input
	human.ActorID, human.DedupeKey = f.actorA, "human-reply:"+id.String()
	row, created, err = f.repo.Create(ctx, human)
	require.NoError(t, err)
	require.True(t, created)
	require.NotEqual(t, uuid.Nil, row.ID, "ordinary human notifications retain existing behavior")
}

func TestMayaReplyDeliveryRecoversAfterPersistenceWithBackoffAndNoDuplicates(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	id, input := newMayaReplyNotification(t, f, f.recipientA)
	require.NoError(t, f.repo.DeferMayaReply(ctx, input, id))
	dueMayaReplies(t, f)
	count, err := f.repo.DispatchMayaReplies(ctx, 25, func(ctx context.Context, input domain.NewNotification) error {
		require.NoError(t, createMayaNotification(f)(ctx, input))
		return errors.New("queue unavailable after notification committed")
	})
	require.Error(t, err)
	require.Equal(t, 1, count)
	var attempts int
	var deferred bool
	require.NoError(t, f.postgres.Pool.QueryRow(ctx, `SELECT attempts,available_at>clock_timestamp() FROM maya_reply_notification_deliveries WHERE dedupe_key=$1`, input.DedupeKey).Scan(&attempts, &deferred))
	require.Equal(t, 1, attempts)
	require.True(t, deferred)
	count, err = f.repo.DispatchMayaReplies(ctx, 25, createMayaNotification(f))
	require.NoError(t, err)
	require.Zero(t, count, "a failed delivery cannot hot-loop")
	dueMayaReplies(t, f)
	var calls atomic.Int32
	var workers sync.WaitGroup
	failures := make(chan error, 4)
	for range 4 {
		workers.Go(func() {
			_, err := f.repo.DispatchMayaReplies(ctx, 25, func(ctx context.Context, input domain.NewNotification) error {
				calls.Add(1)
				return createMayaNotification(f)(ctx, input)
			})
			failures <- err
		})
	}
	workers.Wait()
	close(failures)
	for err := range failures {
		require.NoError(t, err)
	}
	require.Equal(t, int32(1), calls.Load())
	var rows int
	require.NoError(t, f.postgres.Pool.QueryRow(ctx, `SELECT count(*) FROM notifications WHERE dedupe_key=$1`, input.DedupeKey).Scan(&rows))
	require.Equal(t, 1, rows, "retry reuses the already committed notification event receipt")
}

func TestMayaReplyDeliveryRetiresRevokedAccessAndDeletedReplies(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	id, input := newMayaReplyNotification(t, f, f.revocableA)
	require.NoError(t, f.repo.DeferMayaReply(ctx, input, id))
	mustNotificationExec(t, ctx, f.postgres.Pool, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.teamA, f.revocableA)
	dueMayaReplies(t, f)
	count, err := f.repo.DispatchMayaReplies(ctx, 25, createMayaNotification(f))
	require.NoError(t, err)
	require.Equal(t, 1, count)
	id, input = newMayaReplyNotification(t, f, f.recipientA)
	require.NoError(t, f.repo.DeferMayaReply(ctx, input, id))
	mustNotificationExec(t, ctx, f.postgres.Pool, `DELETE FROM story_comments WHERE comment_id=$1`, id)
	dueMayaReplies(t, f)
	count, err = f.repo.DispatchMayaReplies(ctx, 25, createMayaNotification(f))
	require.NoError(t, err)
	require.Zero(t, count)
	var rows int
	require.NoError(t, f.postgres.Pool.QueryRow(ctx, `SELECT count(*) FROM notifications WHERE workspace_id=$1`, f.workspaceA).Scan(&rows))
	require.Zero(t, rows)
}

func TestMayaReplyReplayRechecksAccessAfterNotificationWasPersisted(t *testing.T) {
	for _, test := range []struct {
		name   string
		revoke func(*testing.T, notificationIntegrationFixture)
	}{
		{name: "team removed", revoke: func(t *testing.T, f notificationIntegrationFixture) {
			mustNotificationExec(t, t.Context(), f.postgres.Pool, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.teamA, f.recipientA)
		}},
		{name: "workspace removed", revoke: func(t *testing.T, f notificationIntegrationFixture) {
			mustNotificationExec(t, t.Context(), f.postgres.Pool, `DELETE FROM workspace_members WHERE workspace_id=$1 AND user_id=$2`, f.workspaceA, f.recipientA)
		}},
		{name: "story deleted", revoke: func(t *testing.T, f notificationIntegrationFixture) {
			mustNotificationExec(t, t.Context(), f.postgres.Pool, `UPDATE stories SET deleted_at=clock_timestamp() WHERE id=$1`, f.storyA)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			f := newNotificationIntegrationFixture(t, t.Context())
			id, input := newMayaReplyNotification(t, f, f.recipientA)
			require.NoError(t, f.repo.DeferMayaReply(t.Context(), input, id))
			dueMayaReplies(t, f)
			_, err := f.repo.DispatchMayaReplies(t.Context(), 25, func(ctx context.Context, input domain.NewNotification) error {
				require.NoError(t, createMayaNotification(f)(ctx, input))
				return errors.New("Redis failed after inbox commit")
			})
			require.Error(t, err)
			test.revoke(t, f)
			dueMayaReplies(t, f)
			count, err := f.repo.DispatchMayaReplies(t.Context(), 25, func(context.Context, domain.NewNotification) error {
				t.Fatal("revoked recipient must not receive replayed realtime content")
				return nil
			})
			require.NoError(t, err)
			require.Equal(t, 1, count, "revoked delivery is retired rather than retried indefinitely")
		})
	}
}

func TestMayaReplyReplayUsesCurrentInAppPreference(t *testing.T) {
	t.Parallel()
	f := newNotificationIntegrationFixture(t, t.Context())
	id, input := newMayaReplyNotification(t, f, f.recipientA)
	require.NoError(t, f.repo.DeferMayaReply(t.Context(), input, id))
	dueMayaReplies(t, f)
	_, err := f.repo.DispatchMayaReplies(t.Context(), 25, func(ctx context.Context, input domain.NewNotification) error {
		require.NotNil(t, input.InAppEnabled)
		require.True(t, *input.InAppEnabled)
		require.NoError(t, createMayaNotification(f)(ctx, input))
		return errors.New("Redis failed after inbox commit")
	})
	require.Error(t, err)
	mustNotificationExec(t, t.Context(), f.postgres.Pool, `INSERT INTO notification_preferences(user_id,workspace_id,preferences) VALUES($1,$2,CAST('{"comment_reply":{"in_app":false}}' AS jsonb)) ON CONFLICT(user_id,workspace_id) DO UPDATE SET preferences=EXCLUDED.preferences`, f.recipientA, f.workspaceA)
	dueMayaReplies(t, f)
	count, err := f.repo.DispatchMayaReplies(t.Context(), 25, func(ctx context.Context, input domain.NewNotification) error {
		require.NotNil(t, input.InAppEnabled)
		require.False(t, *input.InAppEnabled, "realtime retry must honor newly disabled in-app delivery")
		return createMayaNotification(f)(ctx, input)
	})
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
