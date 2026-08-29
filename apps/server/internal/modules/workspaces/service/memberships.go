package workspaces

import (
	"context"

	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *Service) AddMember(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	s.log.Info(ctx, "business.core.workspaces.addMember")
	ctx, span := startSpan(ctx, "business.core.workspaces.AddMember")
	defer span.End()
	role = normalizeWorkspaceRole(role)
	if err := authorization.ValidateWorkspaceRole(authorization.WorkspaceRole(role)); err != nil {
		return err
	}
	if err := s.repo.AddMember(ctx, workspaceID, userID, role); err != nil {
		span.RecordError(err)
		return err
	}
	if err := s.users.UpdateLastUsedWorkspace(ctx, userID, workspaceID); err != nil {
		s.log.Error(ctx, "failed to update user's last workspace", "error", err)
	}
	if err := s.subscriptions.UpdateWorkspaceSeats(ctx, workspaceID); err != nil {
		s.log.Error(ctx, "failed to update subscription seats", "error", err)
	}
	span.AddEvent("workspace member added.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()), attribute.String("user_id", userID.String()),
		attribute.String("role", role),
	))
	return nil
}

func (s *Service) RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	s.log.Info(ctx, "business.core.workspaces.removeMember")
	ctx, span := startSpan(ctx, "business.core.workspaces.RemoveMember")
	defer span.End()
	if err := s.repo.RemoveMember(ctx, workspaceID, userID); err != nil {
		span.RecordError(err)
		return err
	}
	if err := s.subscriptions.UpdateWorkspaceSeats(ctx, workspaceID); err != nil {
		s.log.Error(ctx, "failed to update subscription seats", "error", err)
	}
	span.AddEvent("workspace member removed.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()), attribute.String("user_id", userID.String()),
	))
	return nil
}

func (s *Service) UpdateMemberRole(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	s.log.Info(ctx, "business.core.workspaces.updateMemberRole")
	ctx, span := startSpan(ctx, "business.core.workspaces.UpdateMemberRole")
	defer span.End()
	role = normalizeWorkspaceRole(role)
	if err := authorization.ValidateWorkspaceRole(authorization.WorkspaceRole(role)); err != nil {
		return err
	}
	if err := s.repo.UpdateMemberRole(ctx, workspaceID, userID, role); err != nil {
		span.RecordError(err)
		return err
	}
	if err := s.subscriptions.UpdateWorkspaceSeats(ctx, workspaceID); err != nil {
		s.log.Error(ctx, "failed to update subscription seats", "error", err)
	}
	span.AddEvent("workspace member role updated.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()), attribute.String("user_id", userID.String()),
		attribute.String("role", role),
	))
	return nil
}

func normalizeWorkspaceRole(role string) string {
	if role == "" {
		return string(authorization.WorkspaceRoleMember)
	}
	return role
}
