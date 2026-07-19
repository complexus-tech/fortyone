package consumer

import (
	"testing"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

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
