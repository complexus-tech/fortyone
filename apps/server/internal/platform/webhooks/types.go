package webhooks

import (
	"context"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/google/uuid"
)

const CurrentEnvelopeVersion int16 = 1

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusIgnored    Status = "ignored"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

func (status Status) Terminal() bool {
	switch status {
	case StatusCompleted, StatusIgnored, StatusCancelled:
		return true
	default:
		return false
	}
}

func ParseStatus(value string) (Status, bool) {
	status := Status(strings.TrimSpace(value))
	switch status {
	case StatusPending, StatusProcessing, StatusCompleted, StatusIgnored, StatusFailed, StatusCancelled:
		return status, true
	default:
		return "", false
	}
}

// Headers is a provider-neutral copy of the HTTP fields needed for request
// verification. Values preserves repeated header fields.
type Headers map[string][]string

func (headers Headers) Values(name string) []string {
	for key, values := range headers {
		if strings.EqualFold(key, name) {
			return append([]string(nil), values...)
		}
	}
	return nil
}

func (headers Headers) First(name string) string {
	values := headers.Values(name)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

type SignedRequest struct {
	Method        string
	RequestTarget string
	Headers       Headers
	Body          []byte
	ReceivedAt    time.Time
}

// VerifiedDelivery is returned by one provider adapter after signature,
// replay, event identity, and installation checks have passed.
type VerifiedDelivery struct {
	DeliveryID             string
	EventType              string
	ExternalAccountID      string
	WorkspaceID            uuid.UUID
	InstallationID         uuid.UUID
	InstallationGeneration uuid.UUID
	TraceID                string
}

type Envelope struct {
	Version                int16
	Provider               integrations.ProviderKey
	DeliveryID             string
	EventType              string
	ExternalAccountID      string
	WorkspaceID            uuid.UUID
	InstallationID         uuid.UUID
	InstallationGeneration uuid.UUID
	TraceID                string
	ReceivedAt             time.Time
}

type PayloadBinding struct {
	Provider               integrations.ProviderKey
	DeliveryID             string
	WorkspaceID            uuid.UUID
	InstallationID         uuid.UUID
	InstallationGeneration uuid.UUID
}

type Record struct {
	Envelope
	ID                 uuid.UUID
	Status             Status
	AttemptCount       int32
	RecoveryGeneration int32
	RecoveryEnqueuedAt *time.Time
	ProcessedAt        *time.Time
	UpdatedAt          time.Time
	EncryptedPayload   *string
	PayloadExpiresAt   *time.Time
}

type Receipt struct {
	ID      uuid.UUID
	Status  Status
	Created bool
	Queued  bool
	Ignored bool
}

type Task struct {
	InboxID  uuid.UUID
	Provider integrations.ProviderKey
}

type WebhookVerifier interface {
	Verify(ctx context.Context, request SignedRequest) (VerifiedDelivery, error)
}

type PayloadProtector interface {
	Seal(ctx context.Context, binding PayloadBinding, payload []byte) (string, error)
}

type Dispatcher interface {
	Enqueue(ctx context.Context, task Task) error
}

type Inbox interface {
	Register(ctx context.Context, envelope Envelope, encryptedPayload string, expiresAt time.Time) (Record, bool, error)
	MarkQueued(ctx context.Context, id uuid.UUID, queuedAt time.Time) error
	ClaimRecoverable(ctx context.Context, provider integrations.ProviderKey, policy RecoveryPolicy, now time.Time) ([]Record, error)
	ReleaseRecovery(ctx context.Context, id uuid.UUID, generation int32, releasedAt time.Time) error
	GetByID(ctx context.Context, id uuid.UUID) (Record, error)
	GetByExternalKey(ctx context.Context, provider integrations.ProviderKey, externalAccountID, deliveryID string) (Record, error)
	Start(ctx context.Context, id uuid.UUID, now time.Time, lease time.Duration) (Record, bool, error)
	Complete(ctx context.Context, id uuid.UUID, status Status, outcomeCode string, completedAt time.Time) error
	ExpirePayloads(ctx context.Context, now time.Time, limit int32) ([]uuid.UUID, error)
}
