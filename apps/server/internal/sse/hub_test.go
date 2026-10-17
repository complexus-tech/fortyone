package sse

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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

func TestClientSubscribesToAllScopesOnOneSocketBeforeAcknowledgement(t *testing.T) {
	server := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer redisClient.Close()
	hub := NewHub(logger.NewWithText(io.Discard, slog.LevelError, "sse-test"), redisClient)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- hub.Run(ctx) }()
	require.Eventually(t, func() bool { _, running := hub.lifecycleContext(); return running }, time.Second, time.Millisecond)
	client, err := hub.RegisterNewClient(uuid.New(), uuid.New())
	require.NoError(t, err)
	notificationsChannel := pubSubChannelPrefix + client.UserID.String()
	updatesChannel := publisher.UserUpdatesChannel(client.UserID)
	workspaceChannel := workspaceChannelPrefix + client.WorkspaceID.String()
	require.Equal(t, map[string]int{notificationsChannel: 1, updatesChannel: 1, workspaceChannel: 1}, server.PubSubNumSub(notificationsChannel, updatesChannel, workspaceChannel))
	require.Equal(t, 1, server.CurrentConnectionCount(), "one Pub/Sub connection owns all three scopes")
	publish := func(channel string, payload any) {
		encoded, err := json.Marshal(payload)
		require.NoError(t, err)
		require.NoError(t, redisClient.Publish(ctx, channel, encoded).Err())
	}
	notification := notifications.CoreNotification{ID: uuid.New(), RecipientID: client.UserID, WorkspaceID: client.WorkspaceID, Message: notifications.NotificationMessage{Template: "Changed", IdentityReferences: map[string]uuid.UUID{"actor": uuid.New()}}}
	publish(notificationsChannel, notification)
	select {
	case payload := <-client.Send:
		require.NotContains(t, string(payload), "identityReferences")
		require.Contains(t, string(payload), notification.ID.String())
	case <-time.After(time.Second):
		t.Fatal("notification was not forwarded")
	}
	update := publisher.UserUpdate{Type: "chat.updated", UserID: client.UserID, WorkspaceID: client.WorkspaceID}
	publish(updatesChannel, update)
	select {
	case payload := <-client.Send:
		require.Contains(t, string(payload), "chat.updated")
	case <-time.After(time.Second):
		t.Fatal("user update was not forwarded")
	}
	update.Type = "chat.typing"
	publish(updatesChannel, update)
	select {
	case payload := <-client.Send:
		require.Contains(t, string(payload), "chat.typing")
	case <-time.After(time.Second):
		t.Fatal("typing update was not forwarded")
	}
	publish(workspaceChannel, map[string]string{"type": "story.updated"})
	select {
	case payload := <-client.Send:
		require.Contains(t, string(payload), "story.updated")
	case <-time.After(time.Second):
		t.Fatal("workspace update was not forwarded")
	}
	notification.WorkspaceID = uuid.New()
	publish(notificationsChannel, notification)
	notification.WorkspaceID = client.WorkspaceID
	notification.RecipientID = uuid.New()
	publish(notificationsChannel, notification)
	update.UserID = uuid.New()
	publish(updatesChannel, update)
	update.UserID = client.UserID
	update.WorkspaceID = uuid.New()
	publish(updatesChannel, update)
	require.Never(t, func() bool { return len(client.Send) > 0 }, 50*time.Millisecond, time.Millisecond, "out-of-scope events must not leak through a multiplexed socket")
	hub.UnregisterClient(client)
	require.Eventually(t, func() bool {
		return server.PubSubNumSub(notificationsChannel, updatesChannel, workspaceChannel)[updatesChannel] == 0
	}, time.Second, time.Millisecond)
	cancel()
	require.NoError(t, <-done)
}

func TestClientSubscriptionFailureFailsRegistrationAndCleansClient(t *testing.T) {
	server := miniredis.RunT(t)
	server.RequireAuth("not-configured-in-client")
	redisClient := redis.NewClient(&redis.Options{Addr: server.Addr(), MaxRetries: -1})
	defer redisClient.Close()
	hub := NewHub(logger.NewWithText(io.Discard, slog.LevelError, "sse-test"), redisClient)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- hub.Run(ctx) }()
	require.Eventually(t, func() bool { _, running := hub.lifecycleContext(); return running }, time.Second, time.Millisecond)
	client, err := hub.RegisterNewClient(uuid.New(), uuid.New())
	require.Error(t, err)
	require.Nil(t, client)
	require.Eventually(t, func() bool { hub.mu.RLock(); defer hub.mu.RUnlock(); return len(hub.clients) == 0 }, time.Second, time.Millisecond)
	cancel()
	require.NoError(t, <-done)
}

func TestRedisReconnectClosesStreamForAuthoritativeReconciliation(t *testing.T) {
	server := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer redisClient.Close()
	hub := NewHub(logger.NewWithText(io.Discard, slog.LevelError, "sse-test"), redisClient)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- hub.Run(ctx) }()
	require.Eventually(t, func() bool { _, running := hub.lifecycleContext(); return running }, time.Second, time.Millisecond)
	client, err := hub.RegisterNewClient(uuid.New(), uuid.New())
	require.NoError(t, err)
	defer hub.UnregisterClient(client)

	// go-redis automatically reconnects subscriptions. The browser must learn
	// that hints may have been missed rather than retaining a healthy-looking
	// SSE stream and stale data after Redis has returned.
	server.Close()
	require.NoError(t, server.Restart())
	select {
	case <-client.Ctx().Done():
	case <-time.After(5 * time.Second):
		t.Fatal("Redis reconnection did not close SSE for history reconciliation")
	}
	cancel()
	require.NoError(t, <-done)
}

func TestSlowClientIsDisconnectedToRecoverDroppedHints(t *testing.T) {
	server := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer redisClient.Close()
	hub := NewHub(logger.NewWithText(io.Discard, slog.LevelError, "sse-test"), redisClient)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- hub.Run(ctx) }()
	require.Eventually(t, func() bool { _, running := hub.lifecycleContext(); return running }, time.Second, time.Millisecond)
	client, err := hub.RegisterNewClient(uuid.New(), uuid.New())
	require.NoError(t, err)
	for range cap(client.Send) {
		client.Send <- []byte("buffer occupied")
	}
	require.NoError(t, redisClient.Publish(ctx, workspaceChannelPrefix+client.WorkspaceID.String(), `{"type":"story.updated"}`).Err())
	select {
	case <-client.Ctx().Done():
	case <-time.After(2 * clientSendTimeout):
		t.Fatal("slow client stayed connected after losing an update")
	}
	hub.UnregisterClient(client)
	cancel()
	require.NoError(t, <-done)
}
