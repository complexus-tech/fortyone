package sse

import (
	"testing"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNotificationMatchesWorkspace(t *testing.T) {
	workspaceID := uuid.New()
	notification := notifications.CoreNotification{WorkspaceID: workspaceID}

	require.True(t, notificationMatchesWorkspace(notification, workspaceID))
	require.False(t, notificationMatchesWorkspace(notification, uuid.New()))
	require.False(t, notificationMatchesWorkspace(notification, uuid.Nil))
}
