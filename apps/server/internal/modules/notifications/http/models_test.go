package notificationshttp

import (
	"testing"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestToAppPortalNotificationsResponseIncludesSelfContainedFeedbackMetadata(t *testing.T) {
	notificationID := uuid.New()
	actorID := uuid.New()
	feedbackID := uuid.New()
	avatarURL := "https://cdn.example.com/avatar.png"

	response := toAppPortalNotificationsResponse([]notifications.CorePortalNotification{{
		Notification: notifications.CoreNotification{
			ID:         notificationID,
			Type:       "feedback_comment",
			EntityType: "feedback",
			EntityID:   feedbackID,
			ActorID:    actorID,
			Title:      "Export roadmap to PDF",
		},
		ActorName:     "Maya Chen",
		ActorAvatar:   &avatarURL,
		FeedbackTitle: "Export roadmap to PDF",
		FeedbackSlug:  "export-roadmap-to-pdf",
	}}, 1, 20, false)

	require.Len(t, response.Notifications, 1)
	result := response.Notifications[0]
	require.Equal(t, notificationID, result.ID)
	require.Equal(t, actorID, result.Actor.ID)
	require.Equal(t, "Maya Chen", result.Actor.Name)
	require.Equal(t, &avatarURL, result.Actor.AvatarURL)
	require.Equal(t, feedbackID, result.Feedback.ID)
	require.Equal(t, "/feedback/export-roadmap-to-pdf", result.Feedback.Path)
}

func TestToAppNotificationRedactsEmailDeliveryContext(t *testing.T) {
	source := notifications.CoreNotification{
		ID:         uuid.New(),
		EntityType: "strategy",
		Message: notifications.NotificationMessage{
			Template: "Your weekly strategy check-in",
			Strategy: &notifications.StrategyNotificationSnapshot{
				Version: 1,
				Kind:    notifications.StrategyNotificationKindWeeklyCheckIn,
			},
		},
	}
	appNotification := toAppNotification(source)

	require.Nil(t, appNotification.Message.Strategy)
	require.Equal(t, "Strategy guidance is ready to review.", appNotification.Message.Template)
	require.Empty(t, appNotification.Message.Variables)
	require.Equal(t, "Your weekly strategy check-in", source.Message.Template)
	require.NotNil(t, source.Message.Strategy)
}

func TestToAppNotificationIncludesActorAndRemovesSystemActorMessagePrefix(t *testing.T) {
	actorID := uuid.New()
	avatarURL := "https://cdn.example.com/maya.png"

	appNotification := toAppNotification(notifications.CoreNotification{
		ActorID: actorID,
		Actor: &notifications.CoreNotificationActor{
			ID:        actorID,
			Username:  "maya",
			FullName:  "Maya",
			AvatarURL: avatarURL,
			IsActive:  true,
			IsSystem:  true,
		},
		Message: notifications.NotificationMessage{
			Template: "Maya scheduled this task to {scheduled_for}",
			Variables: map[string]notifications.Variable{
				"scheduled_for": {Value: "Wednesday at 11:00", Type: "date"},
			},
		},
	})

	require.NotNil(t, appNotification.Actor)
	require.Equal(t, actorID, appNotification.Actor.ID)
	require.Equal(t, avatarURL, appNotification.Actor.AvatarURL)
	require.True(t, appNotification.Actor.IsSystem)
	require.Equal(t, "scheduled this task to {scheduled_for}", appNotification.Message.Template)
}

func TestToAppPortalNotificationRedactsDeliveryContext(t *testing.T) {
	source := notifications.CoreNotification{
		ID:         uuid.New(),
		EntityType: "feedback",
		EntityID:   uuid.New(),
		Message: notifications.NotificationMessage{
			Template: "Maya commented on your feedback",
			Strategy: &notifications.StrategyNotificationSnapshot{Version: 1},
		},
	}

	response := toAppPortalNotificationsResponse([]notifications.CorePortalNotification{{
		Notification: source,
		FeedbackSlug: "improve-onboarding",
	}}, 1, 20, false)

	require.Len(t, response.Notifications, 1)
	require.Nil(t, response.Notifications[0].Message.Strategy)
	require.Equal(t, source.Message.Template, response.Notifications[0].Message.Template)
	require.NotNil(t, source.Message.Strategy)
}
