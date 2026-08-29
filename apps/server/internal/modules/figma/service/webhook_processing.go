package figma

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	figmaprovider "github.com/complexus-tech/projects-api/internal/modules/figma"
	figmadomain "github.com/complexus-tech/projects-api/internal/modules/figma/domain"
	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/google/uuid"
)

const (
	figmaWebhookProcessingLease   = 10 * time.Minute
	figmaWebhookCompletionTimeout = 5 * time.Second
)

func (s *Service) ReceiveWebhook(
	ctx context.Context,
	request webhooks.SignedRequest,
) (webhooks.Receipt, error) {
	if s == nil || s.webhookGateway == nil {
		return webhooks.Receipt{}, webhooks.ErrNotConfigured
	}
	request.Body = append([]byte(nil), request.Body...)
	return s.webhookGateway.Receive(ctx, figmaprovider.ProviderKey, request)
}

func (s *Service) ProcessWebhook(
	ctx context.Context,
	provider integrations.ProviderKey,
	inboxID uuid.UUID,
) error {
	if s == nil || provider != figmaprovider.ProviderKey || inboxID == uuid.Nil ||
		s.webhookInbox == nil || s.webhookPayloads == nil {
		return webhooks.ErrNotConfigured
	}
	record, process, err := s.webhookInbox.Start(
		ctx,
		inboxID,
		s.now().UTC(),
		figmaWebhookProcessingLease,
	)
	if errors.Is(err, webhooks.ErrLeaseBusy) {
		return nil
	}
	if err != nil || !process {
		return err
	}
	complete := func(status webhooks.Status, outcome string) error {
		completionContext, cancel := context.WithTimeout(
			context.WithoutCancel(ctx),
			figmaWebhookCompletionTimeout,
		)
		defer cancel()
		return s.webhookInbox.Complete(
			completionContext,
			record.ID,
			status,
			outcome,
			s.now().UTC(),
		)
	}
	fail := func(processErr error, outcome string) error {
		return errors.Join(processErr, complete(webhooks.StatusFailed, outcome))
	}
	if record.Provider != provider || record.ID != inboxID || record.EncryptedPayload == nil {
		return fail(webhooks.ErrInvalidDelivery, "figma.invalid_receipt")
	}
	body, err := s.webhookPayloads.Open(record, *record.EncryptedPayload)
	if err != nil {
		return fail(err, "figma.unreadable_payload")
	}
	defer clear(body)
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fail(err, "figma.invalid_payload")
	}
	event.Passcode = ""
	current, err := s.currentWebhookGrant(ctx, record, event)
	if err != nil {
		return fail(err, "figma.grant_lookup_failed")
	}
	if !current {
		return complete(webhooks.StatusCancelled, "figma.stale_installation")
	}
	if err := s.processWebhookEvent(ctx, record.WorkspaceID, event); err != nil {
		return fail(err, "figma.processing_failed")
	}
	return complete(webhooks.StatusCompleted, "figma.processed")
}

func (s *Service) currentWebhookGrant(
	ctx context.Context,
	record webhooks.Record,
	event WebhookEvent,
) (bool, error) {
	webhookID := int64(event.WebhookID)
	if webhookID <= 0 || record.ExternalAccountID != strconv.FormatInt(webhookID, 10) {
		return false, nil
	}
	webhook, err := s.repo.GetCurrentWebhook(
		ctx,
		record.InstallationID,
		record.InstallationGeneration,
		webhookID,
	)
	if err != nil {
		if errors.Is(err, figmadomain.ErrNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("resolve current Figma webhook installation: %w", err)
	}
	return webhook.WorkspaceID == record.WorkspaceID &&
		webhook.ConnectionID == record.InstallationID &&
		webhook.InstallationGeneration == record.InstallationGeneration &&
		webhook.FigmaWebhookID == webhookID &&
		webhook.IsActive, nil
}

func (s *Service) processWebhookEvent(
	ctx context.Context,
	workspaceID uuid.UUID,
	event WebhookEvent,
) error {
	if event.EventType == "PING" {
		return nil
	}
	links, err := s.repo.ListLinksByFile(ctx, workspaceID, event.FileKey)
	if err != nil {
		return err
	}
	for _, link := range links {
		if event.EventType == EventDevModeStatusUpdate &&
			link.Artifact.NodeID != nil && *link.Artifact.NodeID == event.NodeID {
			status := event.Status
			link.DevStatus = &status
			if err := s.repo.UpdateStoryLink(ctx, link); err != nil {
				return err
			}
			_ = s.stories.RecordActivity(ctx, StoryActivity{
				StoryID: link.StoryID, ActorID: link.CreatedByUserID,
				Type: "update", Field: "figma_dev_status",
				Previous: event.ChangeMessage, Current: event.Status,
				WorkspaceID: link.WorkspaceID,
			})
			continue
		}
		if event.EventType != EventFileUpdate {
			continue
		}
		artifact, resolveErr := s.ResolveLink(ctx, link.WorkspaceID, link.Artifact.CanonicalURL)
		if resolveErr != nil {
			now := s.now().UTC()
			link.UnavailableAt = &now
		} else {
			link.Artifact = artifact
			link.UnavailableAt = nil
		}
		if err := s.repo.UpdateStoryLink(ctx, link); err != nil {
			return err
		}
		_ = s.stories.RecordActivity(ctx, StoryActivity{
			StoryID: link.StoryID, ActorID: link.CreatedByUserID,
			Type: "update", Field: "figma_design",
			Previous: event.FileName, Current: link.Artifact.CanonicalURL,
			WorkspaceID: link.WorkspaceID,
		})
	}
	return nil
}

func (s *Service) RecoverPendingWebhooks(ctx context.Context) (int, error) {
	if s == nil || s.webhookGateway == nil {
		return 0, webhooks.ErrNotConfigured
	}
	report, err := s.webhookGateway.Recover(
		ctx,
		figmaprovider.ProviderKey,
		webhooks.DefaultRecoveryPolicy(),
	)
	if err != nil {
		return report.Dispatched, fmt.Errorf("recover Figma webhook deliveries: %w", err)
	}
	return report.Dispatched, nil
}
