package teamsettingsrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *repo) GetSprintSettings(ctx context.Context, teamID, workspaceID uuid.UUID) (teamsettings.CoreTeamSprintSettings, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.teamsettings.GetSprintSettings")
	defer span.End()

	var settings dbTeamSprintSettings
	query := `
		SELECT
			team_id,
			workspace_id,
			auto_create_sprints,
			upcoming_sprints_count,
			sprint_duration_weeks,
			sprint_start_day,
			move_incomplete_stories_enabled,
			last_auto_sprint_number,
			created_at,
			updated_at
		FROM
			team_sprint_settings
		WHERE
			team_id = :team_id
			AND workspace_id = :workspace_id
	`

	params := map[string]any{
		"team_id":      teamID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teamsettings.CoreTeamSprintSettings{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching team sprint settings.")
	if err := stmt.GetContext(ctx, &settings, params); err != nil {
		if err == sql.ErrNoRows {
			// Return default settings if not found
			return r.createDefaultSprintSettings(ctx, teamID, workspaceID)
		}
		errMsg := fmt.Sprintf("Failed to retrieve team sprint settings from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("team sprint settings not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teamsettings.CoreTeamSprintSettings{}, err
	}

	r.log.Info(ctx, "Team sprint settings retrieved successfully.")
	span.AddEvent("sprint settings retrieved.", trace.WithAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))

	return toCoreTeamSprintSettings(settings), nil
}

func (r *repo) GetStoryAutomationSettings(ctx context.Context, teamID, workspaceID uuid.UUID) (teamsettings.CoreTeamStoryAutomationSettings, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.teamsettings.GetStoryAutomationSettings")
	defer span.End()

	var settings dbTeamStoryAutomationSettings
	query := `
		SELECT
			team_id,
			workspace_id,
			auto_close_inactive_enabled,
			auto_close_inactive_months,
			auto_archive_enabled,
			auto_archive_months,
			created_at,
			updated_at
		FROM
			team_story_automation_settings
		WHERE
			team_id = :team_id
			AND workspace_id = :workspace_id
	`

	params := map[string]any{
		"team_id":      teamID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teamsettings.CoreTeamStoryAutomationSettings{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "Fetching team story automation settings.")
	if err := stmt.GetContext(ctx, &settings, params); err != nil {
		if err == sql.ErrNoRows {
			// Return default settings if not found
			return r.createDefaultStoryAutomationSettings(ctx, teamID, workspaceID)
		}
		errMsg := fmt.Sprintf("Failed to retrieve team story automation settings from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("team story automation settings not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teamsettings.CoreTeamStoryAutomationSettings{}, err
	}

	r.log.Info(ctx, "Team story automation settings retrieved successfully.")
	span.AddEvent("story automation settings retrieved.", trace.WithAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))

	return toCoreTeamStoryAutomationSettings(settings), nil
}

func (r *repo) GetTeamsWithAutoSprintCreation(ctx context.Context) ([]teamsettings.CoreTeamSprintSettings, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.teamsettings.GetTeamsWithAutoSprintCreation")
	defer span.End()

	var settings []dbTeamSprintSettings
	query := `
		SELECT
			team_id,
			workspace_id,
			auto_create_sprints,
			upcoming_sprints_count,
			sprint_duration_weeks,
			sprint_start_day,
			move_incomplete_stories_enabled,
			last_auto_sprint_number,
			created_at,
			updated_at
		FROM
			team_sprint_settings
		WHERE
			auto_create_sprints = true
	`

	r.log.Info(ctx, "Fetching teams with auto sprint creation.")
	if err := r.db.SelectContext(ctx, &settings, query); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve teams with auto sprint creation: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("teams with auto sprint creation not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "Teams with auto sprint creation retrieved successfully.")
	span.AddEvent("teams with auto sprint creation retrieved.", trace.WithAttributes(
		attribute.Int("teams.count", len(settings)),
	))

	return toCoreTeamSprintSettingsList(settings), nil
}
