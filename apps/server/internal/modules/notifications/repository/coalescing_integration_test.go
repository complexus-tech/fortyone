//go:build integration

package notificationsrepository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUnreadStoryUpdatesCoalesceWithOrderedEventsAndReadHistory(t *testing.T) {
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	base := time.Now().UTC().Add(-4 * time.Hour).Truncate(time.Microsecond)
	input := f.storyNotification(f.recipientA, notificationDedupeKey("status"))
	input.OccurredAt = base
	first, changed, err := f.repo.Create(ctx, input)
	require.NoError(t, err)
	require.True(t, changed)

	comment := f.storyNotification(f.recipientA, notificationDedupeKey("comment"))
	comment.Type = notifications.NotificationTypeStoryComment
	comment.OccurredAt = base.Add(time.Minute)
	comment.Message.Template = "A comment that must remain visible"
	commentRow, _, err := f.repo.Create(ctx, comment)
	require.NoError(t, err)

	priority := f.storyNotification(f.recipientA, notificationDedupeKey("priority"))
	priority.OccurredAt = base.Add(2 * time.Minute)
	priority.ActorID = f.system
	priority.Title = "Latest story title"
	priority.Message.Template = "Priority changed to urgent"
	updated, changed, err := f.repo.Create(ctx, priority)
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, first.ID, updated.ID)
	require.Equal(t, f.system, updated.ActorID)
	require.Equal(t, priority.Message.Template, updated.Message.Template)
	require.True(t, updated.CreatedAt.Equal(priority.OccurredAt))
	list, err := f.repo.List(ctx, notifications.ListQuery{Access: f.accessA(f.recipientA), Limit: 20})
	require.NoError(t, err)
	require.Len(t, list, 2)
	require.Equal(t, first.ID, list[0].ID, "refresh moves the story above intervening comment activity")
	require.Equal(t, commentRow.ID, list[1].ID)

	replayed, changed, err := f.repo.Create(ctx, input)
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, updated.Message.Template, replayed.Message.Template, "older exact replay cannot revert the inbox")
	conflicting := input
	conflicting.Title = "Changed original event"
	_, _, err = f.repo.Create(ctx, conflicting)
	require.ErrorIs(t, err, notifications.ErrConflict, "immutable event payload remains protected after refresh")
	stale := f.storyNotification(f.recipientA, notificationDedupeKey("late-event"))
	stale.OccurredAt = base.Add(time.Minute)
	_, changed, err = f.repo.Create(ctx, stale)
	require.NoError(t, err)
	require.False(t, changed)

	otherRecipient := priority
	otherRecipient.DedupeKey = notificationDedupeKey("independent-recipient")
	otherRecipient.RecipientID = f.guestA
	separate, _, err := f.repo.Create(ctx, otherRecipient)
	require.NoError(t, err)
	require.NotEqual(t, first.ID, separate.ID)

	observed := first.CreatedAt
	err = f.repo.Mutate(ctx, notifications.NotificationMutation{Access: f.accessA(f.recipientA), NotificationID: first.ID, Kind: notifications.NotificationMutationRead, At: time.Now().UTC(), ExpectedCreatedAt: &observed})
	require.ErrorIs(t, err, notifications.ErrConflict, "stale client cannot consume unseen content")
	observed = updated.CreatedAt
	require.NoError(t, f.repo.Mutate(ctx, notifications.NotificationMutation{Access: f.accessA(f.recipientA), NotificationID: first.ID, Kind: notifications.NotificationMutationRead, At: time.Now().UTC(), ExpectedCreatedAt: &observed}))
	deadline := priority
	deadline.DedupeKey = notificationDedupeKey("deadline")
	deadline.OccurredAt = base.Add(3 * time.Minute)
	deadline.Message.Template = "Deadline changed"
	fresh, _, err := f.repo.Create(ctx, deadline)
	require.NoError(t, err)
	require.NotEqual(t, first.ID, fresh.ID, "read history receives a new inbox row")
	require.ErrorIs(t, f.repo.Mutate(ctx, notifications.NotificationMutation{Access: f.accessA(f.recipientA), NotificationID: first.ID, Kind: notifications.NotificationMutationUnread, At: time.Now().UTC()}), notifications.ErrConflict)

	var writers sync.WaitGroup
	errors := make(chan error, 8)
	for index := range 8 {
		writers.Add(1)
		go func() {
			defer writers.Done()
			event := deadline
			event.DedupeKey = notificationDedupeKey(fmt.Sprintf("concurrent-%d", index))
			event.OccurredAt = base.Add(time.Duration(index+4) * time.Minute)
			event.Message.Template = fmt.Sprintf("Update %d", index)
			_, _, err := f.repo.Create(ctx, event)
			errors <- err
		}()
	}
	writers.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
	list, err = f.repo.List(ctx, notifications.ListQuery{Access: f.accessA(f.recipientA), Limit: 20})
	require.NoError(t, err)
	require.Len(t, list, 3, "one read history row, one unread story update, and the comment")
	require.Equal(t, fresh.ID, list[0].ID)
	require.Equal(t, "Update 7", list[0].Message.Template)
	var receiptCount int
	require.NoError(t, f.postgres.Pool.QueryRow(ctx, "SELECT count(*) FROM notification_event_receipts WHERE recipient_id=$1 AND notification_type='story_update'", f.recipientA).Scan(&receiptCount))
	require.Equal(t, 12, receiptCount, "all accepted events retain independent retry receipts")

	require.NoError(t, f.repo.Mutate(ctx, notifications.NotificationMutation{Access: f.accessA(f.recipientA), NotificationID: fresh.ID, Kind: notifications.NotificationMutationDelete, At: time.Now().UTC()}))
	deletedReplay, changed, err := f.repo.Create(ctx, deadline)
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, uuid.Nil, deletedReplay.ID, "deleted inbox row must not be resurrected by replay")
}

