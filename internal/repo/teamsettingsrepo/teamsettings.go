package teamsettingsrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/internal/core/teamsettings"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}

func (r *repo) GetSprintSettings(ctx context.Context, teamID, workspaceID uuid.UUID) (teamsettings.CoreTeamSprintSettings, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.teamsettings.GetSprintSettings")
	defer span.End()

	var settings dbTeamSprintSettings
	query := `
		SELECT
			team_id,
			workspace_id,
			sprints_enabled,
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

func (r *repo) UpdateSprintSettings(ctx context.Context, teamID, workspaceID uuid.UUID, updates teamsettings.CoreUpdateTeamSprintSettings) (teamsettings.CoreTeamSprintSettings, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.teamsettings.UpdateSprintSettings")
	defer span.End()

	query := "UPDATE team_sprint_settings SET "
	var setClauses []string
	params := map[string]any{
		"team_id":      teamID,
		"workspace_id": workspaceID,
	}

	if updates.SprintsEnabled != nil {
		setClauses = append(setClauses, "sprints_enabled = :sprints_enabled")
		params["sprints_enabled"] = *updates.SprintsEnabled
	}
	if updates.AutoCreateSprints != nil {
		setClauses = append(setClauses, "auto_create_sprints = :auto_create_sprints")
		params["auto_create_sprints"] = *updates.AutoCreateSprints
	}
	if updates.UpcomingSprintsCount != nil {
		setClauses = append(setClauses, "upcoming_sprints_count = :upcoming_sprints_count")
		params["upcoming_sprints_count"] = *updates.UpcomingSprintsCount
	}
	if updates.SprintDurationWeeks != nil {
		setClauses = append(setClauses, "sprint_duration_weeks = :sprint_duration_weeks")
		params["sprint_duration_weeks"] = *updates.SprintDurationWeeks
	}
	if updates.SprintStartDay != nil {
		setClauses = append(setClauses, "sprint_start_day = :sprint_start_day")
		params["sprint_start_day"] = *updates.SprintStartDay
	}
	if updates.MoveIncompleteStoriesEnabled != nil {
		setClauses = append(setClauses, "move_incomplete_stories_enabled = :move_incomplete_stories_enabled")
		params["move_incomplete_stories_enabled"] = *updates.MoveIncompleteStoriesEnabled
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	query += strings.Join(setClauses, ", ")
	query += " WHERE team_id = :team_id AND workspace_id = :workspace_id;"

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teamsettings.CoreTeamSprintSettings{}, err
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		// Check for constraint violations and return user-friendly errors
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Constraint {
			case "team_sprint_settings_sprint_start_day_check":
				return teamsettings.CoreTeamSprintSettings{}, teamsettings.ErrInvalidSprintStartDay
			case "team_sprint_settings_sprint_duration_weeks_check":
				return teamsettings.CoreTeamSprintSettings{}, teamsettings.ErrInvalidSprintDuration
			case "team_sprint_settings_upcoming_sprints_count_check":
				return teamsettings.CoreTeamSprintSettings{}, teamsettings.ErrInvalidUpcomingCount
			}
		}
		errMsg := fmt.Sprintf("Failed to update team sprint settings: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update team sprint settings"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teamsettings.CoreTeamSprintSettings{}, err
	}

	// Get the updated settings
	updatedSettings, err := r.GetSprintSettings(ctx, teamID, workspaceID)
	if err != nil {
		return teamsettings.CoreTeamSprintSettings{}, err
	}

	r.log.Info(ctx, "Team sprint settings updated successfully.")
	span.AddEvent("sprint settings updated.", trace.WithAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))

	return updatedSettings, nil
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

func (r *repo) UpdateStoryAutomationSettings(ctx context.Context, teamID, workspaceID uuid.UUID, updates teamsettings.CoreUpdateTeamStoryAutomationSettings) (teamsettings.CoreTeamStoryAutomationSettings, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.teamsettings.UpdateStoryAutomationSettings")
	defer span.End()

	query := "UPDATE team_story_automation_settings SET "
	var setClauses []string
	params := map[string]any{
		"team_id":      teamID,
		"workspace_id": workspaceID,
	}

	if updates.AutoCloseInactiveEnabled != nil {
		setClauses = append(setClauses, "auto_close_inactive_enabled = :auto_close_inactive_enabled")
		params["auto_close_inactive_enabled"] = *updates.AutoCloseInactiveEnabled
	}
	if updates.AutoCloseInactiveMonths != nil {
		setClauses = append(setClauses, "auto_close_inactive_months = :auto_close_inactive_months")
		params["auto_close_inactive_months"] = *updates.AutoCloseInactiveMonths
	}
	if updates.AutoArchiveEnabled != nil {
		setClauses = append(setClauses, "auto_archive_enabled = :auto_archive_enabled")
		params["auto_archive_enabled"] = *updates.AutoArchiveEnabled
	}
	if updates.AutoArchiveMonths != nil {
		setClauses = append(setClauses, "auto_archive_months = :auto_archive_months")
		params["auto_archive_months"] = *updates.AutoArchiveMonths
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	query += strings.Join(setClauses, ", ")
	query += " WHERE team_id = :team_id AND workspace_id = :workspace_id;"

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teamsettings.CoreTeamStoryAutomationSettings{}, err
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		// Check for constraint violations and return user-friendly errors
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Constraint {
			case "team_story_automation_settings_auto_close_inactive_months_check":
				return teamsettings.CoreTeamStoryAutomationSettings{}, teamsettings.ErrInvalidCloseMonths
			case "team_story_automation_settings_auto_archive_months_check":
				return teamsettings.CoreTeamStoryAutomationSettings{}, teamsettings.ErrInvalidArchiveMonths
			}
		}
		errMsg := fmt.Sprintf("Failed to update team story automation settings: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update team story automation settings"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teamsettings.CoreTeamStoryAutomationSettings{}, err
	}

	// Get the updated settings
	updatedSettings, err := r.GetStoryAutomationSettings(ctx, teamID, workspaceID)
	if err != nil {
		return teamsettings.CoreTeamStoryAutomationSettings{}, err
	}

	r.log.Info(ctx, "Team story automation settings updated successfully.")
	span.AddEvent("story automation settings updated.", trace.WithAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))

	return updatedSettings, nil
}

func (r *repo) GetTeamsWithAutoSprintCreation(ctx context.Context) ([]teamsettings.CoreTeamSprintSettings, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.teamsettings.GetTeamsWithAutoSprintCreation")
	defer span.End()

	var settings []dbTeamSprintSettings
	query := `
		SELECT
			team_id,
			workspace_id,
			sprints_enabled,
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
			sprints_enabled = true
			AND auto_create_sprints = true
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

func (r *repo) IncrementAutoSprintNumber(ctx context.Context, teamID, workspaceID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.teamsettings.IncrementAutoSprintNumber")
	defer span.End()

	query := `
		UPDATE team_sprint_settings
		SET 
			last_auto_sprint_number = last_auto_sprint_number + 1,
			updated_at = NOW()
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
		return err
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		errMsg := fmt.Sprintf("Failed to increment auto sprint number: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to increment auto sprint number"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	r.log.Info(ctx, "Auto sprint number incremented successfully.")
	span.AddEvent("auto sprint number incremented.", trace.WithAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))

	return nil
}

// createDefaultSprintSettings creates and returns default sprint settings
func (r *repo) createDefaultSprintSettings(ctx context.Context, teamID, workspaceID uuid.UUID) (teamsettings.CoreTeamSprintSettings, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.teamsettings.createDefaultSprintSettings")
	defer span.End()

	query := `
		INSERT INTO team_sprint_settings (
			team_id,
			workspace_id,
			sprints_enabled,
			auto_create_sprints,
			upcoming_sprints_count,
			sprint_duration_weeks,
			sprint_start_day,
			move_incomplete_stories_enabled,
			last_auto_sprint_number
		) VALUES (
			:team_id,
			:workspace_id,
			true,
			false,
			2,
			2,
			'Monday',
			true,
			0
		)
		RETURNING
			team_id,
			workspace_id,
			sprints_enabled,
			auto_create_sprints,
			upcoming_sprints_count,
			sprint_duration_weeks,
			sprint_start_day,
			move_incomplete_stories_enabled,
			last_auto_sprint_number,
			created_at,
			updated_at
	`

	params := map[string]any{
		"team_id":      teamID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return teamsettings.CoreTeamSprintSettings{}, err
	}
	defer stmt.Close()

	var settings dbTeamSprintSettings
	if err := stmt.GetContext(ctx, &settings, params); err != nil {
		return teamsettings.CoreTeamSprintSettings{}, err
	}

	span.AddEvent("default sprint settings created")
	return toCoreTeamSprintSettings(settings), nil
}

// createDefaultStoryAutomationSettings creates and returns default story automation settings
func (r *repo) createDefaultStoryAutomationSettings(ctx context.Context, teamID, workspaceID uuid.UUID) (teamsettings.CoreTeamStoryAutomationSettings, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.teamsettings.createDefaultStoryAutomationSettings")
	defer span.End()

	query := `
		INSERT INTO team_story_automation_settings (
			team_id,
			workspace_id,
			auto_close_inactive_enabled,
			auto_close_inactive_months,
			auto_archive_enabled,
			auto_archive_months
		) VALUES (
			:team_id,
			:workspace_id,
			false,
			6,
			false,
			6
		)
		RETURNING
			team_id,
			workspace_id,
			auto_close_inactive_enabled,
			auto_close_inactive_months,
			auto_archive_enabled,
			auto_archive_months,
			created_at,
			updated_at
	`

	params := map[string]any{
		"team_id":      teamID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return teamsettings.CoreTeamStoryAutomationSettings{}, err
	}
	defer stmt.Close()

	var settings dbTeamStoryAutomationSettings
	if err := stmt.GetContext(ctx, &settings, params); err != nil {
		return teamsettings.CoreTeamStoryAutomationSettings{}, err
	}

	span.AddEvent("default story automation settings created")
	return toCoreTeamStoryAutomationSettings(settings), nil
}
