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
		Template: "{actor} moved the task to {value}",
		Variables: map[string]Variable{
			"actor": {Value: "Maya Chen", Type: "actor"},
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
	require.Contains(t, rendered, "Here&#39;s what changed while you were away.")
	require.Contains(t, rendered, "Maya Chen")
	require.Contains(t, rendered, "moved the task to")
	require.Contains(t, rendered, "for task")
	require.Contains(t, rendered, ">Ship billing states</a>")
	require.Contains(t, rendered, ">Review onboarding copy</a>")
	require.Contains(t, rendered, "https://product.fortyone.app/notifications/")
	require.Contains(t, rendered, "<strong")
	require.Contains(t, rendered, "font-weight: 600")
	require.Contains(t, rendered, "padding: 0 0 10px")
	require.Contains(t, rendered, "padding: 10px 0")
	require.Contains(t, rendered, "font-size: 15px")
	require.Contains(t, rendered, "border-top: 0")
	require.Contains(t, rendered, "border-top: 1px solid #e5e5e5")
	require.NotContains(t, rendered, "display: block; margin: 0 0 4px")
	require.NotContains(t, rendered, "font-size: 14px")
}