func TestEmailCompletionCoversTheSentSnapshotRatherThanNewInboxContent(t *testing.T) {
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	scope := notifications.DeliveryScope{RecipientID: f.recipientA, WorkspaceID: f.workspaceA}
	input := f.storyNotification(f.recipientA, notificationDedupeKey("snapshot-old"))
	original, _, err := f.repo.Create(ctx, input)
	require.NoError(t, err)
	before, err := f.repo.ListEmailDigest(ctx, scope)
	require.NoError(t, err)
	require.Len(t, before.Items, 1)
	oldSnapshot := notifications.EmailSnapshot{NotificationID: original.ID, ContentHash: before.Items[0].ContentHash}
	input.DedupeKey = notificationDedupeKey("snapshot-new")
	input.Message.Template = "A newer unread deadline update"
	updated, _, err := f.repo.Create(ctx, input)
	require.NoError(t, err)
	require.Equal(t, original.ID, updated.ID)
	require.NoError(t, f.repo.MarkEmailSent(ctx, notifications.MarkEmailSent{Scope: scope, NotificationSnapshots: []notifications.EmailSnapshot{oldSnapshot}, At: time.Now().UTC()}))
	after, err := f.repo.ListEmailDigest(ctx, scope)
	require.NoError(t, err)
	require.NotNil(t, after, "old send must not consume new content on the reused ID")
	require.Len(t, after.Items, 1)
	require.NotEqual(t, oldSnapshot.ContentHash, after.Items[0].ContentHash)
	require.NoError(t, f.repo.MarkEmailSent(ctx, notifications.MarkEmailSent{Scope: scope, NotificationSnapshots: []notifications.EmailSnapshot{{NotificationID: updated.ID, ContentHash: after.Items[0].ContentHash}}, At: time.Now().UTC()}))
	empty, err := f.repo.ListEmailDigest(ctx, scope)
	require.NoError(t, err)
	require.Nil(t, empty)
}

