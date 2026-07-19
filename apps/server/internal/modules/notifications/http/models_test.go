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
