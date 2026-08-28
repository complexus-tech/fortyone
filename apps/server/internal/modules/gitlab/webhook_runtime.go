package gitlab

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/codehost"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/google/uuid"
)

const (
	gitLabWebhookPayloadPrefix = "gitlab-webhook.v1."
	defaultReplayWindow        = 5 * time.Minute
	maximumReplayWindow        = 15 * time.Minute
	maximumFutureSkew          = time.Minute
)

type WebhookInstallationResolver interface {
	ResolveGitLabWebhookInstallation(
		ctx context.Context,
		instanceURL, externalRepositoryID string,
	) (codehost.InstallationRef, error)
}

type WebhookTask struct {
	InboxID uuid.UUID
}

type WebhookQueue interface {
	EnqueueGitLabWebhook(ctx context.Context, task WebhookTask) error
}

type WebhookRuntime struct {
	Registration webhooks.RuntimeRegistration
	Payloads     *webhooks.BoundPayloadCodec
}

func NewWebhookRuntime(
	resolver WebhookInstallationResolver,
	queue WebhookQueue,
	config Config,
) (WebhookRuntime, error) {
	if resolver == nil || queue == nil || strings.TrimSpace(config.WebhookPayloadSecret) == "" {
		return WebhookRuntime{}, webhooks.ErrNotConfigured
	}
	signingKey, err := parseSigningToken(config.WebhookSigningToken)
	if err != nil {
		return WebhookRuntime{}, err
	}
	instanceURL, err := normalizeInstanceURL(config.BaseURL)
	if err != nil {
		return WebhookRuntime{}, fmt.Errorf("configure GitLab webhook instance: %w", webhooks.ErrNotConfigured)
	}
	replayWindow := config.WebhookReplayWindow
	if replayWindow == 0 {
		replayWindow = defaultReplayWindow
	}
	if replayWindow < time.Minute || replayWindow > maximumReplayWindow {
		return WebhookRuntime{}, fmt.Errorf("configure GitLab webhook replay window: %w", webhooks.ErrNotConfigured)
	}
	now := config.Now
	if now == nil {
		now = time.Now
	}
	codec, err := webhooks.NewBoundPayloadCodec(ProviderKey, gitLabWebhookPayloadPrefix, config.WebhookPayloadSecret)
	if err != nil {
		return WebhookRuntime{}, err
	}
	return WebhookRuntime{
		Registration: webhooks.RuntimeRegistration{
			Provider: ProviderKey,
			Verifier: &gitLabWebhookVerifier{
				resolver:     resolver,
				signingKey:   signingKey,
				instanceURL:  instanceURL,
				replayWindow: replayWindow,
				now:          now,
			},
			Protector:  codec,
			Dispatcher: gitLabWebhookDispatcher{queue: queue},
		},
		Payloads: codec,
	}, nil
}

type gitLabWebhookVerifier struct {
	resolver     WebhookInstallationResolver
	signingKey   []byte
	instanceURL  string
	replayWindow time.Duration
	now          func() time.Time
}

type webhookIdentity struct {
	Project struct {
		ID int64 `json:"id"`
	} `json:"project"`
}

