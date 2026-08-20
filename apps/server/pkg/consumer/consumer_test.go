package consumer

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type retryingNotificationCreator struct {
	attempts  map[string]int
	created   map[string]notifications.CoreNotification
	failFirst map[uuid.UUID]error
}

func (c *retryingNotificationCreator) Create(_ context.Context, notification notifications.CoreNewNotification) (notifications.CoreNotification, error) {
	if c.attempts == nil {
		c.attempts = make(map[string]int)
	}
	if c.created == nil {
		c.created = make(map[string]notifications.CoreNotification)
	}
	c.attempts[notification.DedupeKey]++
	if err := c.failFirst[notification.RecipientID]; err != nil && c.attempts[notification.DedupeKey] == 1 {
		return notifications.CoreNotification{}, err
	}
	if existing, ok := c.created[notification.DedupeKey]; ok {
		return existing, nil
	}
	created := notifications.CoreNotification{
		ID:          uuid.New(),
		RecipientID: notification.RecipientID,
		WorkspaceID: notification.WorkspaceID,
		Type:        notification.Type,
		EntityType:  notification.EntityType,
		EntityID:    notification.EntityID,
		ActorID:     notification.ActorID,
		Title:       notification.Title,
		Message:     notification.Message,
	}
	c.created[notification.DedupeKey] = created
	return created, nil
}

func TestWithEventDedupeKeyKeepsDistinctStoryEventsSeparate(t *testing.T) {
	recipientID := uuid.New()
	storyID := uuid.New()
	notification := notifications.CoreNewNotification{
		RecipientID: recipientID,
		Type:        "story_comment",
		EntityType:  "story",
		EntityID:    storyID,
	}
	firstEvent := events.Event{
		Type:      events.CommentCreated,
		Timestamp: time.Date(2026, 7, 20, 9, 0, 0, 0, time.UTC),
	}
	secondEvent := events.Event{
		Type:      events.CommentCreated,
		Timestamp: firstEvent.Timestamp.Add(time.Second),
	}

	first := withEventDedupeKey(firstEvent, notification, 0)
	retry := withEventDedupeKey(firstEvent, notification, 0)
	second := withEventDedupeKey(secondEvent, notification, 0)

	require.Equal(t, first.DedupeKey, retry.DedupeKey)
	require.NotEqual(t, first.DedupeKey, second.DedupeKey)
}

func TestWithEventDedupeKeyPreservesExplicitSourceKey(t *testing.T) {
	notification := notifications.CoreNewNotification{DedupeKey: "feedback-comment:source-id"}

	result := withEventDedupeKey(events.Event{}, notification, 0)

	require.Equal(t, notification.DedupeKey, result.DedupeKey)
}

func TestCreateStoryUpdateNotificationsReturnsPartialFailuresAndRetriesIdempotently(t *testing.T) {
	t.Parallel()

	firstRecipientID := uuid.New()
	secondRecipientID := uuid.New()
	thirdRecipientID := uuid.New()
	firstErr := errors.New("first notification unavailable")
	thirdErr := errors.New("third notification unavailable")
	creator := &retryingNotificationCreator{failFirst: map[uuid.UUID]error{
		firstRecipientID: firstErr,
		thirdRecipientID: thirdErr,
	}}
	consumer := &Consumer{
		log:           logger.NewWithText(io.Discard, slog.LevelDebug, "consumer-test"),
		notifications: creator,
	}
	event := events.Event{
		Type:      events.StoryUpdated,
		Timestamp: time.Date(2026, 8, 18, 8, 0, 0, 0, time.UTC),
		ActorID:   uuid.New(),
	}
	batch := []notifications.CoreNewNotification{
		{RecipientID: firstRecipientID, Type: "story_update"},
		{RecipientID: secondRecipientID, Type: "story_update"},
		{RecipientID: thirdRecipientID, Type: "story_update"},
	}

	err := consumer.createStoryUpdateNotifications(context.Background(), event, batch)
	require.ErrorIs(t, err, firstErr)
	require.ErrorIs(t, err, thirdErr)
	require.Len(t, creator.attempts, 3, "the first failure must not prevent later recipients from being attempted")
	require.Len(t, creator.created, 1)

	require.NoError(t, consumer.createStoryUpdateNotifications(context.Background(), event, batch))
	require.Len(t, creator.created, 3, "retry must insert only notifications that failed previously")
	for dedupeKey, attempts := range creator.attempts {
		require.NotEmpty(t, dedupeKey)
		require.Equal(t, 2, attempts)
	}
}

