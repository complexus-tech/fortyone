package outboundwebhooksdomain

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
)

type DeliveryStatus string

const (
	DeliveryPending        DeliveryStatus = "pending"
	DeliveryDelivering     DeliveryStatus = "delivering"
	DeliveryRetryScheduled DeliveryStatus = "retry_scheduled"
	DeliverySucceeded      DeliveryStatus = "succeeded"
	DeliveryFailed         DeliveryStatus = "failed"
	DeliveryCancelled      DeliveryStatus = "cancelled"
)

type ClaimedDelivery struct {
	ID                       uuid.UUID
	WorkspaceID              uuid.UUID
	EventID                  uuid.UUID
	EventType                EventType
	EndpointID               uuid.UUID
	EndpointURL              string
	SigningSecretEnvelope    string
	SecretGeneration         int
	PreviousSecretEnvelope   *string
	PreviousSecretGeneration *int
	PreviousSecretExpiresAt  *time.Time
	SubscriptionGeneration   int
	PayloadBody              []byte
	AttemptNumber            int
	LeaseToken               uuid.UUID
	LeaseExpiresAt           time.Time
	CreatedAt                time.Time
}

type AttemptOutcome string

const (
	AttemptSucceeded      AttemptOutcome = "succeeded"
	AttemptRetryScheduled AttemptOutcome = "retry_scheduled"
	AttemptFailed         AttemptOutcome = "failed"
	AttemptCancelled      AttemptOutcome = "cancelled"
)

type DeliveryAttempt struct {
	ID                   uuid.UUID
	DeliveryID           uuid.UUID
	LeaseToken           uuid.UUID
	AttemptNumber        int
	Outcome              AttemptOutcome
	ResolvedIP           *netip.Addr
	HTTPStatus           *int
	ResponseBytes        *int
	ResponseDigest       []byte
	ErrorCode            string
	Duration             time.Duration
	StartedAt            time.Time
	FinishedAt           time.Time
	NextAttemptAt        *time.Time
	DisableEndpoint      bool
	CountEndpointFailure bool
	DisableAfterFailures int
}
