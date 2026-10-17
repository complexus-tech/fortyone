package workspacesrepository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"

	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	workspacesql "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
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
	parameters := workspacesql.UpdateWorkspaceSettingsParams{
		ExpectedUpdatedAt:  settings.UpdatedAt,
		StoryTerm:          settings.StoryTerm,
		SprintTerm:         settings.SprintTerm,
		ObjectiveTerm:      settings.ObjectiveTerm,
		KeyResultTerm:      settings.KeyResultTerm,
		ChatEnabled:        settings.ChatEnabled,
		ObjectiveEnabled:   settings.ObjectiveEnabled,
		KeyResultEnabled:   settings.KeyResultEnabled,
		WorkingDays:        workingDays,
		WorkingStartMinute: workingStartMinute,
		WorkingEndMinute:   workingEndMinute,
		WorkspaceID:        workspaceID,
	}
	var row workspacesql.UpdateWorkspaceSettingsRow
	update := func(queries workspacesql.Querier) error {
		// Serialize feature changes with chat mutations and Maya actions. Both
		// lock the workspace first, including legacy workspaces without settings.
		if _, err := queries.LockWorkspaceSettingsUpdate(ctx, workspacesql.LockWorkspaceSettingsUpdateParams{WorkspaceID: workspaceID}); err != nil {
			return err
		}
		var err error
		row, err = queries.UpdateWorkspaceSettings(ctx, parameters)
		if errors.Is(err, pgx.ErrNoRows) {
			return workspacedomain.ErrSettingsConflict
		}
		return err
	}
	if tx, ok := platformdatabase.UnitTransaction(ctx); ok {
		err = update(workspacesql.New(tx))
	} else {
		err = r.withinTransaction(ctx, update)
	}

	if err != nil {
		return workspacedomain.WorkspaceSettings{}, mapWorkspaceNotFound("update workspace settings", err)
	}
	return workspaceSettingsFromRecord(workspaceSettingsRecordFromUpdate(row)), nil
}

// ChatEnabled reads the current setting without creating or rewriting settings.
// Workspaces predating settings initialization retain the default-on behavior.
func (r *repo) ChatEnabled(ctx context.Context, workspaceID uuid.UUID) (bool, error) {
	queries := r.queries
	if tx, ok := platformdatabase.UnitTransaction(ctx); ok {
		queries = workspacesql.New(tx)
	}
	enabled, err := queries.GetWorkspaceChatEnabled(ctx, workspacesql.GetWorkspaceChatEnabledParams{WorkspaceID: workspaceID})
	if err != nil {
		return false, mapWorkspaceNotFound("read workspace chat setting", err)
	}
	return enabled, nil
}
