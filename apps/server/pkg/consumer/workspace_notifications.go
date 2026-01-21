package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/mailer"
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

	data := map[string]any{
		"WorkspaceName": payload.WorkspaceName,
		"WorkspaceSlug": payload.WorkspaceSlug,
		"WorkspaceURL":  fmt.Sprintf("https://%s.fortyone.app", payload.WorkspaceSlug),
		"RestoreURL":    fmt.Sprintf("https://%s.fortyone.app/settings", payload.WorkspaceSlug),
		"DeletionTime":  "48 hours",
	}

	subject := fmt.Sprintf("You have scheduled workspace %s for deletion", payload.WorkspaceName)
	if err := c.mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:       []string{payload.ActorEmail},
		Template: "workspaces/deletion_scheduled_confirmation",
		Subject:  subject,
		Data:     data,
	}); err != nil {
		c.log.Error(ctx, "failed to send workspace deletion confirmation email", "error", err, "email", payload.ActorEmail)
		return fmt.Errorf("failed to send workspace deletion confirmation email: %w", err)
	}

	c.log.Info(ctx, "successfully sent workspace deletion confirmation email", "email", payload.ActorEmail)
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

	data := map[string]any{
		"ActorName":     payload.ActorName,
		"ActorEmail":    payload.ActorEmail,
		"WorkspaceName": payload.WorkspaceName,
		"WorkspaceSlug": payload.WorkspaceSlug,
		"WorkspaceURL":  fmt.Sprintf("https://%s.fortyone.app", payload.WorkspaceSlug),
		"RestoreURL":    fmt.Sprintf("https://%s.fortyone.app/settings", payload.WorkspaceSlug),
		"DeletionTime":  "48 hours",
	}

	subject := fmt.Sprintf("%s has scheduled your %s workspace for deletion", payload.ActorName, payload.WorkspaceName)
	if err := c.mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:       payload.AdminEmails,
		Template: "workspaces/deletion_scheduled_notification",
		Subject:  subject,
		Data:     data,
	}); err != nil {
		c.log.Error(ctx, "failed to send workspace deletion notification email", "error", err, "workspace_id", payload.WorkspaceID)
		return fmt.Errorf("failed to send workspace deletion notification email: %w", err)
	}

	c.log.Info(ctx, "successfully sent workspace deletion notification email", "workspace_id", payload.WorkspaceID, "admin_count", len(payload.AdminEmails))
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

	data := map[string]any{
		"WorkspaceName": payload.WorkspaceName,
		"WorkspaceURL":  fmt.Sprintf("https://%s.fortyone.app", payload.WorkspaceSlug),
		"ActorName":     payload.ActorName,
		"ActorEmail":    payload.ActorEmail,
		"WorkspaceSlug": payload.WorkspaceSlug,
	}

	subject := fmt.Sprintf("You have restored workspace %s", payload.WorkspaceName)
	if err := c.mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:       []string{payload.ActorEmail},
		Template: "workspaces/restored_confirmation",
		Subject:  subject,
		Data:     data,
	}); err != nil {
		c.log.Error(ctx, "failed to send workspace restore confirmation email", "error", err, "email", payload.ActorEmail)
		return fmt.Errorf("failed to send workspace restore confirmation email: %w", err)
	}

	c.log.Info(ctx, "successfully sent workspace restore confirmation email", "email", payload.ActorEmail)
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

	data := map[string]any{
		"WorkspaceName": payload.WorkspaceName,
		"WorkspaceURL":  fmt.Sprintf("https://%s.fortyone.app", payload.WorkspaceSlug),
		"ActorName":     payload.ActorName,
		"ActorEmail":    payload.ActorEmail,
		"WorkspaceSlug": payload.WorkspaceSlug,
	}

	subject := fmt.Sprintf("%s has restored your %s workspace", payload.ActorName, payload.WorkspaceName)
	if err := c.mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:       payload.AdminEmails,
		Template: "workspaces/restored_notification",
		Subject:  subject,
		Data:     data,
	}); err != nil {
		c.log.Error(ctx, "failed to send workspace restore notification email", "error", err, "workspace_id", payload.WorkspaceID)
		return fmt.Errorf("failed to send workspace restore notification email: %w", err)
	}

	c.log.Info(ctx, "successfully sent workspace restore notification email", "workspace_id", payload.WorkspaceID, "admin_count", len(payload.AdminEmails))
	return nil
}
