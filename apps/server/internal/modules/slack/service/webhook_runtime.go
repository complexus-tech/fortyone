package slack

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
)

var slackWebhookProvider = integrations.ProviderKey(slackProviderMessaging)

type WebhookInstallationRepository interface {
	GetSlackWorkspaceByTeamID(ctx context.Context, slackTeamID string) (slackWorkspaceRecord, error)
}

// NewWebhookRuntime builds Slack's provider-owned adapters for the shared
// durable webhook gateway. The gateway remains responsible for persistence,
// deduplication, dispatch ordering, leases, and recovery.
func NewWebhookRuntime(repo WebhookInstallationRepository, queue EventQueue, cfg Config) (webhooks.RuntimeRegistration, error) {
	if repo == nil || queue == nil {
		return webhooks.RuntimeRegistration{}, webhooks.ErrNotConfigured
	}
	codec, err := newSlackWebhookPayloadCodec(cfg.WebhookPayloadSecret)
	if err != nil {
		return webhooks.RuntimeRegistration{}, err
	}
	if strings.TrimSpace(cfg.SigningSecret) == "" {
		return webhooks.RuntimeRegistration{}, ErrSlackSigningSecretNotConfigured
	}
	return webhooks.RuntimeRegistration{
		Provider: slackWebhookProvider,
		Verifier: &slackWebhookVerifier{
			installations: repo,
			signingSecret: cfg.SigningSecret,
		},
		Protector:  codec,
		Dispatcher: slackWebhookDispatcher{queue: queue},
	}, nil
}

type slackWebhookVerifier struct {
	installations WebhookInstallationRepository
	signingSecret string
}

func (verifier *slackWebhookVerifier) Verify(
	ctx context.Context,
	request webhooks.SignedRequest,
) (webhooks.VerifiedDelivery, error) {
	if verifier == nil || verifier.installations == nil {
		return webhooks.VerifiedDelivery{}, webhooks.ErrNotConfigured
	}
	now := request.ReceivedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if err := verifySlackSignature(
		verifier.signingSecret,
		now,
		request.Body,
		request.Headers.First("X-Slack-Request-Timestamp"),
		request.Headers.First("X-Slack-Signature"),
	); err != nil {
		switch {
		case errors.Is(err, ErrSlackRequestExpired):
			return webhooks.VerifiedDelivery{}, webhooks.ErrReplay
		case errors.Is(err, ErrSlackInvalidSignature), errors.Is(err, ErrSlackSigningSecretNotConfigured):
			return webhooks.VerifiedDelivery{}, webhooks.ErrUnauthenticated
		default:
			return webhooks.VerifiedDelivery{}, err
		}
	}

	payload, err := decodeSlackEvent(request.Body)
	if err != nil {
		return webhooks.VerifiedDelivery{}, ErrSlackInvalidEventPayload
	}
	if payload.Type != slackEventCallback || strings.TrimSpace(payload.EventID) == "" || strings.TrimSpace(payload.TeamID) == "" {
		return webhooks.VerifiedDelivery{}, webhooks.ErrDeliveryIgnored
	}
	event, supported := normalizeSlackEvent(payload)
	if !supported || event.Kind == slackEventKindEntityDetails {
		return webhooks.VerifiedDelivery{}, webhooks.ErrDeliveryIgnored
	}
	installation, err := verifier.installations.GetSlackWorkspaceByTeamID(ctx, payload.TeamID)
	if err != nil {
		if isSlackRepositoryNotFound(err) {
			return webhooks.VerifiedDelivery{}, webhooks.ErrDeliveryIgnored
		}
		return webhooks.VerifiedDelivery{}, webhooks.ErrVerificationUnavailable
	}
	if !installation.IsActive || installation.ID == uuid.Nil || installation.WorkspaceID == uuid.Nil || installation.InstallGeneration == uuid.Nil {
		return webhooks.VerifiedDelivery{}, webhooks.ErrDeliveryIgnored
	}
	return webhooks.VerifiedDelivery{
		DeliveryID:             strings.TrimSpace(payload.EventID),
		EventType:              strings.TrimSpace(payload.Event.Type),
		ExternalAccountID:      strings.TrimSpace(payload.TeamID),
		WorkspaceID:            installation.WorkspaceID,
		InstallationID:         installation.ID,
		InstallationGeneration: installation.InstallGeneration,
	}, nil
}

type slackWebhookDispatcher struct {
	queue EventQueue
}

func (dispatcher slackWebhookDispatcher) Enqueue(ctx context.Context, task webhooks.Task) error {
	if dispatcher.queue == nil || task.Provider != slackWebhookProvider || task.InboxID == uuid.Nil {
		return webhooks.ErrInvalidDelivery
	}
	return dispatcher.queue.EnqueueSlackEvent(ctx, tasks.SlackEventPayload{
		Provider: string(task.Provider),
		InboxID:  task.InboxID,
	})
}

var (
	_ webhooks.WebhookVerifier  = (*slackWebhookVerifier)(nil)
	_ webhooks.PayloadProtector = (*webhooks.BoundPayloadCodec)(nil)
	_ webhooks.Dispatcher       = slackWebhookDispatcher{}
)
