package workspacesrepository

import (
	"context"
	"fmt"

	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	workspacesql "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

func (r *repo) GetWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID) (workspacedomain.WorkspaceSettings, error) {
	row, err := r.queries.GetWorkspaceSettings(ctx, workspacesql.GetWorkspaceSettingsParams{WorkspaceID: workspaceID})
	if err != nil {
		return workspacedomain.WorkspaceSettings{}, mapWorkspaceNotFound("get workspace settings", err)
	}
	return workspaceSettingsFromRecord(workspaceSettingsRecordFromGet(row)), nil
}

func (r *repo) GetOrCreateWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID) (workspacedomain.WorkspaceSettings, error) {
	row, err := r.queries.GetOrCreateWorkspaceSettings(
		ctx,
		workspacesql.GetOrCreateWorkspaceSettingsParams{WorkspaceID: workspaceID},
	)
	if err != nil {
		return workspacedomain.WorkspaceSettings{}, fmt.Errorf("get or create workspace settings: %w", err)
	}
	return workspaceSettingsFromRecord(workspaceSettingsRecordFromGetOrCreate(row)), nil
}

func (r *repo) UpdateWorkspaceSettings(
	ctx context.Context,
	workspaceID uuid.UUID,
	settings workspacedomain.WorkspaceSettings,
) (workspacedomain.WorkspaceSettings, error) {
	workingDays := make([]int16, len(settings.WorkingDays))
	for index, day := range settings.WorkingDays {
		converted, err := safecast.Int16(day)
		if err != nil {
			return workspacedomain.WorkspaceSettings{}, fmt.Errorf("validate working day: %w", err)
		}
		workingDays[index] = converted
	}
	workingStartMinute, err := safecast.Int16(settings.WorkingStartMinute)
	if err != nil {
		return workspacedomain.WorkspaceSettings{}, fmt.Errorf("validate working start minute: %w", err)
	}
	workingEndMinute, err := safecast.Int16(settings.WorkingEndMinute)
	if err != nil {
		return workspacedomain.WorkspaceSettings{}, fmt.Errorf("validate working end minute: %w", err)
	}
	row, err := r.queries.UpdateWorkspaceSettings(ctx, workspacesql.UpdateWorkspaceSettingsParams{
		StoryTerm:          settings.StoryTerm,
		SprintTerm:         settings.SprintTerm,
		ObjectiveTerm:      settings.ObjectiveTerm,
		KeyResultTerm:      settings.KeyResultTerm,
		ObjectiveEnabled:   settings.ObjectiveEnabled,
		KeyResultEnabled:   settings.KeyResultEnabled,
		WorkingDays:        workingDays,
		WorkingStartMinute: workingStartMinute,
		WorkingEndMinute:   workingEndMinute,
		WorkspaceID:        workspaceID,
	})
	if err != nil {
		return workspacedomain.WorkspaceSettings{}, mapWorkspaceNotFound("update workspace settings", err)
	}
	return workspaceSettingsFromRecord(workspaceSettingsRecordFromUpdate(row)), nil
}
