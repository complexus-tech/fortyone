package providers

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type outboundStoryEventPublisherStub struct {
	events []outboundwebhooksdomain.PublishEvent
	err    error
}

func (publisher *outboundStoryEventPublisherStub) Publish(
	_ context.Context,
	event outboundwebhooksdomain.PublishEvent,
) (int, error) {
	publisher.events = append(publisher.events, event)
	return 1, publisher.err
}

func TestOutboundStoryMutationPublisherPreservesCommittedEventIdentity(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	eventID := uuid.New()
	storyID := uuid.New()
	actorID := uuid.New()
	occurredAt := time.Date(2026, time.August, 28, 8, 30, 0, 0, time.UTC)
	payload := []byte(`{"story_id":"` + storyID.String() + `"}`)
	actor, err := platformauth.NewHumanActor(actorID).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("scope actor to workspace: %v", err)
	}
	downstream := &outboundStoryEventPublisherStub{}
	adapter, err := NewOutboundStoryMutationPublisher(downstream)
	if err != nil {
		t.Fatalf("build adapter: %v", err)
	}

	err = adapter.PublishStoryMutationEvent(context.Background(), stories.StoryMutationPublication{
		ID: eventID, WorkspaceID: workspaceID, StoryID: storyID,
		Type: "story.updated", Actor: actor, Payload: payload, OccurredAt: occurredAt,
	})
	if err != nil {
		t.Fatalf("publish story mutation event: %v", err)
	}
	if len(downstream.events) != 1 {
		t.Fatalf("expected one downstream event, got %d", len(downstream.events))
	}
	event := downstream.events[0]
	if event.ID != eventID || event.WorkspaceID != workspaceID || event.SubjectID != storyID {
		t.Fatalf("committed identity was not preserved: %#v", event)
	}
	if event.Type != outboundwebhooksdomain.EventStoryUpdated || !reflect.DeepEqual(event.Actor, actor) || !event.OccurredAt.Equal(occurredAt) {
		t.Fatalf("committed metadata was not preserved: %#v", event)
	}
	if string(event.Payload) != string(payload) {
		t.Fatalf("payload changed: got %s want %s", event.Payload, payload)
	}

	payload[0] = '['
	if string(event.Payload) == string(payload) {
		t.Fatal("adapter retained the caller payload backing array")
	}
}

func TestOutboundStoryMutationPublisherRejectsUnknownEventType(t *testing.T) {
	t.Parallel()

	downstream := &outboundStoryEventPublisherStub{}
	adapter, err := NewOutboundStoryMutationPublisher(downstream)
	if err != nil {
		t.Fatalf("build adapter: %v", err)
	}

	err = adapter.PublishStoryMutationEvent(context.Background(), stories.StoryMutationPublication{
		Type: "story.archived",
	})
	if err == nil {
		t.Fatal("expected an unsupported event type error")
	}
	if len(downstream.events) != 0 {
		t.Fatalf("unsupported event reached downstream publisher: %#v", downstream.events)
	}
}

func TestOutboundStoryMutationPublisherWrapsDownstreamFailure(t *testing.T) {
	t.Parallel()

	downstreamErr := errors.New("database unavailable")
	downstream := &outboundStoryEventPublisherStub{err: downstreamErr}
	adapter, err := NewOutboundStoryMutationPublisher(downstream)
	if err != nil {
		t.Fatalf("build adapter: %v", err)
	}

	err = adapter.PublishStoryMutationEvent(context.Background(), stories.StoryMutationPublication{
		Type: "story.deleted",
	})
	if !errors.Is(err, downstreamErr) {
		t.Fatalf("expected wrapped downstream error, got %v", err)
	}
}
