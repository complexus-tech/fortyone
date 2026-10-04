//go:build integration

package notificationsrepository

import (
	"testing"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/stretchr/testify/require"
)

func TestRoutineSweepRecipientsHavePendingWorkInTheirWorkspace(t *testing.T) {
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	recipients, err := f.repo.ListRoutineRecipients(ctx, nil, 100)
	require.NoError(t, err)
	require.Empty(t, recipients, "inactive workspaces with no pending notifications or subscriptions need no hourly wakeup")
	scope := notifications.DeliveryScope{RecipientID: f.recipientA, WorkspaceID: f.workspaceA}
	recipient, err := f.repo.GetRoutineRecipient(ctx, scope)
	require.NoError(t, err)
	require.NotNil(t, recipient, "direct delivery eligibility remains independent of sweep candidates")

	pending, _, err := f.repo.Create(ctx, f.storyNotification(f.recipientA, notificationDedupeKey("pending-summary")))
	require.NoError(t, err)
	read, _, err := f.repo.Create(ctx, f.storyNotification(f.guestA, notificationDedupeKey("read-summary")))
	require.NoError(t, err)
	sent, _, err := f.repo.Create(ctx, f.storyNotification(f.revocableA, notificationDedupeKey("sent-summary")))
	require.NoError(t, err)
	mustNotificationExec(t, ctx, f.postgres.Pool, "UPDATE notifications SET read_at=now() WHERE notification_id=$1", read.ID)
	mustNotificationExec(t, ctx, f.postgres.Pool, "UPDATE notifications SET email_sent_at=now() WHERE notification_id=$1", sent.ID)
	mustNotificationExec(t, ctx, f.postgres.Pool, `
		INSERT INTO feedback_board_subscriptions (board_id, user_id, email_frequency)
		VALUES ($1, $2, 'weekly'), ($1, $3, 'daily')
	`, f.boardB, f.recipientA, f.recipientB)

	recipients, err = f.repo.ListRoutineRecipients(ctx, nil, 100)
	require.NoError(t, err)
	require.Len(t, recipients, 2, "only unread unsent work or a subscription in this workspace wakes a member")
	keys := make([]string, 0, len(recipients))
	for _, row := range recipients {
		keys = append(keys, row.WorkspaceID.String()+":"+row.UserID.String())
	}
	require.ElementsMatch(t, []string{
		f.workspaceA.String() + ":" + f.recipientA.String(),
		f.workspaceB.String() + ":" + f.recipientB.String(),
	}, keys)

	mustNotificationExec(t, ctx, f.postgres.Pool, "UPDATE notifications SET read_at=now() WHERE notification_id=$1", pending.ID)
	recipients, err = f.repo.ListRoutineRecipients(ctx, nil, 100)
	require.NoError(t, err)
	require.Len(t, recipients, 1, "a subscription in another workspace cannot keep the first workspace awake")
	require.Equal(t, f.recipientB, recipients[0].UserID)
	require.Equal(t, f.workspaceB, recipients[0].WorkspaceID)
}
