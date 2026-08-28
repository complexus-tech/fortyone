package workspaces

import (
	"context"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *Service) Delete(ctx context.Context, workspaceID, deletedBy uuid.UUID) error {
	s.log.Info(ctx, "business.core.workspaces.delete")
	ctx, span := startSpan(ctx, "business.core.workspaces.Delete")
	defer span.End()
	if err := s.repo.Delete(ctx, workspaceID, deletedBy); err != nil {
		span.RecordError(err)
		return err
	}
	if err := s.subscriptions.CancelWorkspaceSubscription(ctx, workspaceID); err != nil {
		s.log.Error(ctx, "failed to cancel workspace subscription", "error", err, "workspace_id", workspaceID)
	}
	if err := s.publishWorkspaceDeletionEvents(ctx, workspaceID, deletedBy); err != nil {
		s.log.Error(ctx, "failed to publish workspace deletion events", "error", err, "workspace_id", workspaceID)
	}
	span.AddEvent("workspace scheduled for deletion.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()), attribute.String("deleted_by", deletedBy.String()),
	))
	return nil
}

func (s *Service) Restore(ctx context.Context, workspaceID, restoredBy uuid.UUID) error {
	s.log.Info(ctx, "business.core.workspaces.restore")
	ctx, span := startSpan(ctx, "business.core.workspaces.Restore")
	defer span.End()
	if err := s.repo.Restore(ctx, workspaceID, restoredBy); err != nil {
		span.RecordError(err)
		return err
	}
	if err := s.publishWorkspaceRestoreEvents(ctx, workspaceID, restoredBy); err != nil {
		s.log.Error(ctx, "failed to publish workspace restore events", "error", err, "workspace_id", workspaceID)
	}
	span.AddEvent("workspace restored.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()), attribute.String("restored_by", restoredBy.String()),
	))
	return nil
}

func (s *Service) publishWorkspaceDeletionEvents(ctx context.Context, workspaceID, actorID uuid.UUID) error {
	workspace, actorName, actorEmail, err := s.workspaceEventContext(ctx, workspaceID, actorID)
	if err != nil {
		return err
	}
	confirmation := events.Event{
		Type: events.WorkspaceDeletionScheduledConfirmation,
		Payload: events.WorkspaceDeletionScheduledConfirmationPayload{
			WorkspaceID: workspaceID, WorkspaceName: workspace.Name, WorkspaceSlug: workspace.Slug,
			ActorEmail: actorEmail, ActorName: actorName,
		},
		Timestamp: time.Now(), ActorID: actorID,
	}
	if err := s.publisher.Publish(ctx, confirmation); err != nil {
		return fmt.Errorf("publish workspace deletion confirmation: %w", err)
	}
	admins := s.workspaceAdminEmails(ctx, workspaceID, actorID)
	notification := events.Event{
		Type: events.WorkspaceDeletionScheduledNotification,
		Payload: events.WorkspaceDeletionScheduledNotificationPayload{
			WorkspaceID: workspaceID, WorkspaceName: workspace.Name, WorkspaceSlug: workspace.Slug,
			ActorID: actorID, ActorName: actorName, ActorEmail: actorEmail, AdminEmails: admins,
		},
		Timestamp: time.Now(), ActorID: actorID,
	}
	if err := s.publisher.Publish(ctx, notification); err != nil {
		return fmt.Errorf("publish workspace deletion notification: %w", err)
	}
	return nil
}

func (s *Service) publishWorkspaceRestoreEvents(ctx context.Context, workspaceID, actorID uuid.UUID) error {
	workspace, actorName, actorEmail, err := s.workspaceEventContext(ctx, workspaceID, actorID)
	if err != nil {
		return err
	}
	confirmation := events.Event{
		Type: events.WorkspaceRestoredConfirmation,
		Payload: events.WorkspaceRestoredConfirmationPayload{
			WorkspaceID: workspaceID, WorkspaceName: workspace.Name, WorkspaceSlug: workspace.Slug,
			ActorEmail: actorEmail, ActorName: actorName,
		},
		Timestamp: time.Now(), ActorID: actorID,
	}
	if err := s.publisher.Publish(ctx, confirmation); err != nil {
		return fmt.Errorf("publish workspace restore confirmation: %w", err)
	}
	admins := s.workspaceAdminEmails(ctx, workspaceID, actorID)
	notification := events.Event{
		Type: events.WorkspaceRestoredNotification,
		Payload: events.WorkspaceRestoredNotificationPayload{
			WorkspaceID: workspaceID, WorkspaceName: workspace.Name, WorkspaceSlug: workspace.Slug,
			ActorID: actorID, ActorName: actorName, ActorEmail: actorEmail, AdminEmails: admins,
		},
		Timestamp: time.Now(), ActorID: actorID,
	}
	if err := s.publisher.Publish(ctx, notification); err != nil {
		return fmt.Errorf("publish workspace restore notification: %w", err)
	}
	return nil
}

func (s *Service) workspaceEventContext(ctx context.Context, workspaceID, actorID uuid.UUID) (CoreWorkspace, string, string, error) {
	workspace, err := s.repo.Get(ctx, workspaceID, actorID)
	if err != nil {
		return CoreWorkspace{}, "", "", fmt.Errorf("get workspace event context: %w", err)
	}
	actor, err := s.users.GetWorkspaceUser(ctx, actorID)
	if err != nil {
		return CoreWorkspace{}, "", "", fmt.Errorf("get workspace event actor: %w", err)
	}
	name := actor.FullName
	if name == "" {
		name = actor.Username
	}
	return workspace, name, actor.Email, nil
}

func (s *Service) workspaceAdminEmails(ctx context.Context, workspaceID, actorID uuid.UUID) []string {
	emails, err := s.repo.GetWorkspaceAdminEmails(ctx, workspaceID, actorID)
	if err != nil {
		s.log.Error(ctx, "failed to get workspace admin emails", "error", err, "workspace_id", workspaceID)
		return []string{}
	}
	return emails
}
