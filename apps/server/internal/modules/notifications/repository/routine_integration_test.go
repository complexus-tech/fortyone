//go:build integration

package notificationsrepository

import (
	"sync"
	"testing"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRoutineEmailClaimsSerializeAndFenceDeliveryAttempts(t *testing.T) {
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	notification, _, err := f.repo.Create(ctx, f.storyNotification(f.recipientA, notificationDedupeKey("routine")))
	require.NoError(t, err)
	now := time.Now().UTC()
	claim := notifications.RoutineClaim{RecipientID: f.recipientA, WorkspaceID: f.workspaceA, Key: "briefing:today", Kind: "briefing", LocalDate: now, Now: now}
	var lock sync.Mutex
	var winners []uuid.UUID
	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			id, err := f.repo.ClaimRoutine(ctx, claim)
			if err == nil && id != uuid.Nil {
				lock.Lock()
				winners = append(winners, id)
				lock.Unlock()
			}
		}()
	}
	group.Wait()
	require.Len(t, winners, 1, "only one worker can own a person's send window")
	id := winners[0]
	scope := notifications.DeliveryScope{RecipientID: f.recipientA, WorkspaceID: f.workspaceA}
	completion := notifications.RoutineCompletion{ID: id, Scope: scope, NotificationIDs: []uuid.UUID{notification.ID}, GuidanceDate: &now, Sent: true, Now: now}
	wrongScope := completion
	wrongScope.Scope.WorkspaceID = f.workspaceB
	require.Error(t, f.repo.CompleteRoutine(ctx, wrongScope))
	require.NoError(t, f.repo.CompleteRoutine(ctx, completion))
	covered, err := f.repo.HasRoutineGuidance(ctx, scope, now)
	require.NoError(t, err)
	require.True(t, covered, "activity and guidance coverage commit together")
	digest, err := f.repo.ListEmailDigest(ctx, scope)
	require.NoError(t, err)
	require.Nil(t, digest, "completion covers notifications in the same transaction")
	replay, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	require.Equal(t, uuid.Nil, replay)
	claim.Key, claim.Kind = "activity:new", "activity"
	old, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	require.NoError(t, f.repo.FailRoutine(ctx, old))
	current, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	require.NotEqual(t, old, current, "a retry must fence the old owner")
	completion.ID = old
	require.Error(t, f.repo.CompleteRoutine(ctx, completion))
	claim.Now = now.Add(11 * time.Minute)
	reclaimed, err := f.repo.ClaimRoutine(ctx, claim)
	require.NoError(t, err)
	require.NotEqual(t, current, reclaimed)
	completion.ID = current
	require.Error(t, f.repo.CompleteRoutine(ctx, completion))
	completion.ID = reclaimed
	require.NoError(t, f.repo.CompleteRoutine(ctx, completion))
}

func TestSentNotificationsStayOutOfLaterDigests(t *testing.T) {
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	scope := notifications.DeliveryScope{RecipientID: f.recipientA, WorkspaceID: f.workspaceA}
	now := time.Now().UTC()
	var inputs []notifications.NewNotification
	var ids []uuid.UUID
	for range 4 {
		input := f.storyNotification(f.recipientA, notificationDedupeKey("digest-update"))
		item, inserted, err := f.repo.Create(ctx, input)
		require.NoError(t, err)
		require.True(t, inserted)
		inputs = append(inputs, input)
		ids = append(ids, item.ID)
	}
	claim, err := f.repo.ClaimRoutine(ctx, notifications.RoutineClaim{
		RecipientID: scope.RecipientID, WorkspaceID: scope.WorkspaceID, Key: "activity:first",
		Kind: "activity", LocalDate: now, Now: now,
	})
	require.NoError(t, err)
	require.NoError(t, f.repo.CompleteRoutine(ctx, notifications.RoutineCompletion{
		ID: claim, Scope: scope, NotificationIDs: ids, Sent: true, Now: now,
	}))
	for hour := 1; hour <= 3; hour++ {
		for _, input := range inputs {
			_, inserted, err := f.repo.Create(ctx, input)
			require.NoError(t, err)
			require.False(t, inserted)
		}
		digest, err := f.repo.ListEmailDigest(ctx, scope)
		require.NoError(t, err)
		require.Nil(t, digest, "later queue runs must remain empty even while sent notifications are unread")
	}
	newUpdate := f.storyNotification(f.recipientA, notificationDedupeKey("new-update"))
	newUpdate.Message.Template = "{actor} changed the deadline"
	item, _, err := f.repo.Create(ctx, newUpdate)
	require.NoError(t, err)
	digest, err := f.repo.ListEmailDigest(ctx, scope)
	require.NoError(t, err)
	require.NotNil(t, digest)
	require.Len(t, digest.Items, 1, "the next digest must contain only new information")
	require.Equal(t, item.ID, digest.Items[0].NotificationID)
}

func TestIdenticalContentWithFreshEventIDsIsCoveredAfterTwoHours(t *testing.T) {
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	scope := notifications.DeliveryScope{RecipientID: f.recipientA, WorkspaceID: f.workspaceA}
	original, _, err := f.repo.Create(ctx, f.storyNotification(f.recipientA, notificationDedupeKey("original")))
	require.NoError(t, err)
	sentAt := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Microsecond)
	require.NoError(t, f.repo.MarkEmailSent(ctx, notifications.MarkEmailSent{Scope: scope, NotificationIDs: []uuid.UUID{original.ID}, At: sentAt}))
	// Removing an inbox item must not remove its delivery receipt.
	_, err = f.postgres.Pool.Exec(ctx, "DELETE FROM public.notifications WHERE notification_id = $1", original.ID)
	require.NoError(t, err)
	for range 3 {
		repeated, inserted, err := f.repo.Create(ctx, f.storyNotification(f.recipientA, notificationDedupeKey("repeated")))
		require.NoError(t, err)
		require.True(t, inserted)
		digest, err := f.repo.ListEmailDigest(ctx, scope)
		require.NoError(t, err)
		require.Nil(t, digest)
		var coveredAt time.Time
		require.NoError(t, f.postgres.Pool.QueryRow(ctx, "SELECT email_sent_at FROM public.notifications WHERE notification_id = $1", repeated.ID).Scan(&coveredAt))
		require.True(t, sentAt.Equal(coveredAt))
		single, err := f.repo.GetEmailDelivery(ctx, notifications.EmailNotificationQuery{Scope: scope, NotificationID: repeated.ID})
		require.NoError(t, err)
		require.Nil(t, single)
	}
	// Content coverage is specific to a recipient, actor, entity and message.
	for _, change := range []string{"recipient", "actor", "entity", "message"} {
		input := f.storyNotification(f.recipientA, notificationDedupeKey(change))
		switch change {
		case "recipient":
			input.RecipientID = f.guestA
		case "actor":
			input.ActorID = f.revocableA
		case "entity":
			input.EntityID = uuid.New()
			insertNotificationStory(t, ctx, f.postgres.Pool, input.EntityID, f.teamA, f.workspaceA, "Another task")
		case "message":
			input.Message.Variables["date"] = notifications.Variable{Value: "15 Sep 2026", Type: "date"}
			input.Message.Template = "{actor} moved this task to {date}"
		}
		item, _, err := f.repo.Create(ctx, input)
		require.NoError(t, err)
		single, err := f.repo.GetEmailDelivery(ctx, notifications.EmailNotificationQuery{Scope: notifications.DeliveryScope{RecipientID: input.RecipientID, WorkspaceID: input.WorkspaceID}, NotificationID: item.ID})
		require.NoError(t, err)
		require.NotNil(t, single, change)
	}
}
