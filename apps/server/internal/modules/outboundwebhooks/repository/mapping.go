package outboundwebhooksrepository

import (
	"fmt"
	"time"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	outboundwebhookssql "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/repository/sqlc"
	"github.com/google/uuid"
)

type endpointRecord struct {
	ID                     uuid.UUID
	WorkspaceID            uuid.UUID
	OwnerPrincipalID       uuid.UUID
	Name                   string
	URL                    string
	Status                 string
	SecretGeneration       int32
	SubscriptionGeneration int32
	ConsecutiveFailures    int32
	LastSuccessAt          *time.Time
	DisabledAt             *time.Time
	DisabledReason         *string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

func endpointFromCreate(row outboundwebhookssql.CreateOutboundWebhookEndpointRow, subscriptions []outboundwebhooksdomain.EventType) (outboundwebhooksdomain.Endpoint, error) {
	return mapEndpoint(endpointRecord{
		ID: row.EndpointID, WorkspaceID: row.WorkspaceID, OwnerPrincipalID: row.OwnerPrincipalID,
		Name: row.Name, URL: row.EndpointURL, Status: row.Status,
		SecretGeneration: row.SecretGeneration, SubscriptionGeneration: row.SubscriptionGeneration,
		ConsecutiveFailures: row.ConsecutiveFailures, LastSuccessAt: row.LastSuccessAt,
		DisabledAt: row.DisabledAt, DisabledReason: row.DisabledReason,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, subscriptions)
}

func endpointFromGet(row outboundwebhookssql.GetOutboundWebhookEndpointRow, subscriptions []outboundwebhooksdomain.EventType) (outboundwebhooksdomain.Endpoint, error) {
	return mapEndpoint(endpointRecord{
		ID: row.EndpointID, WorkspaceID: row.WorkspaceID, OwnerPrincipalID: row.OwnerPrincipalID,
		Name: row.Name, URL: row.EndpointURL, Status: row.Status,
		SecretGeneration: row.SecretGeneration, SubscriptionGeneration: row.SubscriptionGeneration,
		ConsecutiveFailures: row.ConsecutiveFailures, LastSuccessAt: row.LastSuccessAt,
		DisabledAt: row.DisabledAt, DisabledReason: row.DisabledReason,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, subscriptions)
}

func endpointFromList(row outboundwebhookssql.ListOutboundWebhookEndpointsRow, subscriptions []outboundwebhooksdomain.EventType) (outboundwebhooksdomain.Endpoint, error) {
	return mapEndpoint(endpointRecord{
		ID: row.EndpointID, WorkspaceID: row.WorkspaceID, OwnerPrincipalID: row.OwnerPrincipalID,
		Name: row.Name, URL: row.EndpointURL, Status: row.Status,
		SecretGeneration: row.SecretGeneration, SubscriptionGeneration: row.SubscriptionGeneration,
		ConsecutiveFailures: row.ConsecutiveFailures, LastSuccessAt: row.LastSuccessAt,
		DisabledAt: row.DisabledAt, DisabledReason: row.DisabledReason,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, subscriptions)
}

func mapEndpoint(record endpointRecord, subscriptions []outboundwebhooksdomain.EventType) (outboundwebhooksdomain.Endpoint, error) {
	status := outboundwebhooksdomain.EndpointStatus(record.Status)
	if status != outboundwebhooksdomain.EndpointActive && status != outboundwebhooksdomain.EndpointDisabled {
		return outboundwebhooksdomain.Endpoint{}, fmt.Errorf("map outbound webhook endpoint status %q", record.Status)
	}
	if record.SecretGeneration <= 0 || record.SubscriptionGeneration <= 0 || record.ConsecutiveFailures < 0 {
		return outboundwebhooksdomain.Endpoint{}, fmt.Errorf("map outbound webhook endpoint counters")
	}
	return outboundwebhooksdomain.Endpoint{
		ID: record.ID, WorkspaceID: record.WorkspaceID, OwnerPrincipalID: record.OwnerPrincipalID,
		Name: record.Name, URL: record.URL, Status: status,
		SecretGeneration: int(record.SecretGeneration), SubscriptionGeneration: int(record.SubscriptionGeneration),
		Subscriptions:       append([]outboundwebhooksdomain.EventType(nil), subscriptions...),
		ConsecutiveFailures: int(record.ConsecutiveFailures), LastSuccessAt: record.LastSuccessAt,
		DisabledAt: record.DisabledAt, DisabledReason: record.DisabledReason,
		CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}, nil
}

func mapSubscriptions(values []string) ([]outboundwebhooksdomain.EventType, error) {
	result := make([]outboundwebhooksdomain.EventType, 0, len(values))
	for _, value := range values {
		eventType := outboundwebhooksdomain.EventType(value)
		if err := eventType.Validate(); err != nil {
			return nil, fmt.Errorf("map outbound webhook subscription: %w", err)
		}
		result = append(result, eventType)
	}
	return result, nil
}

func mapClaimedDelivery(row outboundwebhookssql.ClaimNextOutboundWebhookDeliveryRow) (outboundwebhooksdomain.ClaimedDelivery, error) {
	if row.LeaseToken == nil || row.LeaseExpiresAt == nil || row.AttemptCount <= 0 {
		return outboundwebhooksdomain.ClaimedDelivery{}, fmt.Errorf("map outbound webhook delivery lease")
	}
	previousFields := 0
	if row.PreviousSecretEnvelope != nil {
		previousFields++
	}
	if row.PreviousSecretGeneration != nil {
		previousFields++
	}
	if row.PreviousSecretExpiresAt != nil {
		previousFields++
	}
	if previousFields != 0 && previousFields != 3 {
		return outboundwebhooksdomain.ClaimedDelivery{}, fmt.Errorf("map outbound webhook previous signing secret")
	}
	if row.SecretGeneration <= 0 || (row.PreviousSecretGeneration != nil &&
		(*row.PreviousSecretGeneration <= 0 || *row.PreviousSecretGeneration >= row.SecretGeneration)) {
		return outboundwebhooksdomain.ClaimedDelivery{}, fmt.Errorf("map outbound webhook signing secret generations")
	}
	eventType := outboundwebhooksdomain.EventType(row.EventType)
	if err := eventType.Validate(); err != nil {
		return outboundwebhooksdomain.ClaimedDelivery{}, err
	}
	return outboundwebhooksdomain.ClaimedDelivery{
		ID: row.DeliveryID, WorkspaceID: row.WorkspaceID, EventID: row.EventID,
		EventType: eventType, EndpointID: row.EndpointID, EndpointURL: row.EndpointURL,
		SigningSecretEnvelope: row.SigningSecretEnvelope, SecretGeneration: int(row.SecretGeneration),
		PreviousSecretEnvelope:   row.PreviousSecretEnvelope,
		PreviousSecretGeneration: optionalInt(row.PreviousSecretGeneration),
		PreviousSecretExpiresAt:  row.PreviousSecretExpiresAt,
		SubscriptionGeneration:   int(row.SubscriptionGeneration), PayloadBody: append([]byte(nil), row.PayloadBody...),
		AttemptNumber: int(row.AttemptCount), LeaseToken: *row.LeaseToken, LeaseExpiresAt: *row.LeaseExpiresAt,
		CreatedAt: row.CreatedAt,
	}, nil
}

func optionalInt(value *int32) *int {
	if value == nil {
		return nil
	}
	converted := int(*value)
	return &converted
}
