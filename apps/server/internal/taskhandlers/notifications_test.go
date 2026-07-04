package taskhandlers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBuildNotificationDigestSubject(t *testing.T) {
	require.Equal(t, "1 update in Product", buildNotificationDigestSubject("Product", 1))
	require.Equal(t, "3 updates in Product", buildNotificationDigestSubject("Product", 3))
}

func TestFormatNotificationDigestMessageGroupsUnreadNotifications(t *testing.T) {
	message := NotificationMessage{
		Template: "{actor} moved {field} to {value}",
		Variables: map[string]Variable{
			"actor": {Value: "Maya Chen", Type: "actor"},
			"field": {Value: "status", Type: "field"},
			"value": {Value: "In review", Type: "value"},
		},
	}
	rawMessage, err := json.Marshal(message)
	require.NoError(t, err)

	items := []NotificationEmailDigestItem{
		{
			NotificationID: uuid.New(),
			Title:          "Ship billing states",
			Message:        rawMessage,
			CreatedAt:      time.Date(2026, 7, 4, 8, 30, 0, 0, time.UTC),
		},
		{
			NotificationID: uuid.New(),
			Title:          "Review onboarding copy",
			Message:        rawMessage,
			CreatedAt:      time.Date(2026, 7, 4, 8, 40, 0, 0, time.UTC),
		},
	}

	rendered, err := formatNotificationDigestMessage(items, "https://product.fortyone.app")

	require.NoError(t, err)
	require.Contains(t, rendered, "You have 2 unread updates")
	require.Contains(t, rendered, "Ship billing states")
	require.Contains(t, rendered, "Review onboarding copy")
	require.Contains(t, rendered, "https://product.fortyone.app/notifications/")
	require.Contains(t, rendered, "<strong")
}