func TestObservedReadRacingStoryRefreshNeverConsumesTheNewUpdate(t *testing.T) {
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	base := time.Now().UTC().Add(-time.Hour).Truncate(time.Microsecond)
	for iteration := range 8 {
		_, err := f.repo.MutateAll(ctx, notifications.WorkspaceMutation{Access: f.accessA(f.recipientA), Kind: notifications.WorkspaceMutationReadAll, At: time.Now().UTC()})
		require.NoError(t, err)
		input := f.storyNotification(f.recipientA, notificationDedupeKey("race-before"))
		input.OccurredAt = base.Add(time.Duration(iteration*2) * time.Minute)
		before, _, err := f.repo.Create(ctx, input)
		require.NoError(t, err)
		next := f.storyNotification(f.recipientA, notificationDedupeKey("race-after"))
		next.OccurredAt = input.OccurredAt.Add(time.Minute)
		next.Message.Template = "The update that was not displayed"
		var group sync.WaitGroup
		group.Add(2)
		readErrors := make(chan error, 1)
		updateErrors := make(chan error, 1)
		go func() {
			defer group.Done()
			readErrors <- f.repo.Mutate(ctx, notifications.NotificationMutation{Access: f.accessA(f.recipientA), NotificationID: before.ID, Kind: notifications.NotificationMutationRead, At: time.Now().UTC(), ExpectedCreatedAt: &before.CreatedAt})
		}()
		go func() {
			defer group.Done()
			_, _, err := f.repo.Create(ctx, next)
			updateErrors <- err
		}()
		group.Wait()
		require.NoError(t, <-updateErrors)
		if err := <-readErrors; err != nil {
			require.ErrorIs(t, err, notifications.ErrConflict)
		}
		count, err := f.repo.CountUnread(ctx, f.accessA(f.recipientA))
		require.NoError(t, err)
		require.Equal(t, 1, count, "whichever operation wins the row lock, the unseen update remains unread")
	}
}

func TestScheduleIssueInboxAndEmailStayBoundToLiveEpisodeAndOwner(t *testing.T) {
	ctx := t.Context()
	f := newNotificationIntegrationFixture(t, ctx)
	issueID := uuid.New()
	mustNotificationExec(t, ctx, f.postgres.Pool, "UPDATE stories SET assignee_id=$1,auto_scheduling_enabled=TRUE WHERE id=$2", f.recipientA, f.storyA)
	mustNotificationExec(t, ctx, f.postgres.Pool, "INSERT INTO story_schedule_issues(workspace_id,story_id,issue_id,owner_id,cause_code,opened_at,updated_at) VALUES($1,$2,$3,$4,'cannot_fit',now(),now())", f.workspaceA, f.storyA, issueID, f.recipientA)
	input := f.storyNotification(f.recipientA, notificationDedupeKey("issue"))
	input.Message.Template = "Maya cannot fit this task"
	input.Message.ScheduleIssue = &notifications.StoryScheduleIssueSnapshot{ID: issueID, Code: "cannot_fit", OwnerID: f.recipientA}
	issue, _, err := f.repo.Create(ctx, input)
	require.NoError(t, err)
	_, _, err = f.repo.Create(ctx, f.storyNotification(f.recipientA, notificationDedupeKey("ordinary-priority")))
	require.NoError(t, err)
	count, err := f.repo.CountUnread(ctx, f.accessA(f.recipientA))
	require.NoError(t, err)
	require.Equal(t, 2, count, "routine story update cannot erase unresolved scheduling action")
	scope := notifications.DeliveryScope{RecipientID: f.recipientA, WorkspaceID: f.workspaceA}
	digest, err := f.repo.ListEmailDigest(ctx, scope)
	require.NoError(t, err)
	require.Len(t, digest.Items, 2)
	mustNotificationExec(t, ctx, f.postgres.Pool, "UPDATE stories SET assignee_id=$1 WHERE id=$2", f.guestA, f.storyA)
	requireNotificationIssueHidden(t, ctx, f, issue.ID, scope)
	mustNotificationExec(t, ctx, f.postgres.Pool, "UPDATE stories SET assignee_id=$1 WHERE id=$2", f.recipientA, f.storyA)
	mustNotificationExec(t, ctx, f.postgres.Pool, "UPDATE story_schedule_issues SET resolved_at=now() WHERE story_id=$1", f.storyA)
	requireNotificationIssueHidden(t, ctx, f, issue.ID, scope)
}

func requireNotificationIssueHidden(t *testing.T, ctx context.Context, f notificationIntegrationFixture, issueNotificationID uuid.UUID, scope notifications.DeliveryScope) {
	t.Helper()
	count, err := f.repo.CountUnread(ctx, f.accessA(f.recipientA))
	require.NoError(t, err)
	require.Equal(t, 1, count)
	list, err := f.repo.List(ctx, notifications.ListQuery{Access: f.accessA(f.recipientA), Limit: 20})
	require.NoError(t, err)
	require.Len(t, list, 1)
	delivery, err := f.repo.GetEmailDelivery(ctx, notifications.EmailNotificationQuery{Scope: scope, NotificationID: issueNotificationID})
	require.NoError(t, err)
	require.Nil(t, delivery)
	push, err := f.repo.GetPushDelivery(ctx, issueNotificationID)
	require.NoError(t, err)
	require.Nil(t, push)
}
