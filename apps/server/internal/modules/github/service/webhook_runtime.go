package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
)

const (
	githubWebhookPayloadPrefix = "github-webhook.v1."
)

var githubWebhookProvider = integrations.ProviderKey("github")

type WebhookInstallationRepository interface {
	GetAuthorizedWebhookInstallation(ctx context.Context, externalInstallationID, externalRepositoryID int64) (githubshared.WebhookInstallation, error)
	GetCurrentWebhookInstallation(ctx context.Context, installationID, installationGeneration uuid.UUID, externalRepositoryID int64) (githubshared.WebhookInstallation, error)
}

type WebhookQueue interface {
	EnqueueGitHubWebhook(ctx context.Context, payload tasks.GitHubWebhookPayload) error
}

type WebhookPayloadOpener interface {
	Open(record webhooks.Record, value string) ([]byte, error)
}

type WebhookRuntime struct {
	Registration webhooks.RuntimeRegistration
	Payloads     WebhookPayloadOpener
}

// WebhookWorkerRuntime contains only the capabilities needed after ingress has
// authenticated and durably recorded a GitHub delivery. It intentionally does
// not require the provider signing secret.
type WebhookWorkerRuntime struct {
	Dispatcher webhooks.Dispatcher
	Payloads   WebhookPayloadOpener
}

func NewWebhookRuntime(repo WebhookInstallationRepository, queue WebhookQueue, cfg Config) (WebhookRuntime, error) {
	if repo == nil || queue == nil || strings.TrimSpace(cfg.WebhookSecret) == "" || strings.TrimSpace(cfg.WebhookPayloadSecret) == "" {
		return WebhookRuntime{}, webhooks.ErrNotConfigured
	}
	codec, err := webhooks.NewBoundPayloadCodec(githubWebhookProvider, githubWebhookPayloadPrefix, cfg.WebhookPayloadSecret)
	if err != nil {
		return WebhookRuntime{}, err
	}
	return WebhookRuntime{
		Registration: webhooks.RuntimeRegistration{
			Provider: githubWebhookProvider,
			Verifier: &githubWebhookVerifier{
				installations: repo,
				secret:        cfg.WebhookSecret,
			},
			Protector:  codec,
			Dispatcher: githubWebhookDispatcher{queue: queue},
		},
		Payloads: codec,
	}, nil
}

func NewWebhookWorkerRuntime(queue WebhookQueue, payloadSecret string) (WebhookWorkerRuntime, error) {
	if queue == nil || strings.TrimSpace(payloadSecret) == "" {
		return WebhookWorkerRuntime{}, webhooks.ErrNotConfigured
	}
	codec, err := webhooks.NewBoundPayloadCodec(
		githubWebhookProvider,
		githubWebhookPayloadPrefix,
		payloadSecret,
	)
	if err != nil {
		return WebhookWorkerRuntime{}, err
	}
	return WebhookWorkerRuntime{
		Dispatcher: githubWebhookDispatcher{queue: queue},
		Payloads:   codec,
	}, nil
}

type githubWebhookVerifier struct {
	installations WebhookInstallationRepository
	secret        string
}

func (verifier *githubWebhookVerifier) Verify(ctx context.Context, request webhooks.SignedRequest) (webhooks.VerifiedDelivery, error) {
	if verifier == nil || verifier.installations == nil || strings.TrimSpace(verifier.secret) == "" {
		return webhooks.VerifiedDelivery{}, webhooks.ErrNotConfigured
	}
	if request.Method != "POST" {
		return webhooks.VerifiedDelivery{}, webhooks.ErrInvalidRequest
	}
	signature, ok := exactGitHubHeader(request.Headers, "X-Hub-Signature-256")
	if !ok || !verifyGitHubWebhookSignature(verifier.secret, request.Body, signature) {
		return webhooks.VerifiedDelivery{}, webhooks.ErrUnauthenticated
	}
	deliveryID, deliveryOK := exactGitHubHeader(request.Headers, "X-GitHub-Delivery")
	eventType, eventOK := exactGitHubHeader(request.Headers, "X-GitHub-Event")
	if !deliveryOK || !eventOK || !supportedGitHubWebhookEvent(eventType) {
		return webhooks.VerifiedDelivery{}, webhooks.ErrDeliveryIgnored
	}

	var identity githubWebhookIdentity
	if err := json.Unmarshal(request.Body, &identity); err != nil {
		return webhooks.VerifiedDelivery{}, webhooks.ErrInvalidDelivery
	}
	if identity.Installation.ID <= 0 || identity.Repository.ID <= 0 {
		return webhooks.VerifiedDelivery{}, webhooks.ErrDeliveryIgnored
	}
	installation, err := verifier.installations.GetAuthorizedWebhookInstallation(ctx, identity.Installation.ID, identity.Repository.ID)
	if err != nil {
		if errors.Is(err, githubshared.ErrWebhookInstallationNotFound) {
			return webhooks.VerifiedDelivery{}, webhooks.ErrDeliveryIgnored
		}
		return webhooks.VerifiedDelivery{}, webhooks.ErrVerificationUnavailable
	}
	return webhooks.VerifiedDelivery{
		DeliveryID:             deliveryID,
		EventType:              eventType,
		ExternalAccountID:      strconv.FormatInt(identity.Installation.ID, 10),
		WorkspaceID:            installation.WorkspaceID,
		InstallationID:         installation.ID,
		InstallationGeneration: installation.InstallationGeneration,
	}, nil
}

func exactGitHubHeader(headers webhooks.Headers, name string) (string, bool) {
	values := headers.Values(name)
	if len(values) != 1 {
		return "", false
	}
	value := strings.TrimSpace(values[0])
	return value, value != ""
}

type githubWebhookIdentity struct {
	Installation struct {
		ID int64 `json:"id"`
	} `json:"installation"`
	Repository struct {
		ID int64 `json:"id"`
	} `json:"repository"`
}

func verifyGitHubWebhookSignature(secret string, body []byte, signature string) bool {
	signature = strings.TrimSpace(signature)
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	provided, err := hex.DecodeString(strings.TrimPrefix(signature, "sha256="))
	if err != nil || len(provided) != sha256.Size {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return hmac.Equal(mac.Sum(nil), provided)
}

func supportedGitHubWebhookEvent(eventType string) bool {
	switch strings.TrimSpace(eventType) {
	case "issues", "pull_request", "pull_request_review", "issue_comment", "check_run", "create", "push":
		return true
	default:
		return false
	}
}

type githubWebhookDispatcher struct {
	queue WebhookQueue
}

func (dispatcher githubWebhookDispatcher) Enqueue(ctx context.Context, task webhooks.Task) error {
	if dispatcher.queue == nil || task.Provider != githubWebhookProvider || task.InboxID == uuid.Nil {
		return webhooks.ErrInvalidDelivery
	}
	return dispatcher.queue.EnqueueGitHubWebhook(ctx, tasks.GitHubWebhookPayload{
		InboxID: task.InboxID,
	})
}

var (
	_ webhooks.WebhookVerifier = (*githubWebhookVerifier)(nil)
	_ webhooks.Dispatcher      = githubWebhookDispatcher{}
)
