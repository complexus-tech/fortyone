package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/google/uuid"
)

const (
	githubWebhookProcessingLease   = 10 * time.Minute
	githubWebhookCompletionTimeout = 5 * time.Second
)

func (s *Service) ReceiveWebhook(ctx context.Context, request webhooks.SignedRequest) (webhooks.Receipt, error) {
	if s == nil || s.webhookGateway == nil {
		return webhooks.Receipt{}, webhooks.ErrNotConfigured
	}
	request.Body = append([]byte(nil), request.Body...)
	return s.webhookGateway.Receive(ctx, githubWebhookProvider, request)
}

// HandleWebhook is retained as a compatibility boundary for callers that do
// not have an HTTP request object. Processing is always durable and async.
func (s *Service) HandleWebhook(ctx context.Context, deliveryID, eventName, signature string, body []byte) error {
	_, err := s.ReceiveWebhook(ctx, webhooks.SignedRequest{
		Method: "POST",
		Headers: webhooks.Headers{
			"X-GitHub-Delivery":   {deliveryID},
			"X-GitHub-Event":      {eventName},
			"X-Hub-Signature-256": {signature},
		},
		Body: body,
	})
	return err
}

func (s *Service) ProcessWebhook(ctx context.Context, provider integrations.ProviderKey, inboxID uuid.UUID) error {
	if s == nil || provider != githubWebhookProvider || inboxID == uuid.Nil ||
		s.webhookInbox == nil || s.webhookPayloads == nil {
		return webhooks.ErrNotConfigured
	}
	record, process, err := s.webhookInbox.Start(ctx, inboxID, s.now().UTC(), githubWebhookProcessingLease)
	if errors.Is(err, webhooks.ErrLeaseBusy) {
		return nil
	}
	if err != nil || !process {
		return err
	}
	complete := func(status webhooks.Status, outcome string) error {
		completionContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), githubWebhookCompletionTimeout)
		defer cancel()
		return s.webhookInbox.Complete(completionContext, record.ID, status, outcome, s.now().UTC())
	}
	fail := func(processErr error, outcome string) error {
		return errors.Join(processErr, complete(webhooks.StatusFailed, outcome))
	}
	if record.Provider != provider || record.ID != inboxID || record.EncryptedPayload == nil {
		return fail(webhooks.ErrInvalidDelivery, "github.invalid_receipt")
	}
	body, err := s.webhookPayloads.Open(record, *record.EncryptedPayload)
	if err != nil {
		return fail(err, "github.unreadable_payload")
	}
	defer clear(body)
	var payload webhookEnvelope
	if err := json.Unmarshal(body, &payload); err != nil {
		return fail(err, "github.invalid_payload")
	}
	current, err := s.currentWebhookGrant(ctx, record, payload)
	if err != nil {
		return fail(err, "github.grant_lookup_failed")
	}
	if !current {
		return complete(webhooks.StatusCancelled, "github.stale_installation")
	}
	processingContext, err := s.webhookActorContext(ctx, record)
	if err != nil {
		return fail(err, "github.actor_context_failed")
	}
	if err := s.processWebhook(processingContext, record.EventType, payload); err != nil {
		return fail(err, "github.processing_failed")
	}
	return complete(webhooks.StatusCompleted, "github.processed")
}

func (s *Service) webhookActorContext(ctx context.Context, record webhooks.Record) (context.Context, error) {
	if s == nil || s.cfg.GitHubUserID == uuid.Nil || record.WorkspaceID == uuid.Nil || record.InstallationID == uuid.Nil {
		return nil, errors.New("github webhook actor identity is incomplete")
	}
	actor, err := platformauth.NewActor(
		s.cfg.GitHubUserID,
		platformauth.PrincipalSystem,
		record.InstallationID,
		platformauth.MustScopeSet(platformauth.ScopeStoriesRead, platformauth.ScopeStoriesWrite),
		platformauth.UnrestrictedTeamAccess(),
	)
	if err != nil {
		return nil, fmt.Errorf("construct GitHub installation actor: %w", err)
	}
	actor, err = actor.WithWorkspace(record.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("bind GitHub installation actor: %w", err)
	}
	return platformauth.SetActor(ctx, actor)
}

func (s *Service) currentWebhookGrant(ctx context.Context, record webhooks.Record, payload webhookEnvelope) (bool, error) {
	if s == nil || s.webhookInstallations == nil || payload.Installation.ID <= 0 || payload.Repository.ID <= 0 ||
		record.ExternalAccountID != strconv.FormatInt(payload.Installation.ID, 10) {
		return false, nil
	}
	installation, err := s.webhookInstallations.GetCurrentWebhookInstallation(
		ctx,
		record.InstallationID,
		record.InstallationGeneration,
		payload.Repository.ID,
	)
	if errors.Is(err, githubshared.ErrWebhookInstallationNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("resolve current GitHub webhook installation: %w", err)
	}
	return installation.WorkspaceID == record.WorkspaceID &&
		installation.ID == record.InstallationID &&
		installation.InstallationGeneration == record.InstallationGeneration &&
		installation.ExternalInstallationID == payload.Installation.ID &&
		installation.ExternalRepositoryID == payload.Repository.ID, nil
}

func (s *Service) RecoverPendingWebhooks(ctx context.Context) (int, error) {
	if s == nil || s.webhookInbox == nil || s.webhookDispatcher == nil {
		return 0, webhooks.ErrNotConfigured
	}
	report, err := webhooks.RecoverDeliveries(
		ctx,
		s.webhookInbox,
		githubWebhookProvider,
		s.webhookDispatcher,
		webhooks.DefaultRecoveryPolicy(),
		s.now().UTC(),
	)
	if err != nil {
		return report.Dispatched, fmt.Errorf("recover GitHub webhook deliveries: %w", err)
	}
	return report.Dispatched, nil
}
