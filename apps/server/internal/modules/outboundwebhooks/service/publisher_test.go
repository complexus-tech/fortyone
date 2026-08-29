package outboundwebhooksservice

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type eventRepositoryStub struct {
	event      outboundwebhooksdomain.Event
	body       []byte
	deliveries []uuid.UUID
	err        error
	calls      int
}

func (repository *eventRepositoryStub) PublishEvent(_ context.Context, event outboundwebhooksdomain.Event, body []byte) ([]uuid.UUID, error) {
	repository.calls++
	repository.event = event
	repository.body = append([]byte(nil), body...)
	return append([]uuid.UUID(nil), repository.deliveries...), repository.err
}

func TestPublisherBuildsVersionedEnvelopeAndDelegatesStableID(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	actor, err := platformauth.NewHumanActor(uuid.New()).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("create actor: %v", err)
	}
	occurredAt := time.Unix(1_700_000_000, 0).UTC()
	eventID := uuid.New()
	repository := &eventRepositoryStub{deliveries: []uuid.UUID{uuid.New(), uuid.New()}}
	publisher, err := newPublisher(repository, &testClock{values: []time.Time{occurredAt.Add(time.Second)}})
	if err != nil {
		t.Fatalf("create publisher: %v", err)
	}
	count, err := publisher.Publish(t.Context(), outboundwebhooksdomain.PublishEvent{
		ID: eventID, WorkspaceID: workspaceID, Type: outboundwebhooksdomain.EventStoryUpdated,
		SubjectID: uuid.New(), Actor: actor, Payload: json.RawMessage(`{"title":"Updated"}`), OccurredAt: occurredAt,
	})
	if err != nil || count != 2 {
		t.Fatalf("Publish() = %d, %v", count, err)
	}
	if repository.event.ID != eventID || repository.event.PayloadVersion != outboundwebhooksdomain.PayloadVersion || repository.calls != 1 {
		t.Fatalf("published event = %+v", repository.event)
	}
	var envelope outboundwebhooksdomain.Envelope
	if err := json.Unmarshal(repository.body, &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.ID != eventID || envelope.Type != outboundwebhooksdomain.EventStoryUpdated || envelope.PayloadVersion != outboundwebhooksdomain.PayloadVersion {
		t.Fatalf("envelope = %+v", envelope)
	}
}

func TestPublisherRejectsInvalidOrFutureEventsBeforeRepository(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	actor, err := platformauth.NewHumanActor(uuid.New()).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("create actor: %v", err)
	}
	now := time.Unix(1_700_000_000, 0).UTC()
	for _, input := range []outboundwebhooksdomain.PublishEvent{
		{ID: uuid.New(), WorkspaceID: workspaceID, Type: outboundwebhooksdomain.EventStoryCreated, SubjectID: uuid.New(), Actor: actor, Payload: json.RawMessage(`[]`), OccurredAt: now},
		{ID: uuid.New(), WorkspaceID: workspaceID, Type: outboundwebhooksdomain.EventStoryCreated, SubjectID: uuid.New(), Actor: actor, Payload: json.RawMessage(`{}`), OccurredAt: now.Add(time.Second)},
	} {
		repository := &eventRepositoryStub{err: errors.New("must not be called")}
		publisher, err := newPublisher(repository, &testClock{values: []time.Time{now}})
		if err != nil {
			t.Fatalf("create publisher: %v", err)
		}
		if _, err := publisher.Publish(t.Context(), input); err == nil || repository.calls != 0 {
			t.Fatalf("invalid Publish() error=%v calls=%d", err, repository.calls)
		}
	}
}
