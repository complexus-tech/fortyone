package figma

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	figmaprovider "github.com/complexus-tech/projects-api/internal/modules/figma"
	figmadomain "github.com/complexus-tech/projects-api/internal/modules/figma/domain"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
)

const figmaWebhookPayloadPrefix = "figma-webhook.v1."

type WebhookRuntime struct {
	Registration webhooks.RuntimeRegistration
	Payloads     WebhookPayloadOpener
}

type webhookInstallationStore interface {
	GetWebhook(ctx context.Context, figmaWebhookID int64) (Webhook, error)
}

func NewWebhookRuntime(
	repository webhookInstallationStore,
	queue WebhookQueue,
	config Config,
) (WebhookRuntime, error) {
	if repository == nil || queue == nil || strings.TrimSpace(config.WebhookPayloadSecret) == "" {
		return WebhookRuntime{}, webhooks.ErrNotConfigured
	}
	codec, err := webhooks.NewBoundPayloadCodec(
		figmaprovider.ProviderKey,
		figmaWebhookPayloadPrefix,
		config.WebhookPayloadSecret,
	)
	if err != nil {
		return WebhookRuntime{}, err
	}
	return WebhookRuntime{
		Registration: webhooks.RuntimeRegistration{
			Provider:   figmaprovider.ProviderKey,
			Verifier:   &figmaWebhookVerifier{installations: repository},
			Protector:  codec,
			Dispatcher: figmaWebhookDispatcher{queue: queue},
		},
		Payloads: codec,
	}, nil
}

type figmaWebhookVerifier struct {
	installations webhookInstallationStore
}

func (verifier *figmaWebhookVerifier) Verify(
	ctx context.Context,
	request webhooks.SignedRequest,
) (webhooks.VerifiedDelivery, error) {
	if verifier == nil || verifier.installations == nil {
		return webhooks.VerifiedDelivery{}, webhooks.ErrNotConfigured
	}
	var event WebhookEvent
	if err := json.Unmarshal(request.Body, &event); err != nil {
		return webhooks.VerifiedDelivery{}, webhooks.ErrInvalidDelivery
	}
	webhookID := int64(event.WebhookID)
	if webhookID <= 0 {
		return webhooks.VerifiedDelivery{}, webhooks.ErrInvalidDelivery
	}
	installation, err := verifier.installations.GetWebhook(ctx, webhookID)
	if err != nil {
		if errors.Is(err, figmadomain.ErrNotFound) {
			return webhooks.VerifiedDelivery{}, webhooks.ErrDeliveryIgnored
		}
		return webhooks.VerifiedDelivery{}, webhooks.ErrVerificationUnavailable
	}
	if installation.ConnectionID == uuid.Nil || installation.WorkspaceID == uuid.Nil ||
		installation.InstallationGeneration == uuid.Nil || !installation.IsActive {
		return webhooks.VerifiedDelivery{}, webhooks.ErrDeliveryIgnored
	}
	if subtle.ConstantTimeCompare(
		[]byte(digest(event.Passcode)),
		[]byte(installation.PasscodeHash),
	) != 1 {
		return webhooks.VerifiedDelivery{}, webhooks.ErrUnauthenticated
	}
	event.EventType = strings.TrimSpace(event.EventType)
	if event.EventType != "PING" && event.EventType != installation.EventType {
		return webhooks.VerifiedDelivery{}, webhooks.ErrInvalidDelivery
	}
	if event.EventType != "PING" && strings.TrimSpace(event.FileKey) != installation.FileKey {
		return webhooks.VerifiedDelivery{}, webhooks.ErrInvalidDelivery
	}
	if event.EventType == "" {
		return webhooks.VerifiedDelivery{}, webhooks.ErrInvalidDelivery
	}
	deliveryDigest := sha256.Sum256(request.Body)
	return webhooks.VerifiedDelivery{
		DeliveryID:             hex.EncodeToString(deliveryDigest[:]),
		EventType:              event.EventType,
		ExternalAccountID:      strconv.FormatInt(webhookID, 10),
		WorkspaceID:            installation.WorkspaceID,
		InstallationID:         installation.ConnectionID,
		InstallationGeneration: installation.InstallationGeneration,
	}, nil
}

type figmaWebhookDispatcher struct {
	queue WebhookQueue
}

func (dispatcher figmaWebhookDispatcher) Enqueue(ctx context.Context, task webhooks.Task) error {
	if dispatcher.queue == nil || task.Provider != figmaprovider.ProviderKey || task.InboxID == uuid.Nil {
		return webhooks.ErrInvalidDelivery
	}
	return dispatcher.queue.EnqueueFigmaWebhook(ctx, tasks.FigmaWebhookPayload{InboxID: task.InboxID})
}

var (
	_ webhooks.WebhookVerifier = (*figmaWebhookVerifier)(nil)
	_ webhooks.Dispatcher      = figmaWebhookDispatcher{}
)
