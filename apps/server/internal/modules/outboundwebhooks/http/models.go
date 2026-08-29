package outboundwebhookshttp

import (
	"time"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	"github.com/google/uuid"
)

type createEndpointRequest struct {
	Name          string                             `json:"name"`
	URL           string                             `json:"url"`
	Subscriptions []outboundwebhooksdomain.EventType `json:"subscriptions"`
}

type replaceSubscriptionsRequest struct {
	Subscriptions []outboundwebhooksdomain.EventType `json:"subscriptions"`
}

type disableEndpointRequest struct {
	Reason string `json:"reason"`
}

type endpointResponse struct {
	ID                     uuid.UUID                             `json:"id"`
	WorkspaceID            uuid.UUID                             `json:"workspaceId"`
	Name                   string                                `json:"name"`
	URL                    string                                `json:"url"`
	Status                 outboundwebhooksdomain.EndpointStatus `json:"status"`
	SecretGeneration       int                                   `json:"secretGeneration"`
	SubscriptionGeneration int                                   `json:"subscriptionGeneration"`
	Subscriptions          []outboundwebhooksdomain.EventType    `json:"subscriptions"`
	ConsecutiveFailures    int                                   `json:"consecutiveFailures"`
	LastSuccessAt          *time.Time                            `json:"lastSuccessAt,omitempty"`
	DisabledAt             *time.Time                            `json:"disabledAt,omitempty"`
	DisabledReason         *string                               `json:"disabledReason,omitempty"`
	CreatedAt              time.Time                             `json:"createdAt"`
	UpdatedAt              time.Time                             `json:"updatedAt"`
}

type createdEndpointResponse struct {
	Endpoint      endpointResponse `json:"endpoint"`
	SigningSecret string           `json:"signingSecret"`
}

type endpointPageResponse struct {
	Items      []endpointResponse `json:"items"`
	NextCursor *string            `json:"nextCursor,omitempty"`
}

type endpointCursor struct {
	Version     int       `json:"version"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	PrincipalID uuid.UUID `json:"principalId"`
	CreatedAt   time.Time `json:"createdAt"`
	EndpointID  uuid.UUID `json:"endpointId"`
	Limit       int       `json:"limit"`
}

type rotatedSecretResponse struct {
	SigningSecret           string    `json:"signingSecret"`
	Generation              int       `json:"generation"`
	PreviousSecretExpiresAt time.Time `json:"previousSecretExpiresAt"`
}

func endpointModel(value outboundwebhooksdomain.Endpoint) endpointResponse {
	subscriptions := append([]outboundwebhooksdomain.EventType(nil), value.Subscriptions...)
	if subscriptions == nil {
		subscriptions = []outboundwebhooksdomain.EventType{}
	}
	return endpointResponse{
		ID: value.ID, WorkspaceID: value.WorkspaceID, Name: value.Name, URL: value.URL,
		Status: value.Status, SecretGeneration: value.SecretGeneration,
		SubscriptionGeneration: value.SubscriptionGeneration, Subscriptions: subscriptions,
		ConsecutiveFailures: value.ConsecutiveFailures, LastSuccessAt: value.LastSuccessAt,
		DisabledAt: value.DisabledAt, DisabledReason: value.DisabledReason,
		CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt,
	}
}

func endpointModels(values []outboundwebhooksdomain.Endpoint) []endpointResponse {
	models := make([]endpointResponse, len(values))
	for index, value := range values {
		models[index] = endpointModel(value)
	}
	return models
}
