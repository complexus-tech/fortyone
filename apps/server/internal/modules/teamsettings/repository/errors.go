package teamsettingsrepository

import (
	"errors"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func mapDatabaseError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return teamsettings.ErrTeamSettingsNotFound
	}
	if platformdatabase.IsRetryableTransactionError(err) {
		return teamsettings.ErrConcurrentUpdate
	}

	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return err
	}
	switch postgresError.ConstraintName {
	case "team_sprint_settings_sprint_start_day_check":
		return teamsettings.ErrInvalidSprintStartDay
	case "team_sprint_settings_sprint_duration_weeks_check":
		return teamsettings.ErrInvalidSprintDuration
	case "team_sprint_settings_upcoming_sprints_count_check":
		return teamsettings.ErrInvalidUpcomingCount
	case "team_sprint_settings_working_days_check":
		return teamsettings.ErrInvalidWorkingDays
	case "team_sprint_settings_next_auto_sprint_number_check":
		return teamsettings.ErrInvalidNextAutoNumber
	case "team_story_automation_settings_auto_close_inactive_months_check":
		return teamsettings.ErrInvalidCloseMonths
	case "team_story_automation_settings_auto_archive_months_check":
		return teamsettings.ErrInvalidArchiveMonths
	case "team_estimation_settings_scheme_check":
		return teamsettings.ErrInvalidEstimateScheme
	case "audit_events_actor_id_fkey":
		return teamsettings.ErrTeamSettingsNotFound
	default:
		return err
	}
}
