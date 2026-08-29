package stories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

const (
	defaultStoryMutationEventBatch = 25
	storyMutationClaimTimeout      = 2 * time.Minute
	maximumMutationEventErrorRunes = 1000
)

var storyMutationRetrySchedule = []time.Duration{
	time.Minute,
	5 * time.Minute,
	30 * time.Minute,
	2 * time.Hour,
	8 * time.Hour,
	24 * time.Hour,
}

// StoryMutationPublication is the narrow caller-owned publication contract.
// Bootstrap may adapt this to outbound webhooks without importing that module
// into the story domain or service.
type StoryMutationPublication struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	StoryID     uuid.UUID
	Type        string
	Actor       platformauth.Actor
	Payload     []byte
	OccurredAt  time.Time
}

type StoryMutationEventPublisher interface {
	PublishStoryMutationEvent(context.Context, StoryMutationPublication) error
}

type StoryMutationEventRepository interface {
	ClaimStoryMutationEvents(context.Context, int, time.Time, time.Time) ([]storydomain.MutationEvent, error)
	CompleteStoryMutationEvent(context.Context, uuid.UUID, uuid.UUID, time.Time) error
	RetryStoryMutationEvent(context.Context, uuid.UUID, uuid.UUID, time.Time, time.Time, string) error
}

type mutationEventClock interface {
	Now() time.Time
}

type systemMutationEventClock struct{}

func (systemMutationEventClock) Now() time.Time { return time.Now() }

type StoryMutationEventDispatcher struct {
	repository StoryMutationEventRepository
	publisher  StoryMutationEventPublisher
	clock      mutationEventClock
	batchSize  int
}

func NewStoryMutationEventDispatcher(
	repository StoryMutationEventRepository,
	publisher StoryMutationEventPublisher,
) (*StoryMutationEventDispatcher, error) {
	return newStoryMutationEventDispatcher(
		repository, publisher, systemMutationEventClock{}, defaultStoryMutationEventBatch,
	)
}

func newStoryMutationEventDispatcher(
	repository StoryMutationEventRepository,
	publisher StoryMutationEventPublisher,
	clock mutationEventClock,
	batchSize int,
) (*StoryMutationEventDispatcher, error) {
	if repository == nil || publisher == nil || clock == nil || batchSize < 1 || batchSize > 100 {
		return nil, errors.New("story mutation event dispatcher dependencies are invalid")
	}
	return &StoryMutationEventDispatcher{
		repository: repository, publisher: publisher, clock: clock, batchSize: batchSize,
	}, nil
}

// DispatchBatch transfers a bounded set of committed story intents to the
// product event publisher. Publication retries always reuse the persisted event
// ID; a failed downstream write never replays the story mutation.
func (dispatcher *StoryMutationEventDispatcher) DispatchBatch(ctx context.Context) (int, error) {
	now := dispatcher.clock.Now().UTC()
	events, err := dispatcher.repository.ClaimStoryMutationEvents(
		ctx, dispatcher.batchSize, now, now.Add(-storyMutationClaimTimeout),
	)
	if err != nil {
		return 0, fmt.Errorf("claim story mutation events: %w", err)
	}
	processed := 0
	for _, event := range events {
		publication := StoryMutationPublication{
			ID: event.ID, WorkspaceID: event.WorkspaceID, StoryID: event.StoryID,
			Type: string(event.Type), Actor: event.Actor,
			Payload: append([]byte(nil), event.Payload...), OccurredAt: event.OccurredAt.UTC(),
		}
		publishErr := dispatcher.publisher.PublishStoryMutationEvent(ctx, publication)
		completedAt := dispatcher.clock.Now().UTC()
		if completedAt.Before(now) {
			completedAt = now
		}
		if publishErr == nil {
			if err := dispatcher.repository.CompleteStoryMutationEvent(
				ctx, event.ID, event.ClaimToken, completedAt,
			); err != nil {
				return processed, fmt.Errorf("complete story mutation event %s: %w", event.ID, err)
			}
			processed++
			continue
		}

		nextAttemptAt := completedAt.Add(storyMutationRetryDelay(event.AttemptCount))
		if err := dispatcher.repository.RetryStoryMutationEvent(
			ctx,
			event.ID,
			event.ClaimToken,
			nextAttemptAt,
			completedAt,
			sanitizeMutationEventError(publishErr),
		); err != nil {
			return processed, fmt.Errorf("schedule story mutation event %s retry: %w", event.ID, err)
		}
		processed++
	}
	return processed, nil
}

func storyMutationRetryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	index := min(attempt-1, len(storyMutationRetrySchedule)-1)
	return storyMutationRetrySchedule[index]
}

func sanitizeMutationEventError(err error) string {
	if err == nil {
		return ""
	}
	message := strings.Join(strings.Fields(err.Error()), " ")
	runes := []rune(message)
	if len(runes) > maximumMutationEventErrorRunes {
		message = strings.TrimSpace(string(runes[:maximumMutationEventErrorRunes-3])) + "..."
	}
	return message
}
