package feedback

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CorePublicationOutboxEvent is one claimed Update publication. Payload is
// intentionally retained as raw JSON so a malformed event can be failed in
// isolation without preventing other claimed events from being dispatched.
type CorePublicationOutboxEvent struct {
	EventID             uuid.UUID
	UpdateID            uuid.UUID
	WorkspaceID         uuid.UUID
	PortalID            uuid.UUID
	ActorID             uuid.UUID
	PublicationSequence int64
	PublishedAt         time.Time
	ClaimToken          uuid.UUID
	AttemptCount        int
	Payload             json.RawMessage
}

type NotificationActorResolver interface {
	ResolveNotificationActor(context.Context, uuid.UUID, uuid.UUID) (uuid.UUID, error)
}

// PublicationOutboxRepository is separate from NextPhaseRepository so the
// durable dispatch contract can evolve without widening every HTTP-facing
// repository test double.
type PublicationOutboxRepository interface {
	NotificationActorResolver
	ClaimPublicationOutboxEvents(context.Context, int, time.Duration) ([]CorePublicationOutboxEvent, error)
	CompletePublicationOutboxEvent(context.Context, uuid.UUID, uuid.UUID) error
	RetryPublicationOutboxEvent(context.Context, uuid.UUID, uuid.UUID, string, time.Time, bool) error
	ListPublicationDeliveryRecipients(context.Context, uuid.UUID, []uuid.UUID, []uuid.UUID) ([]CoreDeliveryRecipient, error)
	ListAccountPublicationRecipients(context.Context, uuid.UUID, []CoreAccountUpdateRecipient) ([]CoreAccountUpdateRecipient, error)
	CreateContributorDelivery(context.Context, CoreCreateDeliveryInput) (CoreDelivery, bool, error)
}
