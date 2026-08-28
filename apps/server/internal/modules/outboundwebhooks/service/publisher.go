package outboundwebhooksservice

import (
	"context"
	"fmt"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	"github.com/google/uuid"
)

type EventRepository interface {
	PublishEvent(context.Context, outboundwebhooksdomain.Event, []byte) ([]uuid.UUID, error)
}

type Publisher struct {
	repository EventRepository
	clock      Clock
}

func NewPublisher(repository EventRepository) (*Publisher, error) {
	return newPublisher(repository, systemClock{})
}

func newPublisher(repository EventRepository, clock Clock) (*Publisher, error) {
	if repository == nil || clock == nil {
		return nil, fmt.Errorf("outbound webhook publisher dependencies are required")
	}
	return &Publisher{repository: repository, clock: clock}, nil
}

// Publish durably records one business event and atomically fans it out to the
// active, currently authorized endpoints subscribed at that moment. Event IDs
// are caller-supplied idempotency identities and must be stable across retries.
func (publisher *Publisher) Publish(ctx context.Context, input outboundwebhooksdomain.PublishEvent) (int, error) {
	if err := input.Validate(); err != nil {
		return 0, err
	}
	now := publisher.clock.Now().UTC()
	if now.Before(input.OccurredAt.UTC()) {
		return 0, outboundwebhooksdomain.ErrInvalidPayload
	}
	event := outboundwebhooksdomain.Event{
		ID: input.ID, WorkspaceID: input.WorkspaceID, Type: input.Type,
		PayloadVersion: outboundwebhooksdomain.PayloadVersion,
		SubjectType:    input.Type.SubjectType(), SubjectID: input.SubjectID,
		Actor: input.Actor, Payload: append([]byte(nil), input.Payload...),
		OccurredAt: input.OccurredAt.UTC(), CreatedAt: now,
	}
	body, err := outboundwebhooksdomain.NewEnvelope(event)
	if err != nil {
		return 0, err
	}
	deliveryIDs, err := publisher.repository.PublishEvent(ctx, event, body)
	if err != nil {
		return 0, fmt.Errorf("publish outbound webhook event: %w", err)
	}
	return len(deliveryIDs), nil
}
