package chatadapter

import (
	"context"
	"fmt"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

type TypingPublisher interface {
	PublishChatTyping(context.Context, publisher.ChatTyping) error
}

type TypingBroadcaster struct {
	redis     *redis.Client
	publisher TypingPublisher
}

func NewTypingBroadcaster(client *redis.Client, updates TypingPublisher) *TypingBroadcaster {
	return &TypingBroadcaster{redis: client, publisher: updates}
}

var typingThrottle = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then return {} end
redis.call('SET', KEYS[1], ARGV[1], 'PX', 3000)
return redis.call('TIME')
`)

func (b *TypingBroadcaster) AllowTyping(ctx context.Context, command domain.TypingCommand) (time.Time, error) {
	// One short-lived key per workspace actor bounds memory even when users
	// switch conversations or threads. State transitions pass immediately;
	// repeated activity in the same scope is capped at once per three seconds.
	key := fmt.Sprintf("chat:typing:%s:%s", command.Scope.WorkspaceID, command.Scope.Actor.PrincipalID)
	thread := "main"
	if command.ThreadRootID != nil {
		thread = command.ThreadRootID.String()
	}
	state := fmt.Sprintf("%s:%s:%t", command.ConversationID, thread, command.IsTyping)
	accepted, err := typingThrottle.Run(ctx, b.redis, []string{key}, state).Int64Slice()
	if err != nil {
		return time.Time{}, fmt.Errorf("throttle chat typing: %w", err)
	}
	if len(accepted) == 0 {
		return time.Time{}, nil
	}
	if len(accepted) != 2 {
		return time.Time{}, fmt.Errorf("chat typing throttle returned invalid timestamp")
	}
	return time.Unix(accepted[0], accepted[1]*int64(time.Microsecond)).UTC().Add(6 * time.Second), nil
}

func (b *TypingBroadcaster) PublishTyping(ctx context.Context, activity domain.TypingActivity) error {
	// Start and stop share the same expiry convention so clients can reject
	// delayed events using a server-derived timestamp, including late starts.
	// The timestamp was assigned atomically when Redis accepted the activity,
	// before membership lookup or fanout could delay an older start behind a stop.
	if !activity.ExpiresAt.After(time.Now()) {
		return nil
	}
	group, work := errgroup.WithContext(ctx)
	group.SetLimit(16)
	for _, recipient := range activity.RecipientIDs {
		if work.Err() != nil {
			break
		}
		if recipient == activity.Actor.ID {
			continue
		}
		event := publisher.ChatTyping{
			WorkspaceID: activity.WorkspaceID, UserID: recipient, ConversationID: activity.ConversationID,
			ActorID: activity.Actor.ID, FullName: activity.Actor.FullName, ThreadRootID: activity.ThreadRootID,
			IsTyping: activity.IsTyping, ExpiresAt: activity.ExpiresAt,
		}
		group.Go(func() error { return b.publisher.PublishChatTyping(work, event) })
	}
	if err := group.Wait(); err != nil {
		return err
	}
	return ctx.Err()
}
