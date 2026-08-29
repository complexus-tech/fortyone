package teamsettings

import teamsettingsdomain "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"

var (
	ErrInvalidSprintStartDay  = teamsettingsdomain.ErrInvalidSprintStartDay
	ErrInvalidSprintDuration  = teamsettingsdomain.ErrInvalidSprintDuration
	ErrInvalidWorkingDays     = teamsettingsdomain.ErrInvalidWorkingDays
	ErrInvalidUpcomingCount   = teamsettingsdomain.ErrInvalidUpcomingCount
	ErrInvalidNextAutoNumber  = teamsettingsdomain.ErrInvalidNextAutoNumber
	ErrInvalidCloseMonths     = teamsettingsdomain.ErrInvalidCloseMonths
	ErrInvalidArchiveMonths   = teamsettingsdomain.ErrInvalidArchiveMonths
	ErrInvalidEstimateScheme  = teamsettingsdomain.ErrInvalidEstimateScheme
	ErrSprintScheduleConflict = teamsettingsdomain.ErrSprintScheduleConflict
	ErrTeamSettingsNotFound   = teamsettingsdomain.ErrTeamSettingsNotFound
	ErrTeamMembershipRequired = teamsettingsdomain.ErrTeamMembershipRequired
	ErrNoSettingsChanges      = teamsettingsdomain.ErrNoSettingsChanges
	ErrConcurrentUpdate       = teamsettingsdomain.ErrConcurrentUpdate
)
