package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/events"
)

// handleWorkspaceDeletionScheduledConfirmation processes workspace deletion confirmation events
func (c *Consumer) handleWorkspaceDeletionScheduledConfirmation(ctx context.Context, event events.Event) error {
	var payload events.WorkspaceDeletionScheduledConfirmationPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleWorkspaceDeletionScheduledConfirmation",
		"workspace_id", payload.WorkspaceID,
		"workspace_name", payload.WorkspaceName,
		"actor_email", payload.ActorEmail)

	// Prepare Brevo template parameters
	brevoParams := map[string]any{
		"WORKSPACE_NAME": payload.WorkspaceName,
		"WORKSPACE_URL":  fmt.Sprintf("https://%s.fortyone.app", payload.WorkspaceSlug),
		"RESTORE_URL":    fmt.Sprintf("https://%s.fortyone.app/settings", payload.WorkspaceSlug),
		"DELETION_TIME":  "48 hours",
	}

	// Send templated email via Brevo service
	req := brevo.SendTemplatedEmailRequest{
		TemplateID: brevo.TemplateWorkspaceDeletionScheduledConfirmation,
		To: []brevo.EmailRecipient{
			{
				Email: payload.ActorEmail,
			},
		},
		Subject: fmt.Sprintf("You have scheduled %s for deletion", payload.WorkspaceName),
		Params:  brevoParams,
	}

	if err := c.brevoService.SendTemplatedEmail(ctx, req); err != nil {
		c.log.Error(ctx, "failed to send workspace deletion confirmation email via Brevo", "error", err, "email", payload.ActorEmail)
		return fmt.Errorf("failed to send workspace deletion confirmation email via Brevo: %w", err)
	}

	c.log.Info(ctx, "successfully sent workspace deletion confirmation email via Brevo", "email", payload.ActorEmail)
	return nil
}

// handleWorkspaceDeletionScheduledNotification processes workspace deletion notification events
func (c *Consumer) handleWorkspaceDeletionScheduledNotification(ctx context.Context, event events.Event) error {
	var payload events.WorkspaceDeletionScheduledNotificationPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleWorkspaceDeletionScheduledNotification",
		"workspace_id", payload.WorkspaceID,
		"workspace_name", payload.WorkspaceName,
		"actor_id", payload.ActorID,
		"admin_count", len(payload.AdminEmails))

	if len(payload.AdminEmails) == 0 {
		c.log.Info(ctx, "no other workspace admins to notify", "workspace_id", payload.WorkspaceID)
		return nil
	}

	// Prepare Brevo template parameters
	brevoParams := map[string]any{
		"WORKSPACE_NAME": payload.WorkspaceName,
		"WORKSPACE_URL":  fmt.Sprintf("https://%s.fortyone.app", payload.WorkspaceSlug),
		"RESTORE_URL":    fmt.Sprintf("https://%s.fortyone.app/settings", payload.WorkspaceSlug),
		"DELETION_TIME":  "48 hours",
	}

	// Send templated email to all workspace admins
	req := brevo.SendTemplatedEmailRequest{
		TemplateID: brevo.TemplateWorkspaceDeletionScheduledNotification,
		To:         make([]brevo.EmailRecipient, len(payload.AdminEmails)),
		Subject:    fmt.Sprintf("Workspace %s scheduled for deletion", payload.WorkspaceName),
		Params:     brevoParams,
	}

	for i, email := range payload.AdminEmails {
		req.To[i] = brevo.EmailRecipient{Email: email}
	}

	if err := c.brevoService.SendTemplatedEmail(ctx, req); err != nil {
		c.log.Error(ctx, "failed to send workspace deletion notification email via Brevo", "error", err, "workspace_id", payload.WorkspaceID)
		return fmt.Errorf("failed to send workspace deletion notification email via Brevo: %w", err)
	}

	c.log.Info(ctx, "successfully sent workspace deletion notification email via Brevo", "workspace_id", payload.WorkspaceID, "admin_count", len(payload.AdminEmails))
	return nil
}

// handleWorkspaceRestoredConfirmation processes workspace restore confirmation events
func (c *Consumer) handleWorkspaceRestoredConfirmation(ctx context.Context, event events.Event) error {
	var payload events.WorkspaceRestoredConfirmationPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleWorkspaceRestoredConfirmation",
		"workspace_id", payload.WorkspaceID,
		"workspace_name", payload.WorkspaceName,
		"actor_email", payload.ActorEmail)

	// Prepare Brevo template parameters
	brevoParams := map[string]any{
		"WORKSPACE_NAME": payload.WorkspaceName,
		"WORKSPACE_URL":  fmt.Sprintf("https://%s.fortyone.app", payload.WorkspaceSlug),
	}

	// Send templated email via Brevo service
	req := brevo.SendTemplatedEmailRequest{
		TemplateID: brevo.TemplateWorkspaceRestoredConfirmation,
		To: []brevo.EmailRecipient{
			{
				Email: payload.ActorEmail,
			},
		},
		Subject: fmt.Sprintf("You have restored %s", payload.WorkspaceName),
		Params:  brevoParams,
	}

	if err := c.brevoService.SendTemplatedEmail(ctx, req); err != nil {
		c.log.Error(ctx, "failed to send workspace restore confirmation email via Brevo", "error", err, "email", payload.ActorEmail)
		return fmt.Errorf("failed to send workspace restore confirmation email via Brevo: %w", err)
	}

	c.log.Info(ctx, "successfully sent workspace restore confirmation email via Brevo", "email", payload.ActorEmail)
	return nil
}

// handleWorkspaceRestoredNotification processes workspace restore notification events
func (c *Consumer) handleWorkspaceRestoredNotification(ctx context.Context, event events.Event) error {
	var payload events.WorkspaceRestoredNotificationPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleWorkspaceRestoredNotification",
		"workspace_id", payload.WorkspaceID,
		"workspace_name", payload.WorkspaceName,
		"actor_id", payload.ActorID,
		"admin_count", len(payload.AdminEmails))

	if len(payload.AdminEmails) == 0 {
		c.log.Info(ctx, "no other workspace admins to notify", "workspace_id", payload.WorkspaceID)
		return nil
	}

	// Prepare Brevo template parameters
	brevoParams := map[string]any{
		"WORKSPACE_NAME": payload.WorkspaceName,
		"WORKSPACE_URL":  fmt.Sprintf("https://%s.fortyone.app", payload.WorkspaceSlug),
	}

	// Send templated email to all workspace admins
	req := brevo.SendTemplatedEmailRequest{
		TemplateID: brevo.TemplateWorkspaceRestoredNotification,
		To:         make([]brevo.EmailRecipient, len(payload.AdminEmails)),
		Subject:    fmt.Sprintf("Workspace %s has been restored", payload.WorkspaceName),
		Params:     brevoParams,
	}

	for i, email := range payload.AdminEmails {
		req.To[i] = brevo.EmailRecipient{Email: email}
	}

	if err := c.brevoService.SendTemplatedEmail(ctx, req); err != nil {
		c.log.Error(ctx, "failed to send workspace restore notification email via Brevo", "error", err, "workspace_id", payload.WorkspaceID)
		return fmt.Errorf("failed to send workspace restore notification email via Brevo: %w", err)
	}

	c.log.Info(ctx, "successfully sent workspace restore notification email via Brevo", "workspace_id", payload.WorkspaceID, "admin_count", len(payload.AdminEmails))
	return nil
}