func (verifier *gitLabWebhookVerifier) Verify(
	ctx context.Context,
	request webhooks.SignedRequest,
) (webhooks.VerifiedDelivery, error) {
	if verifier == nil || verifier.resolver == nil || len(verifier.signingKey) == 0 || verifier.now == nil {
		return webhooks.VerifiedDelivery{}, webhooks.ErrNotConfigured
	}
	deliveryID := strings.TrimSpace(request.Headers.First("webhook-id"))
	timestamp := strings.TrimSpace(request.Headers.First("webhook-timestamp"))
	if deliveryID == "" || timestamp == "" || !verifyStandardWebhookSignature(
		verifier.signingKey,
		deliveryID,
		timestamp,
		request.Body,
		request.Headers.First("webhook-signature"),
	) {
		return webhooks.VerifiedDelivery{}, webhooks.ErrUnauthenticated
	}
	if retryID := strings.TrimSpace(request.Headers.First("Idempotency-Key")); retryID != "" && retryID != deliveryID {
		return webhooks.VerifiedDelivery{}, webhooks.ErrInvalidDelivery
	}
	sentAtUnix, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return webhooks.VerifiedDelivery{}, webhooks.ErrReplay
	}
	age := verifier.now().UTC().Sub(time.Unix(sentAtUnix, 0).UTC())
	if age < -maximumFutureSkew || age > verifier.replayWindow {
		return webhooks.VerifiedDelivery{}, webhooks.ErrReplay
	}
	eventType := strings.TrimSpace(request.Headers.First("X-Gitlab-Event"))
	if eventType != "Issue Hook" && eventType != "Note Hook" {
		return webhooks.VerifiedDelivery{}, webhooks.ErrDeliveryIgnored
	}
	instanceURL, err := normalizeInstanceURL(request.Headers.First("X-Gitlab-Instance"))
	if err != nil || instanceURL != verifier.instanceURL {
		return webhooks.VerifiedDelivery{}, webhooks.ErrInvalidDelivery
	}
	var identity webhookIdentity
	if err := json.Unmarshal(request.Body, &identity); err != nil || identity.Project.ID <= 0 {
		return webhooks.VerifiedDelivery{}, webhooks.ErrInvalidDelivery
	}
	installation, err := verifier.resolver.ResolveGitLabWebhookInstallation(
		ctx,
		instanceURL,
		strconv.FormatInt(identity.Project.ID, 10),
	)
	if err != nil {
		if errors.Is(err, codehost.ErrNotFound) || errors.Is(err, codehost.ErrGrantRevoked) {
			return webhooks.VerifiedDelivery{}, webhooks.ErrDeliveryIgnored
		}
		return webhooks.VerifiedDelivery{}, err
	}
	if err := validateInstallation(installation); err != nil {
		return webhooks.VerifiedDelivery{}, webhooks.ErrInvalidDelivery
	}
	return webhooks.VerifiedDelivery{
		DeliveryID:             deliveryID,
		EventType:              eventType,
		ExternalAccountID:      installation.ExternalInstallationID,
		WorkspaceID:            installation.WorkspaceID,
		InstallationID:         installation.InstallationID,
		InstallationGeneration: installation.Generation,
		TraceID:                strings.TrimSpace(request.Headers.First("X-Gitlab-Event-UUID")),
	}, nil
}

func parseSigningToken(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "whsec_") {
		return nil, fmt.Errorf("configure GitLab webhook signing token: %w", webhooks.ErrNotConfigured)
	}
	encoded := strings.TrimPrefix(value, "whsec_")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(encoded)
	}
	if err != nil || len(decoded) < sha256.Size {
		clear(decoded)
		return nil, fmt.Errorf("configure GitLab webhook signing token: %w", webhooks.ErrNotConfigured)
	}
	return decoded, nil
}

func verifyStandardWebhookSignature(key []byte, deliveryID, timestamp string, body []byte, signatures string) bool {
	if len(key) == 0 || deliveryID == "" || timestamp == "" || len(body) == 0 {
		return false
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(deliveryID))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write([]byte(timestamp))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(body)
	expected := mac.Sum(nil)
	for _, signature := range strings.Fields(signatures) {
		if !strings.HasPrefix(signature, "v1,") {
			continue
		}
		provided, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(signature, "v1,"))
		if err == nil && hmac.Equal(expected, provided) {
			return true
		}
	}
	return false
}

type gitLabWebhookDispatcher struct{ queue WebhookQueue }

func (dispatcher gitLabWebhookDispatcher) Enqueue(ctx context.Context, task webhooks.Task) error {
	if dispatcher.queue == nil || task.Provider != ProviderKey || task.InboxID == uuid.Nil {
		return webhooks.ErrInvalidDelivery
	}
	return dispatcher.queue.EnqueueGitLabWebhook(ctx, WebhookTask{
		InboxID: task.InboxID,
	})
}

var (
	_ webhooks.WebhookVerifier = (*gitLabWebhookVerifier)(nil)
	_ webhooks.Dispatcher      = gitLabWebhookDispatcher{}
)
