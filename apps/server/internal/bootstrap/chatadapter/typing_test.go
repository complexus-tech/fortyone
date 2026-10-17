package chatadapter

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type typingPublisherStub struct {
	mu     sync.Mutex
	events []publisher.ChatTyping
}

func (p *typingPublisherStub) PublishChatTyping(_ context.Context, event publisher.ChatTyping) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, event)
	return nil
}

func TestTypingThrottleExpiresAndAllowsImmediateStopWithBoundedKeys(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	broadcaster := NewTypingBroadcaster(client, &typingPublisherStub{})
	workspace := uuid.New()
	actor, err := auth.NewHumanActor(uuid.New()).WithWorkspace(workspace)
	require.NoError(t, err)
	command := domain.TypingCommand{Scope: domain.ActorScope{WorkspaceID: workspace, Actor: actor}, ConversationID: uuid.New(), IsTyping: true}
	allowed, err := broadcaster.AllowTyping(t.Context(), command)
	require.NoError(t, err)
	require.False(t, allowed.IsZero())
	allowed, err = broadcaster.AllowTyping(t.Context(), command)
	require.NoError(t, err)
	require.True(t, allowed.IsZero())
	command.IsTyping = false
	allowed, err = broadcaster.AllowTyping(t.Context(), command)
	require.NoError(t, err)
	require.False(t, allowed.IsZero(), "stop is delivered immediately")
	command.IsTyping = true
	allowed, err = broadcaster.AllowTyping(t.Context(), command)
	require.NoError(t, err)
	require.False(t, allowed.IsZero(), "a new typing session can begin immediately")
	for range 10 {
		root := uuid.New()
		command.ThreadRootID = &root
		_, err = broadcaster.AllowTyping(t.Context(), command)
		require.NoError(t, err)
	}
	require.Len(t, server.Keys(), 1, "thread switches cannot accumulate throttle keys")
	server.FastForward(3 * time.Second)
	allowed, err = broadcaster.AllowTyping(t.Context(), command)
	require.NoError(t, err)
	require.False(t, allowed.IsZero())
	server.FastForward(3 * time.Second)
	require.Empty(t, server.Keys())
}

func TestTypingFanoutScopesActorThreadAndExpiryForStartAndStop(t *testing.T) {
	updates := &typingPublisherStub{}
	broadcaster := NewTypingBroadcaster(nil, updates)
	actor, peer, root := uuid.New(), uuid.New(), uuid.New()
	activity := domain.TypingActivity{WorkspaceID: uuid.New(), ConversationID: uuid.New(), ThreadRootID: &root, Actor: domain.Person{ID: actor, FullName: "Joseph"}, RecipientIDs: []uuid.UUID{actor, peer}, IsTyping: true, ExpiresAt: time.Now().UTC().Add(6 * time.Second)}
	before := time.Now().UTC().Add(6 * time.Second)
	require.NoError(t, broadcaster.PublishTyping(t.Context(), activity))
	activity.IsTyping = false
	activity.ExpiresAt = time.Now().UTC().Add(6 * time.Second)
	require.NoError(t, broadcaster.PublishTyping(t.Context(), activity))
	require.Len(t, updates.events, 2)
	for _, event := range updates.events {
		require.Equal(t, peer, event.UserID)
		require.Equal(t, actor, event.ActorID)
		require.Equal(t, "Joseph", event.FullName)
		require.Equal(t, activity.WorkspaceID, event.WorkspaceID)
		require.Equal(t, activity.ConversationID, event.ConversationID)
		require.Equal(t, &root, event.ThreadRootID)
		require.WithinDuration(t, before, event.ExpiresAt, time.Second)
	}
	require.True(t, updates.events[0].IsTyping)
	require.False(t, updates.events[1].IsTyping)
	require.False(t, updates.events[1].ExpiresAt.Before(updates.events[0].ExpiresAt))
}
