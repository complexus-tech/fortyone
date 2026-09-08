//go:build integration

package storiesrepository

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/complexus-tech/projects-api/internal/eventconsumer"
	"github.com/complexus-tech/projects-api/internal/migrations"
	notificationdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	notificationrepo "github.com/complexus-tech/projects-api/internal/modules/notifications/repository"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	statesrepo "github.com/complexus-tech/projects-api/internal/modules/states/repository"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	usersrepo "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// Covers the actual mutation -> stream -> consumer -> inbox -> realtime path.
// Providers are not invoked; PostgreSQL and Redis state are owned by this test.
func TestStoryPriorityAndStatusUpdatesReachInbox(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	mustStoryReadExec(t, ctx, postgres.Pool, "ALTER TABLE notifications ADD CONSTRAINT notifications_recipient_workspace_entity_unique UNIQUE (recipient_id, workspace_id, entity_type, entity_id)")
	mustStoryReadExec(t, ctx, postgres.Pool, "CREATE UNIQUE INDEX legacy_notification_entity_order ON notifications (recipient_id, workspace_id, entity_id, entity_type)")
	repair, err := migrations.FS.ReadFile("000186_notification_event_uniqueness.up.sql")
	require.NoError(t, err)
	mustStoryReadExec(t, ctx, postgres.Pool, string(repair))
	// Reapplying the repair must preserve the event dedupe index and other indexes.
	mustStoryReadExec(t, ctx, postgres.Pool, string(repair))
	fixture := seedStoryReadFixture(t, ctx, postgres.Pool)
	recipient := uuid.New()
	insertStoryReadUser(t, ctx, postgres.Pool, recipient, true)
	mustStoryReadExec(t, ctx, postgres.Pool, "INSERT INTO workspace_members (workspace_id,user_id,role) VALUES ($1,$2,'member')", fixture.workspaceA, recipient)
	mustStoryReadExec(t, ctx, postgres.Pool, "INSERT INTO team_members (team_id,user_id) VALUES ($1,$2)", fixture.teamA, recipient)
	mustStoryReadExec(t, ctx, postgres.Pool, "UPDATE stories SET assignee_id=$1 WHERE id=$2", recipient, fixture.visible)
	statusID := uuid.New()
	mustStoryReadExec(t, ctx, postgres.Pool, "INSERT INTO statuses (status_id,workspace_id,team_id,name,category) VALUES ($1,$2,$3,'To do','unstarted')", statusID, fixture.workspaceA, fixture.teamA)
	log := logger.NewWithText(io.Discard, slog.LevelDebug, "notification-delivery-test")
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	queue, err := tasks.New(client, log)
	require.NoError(t, err)
	storyService := stories.New(log, New(log, postgres.Pool), publisher.New(client, log), queue)
	inboxRepo := notificationrepo.New(postgres.Pool)
	inbox := notifications.New(log, inboxRepo, client, queue)
	consumer := eventconsumer.New(client, log, "http://localhost", inbox, nil, storyService, nil, users.New(log, usersrepo.New(postgres.Pool), queue), states.New(statesrepo.New(postgres.Pool)), nil, nil, nil)
	require.NoError(t, consumer.Initialize(ctx))
	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- consumer.Run(runCtx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Error("consumer did not stop")
		}
	})
	actorCtx := auth.SetUserID(ctx, fixture.actor)
	subscription := client.Subscribe(ctx, "user-notifications:"+recipient.String())
	defer subscription.Close()
	_, err = subscription.Receive(ctx)
	require.NoError(t, err)
	var lastEvent events.Event
	for index, update := range []map[string]any{{"priority": "Urgent"}, {"status_id": statusID}} {
		require.NoError(t, storyService.Update(actorCtx, fixture.visible, fixture.workspaceA, update))
		rows, err := client.XRevRangeN(ctx, "events-stream", "+", "-", 1).Result()
		require.NoError(t, err)
		require.Len(t, rows, 1)
		var event events.Event
		require.NoError(t, json.Unmarshal([]byte(rows[0].Values["payload"].(string)), &event))
		body, err := json.Marshal(event.Payload)
		require.NoError(t, err)
		var payload events.StoryUpdatedPayload
		require.NoError(t, json.Unmarshal(body, &payload))
		require.Contains(t, payload.AudienceIDs, recipient)
		require.Equal(t, fixture.actor, event.ActorID)
		lastEvent = event
		require.Eventually(t, func() bool {
			count, err := inboxRepo.CountUnread(ctx, notificationdomain.WorkspaceAccess{ActorID: recipient, WorkspaceID: fixture.workspaceA})
			return err == nil && count == index+1
		}, 5*time.Second, 20*time.Millisecond, "no inbox notification for %v; published payload: %s", update, body)
		realtimeCtx, stop := context.WithTimeout(ctx, 2*time.Second)
		message, err := subscription.ReceiveMessage(realtimeCtx)
		stop()
		require.NoError(t, err)
		var notification notificationdomain.Notification
		require.NoError(t, json.Unmarshal([]byte(message.Payload), &notification))
		require.Equal(t, fixture.visible, notification.EntityID)
		require.Equal(t, recipient, notification.RecipientID)
	}
	items, err := inboxRepo.List(ctx, notificationdomain.ListQuery{Access: notificationdomain.WorkspaceAccess{ActorID: recipient, WorkspaceID: fixture.workspaceA}, Limit: 20})
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, "To do", items[0].Message.Variables["value"].Value)
	require.Equal(t, "Urgent", items[1].Message.Variables["value"].Value)

	// Exact event replay remains idempotent after removing resource uniqueness.
	require.NoError(t, publisher.New(client, log).Publish(ctx, lastEvent))
	rows, err := client.XRevRangeN(ctx, "events-stream", "+", "-", 1).Result()
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		groups, err := client.XInfoGroups(ctx, "events-stream").Result()
		return err == nil && len(groups) == 1 && groups[0].LastDeliveredID == rows[0].ID && groups[0].Pending == 0
	}, 5*time.Second, 20*time.Millisecond)
	count, err := inboxRepo.CountUnread(ctx, notificationdomain.WorkspaceAccess{ActorID: recipient, WorkspaceID: fixture.workspaceA})
	require.NoError(t, err)
	require.Equal(t, 2, count)
}
