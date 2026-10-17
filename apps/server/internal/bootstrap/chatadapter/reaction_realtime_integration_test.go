//go:build integration

package chatadapter

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	chat "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatrepository "github.com/complexus-tech/projects-api/internal/modules/chat/repository"
	chatservice "github.com/complexus-tech/projects-api/internal/modules/chat/service"
	notificationsrepository "github.com/complexus-tech/projects-api/internal/modules/notifications/repository"
	notificationsservice "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/sse"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// Exercise the committed reaction -> durable outbox -> Redis -> SSE path using
// the real authorization queries. The control DB must be disposable; testkit
// creates and removes a separately migrated database for this test.
func TestReactionsReachAuthorizedSSEClientsForMessagesAndReplies(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspace, team, author, peer, outsider := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	_, err := postgres.Pool.Exec(ctx, `INSERT INTO workspaces(workspace_id,name,slug) VALUES($1,'Reaction test',$2)`, workspace, workspace.String())
	require.NoError(t, err)
	for _, id := range []uuid.UUID{author, peer, outsider} {
		_, err = postgres.Pool.Exec(ctx, `INSERT INTO users(user_id,username,email,full_name) VALUES($1,$2,$3,'Member')`, id, id.String(), id.String()+"@example.test")
		require.NoError(t, err)
		_, err = postgres.Pool.Exec(ctx, `INSERT INTO workspace_members(workspace_id,user_id,role) VALUES($1,$2,'admin')`, workspace, id)
		require.NoError(t, err)
	}
	_, err = postgres.Pool.Exec(ctx, `INSERT INTO teams(team_id,workspace_id,name,code,color) VALUES($1,$2,'Product','PROD','#111111')`, team, workspace)
	require.NoError(t, err)
	for _, id := range []uuid.UUID{author, peer} {
		_, err = postgres.Pool.Exec(ctx, `INSERT INTO team_members(team_id,user_id) VALUES($1,$2)`, team, id)
		require.NoError(t, err)
	}
	var room uuid.UUID
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT id FROM chat_conversations WHERE workspace_id=$1 AND team_id=$2`, workspace, team).Scan(&room))
	scope := func(id uuid.UUID) chat.ActorScope {
		actor, err := auth.NewHumanActor(id).WithWorkspace(workspace)
		require.NoError(t, err)
		return chat.ActorScope{WorkspaceID: workspace, Actor: actor}
	}
	repository := chatrepository.New(nil, postgres.Pool)
	service := chatservice.New(repository)
	root, err := service.CreateMessage(ctx, chat.CreateMessageCommand{Scope: scope(author), ConversationID: room, CreationID: uuid.New(), Text: "A message to react to"})
	require.NoError(t, err)
	reply, err := service.CreateMessage(ctx, chat.CreateMessageCommand{Scope: scope(author), ConversationID: room, CreationID: uuid.New(), ThreadRootID: &root.ID, Text: "A reply to react to"})
	require.NoError(t, err)
	_, err = service.MarkRead(ctx, chat.ReadCommand{Scope: scope(author), ConversationID: room, ThroughMessageID: root.ID})
	require.NoError(t, err)
	_, err = service.MarkRead(ctx, chat.ReadCommand{Scope: scope(author), ConversationID: room, ThreadRootID: &root.ID, ThroughMessageID: reply.ID})
	require.NoError(t, err)
	// Only the subsequent reaction events belong to this delivery assertion.
	_, err = postgres.Pool.Exec(ctx, `UPDATE chat_outbox SET delivered_at=now() WHERE workspace_id=$1`, workspace)
	require.NoError(t, err)

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { require.NoError(t, redisClient.Close()) })
	log := logger.NewWithText(io.Discard, slog.LevelError, "reaction-test")
	hub := sse.NewHub(log, redisClient)
	hubCtx, stopHub := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- hub.Run(hubCtx) }()
	t.Cleanup(func() {
		stopHub()
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(time.Second):
			t.Error("SSE hub did not stop")
		}
	})
	var authorClient *sse.Client
	require.Eventually(t, func() bool {
		authorClient, err = hub.RegisterNewClient(author, workspace)
		return err == nil
	}, time.Second, time.Millisecond)
	peerClient, err := hub.RegisterNewClient(peer, workspace)
	require.NoError(t, err)
	outsiderClient, err := hub.RegisterNewClient(outsider, workspace)
	require.NoError(t, err)
	foreignClient, err := hub.RegisterNewClient(author, uuid.New())
	require.NoError(t, err)
	queue := &reactionNotificationTasks{}
	notices := notificationsservice.New(log, notificationsrepository.New(postgres.Pool), redisClient, queue)
	dispatcher := Dispatcher{Store: repository, Updates: publisher.New(redisClient, log), Notifications: notices}
	expectedReceipts := 0

	for _, message := range []chat.Message{root, reply} {
		for _, active := range []bool{true, false} {
			_, err = service.SetReaction(ctx, scope(peer), message.ID, "👍", active)
			require.NoError(t, err)
			_, err = service.SetReaction(ctx, scope(peer), message.ID, "👍", active)
			require.NoError(t, err, "repeated PUT/DELETE remains idempotent")
			count, err := dispatcher.Dispatch(ctx, 50)
			require.NoError(t, err)
			require.Equal(t, 1, count, "only an actual reaction change emits a committed event")
			for _, client := range []*sse.Client{authorClient, peerClient} {
				select {
				case payload := <-client.Send:
					var hint publisher.ChatUpdate
					require.NoError(t, json.Unmarshal(payload, &hint))
					require.Equal(t, publisher.ChatUpdatedType, hint.Type)
					require.Equal(t, "reaction.updated", hint.Change)
					require.Equal(t, client.UserID, hint.UserID)
					require.Equal(t, workspace, hint.WorkspaceID)
					require.Equal(t, room, hint.ConversationID)
					require.Equal(t, &message.ID, hint.MessageID)
				case <-time.After(time.Second):
					t.Fatal("reaction did not reach the connected SSE client")
				}
			}
			current, err := service.GetMessage(ctx, scope(author), message.ID)
			require.NoError(t, err)
			if active {
				require.Len(t, current.Reactions, 1)
				require.Equal(t, 1, current.Reactions[0].Count)
				require.False(t, current.Reactions[0].Reacted, "author sees the peer's reaction, not their own")
				expectedReceipts++
				select {
				case payload := <-authorClient.Send:
					var alert publisher.ChatNotification
					require.NoError(t, json.Unmarshal(payload, &alert))
					require.Equal(t, publisher.ChatNotificationType, alert.Type)
					require.Equal(t, author, alert.UserID)
					require.Equal(t, peer, alert.ActorID)
					require.Equal(t, workspace, alert.WorkspaceID)
					require.Equal(t, room, alert.ConversationID)
					require.Equal(t, message.ID, alert.MessageID)
					require.Equal(t, message.ThreadRootID, alert.ThreadRootID)
					require.Contains(t, alert.Body, "Member reacted 👍 to your message: "+message.Text)
					require.NotEqual(t, uuid.Nil, alert.NotificationID)
				case <-time.After(time.Second):
					t.Fatal("reaction notification did not reach the message author through SSE")
				}
			} else {
				require.Empty(t, current.Reactions)
			}
			var receipts int
			require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT count(*) FROM notifications WHERE workspace_id=$1`, workspace).Scan(&receipts))
			require.Equal(t, expectedReceipts, receipts, "duplicates and removals must not create additional notifications")
			require.Never(t, func() bool { return len(authorClient.Send)+len(peerClient.Send) > 0 }, 20*time.Millisecond, time.Millisecond)
		}
	}
	// Self-reactions still update the chat but never create an interruption.
	_, err = service.SetReaction(ctx, scope(author), root.ID, "👀", true)
	require.NoError(t, err)
	count, err := dispatcher.Dispatch(ctx, 50)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	for _, client := range []*sse.Client{authorClient, peerClient} {
		select {
		case payload := <-client.Send:
			var hint publisher.ChatUpdate
			require.NoError(t, json.Unmarshal(payload, &hint))
			require.Equal(t, publisher.ChatUpdatedType, hint.Type)
		case <-time.After(time.Second):
			t.Fatal("self-reaction did not update the chat")
		}
	}

	// Recipient authorization is evaluated again at delivery, not just at write.
	_, err = service.SetReaction(ctx, scope(peer), root.ID, "✅", true)
	require.NoError(t, err)
	_, err = postgres.Pool.Exec(ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, team, peer)
	require.NoError(t, err)
	count, err = dispatcher.Dispatch(ctx, 50)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	select {
	case <-authorClient.Send:
	case <-time.After(time.Second):
		t.Fatal("remaining member did not receive the reaction")
	}
	require.Never(t, func() bool {
		return len(authorClient.Send)+len(peerClient.Send)+len(outsiderClient.Send)+len(foreignClient.Send) > 0
	}, 50*time.Millisecond, time.Millisecond, "revoked, nonmember and different-workspace sessions receive no reaction content")
	var receipts int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT count(*) FROM notifications WHERE workspace_id=$1`, workspace).Scan(&receipts))
	require.Equal(t, 2, receipts, "only the two peer additions to the main message and reply notify the author")
	require.Len(t, queue.pushes, 2)
	require.Zero(t, queue.digests, "reaction alerts must not add email noise")
}

type reactionNotificationTasks struct {
	pushes  []tasks.NotificationPushPayload
	digests int
}

func (q *reactionNotificationTasks) EnqueueNotificationEmailDigest(tasks.NotificationEmailDigestPayload, ...asynq.Option) (*asynq.TaskInfo, error) {
	q.digests++
	return &asynq.TaskInfo{ID: "test-digest"}, nil
}

func (q *reactionNotificationTasks) EnqueueNotificationPush(payload tasks.NotificationPushPayload, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	q.pushes = append(q.pushes, payload)
	return &asynq.TaskInfo{ID: "test-push"}, nil
}

func (q *reactionNotificationTasks) EnqueueNotificationPushTest(tasks.NotificationPushTestPayload, ...asynq.Option) (*asynq.TaskInfo, error) {
	return &asynq.TaskInfo{ID: "test-push"}, nil
}