func TestHandleStoryUpdatedReturnsAggregatedNotificationFailures(t *testing.T) {
	t.Parallel()

	oldAssigneeID := uuid.New()
	newAssigneeID := uuid.New()
	oldAssigneeErr := errors.New("old assignee notification unavailable")
	newAssigneeErr := errors.New("new assignee notification unavailable")
	creator := &retryingNotificationCreator{failFirst: map[uuid.UUID]error{
		oldAssigneeID: oldAssigneeErr,
		newAssigneeID: newAssigneeErr,
	}}
	testLogger := logger.NewWithText(io.Discard, slog.LevelDebug, "consumer-test")
	consumer := &Consumer{
		log:               testLogger,
		notifications:     creator,
		notificationRules: notifications.NewRules(testLogger, nil, nil, nil),
	}
	event := events.Event{
		Type:      events.StoryUpdated,
		Timestamp: time.Date(2026, 8, 18, 8, 0, 0, 0, time.UTC),
		ActorID:   uuid.New(),
		Payload: events.StoryUpdatedPayload{
			StoryID:     uuid.New(),
			WorkspaceID: uuid.New(),
			AssigneeID:  &oldAssigneeID,
			Updates:     map[string]any{"assignee_id": newAssigneeID},
		},
	}

	err := consumer.handleStoryUpdated(context.Background(), event)
	require.ErrorIs(t, err, oldAssigneeErr)
	require.ErrorIs(t, err, newAssigneeErr)
	require.Len(t, creator.attempts, 2, "both reassignment recipients must be attempted before the event is retried")
}

func TestShouldBridgeFeedbackStatusOnlyForCommittedStatusTransitions(t *testing.T) {
	previous := uuid.New()
	require.True(t, shouldBridgeFeedbackStatus(events.StoryUpdatedPayload{
		Updates: map[string]any{"status_id": uuid.New()}, PreviousStatusID: &previous,
	}))
	require.False(t, shouldBridgeFeedbackStatus(events.StoryUpdatedPayload{
		Updates: map[string]any{"status_id": uuid.New()},
	}), "legacy or failed transitions without a captured previous status must not produce a bridge event")
	require.False(t, shouldBridgeFeedbackStatus(events.StoryUpdatedPayload{
		Updates: map[string]any{"title": "Changed"}, PreviousStatusID: &previous,
	}))
}

func TestShouldReconcileStoryScheduleUsesOnlySchedulingFields(t *testing.T) {
	require.True(t, shouldReconcileStorySchedule(map[string]any{"estimated_duration_minutes": 90}))
	require.True(t, shouldReconcileStorySchedule(map[string]any{"minimum_focus_block_minutes": 30}))
	require.True(t, shouldReconcileStorySchedule(map[string]any{"assignee_id": uuid.New()}))
	require.True(t, shouldReconcileStorySchedule(map[string]any{"status_id": uuid.New()}))
	require.True(t, shouldReconcileStorySchedule(map[string]any{"end_date": time.Now()}))
	require.True(t, shouldReconcileStorySchedule(map[string]any{"title": "Copy edit"}))
	require.False(t, shouldReconcileStorySchedule(map[string]any{"description": "Copy edit"}))
}

func TestAutoSchedulingUpdatesBroadcastWithoutRequeueingReconciliation(t *testing.T) {
	updates := map[string]any{
		"auto_scheduling_status":     "at_risk",
		"auto_scheduling_reason":     "A new busy event displaced the reserved time.",
		"auto_scheduling_updated_at": time.Date(2026, 8, 18, 8, 0, 0, 0, time.UTC),
	}

	consumer := &Consumer{}
	require.True(t, consumer.hasSignificantChanges(updates))
	require.False(t, shouldReconcileStorySchedule(updates), "scheduler-owned state must not create a reconciliation loop")
	require.Equal(t, map[string]any{
		"autoSchedulingStatus":    updates["auto_scheduling_status"],
		"autoSchedulingReason":    updates["auto_scheduling_reason"],
		"autoSchedulingUpdatedAt": updates["auto_scheduling_updated_at"],
	}, frontendStoryChanges(updates))
}

func TestCompletionUpdatesBroadcastForCalendarRefresh(t *testing.T) {
	completedAt := time.Date(2026, 8, 20, 13, 0, 0, 0, time.UTC)
	updates := map[string]any{"completed_at": completedAt}

	consumer := &Consumer{}
	require.True(t, consumer.hasSignificantChanges(updates))
	require.True(t, shouldReconcileStorySchedule(updates))
	require.Equal(t, map[string]any{
		"completedAt": completedAt,
	}, frontendStoryChanges(updates))
}
