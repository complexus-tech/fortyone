package outboundwebhooksdomain

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

type EndpointStatus string

const (
	EndpointActive   EndpointStatus = "active"
	EndpointDisabled EndpointStatus = "disabled"
)

type Endpoint struct {
	ID                     uuid.UUID
	WorkspaceID            uuid.UUID
	OwnerPrincipalID       uuid.UUID
	Name                   string
	URL                    string
	Status                 EndpointStatus
	SecretGeneration       int
	SubscriptionGeneration int
	Subscriptions          []EventType
	ConsecutiveFailures    int
	LastSuccessAt          *time.Time
	DisabledAt             *time.Time
	DisabledReason         *string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type SigningSecret struct {
	value string
}

func NewSigningSecret(value string) SigningSecret {
	return SigningSecret{value: value}
}

func (secret SigningSecret) Reveal() string {
	return secret.value
}

func (SigningSecret) String() string {
	return "[REDACTED]"
}

func (SigningSecret) GoString() string {
	return "outboundwebhooksdomain.SigningSecret{[REDACTED]}"
}

func (SigningSecret) MarshalJSON() ([]byte, error) {
	return []byte(`"[REDACTED]"`), nil
}

func (SigningSecret) LogValue() slog.Value {
	return slog.StringValue("[REDACTED]")
}

type CreatedEndpoint struct {
	Endpoint Endpoint
	Secret   SigningSecret
}

type EndpointCursor struct {
	CreatedAt time.Time `json:"createdAt"`
	ID        uuid.UUID `json:"id"`
}

func (cursor EndpointCursor) Validate() error {
	if cursor.CreatedAt.IsZero() || cursor.ID == uuid.Nil {
		return ErrInvalidEndpoint
	}
	return nil
}

type EndpointPage struct {
	Items      []Endpoint
	NextCursor *EndpointCursor
}

type CreateEndpoint struct {
	ID               uuid.UUID
	AuditID          uuid.UUID
	WorkspaceID      uuid.UUID
	OwnerPrincipalID uuid.UUID
	Actor            platformauth.Actor
	WorkspaceRole    authorization.WorkspaceRole
	Name             string
	URL              string
	Subscriptions    []EventType
	CreatedAt        time.Time
	RequestID        string
}

func (input CreateEndpoint) Validate() error {
	if input.ID == uuid.Nil || input.AuditID == uuid.Nil || input.WorkspaceID == uuid.Nil || input.OwnerPrincipalID == uuid.Nil || input.CreatedAt.IsZero() {
		return ErrInvalidEndpoint
	}
	name := strings.TrimSpace(input.Name)
	if name == "" || len(name) > 120 || name != input.Name {
		return fmt.Errorf("%w: name", ErrInvalidEndpoint)
	}
	if strings.TrimSpace(input.URL) != input.URL || len(input.URL) > 2048 {
		return fmt.Errorf("%w: URL", ErrInvalidEndpoint)
	}
	if err := ValidateSubscriptions(input.Subscriptions); err != nil {
		return err
	}
	return nil
}

func ValidateSubscriptions(subscriptions []EventType) error {
	if len(subscriptions) == 0 || len(subscriptions) > len(eventCatalog) {
		return ErrInvalidSubscription
	}
	seen := make(map[EventType]struct{}, len(subscriptions))
	for _, eventType := range subscriptions {
		if err := eventType.Validate(); err != nil {
			return err
		}
		if _, duplicate := seen[eventType]; duplicate {
			return fmt.Errorf("%w: duplicate %q", ErrInvalidSubscription, eventType)
		}
		seen[eventType] = struct{}{}
	}
	return nil
}
