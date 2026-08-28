package sse

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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

func TestHubRejectsRegistrationOutsideSupervisedRun(t *testing.T) {
	t.Parallel()

	hub, closeRedis := newLifecycleTestHub(t)
	defer closeRedis()

	client, err := hub.RegisterNewClient(uuid.New(), uuid.New())
	require.Nil(t, client)
	require.ErrorContains(t, err, "not accepting clients")
}

func TestHubRunStopsOnContextCancellation(t *testing.T) {
	t.Parallel()

	hub, closeRedis := newLifecycleTestHub(t)
	defer closeRedis()
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		result <- hub.Run(ctx)
	}()

	require.Eventually(t, func() bool {
		_, running := hub.lifecycleContext()
		return running
	}, time.Second, time.Millisecond)
	cancel()

	require.NoError(t, <-result)
	_, running := hub.lifecycleContext()
	require.False(t, running)
}

func TestHubRejectsConcurrentRuns(t *testing.T) {
	t.Parallel()

	hub, closeRedis := newLifecycleTestHub(t)
	defer closeRedis()
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		result <- hub.Run(ctx)
	}()
	require.Eventually(t, func() bool {
		_, running := hub.lifecycleContext()
		return running
	}, time.Second, time.Millisecond)

	require.ErrorContains(t, hub.Run(context.Background()), "already running")
	cancel()
	require.NoError(t, <-result)
}

func TestHubShutdownWaitsForActiveClientListeners(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { require.NoError(t, redisClient.Close()) })
	var output bytes.Buffer
	hub := NewHub(logger.NewWithJSON(&output, slog.LevelDebug, "sse-test"), redisClient)
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- hub.Run(ctx) }()
	require.Eventually(t, func() bool {
		_, running := hub.lifecycleContext()
		return running
	}, time.Second, time.Millisecond)

	client, err := hub.RegisterNewClient(uuid.New(), uuid.New())
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		hub.mu.RLock()
		defer hub.mu.RUnlock()
		return len(hub.clients) == 1
	}, time.Second, time.Millisecond)

	cancel()
	select {
	case err := <-result:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("SSE hub did not stop active client listeners within the shutdown budget")
	}
	require.ErrorIs(t, client.Ctx().Err(), context.Canceled)
	hub.mu.RLock()
	require.Empty(t, hub.clients)
	hub.mu.RUnlock()
}

func newLifecycleTestHub(t *testing.T) (*Hub, func()) {
	t.Helper()

	var output bytes.Buffer
	redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	hub := NewHub(logger.NewWithJSON(&output, slog.LevelDebug, "sse-test"), redisClient)
	return hub, func() {
		require.NoError(t, redisClient.Close())
	}
}
