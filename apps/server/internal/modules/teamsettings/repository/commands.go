package teamsettingsrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *repo) UpdateSprintSettings(ctx context.Context, teamID, workspaceID uuid.UUID, updates teamsettings.CoreUpdateTeamSprintSettings, actorID *uuid.UUID) (teamsettings.CoreTeamSprintSettings, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.teamsettings.UpdateSprintSettings")
	defer span.End()

	// Ensure the default settings row exists before opening the update transaction.
	if _, err := r.GetSprintSettings(ctx, teamID, workspaceID); err != nil {
		return teamsettings.CoreTeamSprintSettings{}, err
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return teamsettings.CoreTeamSprintSettings{}, fmt.Errorf("begin sprint settings update: %w", err)
	}
	defer tx.Rollback()

	query := "UPDATE team_sprint_settings SET "
	var setClauses []string
	params := map[string]any{
		"team_id":      teamID,
		"workspace_id": workspaceID,
	}

	if updates.AutoCreateSprints != nil {
		setClauses = append(setClauses, "auto_create_sprints = :auto_create_sprints")
		params["auto_create_sprints"] = *updates.AutoCreateSprints
		if *updates.AutoCreateSprints {
			setClauses = append(setClauses, "auto_create_disabled_at = NULL")
			setClauses = append(setClauses, "auto_create_disabled_reason = NULL")
		}
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
	if updates.WorkingDays != nil {
		setClauses = append(setClauses, "working_days = :working_days")
		params["working_days"] = pq.Int64Array(intsToInt64s(*updates.WorkingDays))
	}
	if updates.MoveIncompleteStoriesEnabled != nil {
		setClauses = append(setClauses, "move_incomplete_stories_enabled = :move_incomplete_stories_enabled")
		params["move_incomplete_stories_enabled"] = *updates.MoveIncompleteStoriesEnabled
	}
	if updates.NextAutoSprintNumber != nil {
		setClauses = append(setClauses, "next_auto_sprint_number = :next_auto_sprint_number")
		params["next_auto_sprint_number"] = *updates.NextAutoSprintNumber
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	query += strings.Join(setClauses, ", ")
	query += " WHERE team_id = :team_id AND workspace_id = :workspace_id;"

	stmt, err := tx.PrepareNamedContext(ctx, query)
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
			case "team_sprint_settings_working_days_check":
				return teamsettings.CoreTeamSprintSettings{}, teamsettings.ErrInvalidWorkingDays
			}
		}
		errMsg := fmt.Sprintf("Failed to update team sprint settings: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update team sprint settings"), trace.WithAttributes(attribute.String("error", errMsg)))
		return teamsettings.CoreTeamSprintSettings{}, err
	}

	var updated dbTeamSprintSettings
	if err := tx.GetContext(ctx, &updated, `
		SELECT
			team_id, workspace_id, auto_create_sprints, upcoming_sprints_count,
			sprint_duration_weeks, sprint_start_day, working_days, move_incomplete_stories_enabled,
			last_auto_sprint_number, next_auto_sprint_number, auto_create_disabled_at,
			auto_create_disabled_reason, created_at, updated_at
		FROM team_sprint_settings
		WHERE team_id = $1 AND workspace_id = $2
		FOR UPDATE
	`, teamID, workspaceID); err != nil {
		return teamsettings.CoreTeamSprintSettings{}, fmt.Errorf("get updated sprint settings: %w", err)
	}
	updatedSettings := toCoreTeamSprintSettings(updated)

	cadenceChanged := updates.SprintDurationWeeks != nil || updates.SprintStartDay != nil
	if cadenceChanged && updatedSettings.AutoCreateSprints {
		if _, err := r.reconcileSprintScheduleTx(ctx, tx, updatedSettings, actorID); err != nil {
			return teamsettings.CoreTeamSprintSettings{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return teamsettings.CoreTeamSprintSettings{}, fmt.Errorf("commit sprint settings update: %w", err)
	}

	r.log.Info(ctx, "Team sprint settings updated successfully.")
	span.AddEvent("sprint settings updated.", trace.WithAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))

	return updatedSettings, nil
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

func (r *repo) UpdateEstimationSettings(ctx context.Context, teamID, workspaceID uuid.UUID, updates teamsettings.CoreUpdateTeamEstimationSettings) (teamsettings.CoreTeamEstimationSettings, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.teamsettings.UpdateEstimationSettings")
	defer span.End()

	scheme := "hours"
	if updates.Scheme != nil {
		scheme = *updates.Scheme
	}

	query := `
		INSERT INTO team_estimation_settings (
			team_id,
			workspace_id,
			scheme
		) VALUES (
			:team_id,
			:workspace_id,
			:scheme
		)
		ON CONFLICT (team_id) DO UPDATE SET
			workspace_id = EXCLUDED.workspace_id,
			scheme = EXCLUDED.scheme,
			updated_at = NOW()
	`

	params := map[string]any{
		"team_id":      teamID,
		"workspace_id": workspaceID,
		"scheme":       scheme,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return teamsettings.CoreTeamEstimationSettings{}, err
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		return teamsettings.CoreTeamEstimationSettings{}, err
	}

	return r.GetEstimationSettings(ctx, teamID, workspaceID)
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
			auto_create_sprints,
			upcoming_sprints_count,
			sprint_duration_weeks,
			sprint_start_day,
			working_days,
			move_incomplete_stories_enabled,
			last_auto_sprint_number,
			next_auto_sprint_number
		) VALUES (
			:team_id,
			:workspace_id,
			false,
			2,
			2,
			'Monday',
			ARRAY[1, 2, 3, 4, 5]::smallint[],
			true,
			0,
			1
		)
		RETURNING
			team_id,
			workspace_id,
			auto_create_sprints,
			upcoming_sprints_count,
			sprint_duration_weeks,
			sprint_start_day,
			working_days,
			move_incomplete_stories_enabled,
			last_auto_sprint_number,
			next_auto_sprint_number,
			auto_create_disabled_at,
			auto_create_disabled_reason,
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
			true,
			3,
			true,
			3
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

func (r *repo) createDefaultEstimationSettings(ctx context.Context, teamID, workspaceID uuid.UUID) (teamsettings.CoreTeamEstimationSettings, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.teamsettings.createDefaultEstimationSettings")
	defer span.End()

	query := `
		INSERT INTO team_estimation_settings (
			team_id,
			workspace_id,
			scheme
		) VALUES (
			:team_id,
			:workspace_id,
			'hours'
		)
		ON CONFLICT (team_id) DO UPDATE SET
			workspace_id = EXCLUDED.workspace_id,
			updated_at = NOW()
		RETURNING
			team_id,
			workspace_id,
			scheme,
			created_at,
			updated_at
	`

	params := map[string]any{
		"team_id":      teamID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return teamsettings.CoreTeamEstimationSettings{}, err
	}
	defer stmt.Close()

	var settings dbTeamEstimationSettings
	if err := stmt.GetContext(ctx, &settings, params); err != nil {
		return teamsettings.CoreTeamEstimationSettings{}, err
	}

	return toCoreTeamEstimationSettings(settings), nil
}
