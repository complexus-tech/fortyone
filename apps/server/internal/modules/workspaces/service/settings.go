package workspaces

import (
	"context"

	"github.com/complexus-tech/projects-api/internal/platform/workschedule"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *Service) GetWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspaceSettings, error) {
	s.log.Info(ctx, "business.core.workspaces.getSettings")
	ctx, span := startSpan(ctx, "business.core.workspaces.GetWorkspaceSettings")
	defer span.End()
	settings, err := s.repo.GetWorkspaceSettings(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspaceSettings{}, err
	}
	span.AddEvent("settings retrieved.", trace.WithAttributes(attribute.String("workspace_id", workspaceID.String())))
	return settings, nil
}

func (s *Service) GetOrCreateWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID) (CoreWorkspaceSettings, error) {
	s.log.Info(ctx, "business.core.workspaces.getOrCreateSettings")
	ctx, span := startSpan(ctx, "business.core.workspaces.GetOrCreateWorkspaceSettings")
	defer span.End()
	settings, err := s.repo.GetOrCreateWorkspaceSettings(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspaceSettings{}, err
	}
	span.AddEvent("settings retrieved or initialized.", trace.WithAttributes(attribute.String("workspace_id", workspaceID.String())))
	return settings, nil
}

func (s *Service) UpdateWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID, settings CoreWorkspaceSettings) (CoreWorkspaceSettings, error) {
	s.log.Info(ctx, "business.core.workspaces.updateSettings")
	ctx, span := startSpan(ctx, "business.core.workspaces.UpdateWorkspaceSettings")
	defer span.End()
	settings.WorkspaceID = workspaceID
	if err := workschedule.ValidateWorkingDays(settings.WorkingDays); err != nil {
		return CoreWorkspaceSettings{}, err
	}
	if err := workschedule.ValidateHours(settings.WorkingStartMinute, settings.WorkingEndMinute); err != nil {
		return CoreWorkspaceSettings{}, err
	}
	updated, err := s.repo.UpdateWorkspaceSettings(ctx, workspaceID, settings)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspaceSettings{}, err
	}
	span.AddEvent("settings updated.", trace.WithAttributes(attribute.String("workspace_id", workspaceID.String())))
	return updated, nil
}
