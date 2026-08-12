package sse

import (
	"testing"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/complexus-tech/projects-api/pkg/publisher"
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

func TestUserUpdateMatchesClientScope(t *testing.T) {
	userID := uuid.New()
	workspaceID := uuid.New()
	client := &Client{UserID: userID, WorkspaceID: workspaceID}

	require.True(t, userUpdateMatchesClient(publisher.UserUpdate{UserID: userID, WorkspaceID: workspaceID}, client))
	require.False(t, userUpdateMatchesClient(publisher.UserUpdate{UserID: uuid.New(), WorkspaceID: workspaceID}, client))
	require.False(t, userUpdateMatchesClient(publisher.UserUpdate{UserID: userID, WorkspaceID: uuid.New()}, client))
	require.False(t, userUpdateMatchesClient(publisher.UserUpdate{UserID: userID, WorkspaceID: workspaceID}, nil))
}
