package stories

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

type mutationEventRepositoryStub struct {
	events        []storydomain.MutationEvent
	claimErr      error
	completed     []uuid.UUID
	retried       []uuid.UUID
	nextAttemptAt time.Time
	lastError     string
}

func (repository *mutationEventRepositoryStub) ClaimStoryMutationEvents(
	context.Context,
	int,
	time.Time,
	time.Time,
) ([]storydomain.MutationEvent, error) {
	return append([]storydomain.MutationEvent(nil), repository.events...), repository.claimErr
}

func (repository *mutationEventRepositoryStub) CompleteStoryMutationEvent(
	_ context.Context,
	eventID, _ uuid.UUID,
	_ time.Time,
) error {
	repository.completed = append(repository.completed, eventID)
	return nil
}

func (repository *mutationEventRepositoryStub) RetryStoryMutationEvent(
	_ context.Context,
	eventID, _ uuid.UUID,
	nextAttemptAt, _ time.Time,
	lastError string,
) error {
	repository.retried = append(repository.retried, eventID)
	repository.nextAttemptAt = nextAttemptAt
	repository.lastError = lastError
	return nil
}

type mutationEventPublisherStub struct {
	publications []StoryMutationPublication
	err          error
}

func (publisher *mutationEventPublisherStub) PublishStoryMutationEvent(
	_ context.Context,
	event StoryMutationPublication,
) error {
	publisher.publications = append(publisher.publications, event)
	return publisher.err
}

func TestStoryMutationEventDispatcherCompletesPublishedEventWithStableID(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	event := dispatchableStoryMutationEvent(t, now, 1)
	repository := &mutationEventRepositoryStub{events: []storydomain.MutationEvent{event}}
	publisher := &mutationEventPublisherStub{}
	dispatcher, err := newStoryMutationEventDispatcher(
		repository, publisher, testkit.NewFixedClock(now), 5,
	)
	if err != nil {
		t.Fatalf("construct dispatcher: %v", err)
	}
	processed, err := dispatcher.DispatchBatch(t.Context())
	if err != nil || processed != 1 {
		t.Fatalf("DispatchBatch() processed=%d error=%v", processed, err)
	}
	if len(publisher.publications) != 1 || publisher.publications[0].ID != event.ID ||
		len(repository.completed) != 1 || repository.completed[0] != event.ID || len(repository.retried) != 0 {
		t.Fatalf("publication=%#v completed=%#v retried=%#v", publisher.publications, repository.completed, repository.retried)
	}
}

func TestStoryMutationEventDispatcherDurablySchedulesPublishFailure(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	event := dispatchableStoryMutationEvent(t, now, 2)
	repository := &mutationEventRepositoryStub{events: []storydomain.MutationEvent{event}}
	publisher := &mutationEventPublisherStub{err: errors.New("  downstream   unavailable  ")}
	dispatcher, err := newStoryMutationEventDispatcher(
		repository, publisher, testkit.NewFixedClock(now), 5,
	)
	if err != nil {
		t.Fatalf("construct dispatcher: %v", err)
	}
	processed, err := dispatcher.DispatchBatch(t.Context())
	if err != nil || processed != 1 {
		t.Fatalf("DispatchBatch() processed=%d error=%v", processed, err)
	}
	if len(repository.retried) != 1 || len(repository.completed) != 0 {
		t.Fatalf("completed=%#v retried=%#v", repository.completed, repository.retried)
	}
	if want := now.Add(5 * time.Minute); !repository.nextAttemptAt.Equal(want) {
		t.Fatalf("next attempt = %v, want %v", repository.nextAttemptAt, want)
	}
	if repository.lastError != "downstream unavailable" {
		t.Fatalf("sanitized error = %q", repository.lastError)
	}
	if publisher.publications[0].ID != event.ID {
		t.Fatalf("published ID = %s, want stable %s", publisher.publications[0].ID, event.ID)
	}
}

func TestStoryMutationEventDispatcherCompletesInternalOnlyEventWithoutPublishing(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 30, 12, 0, 0, 0, time.UTC)
	event := dispatchableStoryMutationEvent(t, now, 1)
	event.Payload = []byte(`{"storyId":"test","_delivery":"internal_only"}`)
	repository := &mutationEventRepositoryStub{events: []storydomain.MutationEvent{event}}
	publisher := &mutationEventPublisherStub{}
	dispatcher, err := newStoryMutationEventDispatcher(
		repository, publisher, testkit.NewFixedClock(now), 5,
	)
	if err != nil {
		t.Fatalf("construct dispatcher: %v", err)
	}

	processed, err := dispatcher.DispatchBatch(t.Context())
	if err != nil || processed != 1 {
		t.Fatalf("DispatchBatch() processed=%d error=%v", processed, err)
	}
	if len(publisher.publications) != 0 {
		t.Fatalf("internal-only event was published: %#v", publisher.publications)
	}
	if len(repository.completed) != 1 || repository.completed[0] != event.ID || len(repository.retried) != 0 {
		t.Fatalf("completed=%#v retried=%#v", repository.completed, repository.retried)
	}
}

func TestStoryMutationEventDispatcherFailsClosedForUnknownDelivery(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 30, 12, 0, 0, 0, time.UTC)
	event := dispatchableStoryMutationEvent(t, now, 1)
	event.Payload = []byte(`{"storyId":"test","_delivery":"unsupported"}`)
	repository := &mutationEventRepositoryStub{events: []storydomain.MutationEvent{event}}
	publisher := &mutationEventPublisherStub{}
	dispatcher, err := newStoryMutationEventDispatcher(
		repository, publisher, testkit.NewFixedClock(now), 5,
	)
	if err != nil {
		t.Fatalf("construct dispatcher: %v", err)
	}

	processed, err := dispatcher.DispatchBatch(t.Context())
	if err != nil || processed != 1 {
		t.Fatalf("DispatchBatch() processed=%d error=%v", processed, err)
	}
	if len(publisher.publications) != 0 || len(repository.completed) != 0 || len(repository.retried) != 1 {
		t.Fatalf("publications=%#v completed=%#v retried=%#v", publisher.publications, repository.completed, repository.retried)
	}
	if !strings.Contains(repository.lastError, "unsupported story mutation event delivery") {
		t.Fatalf("retry error = %q", repository.lastError)
	}
}

func dispatchableStoryMutationEvent(t *testing.T, now time.Time, attempt int) storydomain.MutationEvent {
	t.Helper()
	workspaceID := uuid.New()
	actor, err := platformauth.NewHumanActor(uuid.New()).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind actor: %v", err)
	}
	return storydomain.MutationEvent{
		ID: uuid.New(), WorkspaceID: workspaceID, StoryID: uuid.New(),
		Type: storydomain.MutationEventStoryUpdated, Actor: actor,
		Payload: []byte(`{"storyId":"test"}`), OccurredAt: now.Add(-time.Minute),
		AttemptCount: attempt, ClaimToken: uuid.New(), ClaimedAt: &now, CreatedAt: now.Add(-time.Minute),
	}
}
